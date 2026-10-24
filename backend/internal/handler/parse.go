package handler

import (
	"encoding/json"
	"net/http"

	"schemaflow-backend/internal/httpx"
	"schemaflow-backend/internal/service"
)

type parseRequest struct {
	Query string `json:"query"`
}

// ParseHandler handles POST /parse.
func (h *Handler) ParseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var req parseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid JSON")
		return
	}

	if req.Query == "" {
		httpx.BadRequest(w, "Missing required field: query")
		return
	}

	parsed, err := service.ParseQuery(req.Query)
	if err != nil {
		httpx.InternalError(w, "Failed to parse query: "+err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"columns": parsed.GetColumnNames(),
		"tables":  parsed.GetTableNames(),
		"counts": map[string]int{
			"columns": len(parsed.Columns),
			"tables":  len(parsed.Tables),
		},
	})
}
