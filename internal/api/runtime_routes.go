package api

import "net/http"

func mountRuntimeMutationRoutes(router Router, deps Dependencies) {
	router.Post("/api/workspaces/runtime-status", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "workspace_runtime_status_not_implemented"})
	})
}
