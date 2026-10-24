// Package service implements the business logic of SchemaFlow.
package service

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"schemaflow-backend/internal/models"
)

// ImpactService finds objects that depend on a column.
type ImpactService struct {
	pool *pgxpool.Pool
}

// NewImpactService returns a new impact service backed by the pool.
func NewImpactService(pool *pgxpool.Pool) *ImpactService {
	return &ImpactService{pool: pool}
}

// GetDependencies returns every view and function that depends on
// the given column, including any historical names in the lineage
// chain. This handles both before-rename and after-rename scenarios.
func (s *ImpactService) GetDependencies(
	ctx context.Context,
	schema, table, column string,
) (*models.ImpactReport, error) {
	log.Printf("Getting dependencies for %s.%s.%s", schema, table, column)

	columnNames, err := s.getAllColumnNames(ctx, schema, table, column)
	if err != nil {
		log.Printf("Warning: Could not get lineage chain: %v", err)
		columnNames = []string{column}
	}
	log.Printf("Checking dependencies for columns: %v", columnNames)

	report := &models.ImpactReport{
		Schema:       schema,
		Table:        table,
		Column:       column,
		Dependencies: []models.ColumnDependency{},
	}

	for _, colName := range columnNames {
		views, err := s.findViews(ctx, schema, table, colName)
		if err != nil {
			log.Printf("Error querying views for column '%s': %v", colName, err)
		} else {
			report.Dependencies = append(report.Dependencies, views...)
		}

		funcs, err := s.findFunctions(ctx, schema, table, colName)
		if err != nil {
			log.Printf("Error querying functions for column '%s': %v", colName, err)
		} else {
			report.Dependencies = append(report.Dependencies, funcs...)
		}
	}

	report.Total = len(report.Dependencies)
	log.Printf("Found %d dependencies for %s.%s.%s", report.Total, schema, table, column)
	return report, nil
}

// findViews returns every view that depends on the column.
func (s *ImpactService) findViews(
	ctx context.Context,
	schema, table, column string,
) ([]models.ColumnDependency, error) {
	query := `
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

	rows, err := s.pool.Query(ctx, query, schema, table, column)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.ColumnDependency{}
	for rows.Next() {
		var dep models.ColumnDependency
		if err := rows.Scan(&dep.ObjectType, &dep.SchemaName, &dep.ObjectName, &dep.Definition); err != nil {
			continue
		}
		dep.DependencyType = "view"
		out = append(out, dep)
	}
	return out, rows.Err()
}

// findFunctions returns every function that depends on the column.
func (s *ImpactService) findFunctions(
	ctx context.Context,
	schema, table, column string,
) ([]models.ColumnDependency, error) {
	query := `
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

	rows, err := s.pool.Query(ctx, query, schema, table, column)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.ColumnDependency{}
	for rows.Next() {
		var dep models.ColumnDependency
		if err := rows.Scan(&dep.ObjectType, &dep.SchemaName, &dep.ObjectName, &dep.Definition); err != nil {
			continue
		}
		dep.DependencyType = "function"
		out = append(out, dep)
	}
	return out, rows.Err()
}

// getAllColumnNames walks the lineage chain and returns every name
// connected to the given column, including cycle protection.
func (s *ImpactService) getAllColumnNames(
	ctx context.Context,
	schema, table, column string,
) ([]string, error) {
	names := []string{column}
	seen := map[string]bool{column: true}

	rows, err := s.pool.Query(ctx, `
		SELECT old_column_name, new_column_name
		FROM column_lineage
		WHERE schema_name = $1 AND table_name = $2
	`, schema, table)
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
