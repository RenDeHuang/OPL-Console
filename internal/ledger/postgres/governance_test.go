package postgres

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

func TestRecordGovernanceAuditAndReceipt(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)

	auditID := testID(t, "audit")
	receiptID := testID(t, "receipt")
	actorID := testID(t, "actor")
	metadata := json.RawMessage(`{"policy":"managed-workspace"}`)
	payload := json.RawMessage(`{"stage":"approved"}`)

	if err := store.RecordAuditEvent(ctx, ledger.AuditEvent{
		ID:          auditID,
		ActorUserID: actorID,
		Action:      "workspace.create.requested",
		ObjectType:  "workspace",
		ObjectID:    "ws-alpha",
		RequestID:   "req-alpha",
		Result:      "succeeded",
		Metadata:    metadata,
	}); err != nil {
		t.Fatalf("RecordAuditEvent: %v", err)
	}
	if err := store.RecordReceipt(ctx, ledger.Receipt{
		ID:          receiptID,
		ReceiptType: "governance.workspace.create",
		SubjectType: "workspace",
		SubjectID:   "ws-alpha",
		OperationID: "op-alpha",
		Payload:     payload,
	}); err != nil {
		t.Fatalf("RecordReceipt: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM receipts WHERE id = $1`, receiptID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM audit_events WHERE id = $1`, auditID)
	})

	var gotAction string
	if err := pool.QueryRow(ctx, `SELECT action FROM audit_events WHERE id = $1`, auditID).Scan(&gotAction); err != nil {
		t.Fatalf("query audit: %v", err)
	}
	if gotAction != "workspace.create.requested" {
		t.Fatalf("action = %q", gotAction)
	}
	var gotReceiptType string
	if err := pool.QueryRow(ctx, `SELECT receipt_type FROM receipts WHERE id = $1`, receiptID).Scan(&gotReceiptType); err != nil {
		t.Fatalf("query receipt: %v", err)
	}
	if gotReceiptType != "governance.workspace.create" {
		t.Fatalf("receipt_type = %q", gotReceiptType)
	}
}
