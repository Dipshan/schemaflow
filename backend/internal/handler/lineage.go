package handler

import (
	"net/http"

	"schemaflow-backend/internal/httpx"
)

// LineageHandler handles GET /lineage.
func (h *Handler) LineageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	records, err := h.LineageRepo.List(r.Context())
	if err != nil {
		httpx.InternalError(w, "Failed to list lineage: "+err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, records)
}
