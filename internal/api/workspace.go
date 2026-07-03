package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/workspace"
)

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, request workspace.CreateWorkspaceRequest) (workspace.CreateWorkspaceResult, error)
}

func mountWorkspaceRoutes(router Router, deps Dependencies) {
	router.Post("/api/workspaces", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Workspace == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "workspace_facade_not_configured"})
			return
		}
		var payload workspace.CreateWorkspaceRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		result, err := deps.Workspace.CreateWorkspace(r.Context(), payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspace_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})
}
