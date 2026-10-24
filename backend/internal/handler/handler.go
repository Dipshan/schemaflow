// Package handler implements the HTTP layer for SchemaFlow.
package handler

import (
	"schemaflow-backend/internal/repository"
	"schemaflow-backend/internal/service"
)

// Handler holds the services used by every HTTP handler.
type Handler struct {
	Migrate     *service.MigrateService
	Rewriter    *service.Rewriter
	Impact      *service.ImpactService
	EventsRepo  *repository.SchemaEventsRepo
	LineageRepo *repository.ColumnLineageRepo
}

// New returns a Handler with all services wired in.
func New(
	migrate *service.MigrateService,
	rewriter *service.Rewriter,
	impact *service.ImpactService,
	eventsRepo *repository.SchemaEventsRepo,
	lineageRepo *repository.ColumnLineageRepo,
) *Handler {
	return &Handler{
		Migrate:     migrate,
		Rewriter:    rewriter,
		Impact:      impact,
		EventsRepo:  eventsRepo,
		LineageRepo: lineageRepo,
	}
}
