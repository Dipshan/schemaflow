package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"schemaflow-backend/internal/httpx"
)

type impactRequest struct {
	Schema string `json:"schema"`
	Table  string `json:"table"`
	Column string `json:"column"`
}

// ImpactHandler handles POST /impact.
func (h *Handler) ImpactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var req impactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid JSON")
		return
	}

	if req.Schema == "" {
		req.Schema = "public"
	}
	if req.Table == "" || req.Column == "" {
		httpx.BadRequest(w, "Missing required fields: table, column")
		return
	}

	report, err := h.Impact.GetDependencies(r.Context(), req.Schema, req.Table, req.Column)
	if err != nil {
		httpx.InternalError(w, "Failed to get dependencies: "+err.Error())
		return
	}

	resp := map[string]interface{}{
		"column":           req.Column,
		"table":            req.Table,
		"schema":           req.Schema,
		"dependencies":     report.Dependencies,
		"total":            report.Total,
		"has_dependencies": report.HasDependencies(),
	}

	if report.HasDependencies() {
		resp["warning"] = fmt.Sprintf("%d dependencies found. Renaming this column may break them.", report.Total)
	} else {
		resp["message"] = "No dependencies found. Safe to rename this column."
	}

	httpx.JSON(w, http.StatusOK, resp)
}
