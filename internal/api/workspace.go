package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/workspace"
	"github.com/go-chi/chi/v5"
)

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, request workspace.CreateWorkspaceRequest) (workspace.CreateWorkspaceResult, error)
	Handoff(ctx context.Context, request workspace.HandoffRequest) (workspace.HandoffResult, error)
}

func mountWorkspaceRoutes(router Router, deps Dependencies) {
	router.Get("/w/{workspaceId}", func(w http.ResponseWriter, r *http.Request) {
		if deps.Workspace == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "workspace_facade_not_configured"})
			return
		}
		token := r.URL.Query().Get("token")
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing_workspace_token"})
			return
		}
		result, err := deps.Workspace.Handoff(r.Context(), workspace.HandoffRequest{
			WorkspaceID: chi.URLParam(r, "workspaceId"),
			Token:       token,
		})
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "workspace_handoff_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	router.Post("/api/workspaces", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
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
		payload.ActorUserID = session.User.ID
		result, err := deps.Workspace.CreateWorkspace(r.Context(), payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspace_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})
}
