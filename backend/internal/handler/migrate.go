package handler

import (
	"encoding/json"
	"net/http"

	"schemaflow-backend/internal/httpx"
	"schemaflow-backend/internal/models"
)

// Migrate handles POST /migrate.
func (h *Handler) MigrateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var req models.MigrateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid JSON")
		return
	}

	if err := h.Migrate.Record(r.Context(), req); err != nil {
		httpx.BadRequest(w, err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, models.MigrateResponse{
		Status:  "saved",
		Message: "Migration recorded successfully",
	})
}
