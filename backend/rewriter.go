// Package main contains the SQL parsing and rewriting logic for SchemaFlow
// This file handles:
// 1. Parsing SQL queries to extract column and table names
// 2. Looking up column names in column_lineage table
// 3. Rewriting queries with new column names
package main

import (
	"context" // For database queries with timeout support
	"fmt"     // For formatted error messages
	"log"     // For logging parser activity
	"strings" // For string manipulation (search, replace, split)
)

// DATA STRUCTURES

// ColumnReference represents a column found in a SQL query
// Example: In "SELECT users.full_name AS name FROM users",
//
//	Name = "full_name", TableName = "users", Alias = "name"
type ColumnReference struct {
	Name      string // The actual column name (e.g., "full_name")
	TableName string // The table qualifier (e.g., "users" in "users.full_name")
	Alias     string // The alias (e.g., "name" in "full_name AS name")
	Position  int    // Order position in the query (for future use)
}

// TableReference represents a table found in a SQL query
// Example: In "SELECT * FROM users u", Name = "users", Alias = "u"
type TableReference struct {
	Name   string // The actual table name (e.g., "users")
	Schema string // The schema name (e.g., "public")
	Alias  string // The table alias (e.g., "u" in "users u")
}

// ParsedQuery holds the original query and all extracted references
// This is returned by ParseQuery() and used by the rewriter
type ParsedQuery struct {
	OriginalQuery string            // The original SQL query string
	Columns       []ColumnReference // All column references found
	Tables        []TableReference  // All table references found
}

// PARSER: Extract columns and tables from SQL queries

// ParseQuery parses a SQL query and extracts column and table references
// For MVP, this uses simple string parsing. Production would use AST.
// Example: ParseQuery("SELECT full_name, email FROM users")
//
//	returns Columns: [full_name, email], Tables: [users]
func ParseQuery(query string) (*ParsedQuery, error) {
	// STEP 1: Initialize the result structure
	result := &ParsedQuery{
		OriginalQuery: query,
		Columns:       []ColumnReference{},
		Tables:        []TableReference{},
	}

	// STEP 2: Convert to lowercase for case-insensitive matching
	lower := strings.ToLower(query)

	// STEP 3: Extract columns from SELECT clause
	// Look for "SELECT ... FROM" pattern
	if strings.Contains(lower, "select") && strings.Contains(lower, "from") {
		// Find the position of "FROM"
		fromIdx := strings.Index(lower, "from")
		if fromIdx > 0 {
			// Extract everything between "SELECT" and "FROM"
			// "SELECT " is 6 characters, so start at position 6
			selectPart := query[6:fromIdx]

			// Split columns by commas
			columns := strings.Split(selectPart, ",")

			// Process each column
			for _, col := range columns {
				col = strings.TrimSpace(col)

				// Skip empty columns or SELECT *
				if col == "" || col == "*" {
					continue
				}

				// Handle "column AS alias" pattern
				colName := col
				alias := ""
				if idx := strings.Index(strings.ToLower(col), " as "); idx > 0 {
					colName = strings.TrimSpace(col[:idx])
					alias = strings.TrimSpace(col[idx+4:])
				}

				// Handle "table.column" pattern
				tableName := ""
				if idx := strings.Index(colName, "."); idx > 0 {
					parts := strings.Split(colName, ".")
					tableName = parts[0]
					colName = parts[1]
				}

				// Add to result
				if colName != "" {
					result.Columns = append(result.Columns, ColumnReference{
						Name:      colName,
						TableName: tableName,
						Alias:     alias,
					})
				}
			}
		}
	}

	// STEP 4: Extract tables from FROM clause
	if strings.Contains(lower, "from") {
		// Find the position of "FROM"
		fromIdx := strings.Index(lower, "from")
		if fromIdx > 0 {
			// Extract everything after "FROM "
			remaining := query[fromIdx+4:]

			// Find where table list ends (at WHERE, GROUP BY, etc.)
			stopWords := []string{"where", "group", "order", "limit", "offset", "join"}
			endIdx := len(remaining)
			for _, word := range stopWords {
				if idx := strings.Index(strings.ToLower(remaining), " "+word+" "); idx > 0 && idx < endIdx {
					endIdx = idx
				}
			}

			// Extract the table part
			tablePart := strings.TrimSpace(remaining[:endIdx])

			// Split tables by commas
			tables := strings.Split(tablePart, ",")

			// Process each table
			for _, t := range tables {
				t = strings.TrimSpace(t)
				if t == "" {
					continue
				}

				// Remove alias if present (e.g., "users u" → "users")
				parts := strings.Fields(t)
				tableName := parts[0]
				alias := ""
				if len(parts) > 1 {
					alias = parts[1]
				}

				// Handle "schema.table" pattern
				schema := ""
				if strings.Contains(tableName, ".") {
					parts := strings.Split(tableName, ".")
					schema = parts[0]
					tableName = parts[1]
				}

				// Add to result
				result.Tables = append(result.Tables, TableReference{
					Name:   tableName,
					Schema: schema,
					Alias:  alias,
				})
			}
		}
	}

	// STEP 5: Log results for debugging
	log.Printf("Parsed query: %d columns, %d tables", len(result.Columns), len(result.Tables))

	return result, nil
}

// HELPER FUNCTIONS FOR PARSED QUERY

// GetColumnNames returns a list of all column names from the parsed query
// Used by the /parse endpoint to return column names
func (pq *ParsedQuery) GetColumnNames() []string {
	names := []string{}
	for _, col := range pq.Columns {
		if col.Name != "*" {
			names = append(names, col.Name)
		}
	}
	return names
}

// GetTableNames returns a list of all table names from the parsed query
// Used by the /parse endpoint to return table names
func (pq *ParsedQuery) GetTableNames() []string {
	names := []string{}
	for _, table := range pq.Tables {
		names = append(names, table.Name)
	}
	return names
}

// HasColumn checks if a specific column exists in the parsed query
// Used by the rewriter to check if a column needs rewriting
func (pq *ParsedQuery) HasColumn(columnName string) bool {
	for _, col := range pq.Columns {
		if col.Name == columnName {
			return true
		}
	}
	return false
}

// QUERY REWRITER: Replace old column names with new ones

// RewriteQuery rewrites a SQL query by replacing old column names with new ones
// It uses the column_lineage table to find mappings from old names to current names
//
// Example:
//
//	Input:  "SELECT full_name, email FROM users WHERE id = 123"
//	Output: "SELECT buyer_name, email FROM users WHERE id = 123"
//
// If no columns need rewriting, the original query is returned unchanged
func RewriteQuery(query string) (string, error) {
	// STEP 1: Parse the query to extract column references
	parsed, err := ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("failed to parse query: %w", err)
	}

	// STEP 2: If no columns found, return original query
	if len(parsed.Columns) == 0 {
		log.Println("No columns found, returning original query")
		return query, nil
	}

	// STEP 3: Start with the original query
	rewritten := query

	// STEP 4: For each column, check if it needs rewriting
	for _, col := range parsed.Columns {
		// Skip SELECT * (no column name to rewrite)
		if col.Name == "*" {
			continue
		}

		// STEP 5: Look up the current column name from column_lineage
		currentName, err := getCurrentColumnName(col.Name, col.TableName, parsed.Tables)
		if err != nil {
			// Column not found in lineage, skip (no rewrite needed)
			log.Printf("No mapping found for column '%s', skipping", col.Name)
			continue
		}

		// STEP 6: If the name changed, rewrite the query
		if currentName != "" && currentName != col.Name {
			// Simple string replacement (production would use AST)
			rewritten = strings.ReplaceAll(rewritten, col.Name, currentName)
			log.Printf("Rewrote: %s → %s", col.Name, currentName)
		}
	}

	// STEP 7: If no changes were made, return original query
	if rewritten == query {
		log.Println("No rewrites needed, returning original query")
	} else {
		log.Printf("Query rewritten: %s → %s", query, rewritten)
	}

	return rewritten, nil
}

// DATABASE LOOKUP: Find current column name from column_lineage

// getCurrentColumnName looks up the current name for a column in column_lineage
// It queries the database for the most recent rename mapping
//
// Parameters:
//   - oldName: The old column name (e.g., "full_name")
//   - tableName: The table name (e.g., "users") - can be empty
//   - tables: List of tables from the parsed query
//
// Returns:
//   - The current column name (e.g., "buyer_name") if found
//   - Empty string if no mapping exists
//   - Error if database query fails
func getCurrentColumnName(oldName, tableName string, tables []TableReference) (string, error) {
	// STEP 1: If tableName is empty, try to find it from the tables list
	// This handles cases like "SELECT full_name FROM users" where "full_name" has no qualifier
	if tableName == "" && len(tables) > 0 {
		// Use the first table as the default (most common case)
		tableName = tables[0].Name
		log.Printf("Using default table '%s' for column '%s'", tableName, oldName)
	}

	// STEP 2: If still empty, return empty (no rewrite possible)
	if tableName == "" {
		log.Printf("No table name found for column '%s', skipping", oldName)
		return "", nil
	}

	// STEP 3: Use 'public' as default schema (MVP)
	schemaName := "public"

	// STEP 4: Query the column_lineage table for the current name
	var currentName string
	query := `
		SELECT new_column_name 
		FROM column_lineage 
		WHERE schema_name = $1 
		  AND table_name = $2 
		  AND old_column_name = $3 
		  AND is_current = true
		LIMIT 1
	`

	// Execute the query
	err := DB.QueryRow(context.Background(), query, schemaName, tableName, oldName).Scan(&currentName)
	if err != nil {
		// Column not found or other error
		log.Printf("Warning: Could not find mapping for %s.%s: %v", tableName, oldName, err)
		return "", nil // Return empty, not error (no rewrite)
	}

	log.Printf("Found mapping: %s.%s → %s", tableName, oldName, currentName)
	return currentName, nil
}
