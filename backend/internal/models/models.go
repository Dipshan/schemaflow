// Package models defines shared domain types.
package models

import "time"

// SchemaEvent represents a single row in the schema_events table.
type SchemaEvent struct {
	EventID   string    `json:"event_id"`
	Version   int       `json:"version"`
	EventType string    `json:"event_type"`
	TableName string    `json:"table_name"`
	OldName   string    `json:"old_name"`
	NewName   string    `json:"new_name"`
	AppliedAt time.Time `json:"applied_at"`
	AppliedBy string    `json:"applied_by"`
}

// LineageRecord represents a single row in the column_lineage table.
// ValidTo is a pointer so it can be null when the mapping is current.
type LineageRecord struct {
	LineageID     int        `json:"lineage_id"`
	SchemaEventID string     `json:"schema_event_id"`
	SchemaName    string     `json:"schema_name"`
	TableName     string     `json:"table_name"`
	OldColumnName string     `json:"old_column_name"`
	NewColumnName string     `json:"new_column_name"`
	IsCurrent     bool       `json:"is_current"`
	ValidFrom     time.Time  `json:"valid_from"`
	ValidTo       *time.Time `json:"valid_to"`
}
