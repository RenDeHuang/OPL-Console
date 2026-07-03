package ledger

import "context"

type Wallet struct {
	BillingAccountID string `json:"billingAccountId"`
	BalanceFen       int64  `json:"balanceFen"`
	FrozenFen        int64  `json:"frozenFen"`
	AvailableFen     int64  `json:"availableFen"`
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

type Port interface {
	GetWallet(ctx context.Context, billingAccountID string) (Wallet, error)
	FreezeHold(ctx context.Context, request HoldRequest) error
	ReleaseHold(ctx context.Context, holdID string, actorUserID string) error
	DebitHold(ctx context.Context, holdID string, actorUserID string) error
	RecordManualTopUp(ctx context.Context, request TopUpRequest) error
}
