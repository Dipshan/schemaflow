package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ColumnLineageRepo provides data access for the column_lineage table.
type ColumnLineageRepo struct {
	pool *pgxpool.Pool
}

// NewColumnLineageRepo returns a new repository backed by the given pool.
func NewColumnLineageRepo(pool *pgxpool.Pool) *ColumnLineageRepo {
	return &ColumnLineageRepo{pool: pool}
}

// MarkOldAsSuperseded sets is_current to false on any existing mapping
// for the given old column name. This is called before inserting a new
// rename for the same column.
func (r *ColumnLineageRepo) MarkOldAsSuperseded(
	ctx context.Context,
	schemaName, tableName, oldName, newName string,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE column_lineage
		SET is_current = false, valid_to = now(), current_column_name = $4
		WHERE schema_name = $1
		  AND table_name = $2
		  AND old_column_name = $3
		  AND is_current = true
	`, schemaName, tableName, oldName, newName)
	return err
}

// Insert records a new old-to-new column mapping.
func (r *ColumnLineageRepo) Insert(
	ctx context.Context,
	eventID, schemaName, tableName, oldName, newName string,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO column_lineage
			(schema_event_id, schema_name, table_name,
			 old_column_name, new_column_name, current_column_name,
			 is_current, valid_from, valid_to, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, now(), NULL, now())
	`, eventID, schemaName, tableName, oldName, newName, newName)
	return err
}

// List returns all lineage records, most recent first.
func (r *ColumnLineageRepo) List(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lineage_id, schema_event_id, schema_name, table_name,
		       old_column_name, new_column_name, is_current,
		       valid_from, valid_to
		FROM column_lineage
		ORDER BY lineage_id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []map[string]interface{}{}

	for rows.Next() {
		var lineageID int
		var eventID, schemaName, tableName, oldName, newName string
		var isCurrent bool
		var validFrom, validTo interface{}

		if err := rows.Scan(
			&lineageID, &eventID, &schemaName, &tableName,
			&oldName, &newName, &isCurrent,
			&validFrom, &validTo,
		); err != nil {
			return nil, err
		}

		out = append(out, map[string]interface{}{
			"lineage_id":      lineageID,
			"schema_event_id": eventID,
			"schema_name":     schemaName,
			"table_name":      tableName,
			"old_column_name": oldName,
			"new_column_name": newName,
			"is_current":      isCurrent,
			"valid_from":      validFrom,
			"valid_to":        validTo,
		})
	}

	return out, rows.Err()
}
