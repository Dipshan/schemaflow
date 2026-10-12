// Package main contains the migration logic for SchemaFlow
package main

import (
	"context"
	"log"
)

// SaveMigrationEvent inserts a new migration record into schema_events
// AND if it's a rename, also inserts into column_lineage
func SaveMigrationEvent(eventType, tableName, oldName, newName, appliedBy string) error {
	// STEP 1: Get the next version number
	var nextVersion int
	queryVersion := `SELECT COALESCE(MAX(version), 0) + 1 FROM schema_events`
	err := DB.QueryRow(context.Background(), queryVersion).Scan(&nextVersion)
	if err != nil {
		log.Println("Error getting next version:", err)
		return err
	}

	// STEP 2: Insert into schema_events and get the generated event_id (UUID)
	var eventID string
	insertQuery := `
		INSERT INTO schema_events (
			version, event_type, table_name, column_name, old_name, new_name, applied_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING event_id
	`

	// For column-level events, column_name is the same as old_name (for now)
	columnName := oldName

	err = DB.QueryRow(context.Background(), insertQuery,
		nextVersion, // $1
		eventType,   // $2
		tableName,   // $3
		columnName,  // $4
		oldName,     // $5
		newName,     // $6
		appliedBy,   // $7
	).Scan(&eventID)

	if err != nil {
		log.Println("Error inserting migration event:", err)
		return err
	}

	log.Printf("Migration saved: version %d, event_id %s, type %s on table %s\n",
		nextVersion, eventID, eventType, tableName)

	// STEP 3: If this is a rename column, also insert into column_lineage
	if eventType == "RENAME_COLUMN" {
		err = insertColumnLineage(eventID, tableName, oldName, newName)
		if err != nil {
			log.Println("Error inserting column lineage:", err)
			return err
		}
		log.Printf("Column lineage saved: %s → %s on table %s\n", oldName, newName, tableName)
	}

	return nil
}

// insertColumnLineage inserts a rename mapping into the column_lineage table
func insertColumnLineage(eventID string, tableName, oldName, newName string) error {
	// For MVP, we assume 'public' schema
	schemaName := "public"

	// STEP 1: If this column was previously renamed, mark that old row as not current
	updateOld := `
		UPDATE column_lineage 
		SET is_current = false, valid_to = now()
		WHERE schema_name = $1 
		  AND table_name = $2 
		  AND old_column_name = $3 
		  AND is_current = true
	`
	_, err := DB.Exec(context.Background(), updateOld, schemaName, tableName, oldName)
	if err != nil {
		log.Println("Warning: Could not update old lineage record:", err)
		// Continue anyway — not a fatal error
	}

	// STEP 2: Insert the new rename mapping
	insertQuery := `
		INSERT INTO column_lineage (
			schema_event_id, schema_name, table_name, old_column_name, new_column_name, 
			is_current, valid_from, valid_to, created_at
		) VALUES ($1, $2, $3, $4, $5, true, now(), NULL, now())
	`
	_, err = DB.Exec(context.Background(), insertQuery,
		eventID,    // $1 - UUID from schema_events
		schemaName, // $2
		tableName,  // $3
		oldName,    // $4
		newName,    // $5
	)
	if err != nil {
		return err
	}

	return nil
}

// ListMigrations returns all migrations from schema_events, most recent first.
func ListMigrations() ([]map[string]interface{}, error) {
	query := `
		SELECT event_id, version, event_type, table_name, old_name, new_name, applied_at, applied_by
		FROM schema_events
		ORDER BY version DESC
	`

	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []map[string]interface{}

	for rows.Next() {
		var eventID, eventType, tableName, oldName, newName, appliedBy string
		var version int
		var appliedAt interface{}

		err := rows.Scan(&eventID, &version, &eventType, &tableName, &oldName, &newName, &appliedAt, &appliedBy)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, map[string]interface{}{
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

	if migrations == nil {
		migrations = []map[string]interface{}{}
	}

	return migrations, nil
}

// ListLineage returns all column lineage records, most recent first.
func ListLineage() ([]map[string]interface{}, error) {
	query := `
		SELECT lineage_id, schema_event_id, schema_name, table_name,
		       old_column_name, new_column_name, is_current, valid_from, valid_to
		FROM column_lineage
		ORDER BY lineage_id DESC
	`

	rows, err := DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}

	for rows.Next() {
		var lineageID int
		var eventID, schemaName, tableName, oldName, newName string
		var isCurrent bool
		var validFrom, validTo interface{}

		err := rows.Scan(&lineageID, &eventID, &schemaName, &tableName,
			&oldName, &newName, &isCurrent, &validFrom, &validTo)
		if err != nil {
			return nil, err
		}

		records = append(records, map[string]interface{}{
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

	if records == nil {
		records = []map[string]interface{}{}
	}

	return records, nil
}
