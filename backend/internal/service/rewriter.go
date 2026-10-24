// Package service implements the business logic of SchemaFlow.
package service

import (
	"context"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Rewriter rewrites SQL queries to use current column names.
type Rewriter struct {
	pool *pgxpool.Pool
}

// NewRewriter returns a rewriter backed by the given pool.
func NewRewriterService(pool *pgxpool.Pool) *Rewriter {
	return &Rewriter{pool: pool}
}

// Rewrite parses the query, looks up each column in column_lineage,
// and returns the query with old column names replaced by current ones.
// If no columns need rewriting, the original query is returned.
func (r *Rewriter) Rewrite(ctx context.Context, query string) (string, error) {
	parsed, err := ParseQuery(query)
	if err != nil {
		return "", err
	}

	if len(parsed.Columns) == 0 {
		return query, nil
	}

	rewritten := query
	changes := 0

	for _, col := range parsed.Columns {
		if col.Name == "*" {
			continue
		}

		tableName := col.TableName
		if tableName == "" && len(parsed.Tables) > 0 {
			tableName = parsed.Tables[0].Name
		}
		if tableName == "" {
			continue
		}

		currentName, err := r.currentColumnName(ctx, tableName, col.Name)
		if err != nil {
			log.Printf("No mapping for %s.%s: %v", tableName, col.Name, err)
			continue
		}
		if currentName == "" || currentName == col.Name {
			continue
		}

		rewritten = strings.ReplaceAll(rewritten, col.Name, currentName)
		changes++
		log.Printf("Rewrote: %s -> %s", col.Name, currentName)
	}

	if changes == 0 {
		log.Println("No rewrites needed, returning original query")
	} else {
		log.Printf("Query rewritten with %d changes", changes)
	}

	return rewritten, nil
}

// currentColumnName returns the current name of a column by querying
// the column_lineage table. Returns an error if no mapping exists.
func (r *Rewriter) currentColumnName(
	ctx context.Context,
	tableName, oldName string,
) (string, error) {
	var currentName string
	err := r.pool.QueryRow(ctx, `
		SELECT current_column_name
		FROM column_lineage
		WHERE schema_name = 'public'
		  AND table_name = $1
		  AND old_column_name = $2
		  AND is_current = true
		LIMIT 1
	`, tableName, oldName).Scan(&currentName)
	if err != nil {
		return "", err
	}
	return currentName, nil
}
