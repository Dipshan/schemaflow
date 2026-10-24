// Package models defines shared domain types.
package models

// MigrateRequest describes a migration instruction sent by a caller.
type MigrateRequest struct {
	EventType string `json:"event_type"`
	TableName string `json:"table_name"`
	OldName   string `json:"old_name"`
	NewName   string `json:"new_name"`
	AppliedBy string `json:"applied_by"`
}

// MigrateResponse is returned to the caller after a migration is recorded.
type MigrateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
