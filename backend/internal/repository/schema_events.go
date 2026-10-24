// Package repository provides data access for SchemaFlow tables.
// Each repository owns the SQL for exactly one table.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SchemaEventsRepo provides data access for the schema_events table.
type SchemaEventsRepo struct {
	pool *pgxpool.Pool
}

// NewSchemaEventsRepo returns a new repository backed by the given pool.
func NewSchemaEventsRepo(pool *pgxpool.Pool) *SchemaEventsRepo {
	return &SchemaEventsRepo{pool: pool}
}

// NextVersion returns the next sequential version number.
// If the table is empty, it returns 1.
func (r *SchemaEventsRepo) NextVersion(ctx context.Context) (int, error) {
	var v int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(version), 0) + 1 FROM schema_events`,
	).Scan(&v)
	return v, err
}

// Insert records a new migration event and returns the generated event_id.
// For column-level events, column_name is set to old_name for now.
func (r *SchemaEventsRepo) Insert(
	ctx context.Context,
	version int,
	eventType, tableName, oldName, newName, appliedBy string,
) (string, error) {
	var eventID string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO schema_events
			(version, event_type, table_name, column_name, old_name, new_name, applied_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING event_id`,
		version, eventType, tableName, oldName, oldName, newName, appliedBy,
	).Scan(&eventID)
	return eventID, err
}

// List returns all migration events, most recent first.
func (r *SchemaEventsRepo) List(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, version, event_type, table_name,
		       old_name, new_name, applied_at, applied_by
		FROM schema_events
		ORDER BY version DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []map[string]interface{}{}

	for rows.Next() {
		var eventID, eventType, tableName, oldName, newName, appliedBy string
		var version int
		var appliedAt interface{}

		if err := rows.Scan(
			&eventID, &version, &eventType, &tableName,
			&oldName, &newName, &appliedAt, &appliedBy,
		); err != nil {
			return nil, err
		}

		out = append(out, map[string]interface{}{
			"event_id":   eventID,
			"version":    version,
			"event_type": eventType,
			"table_name": tableName,
			"old_name":   oldName,
			"new_name":   newName,
			"applied_at": appliedAt,
			"applied_by": appliedBy,
		})
	}

	return out, rows.Err()
}
