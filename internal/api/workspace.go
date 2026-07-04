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
	ConfigureWorkspace(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error)
	SuspendWorkspace(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error)
	DeleteWorkspace(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error)
	ResetWorkspaceToken(ctx context.Context, request workspace.TokenRequest) (workspace.ActionResult, error)
	DeleteWorkspaceToken(ctx context.Context, request workspace.ActionRequest) (workspace.ActionResult, error)
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

	router.Post("/api/workspaces/{id}/configure", func(w http.ResponseWriter, r *http.Request) {
		lifecycleAction(w, r, deps, func(ctx context.Context, service WorkspaceService, request workspace.ActionRequest) (workspace.ActionResult, error) {
			return service.ConfigureWorkspace(ctx, request)
		})
	})
	router.Post("/api/workspaces/{id}/suspend", func(w http.ResponseWriter, r *http.Request) {
		lifecycleAction(w, r, deps, func(ctx context.Context, service WorkspaceService, request workspace.ActionRequest) (workspace.ActionResult, error) {
			return service.SuspendWorkspace(ctx, request)
		})
	})
	router.Post("/api/workspaces/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		lifecycleAction(w, r, deps, func(ctx context.Context, service WorkspaceService, request workspace.ActionRequest) (workspace.ActionResult, error) {
			return service.DeleteWorkspace(ctx, request)
		})
	})
	router.Post("/api/workspaces/{id}/tokens/reset", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireOwner(w, r, deps)
		if !ok {
			return
		}
		if deps.Workspace == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "workspace_facade_not_configured"})
			return
		}
		var payload workspace.TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		payload.WorkspaceID = chi.URLParam(r, "id")
		payload.ActorUserID = session.User.ID
		result, err := deps.Workspace.ResetWorkspaceToken(r.Context(), payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspace_token_reset_failed"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	router.Post("/api/workspaces/{id}/tokens/delete", func(w http.ResponseWriter, r *http.Request) {
		lifecycleAction(w, r, deps, func(ctx context.Context, service WorkspaceService, request workspace.ActionRequest) (workspace.ActionResult, error) {
			return service.DeleteWorkspaceToken(ctx, request)
		})
	})
}

func lifecycleAction(
	w http.ResponseWriter,
	r *http.Request,
	deps Dependencies,
	run func(context.Context, WorkspaceService, workspace.ActionRequest) (workspace.ActionResult, error),
) {
	session, ok := requireOwner(w, r, deps)
	if !ok {
		return
	}
	if deps.Workspace == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "workspace_facade_not_configured"})
		return
	}
	result, err := run(r.Context(), deps.Workspace, workspace.ActionRequest{
		WorkspaceID: chi.URLParam(r, "id"),
		ActorUserID: session.User.ID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspace_lifecycle_failed"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
