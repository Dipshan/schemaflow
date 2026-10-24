package handler

import (
	"encoding/json"
	"net/http"

	"schemaflow-backend/internal/httpx"
)

type rewriteRequest struct {
	Query string `json:"query"`
}

// RewriteHandler handles POST /rewrite.
func (h *Handler) RewriteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var req rewriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid JSON")
		return
	}

	if req.Query == "" {
		httpx.BadRequest(w, "Missing required field: query")
		return
	}

	rewritten, err := h.Rewriter.Rewrite(r.Context(), req.Query)
	if err != nil {
		httpx.InternalError(w, "Failed to rewrite query: "+err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{
		"original":  req.Query,
		"rewritten": rewritten,
	})
}
