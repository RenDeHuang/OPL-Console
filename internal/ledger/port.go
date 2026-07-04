package ledger

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrInsufficientBalance = errors.New("insufficient_balance")
	ErrHoldNotActive       = errors.New("hold_not_active")
)

type Wallet struct {
	BillingAccountID string `json:"billingAccountId"`
	BalanceFen       int64  `json:"balanceFen"`
	FrozenFen        int64  `json:"frozenFen"`
	AvailableFen     int64  `json:"availableFen"`
}

type WorkspaceQuoteRequest struct {
	BillingAccountID string
	PackageID        string
}

type WorkspaceQuote struct {
	BillingAccountID  string `json:"billingAccountId"`
	PackageID         string `json:"packageId"`
	Currency          string `json:"currency"`
	ComputeHourlyFen  int64  `json:"computeHourlyFen"`
	StorageGBMonthFen int64  `json:"storageGbMonthFen"`
	StorageGB         int    `json:"storageGb"`
	HoldDays          int    `json:"holdDays"`
	ComputeHoldFen    int64  `json:"computeHoldFen"`
	StorageHoldFen    int64  `json:"storageHoldFen"`
	TotalHoldFen      int64  `json:"totalHoldFen"`
	BalanceFen        int64  `json:"balanceFen"`
	FrozenFen         int64  `json:"frozenFen"`
	AvailableFen      int64  `json:"availableFen"`
	SufficientBalance bool   `json:"sufficientBalance"`
	Source            string `json:"source"`
}

type HoldRequest struct {
	HoldID           string
	BillingAccountID string
	ResourceType     string
	ResourceID       string
	AmountFen        int64
	ActorUserID      string
}

type TopUpRequest struct {
	TopUpID          string
	BillingAccountID string
	AmountFen        int64
	ActorUserID      string
	Note             string
}

type AuditEvent struct {
	ID          string
	ActorUserID string
	Action      string
	ObjectType  string
	ObjectID    string
	RequestID   string
	Result      string
	Metadata    json.RawMessage
}

type Receipt struct {
	ID          string
	ReceiptType string
	SubjectType string
	SubjectID   string
	OperationID string
	Payload     json.RawMessage
}

type LedgerEntry struct {
	ID               string `json:"id"`
	BillingAccountID string `json:"billingAccountId,omitempty"`
	WorkspaceID      string `json:"workspaceId,omitempty"`
	ResourceType     string `json:"resourceType"`
	ResourceID       string `json:"resourceId,omitempty"`
	AmountFen        int64  `json:"amountFen"`
	Kind             string `json:"kind"`
	Description      string `json:"description"`
	CreatedAt        string `json:"createdAt"`
}

type ReconciliationStatus struct {
	Ready     bool   `json:"ready"`
	Status    string `json:"status"`
	CheckedAt string `json:"checkedAt,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type Port interface {
	QuoteWorkspace(ctx context.Context, request WorkspaceQuoteRequest) (WorkspaceQuote, error)
	GetWallet(ctx context.Context, billingAccountID string) (Wallet, error)
	FreezeHold(ctx context.Context, request HoldRequest) error
	ReleaseHold(ctx context.Context, holdID string, actorUserID string) error
	DebitHold(ctx context.Context, holdID string, actorUserID string) error
	RecordManualTopUp(ctx context.Context, request TopUpRequest) error
	BillingLedger(ctx context.Context, billingAccountID string) ([]LedgerEntry, error)
	AdminBillingLedger(ctx context.Context) ([]LedgerEntry, error)
	ReconciliationStatus(ctx context.Context) (ReconciliationStatus, error)
	RecordAuditEvent(ctx context.Context, event AuditEvent) error
	RecordReceipt(ctx context.Context, receipt Receipt) error
	Receipts(ctx context.Context, subjectID string) ([]Receipt, error)
}
