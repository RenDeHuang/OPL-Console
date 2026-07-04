package api

import (
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

type computeResourceRequest struct {
	ComputeID        string             `json:"computeId"`
	BillingAccountID string             `json:"billingAccountId"`
	Package          fabric.PackagePlan `json:"package"`
}

type storageVolumeRequest struct {
	StorageID        string             `json:"storageId"`
	BillingAccountID string             `json:"billingAccountId"`
	Package          fabric.PackagePlan `json:"package"`
}

type attachmentRequest struct {
	AttachmentID string `json:"attachmentId"`
	ComputeID    string `json:"computeId"`
	StorageID    string `json:"storageId"`
	MountPath    string `json:"mountPath"`
}

func mountFabricRoutes(router Router, deps Dependencies) {
	router.Post("/api/compute-resources", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Fabric == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "fabric_not_configured"})
			return
		}
		var payload computeResourceRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		handle, err := deps.Fabric.CreateCompute(r.Context(), fabric.CreateComputeRequest{
			ComputeID:        payload.ComputeID,
			BillingAccountID: payload.BillingAccountID,
			Package:          payload.Package,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "compute_resource_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, handle)
	})

	router.Post("/api/compute-resources/destroy", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Fabric == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "fabric_not_configured"})
			return
		}
		var payload computeResourceRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		if err := deps.Fabric.DestroyCompute(r.Context(), fabric.DestroyComputeRequest{ComputeID: payload.ComputeID}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "compute_resource_destroy_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	router.Post("/api/storage-volumes", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Fabric == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "fabric_not_configured"})
			return
		}
		var payload storageVolumeRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		handle, err := deps.Fabric.CreateStorage(r.Context(), fabric.CreateStorageRequest{
			StorageID:        payload.StorageID,
			BillingAccountID: payload.BillingAccountID,
			Package:          payload.Package,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "storage_volume_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, handle)
	})

	router.Post("/api/storage-volumes/destroy", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Fabric == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "fabric_not_configured"})
			return
		}
		var payload storageVolumeRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		if err := deps.Fabric.DestroyStorage(r.Context(), fabric.DestroyStorageRequest{StorageID: payload.StorageID}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "storage_volume_destroy_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	router.Post("/api/storage-attachments", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Fabric == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "fabric_not_configured"})
			return
		}
		var payload attachmentRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		handle, err := deps.Fabric.AttachStorage(r.Context(), fabric.AttachStorageRequest{
			AttachmentID: payload.AttachmentID,
			ComputeID:    payload.ComputeID,
			StorageID:    payload.StorageID,
			MountPath:    payload.MountPath,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "storage_attachment_create_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, handle)
	})

	router.Post("/api/storage-attachments/detach", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "storage_attachment_detach_not_implemented"})
	})
}
