package handler

import (
	"net/http"

	"schemaflow-backend/internal/httpx"
)

// MigrationsHandler handles GET /migrations.
func (h *Handler) MigrationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	migrations, err := h.EventsRepo.List(r.Context())
	if err != nil {
		httpx.InternalError(w, "Failed to list migrations: "+err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, migrations)
}
