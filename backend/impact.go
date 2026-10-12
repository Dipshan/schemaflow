// Package main contains the impact prediction logic for SchemaFlow
// This file queries PostgreSQL's system catalog to find objects that depend on a column
package main

import (
	"context"
	"fmt"
	"log"
)

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// ColumnDependency represents a single object that depends on a column
type ColumnDependency struct {
	ObjectType     string `json:"object_type"`     // "VIEW", "FUNCTION", "TRIGGER"
	SchemaName     string `json:"schema_name"`     // Schema of the dependent object
	ObjectName     string `json:"object_name"`     // Name of the dependent object
	Definition     string `json:"definition"`      // Definition (for views, functions)
	DependencyType string `json:"dependency_type"` // "view", "function", "trigger"
}

// ImpactReport holds all dependencies found for a column
type ImpactReport struct {
	Schema       string             `json:"schema"`       // The schema of the column
	Table        string             `json:"table"`        // The table of the column
	Column       string             `json:"column"`       // The column name
	Dependencies []ColumnDependency `json:"dependencies"` // All objects that depend on this column
	Total        int                `json:"total"`        // Total number of dependencies found
}

// ============================================================================
// CORE FUNCTION: Get Column Dependencies
// ============================================================================

// GetColumnDependencies queries PostgreSQL's system catalog to find all objects
// that depend on a specific column — including all historical names in the
// lineage chain (via column_lineage table).
//
// Workflow:
//  1. Get all names in the lineage chain (old names + current name)
//  2. Loop over each name and query pg_catalog
//  3. Merge results and return
//
// This handles both BEFORE-rename and AFTER-rename scenarios.
func GetColumnDependencies(schema, table, column string) (*ImpactReport, error) {
	log.Printf("Getting dependencies for %s.%s.%s", schema, table, column)

	// STEP 1: Get all column names in the lineage chain
	columnNames, err := getAllColumnNames(schema, table, column)
	if err != nil {
		log.Printf("Warning: Could not get lineage chain: %v", err)
		columnNames = []string{column} // Fall back to just the given name
	}
	log.Printf("Checking dependencies for columns: %v", columnNames)

	// STEP 2: Initialize the report
	report := &ImpactReport{
		Schema:       schema,
		Table:        table,
		Column:       column,
		Dependencies: []ColumnDependency{},
	}

	// STEP 3: Loop over every name in the chain
	for _, colName := range columnNames {

		// --------------------------------------------------------------------
		// Query for VIEW dependencies
		// --------------------------------------------------------------------
		viewsQuery := `
			SELECT DISTINCT
				'VIEW' AS object_type,
				n.nspname AS schema_name,
				c.relname AS object_name,
				pg_get_viewdef(c.oid) AS definition
			FROM pg_depend d
			JOIN pg_rewrite r ON d.objid = r.oid
			JOIN pg_class c ON r.ev_class = c.oid
			JOIN pg_namespace n ON c.relnamespace = n.oid
			WHERE d.refobjid = (
				SELECT c.oid FROM pg_class c
				JOIN pg_namespace n ON c.relnamespace = n.oid
				WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind = 'r'
			)
			AND d.refobjsubid = (
				SELECT attnum FROM pg_attribute
				WHERE attrelid = (
					SELECT c.oid FROM pg_class c
					JOIN pg_namespace n ON c.relnamespace = n.oid
					WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind = 'r'
				) AND attname = $3
			)
		`
		rows, err := DB.Query(context.Background(), viewsQuery, schema, table, colName)
		if err != nil {
			log.Printf("Error querying views for column '%s': %v", colName, err)
		} else {
			for rows.Next() {
				var dep ColumnDependency
				if err := rows.Scan(&dep.ObjectType, &dep.SchemaName, &dep.ObjectName, &dep.Definition); err == nil {
					dep.DependencyType = "view"
					report.Dependencies = append(report.Dependencies, dep)
				}
			}
			rows.Close()
		}

		// --------------------------------------------------------------------
		// Query for FUNCTION dependencies
		// --------------------------------------------------------------------
		functionsQuery := `
			SELECT DISTINCT
				'FUNCTION' AS object_type,
				n.nspname AS schema_name,
				p.proname AS object_name,
				pg_get_functiondef(p.oid) AS definition
			FROM pg_depend d
			JOIN pg_proc p ON d.objid = p.oid
			JOIN pg_namespace n ON p.pronamespace = n.oid
			WHERE d.refobjid = (
				SELECT c.oid FROM pg_class c
				JOIN pg_namespace n ON c.relnamespace = n.oid
				WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind = 'r'
			)
			AND d.refobjsubid = (
				SELECT attnum FROM pg_attribute
				WHERE attrelid = (
					SELECT c.oid FROM pg_class c
					JOIN pg_namespace n ON c.relnamespace = n.oid
					WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind = 'r'
				) AND attname = $3
			)
		`
		rows, err = DB.Query(context.Background(), functionsQuery, schema, table, colName)
		if err != nil {
			log.Printf("Error querying functions for column '%s': %v", colName, err)
		} else {
			for rows.Next() {
				var dep ColumnDependency
				if err := rows.Scan(&dep.ObjectType, &dep.SchemaName, &dep.ObjectName, &dep.Definition); err == nil {
					dep.DependencyType = "function"
					report.Dependencies = append(report.Dependencies, dep)
				}
			}
			rows.Close()
		}
	}

	// STEP 4: Set total count
	report.Total = len(report.Dependencies)

	log.Printf("Found %d dependencies for %s.%s.%s", report.Total, schema, table, column)
	return report, nil
}

// ============================================================================
// HELPER FUNCTION: Get All Column Names in Lineage Chain
// ============================================================================

// getAllColumnNames walks the column_lineage table and returns every column name
// connected to the given column — both old names and the current name.
//
// Example: If lineage contains:
//
//	primary_title → movie_title → film_title
//
// Then calling getAllColumnNames("movie_title") returns:
//
//	["movie_title", "primary_title", "film_title"]
//
// This allows /impact to detect dependencies even after a column has been renamed.
// Includes cycle detection to prevent infinite loops if names are reused.
func getAllColumnNames(schema, table, column string) ([]string, error) {
	names := []string{column}
	seen := map[string]bool{column: true}

	// Load all lineage pairs for this table
	query := `
		SELECT old_column_name, new_column_name
		FROM column_lineage
		WHERE schema_name = $1 AND table_name = $2
	`
	rows, err := DB.Query(context.Background(), query, schema, table)
	if err != nil {
		return names, err
	}
	defer rows.Close()

	type pair struct{ old, new string }
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.old, &p.new); err == nil {
			pairs = append(pairs, p)
		}
	}

	// Walk the chain both directions with cycle detection
	// Repeat until no new names are discovered
	changed := true
	for changed {
		changed = false
		for _, p := range pairs {
			if seen[p.old] && !seen[p.new] {
				names = append(names, p.new)
				seen[p.new] = true
				changed = true
			}
			if seen[p.new] && !seen[p.old] {
				names = append(names, p.old)
				seen[p.old] = true
				changed = true
			}
		}
	}

	return names, nil
}

// ============================================================================
// HELPER FUNCTIONS FOR IMPACT REPORT
// ============================================================================

// HasDependencies returns true if the report has any dependencies
func (r *ImpactReport) HasDependencies() bool {
	return r.Total > 0
}

// GetDependencyNames returns a list of all dependent object names
func (r *ImpactReport) GetDependencyNames() []string {
	names := []string{}
	for _, dep := range r.Dependencies {
		names = append(names, dep.ObjectName)
	}
	return names
}

// GetDependenciesByType returns dependencies filtered by type
func (r *ImpactReport) GetDependenciesByType(objectType string) []ColumnDependency {
	result := []ColumnDependency{}
	for _, dep := range r.Dependencies {
		if dep.ObjectType == objectType {
			result = append(result, dep)
		}
	}
	return result
}

// PrettyPrint prints the impact report to stdout (for debugging)
func (r *ImpactReport) PrettyPrint() {
	fmt.Printf("=== Impact Report ===\n")
	fmt.Printf("Schema: %s\n", r.Schema)
	fmt.Printf("Table:  %s\n", r.Table)
	fmt.Printf("Column: %s\n", r.Column)
	fmt.Printf("Total:  %d dependencies\n\n", r.Total)
	for i, dep := range r.Dependencies {
		fmt.Printf("  [%d] %s: %s.%s\n", i+1, dep.ObjectType, dep.SchemaName, dep.ObjectName)
	}
	fmt.Println("=====================")
}
