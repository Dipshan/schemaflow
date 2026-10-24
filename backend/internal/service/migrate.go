// Package service implements the business logic of SchemaFlow.
package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"schemaflow-backend/internal/models"
	"schemaflow-backend/internal/repository"
)

// MigrateService records schema migrations and maintains column lineage.
type MigrateService struct {
	events  *repository.SchemaEventsRepo
	lineage *repository.ColumnLineageRepo
}

// NewMigrateService wires a migrate service with its repositories.
func NewMigrateService(
	events *repository.SchemaEventsRepo,
	lineage *repository.ColumnLineageRepo,
) *MigrateService {
	return &MigrateService{
		events:  events,
		lineage: lineage,
	}
}

// Record stores a migration event and, for column renames, adds a
// corresponding lineage record.
func (s *MigrateService) Record(ctx context.Context, req models.MigrateRequest) error {
	req.EventType = strings.ToUpper(strings.TrimSpace(req.EventType))
	req.TableName = strings.TrimSpace(req.TableName)
	req.OldName = strings.TrimSpace(req.OldName)
	req.NewName = strings.TrimSpace(req.NewName)
	req.AppliedBy = strings.TrimSpace(req.AppliedBy)

	if req.EventType == "" || req.TableName == "" || req.AppliedBy == "" {
		return errors.New("missing required fields: event_type, table_name, applied_by")
	}

	if req.EventType == "RENAME_COLUMN" && req.OldName == req.NewName {
		return errors.New("old_name and new_name cannot be the same")
	}

	version, err := s.events.NextVersion(ctx)
	if err != nil {
		log.Println("Error getting next version:", err)
		return err
	}

	eventID, err := s.events.Insert(
		ctx,
		version,
		req.EventType,
		req.TableName,
		req.OldName,
		req.NewName,
		req.AppliedBy,
	)
	if err != nil {
		log.Println("Error inserting schema event:", err)
		return err
	}

	log.Printf("Migration saved: version %d, event_id %s, type %s on table %s",
		version, eventID, req.EventType, req.TableName)

	if req.EventType == "RENAME_COLUMN" {
		if err := s.lineage.MarkOldAsSuperseded(ctx, "public", req.TableName, req.OldName, req.NewName); err != nil {
			log.Println("Warning: Could not supersede old lineage:", err)
		}

		if err := s.lineage.Insert(ctx, eventID, "public", req.TableName, req.OldName, req.NewName); err != nil {
			log.Println("Error inserting lineage record:", err)
			return err
		}

		log.Printf("Lineage saved: %s -> %s on table %s", req.OldName, req.NewName, req.TableName)
	}

	return nil
}
