// Package service implements the business logic of SchemaFlow.
package service

import (
	"log"
	"strings"

	"schemaflow-backend/internal/models"
)

// ParseQuery extracts column and table references from a SQL query.
// Handles SELECT, INSERT, UPDATE, and DELETE statements using string
// inspection. A full AST parser would be used in a production system.
func ParseQuery(query string) (*models.ParsedQuery, error) {
	result := &models.ParsedQuery{
		OriginalQuery: query,
		Columns:       []models.ColumnReference{},
		Tables:        []models.TableReference{},
	}

	lower := strings.ToLower(query)

	switch {
	case strings.Contains(lower, "select"):
		extractFromSelect(query, lower, result)
	case strings.Contains(lower, "insert"):
		extractFromInsert(query, lower, result)
	case strings.Contains(lower, "update"):
		extractFromUpdate(query, lower, result)
	case strings.Contains(lower, "delete"):
		extractFromDelete(query, lower, result)
	}

	log.Printf("Parsed query: %d columns, %d tables", len(result.Columns), len(result.Tables))
	return result, nil
}

// extractFromSelect handles SELECT statements.
func extractFromSelect(query, lower string, result *models.ParsedQuery) {
	selectIdx := strings.Index(lower, "select")
	fromIdx := strings.Index(lower, "from")

	if selectIdx < 0 || fromIdx <= selectIdx {
		return
	}

	selectPart := strings.TrimSpace(query[selectIdx+6 : fromIdx])

	if selectPart == "*" {
		result.Columns = append(result.Columns, models.ColumnReference{Name: "*"})
	} else {
		for _, raw := range splitOnComma(selectPart) {
			col := strings.TrimSpace(raw)
			if col == "" {
				continue
			}

			alias := ""
			colName := col
			if idx := strings.Index(strings.ToLower(col), " as "); idx > 0 {
				colName = strings.TrimSpace(col[:idx])
				alias = strings.TrimSpace(col[idx+4:])
			}

			tableName := ""
			if idx := strings.Index(colName, "."); idx > 0 {
				parts := strings.SplitN(colName, ".", 2)
				tableName = parts[0]
				colName = parts[1]
			}

			if strings.Contains(colName, "(") {
				continue
			}

			if colName != "" && colName != "*" {
				result.Columns = append(result.Columns, models.ColumnReference{
					Name:      colName,
					TableName: tableName,
					Alias:     alias,
					Position:  len(result.Columns),
				})
			}
		}
	}

	extractTables(query, lower, result)
}

// extractFromInsert handles INSERT statements.
func extractFromInsert(query, lower string, result *models.ParsedQuery) {
	intoIdx := strings.Index(lower, "insert into ")
	if intoIdx < 0 {
		return
	}

	rest := query[intoIdx+len("insert into "):]
	tableEnd := strings.IndexAny(rest, " (")
	if tableEnd < 0 {
		tableEnd = len(rest)
	}
	tableName := strings.TrimSpace(rest[:tableEnd])
	if tableName != "" {
		result.Tables = append(result.Tables, models.TableReference{Name: tableName})
	}

	parenStart := strings.Index(rest, "(")
	parenEnd := strings.Index(rest, ")")
	if parenStart >= 0 && parenEnd > parenStart {
		cols := rest[parenStart+1 : parenEnd]
		for _, raw := range strings.Split(cols, ",") {
			col := strings.TrimSpace(raw)
			if col != "" {
				result.Columns = append(result.Columns, models.ColumnReference{Name: col})
			}
		}
	}
}

// extractFromUpdate handles UPDATE statements.
func extractFromUpdate(query, lower string, result *models.ParsedQuery) {
	updateIdx := strings.Index(lower, "update ")
	if updateIdx < 0 {
		return
	}

	rest := query[updateIdx+len("update "):]
	tableEnd := strings.IndexAny(rest, " ")
	if tableEnd < 0 {
		tableEnd = len(rest)
	}
	tableName := strings.TrimSpace(rest[:tableEnd])
	if tableName != "" {
		result.Tables = append(result.Tables, models.TableReference{Name: tableName})
	}

	setIdx := strings.Index(strings.ToLower(rest), " set ")
	whereIdx := strings.Index(strings.ToLower(rest), " where ")
	if setIdx < 0 {
		return
	}

	setStart := setIdx + len(" set ")
	setEnd := len(rest)
	if whereIdx > setIdx {
		setEnd = whereIdx
	}

	for _, raw := range strings.Split(rest[setStart:setEnd], ",") {
		eq := strings.Index(raw, "=")
		if eq < 0 {
			continue
		}
		col := strings.TrimSpace(raw[:eq])
		if col != "" {
			result.Columns = append(result.Columns, models.ColumnReference{Name: col})
		}
	}
}

// extractFromDelete handles DELETE statements.
func extractFromDelete(query, lower string, result *models.ParsedQuery) {
	idx := strings.Index(lower, "delete from ")
	if idx < 0 {
		return
	}
	rest := query[idx+len("delete from "):]
	end := strings.IndexAny(rest, " ")
	if end < 0 {
		end = len(rest)
	}
	tableName := strings.TrimSpace(rest[:end])
	if tableName != "" {
		result.Tables = append(result.Tables, models.TableReference{Name: tableName})
	}
}

// extractTables finds table names in the FROM clause.
func extractTables(query, lower string, result *models.ParsedQuery) {
	fromIdx := strings.Index(lower, " from ")
	if fromIdx < 0 {
		return
	}

	rest := query[fromIdx+len(" from "):]
	stopWords := []string{" where ", " group ", " order ", " limit ", " offset ", " join "}
	endIdx := len(rest)
	for _, w := range stopWords {
		if idx := strings.Index(strings.ToLower(rest), w); idx >= 0 && idx < endIdx {
			endIdx = idx
		}
	}
	tablePart := strings.TrimSpace(rest[:endIdx])

	for _, raw := range strings.Split(tablePart, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		fields := strings.Fields(raw)
		if len(fields) == 0 {
			continue
		}
		tableName := fields[0]
		alias := ""
		if len(fields) > 1 {
			alias = fields[1]
		}
		schema := ""
		if idx := strings.Index(tableName, "."); idx > 0 {
			parts := strings.SplitN(tableName, ".", 2)
			schema = parts[0]
			tableName = parts[1]
		}
		result.Tables = append(result.Tables, models.TableReference{
			Name:   tableName,
			Schema: schema,
			Alias:  alias,
		})
	}
}

// splitOnComma splits a string by commas, ignoring commas inside parentheses.
func splitOnComma(s string) []string {
	var out []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		switch r {
		case '(':
			depth++
			current.WriteRune(r)
		case ')':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				out = append(out, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		out = append(out, current.String())
	}
	return out
}
