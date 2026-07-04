package api

import (
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

type taskReceiptRequest struct {
	ID          string          `json:"id"`
	ReceiptType string          `json:"receiptType"`
	SubjectType string          `json:"subjectType"`
	SubjectID   string          `json:"subjectId"`
	OperationID string          `json:"operationId"`
	TaskID      string          `json:"taskId"`
	WorkspaceID string          `json:"workspaceId"`
	Payload     json.RawMessage `json:"payload"`
}

func mountLedgerRoutes(router Router, deps Dependencies) {
	router.Get("/api/ledger/task-receipts", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"receipts": []any{}})
	})

	router.Post("/api/ledger/task-receipts", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		if deps.Ledger == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "ledger_not_configured"})
			return
		}
		var payload taskReceiptRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		if payload.ID == "" {
			id, err := apiID("receipt")
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "receipt_id_failed"})
				return
			}
			payload.ID = id
		}
		if payload.ReceiptType == "" {
			payload.ReceiptType = "task.evidence"
		}
		if payload.SubjectType == "" {
			payload.SubjectType = "workspace"
		}
		if payload.SubjectID == "" {
			payload.SubjectID = payload.WorkspaceID
		}
		if payload.OperationID == "" {
			payload.OperationID = payload.TaskID
		}
		if len(payload.Payload) == 0 {
			payload.Payload = []byte(`{}`)
		}
		if err := deps.Ledger.RecordReceipt(r.Context(), ledger.Receipt{
			ID:          payload.ID,
			ReceiptType: payload.ReceiptType,
			SubjectType: payload.SubjectType,
			SubjectID:   payload.SubjectID,
			OperationID: payload.OperationID,
			Payload:     payload.Payload,
		}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "task_receipt_record_failed"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": payload.ID})
	})
}
