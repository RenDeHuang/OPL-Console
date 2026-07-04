package console

import "encoding/json"

type UserView struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

type OrganizationView struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type TeamView struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	Status         string `json:"status"`
}

type RoleView struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId,omitempty"`
	Name           string          `json:"name"`
	Scope          string          `json:"scope"`
	Permissions    json.RawMessage `json:"permissions,omitempty"`
}

type Me struct {
	User         UserView         `json:"user"`
	Organization OrganizationView `json:"organization"`
}

type Package struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	CPU               int    `json:"cpu"`
	MemoryGB          int    `json:"memoryGb"`
	StorageGB         int    `json:"storageGb"`
	ComputeHourlyFen  int64  `json:"computeHourlyFen"`
	StorageGBMonthFen int64  `json:"storageGbMonthFen"`
	Available         bool   `json:"available"`
}

type ManagedWorkspace struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	State            string `json:"state"`
	Policy           string `json:"policy"`
	URL              string `json:"url,omitempty"`
	Provider         string `json:"provider,omitempty"`
	PackageID        string `json:"packageId,omitempty"`
	ComputeStatus    string `json:"computeStatus,omitempty"`
	StorageStatus    string `json:"storageStatus,omitempty"`
	AttachmentStatus string `json:"attachmentStatus,omitempty"`
	TokenStatus      string `json:"tokenStatus,omitempty"`
	EstimatedHoldFen int64  `json:"estimatedHoldFen,omitempty"`
}

type WorkspaceDetail struct {
	ManagedWorkspace
	BillingAccountID string                   `json:"billingAccountId"`
	ComputeID        string                   `json:"computeId,omitempty"`
	StorageID        string                   `json:"storageId,omitempty"`
	AttachmentID     string                   `json:"attachmentId,omitempty"`
	Package          Package                  `json:"package"`
	LifecycleSteps   []LifecycleStepView      `json:"lifecycleSteps"`
	LedgerEntries    []BillingLedgerEntryView `json:"ledgerEntries"`
	Receipts         []ReceiptView            `json:"receipts"`
	SupportTickets   []SupportTicketView      `json:"supportTickets"`
	AuditEvents      []AuditEventView         `json:"auditEvents"`
}

type LifecycleStepView struct {
	StepName           string `json:"stepName"`
	DesiredState       string `json:"desiredState"`
	ActualState        string `json:"actualState"`
	ProviderResourceID string `json:"providerResourceId,omitempty"`
	ErrorCode          string `json:"errorCode,omitempty"`
	LastCheckedAt      string `json:"lastCheckedAt,omitempty"`
}

type ManagedResourceView struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	ResourceType   string          `json:"resourceType"`
	ResourceID     string          `json:"resourceId"`
	DisplayName    string          `json:"displayName"`
	Provider       string          `json:"provider"`
	Status         string          `json:"status"`
	PolicyState    string          `json:"policyState"`
	WorkspaceID    string          `json:"workspaceId,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

type WalletView struct {
	BillingAccountID string `json:"billingAccountId"`
	BalanceFen       int64  `json:"balanceFen"`
	FrozenFen        int64  `json:"frozenFen"`
	AvailableFen     int64  `json:"availableFen"`
}

type BillingLedgerEntryView struct {
	ID           string `json:"id"`
	WorkspaceID  string `json:"workspaceId,omitempty"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId,omitempty"`
	AmountFen    int64  `json:"amountFen"`
	Kind         string `json:"kind"`
	Description  string `json:"description"`
	CreatedAt    string `json:"createdAt"`
}

type WorkspaceQuoteView struct {
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

type WorkspaceQuoteRequest struct {
	BillingAccountID string `json:"billingAccountId"`
	PackageID        string `json:"packageId"`
}

type ReceiptView struct {
	ID          string          `json:"id"`
	ReceiptType string          `json:"receiptType"`
	SubjectType string          `json:"subjectType"`
	SubjectID   string          `json:"subjectId"`
	OperationID string          `json:"operationId,omitempty"`
	Payload     json.RawMessage `json:"payload,omitempty"`
}

type AuditEventView struct {
	ID          string          `json:"id"`
	ActorUserID string          `json:"actorUserId"`
	Action      string          `json:"action"`
	ObjectType  string          `json:"objectType"`
	ObjectID    string          `json:"objectId"`
	Result      string          `json:"result"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
}

type SupportTicketView struct {
	ID                  string          `json:"id"`
	WorkspaceID         string          `json:"workspaceId,omitempty"`
	Subject             string          `json:"subject"`
	Body                string          `json:"body,omitempty"`
	Status              string          `json:"status"`
	Priority            string          `json:"priority,omitempty"`
	AssigneeUserID      string          `json:"assigneeUserId,omitempty"`
	FailedLifecycleStep string          `json:"failedLifecycleStep,omitempty"`
	FabricErrorCode     string          `json:"fabricErrorCode,omitempty"`
	RuntimeStatus       json.RawMessage `json:"runtimeStatus,omitempty"`
	LedgerSummary       json.RawMessage `json:"ledgerSummary,omitempty"`
	CreatedAt           string          `json:"createdAt,omitempty"`
}

type CreateSupportTicketRequest struct {
	WorkspaceID string `json:"workspaceId"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
}

type PolicyView struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	Name           string          `json:"name"`
	PolicyType     string          `json:"policyType"`
	Status         string          `json:"status"`
	Rules          json.RawMessage `json:"rules,omitempty"`
}

type CreatePolicyRequest struct {
	OrganizationID string          `json:"organizationId"`
	Name           string          `json:"name"`
	PolicyType     string          `json:"policyType"`
	Rules          json.RawMessage `json:"rules"`
}

type ApprovalView struct {
	ID              string          `json:"id"`
	OrganizationID  string          `json:"organizationId,omitempty"`
	PolicyID        string          `json:"policyId,omitempty"`
	RequesterUserID string          `json:"requesterUserId,omitempty"`
	ReviewerUserID  string          `json:"reviewerUserId,omitempty"`
	Action          string          `json:"action,omitempty"`
	ObjectType      string          `json:"objectType,omitempty"`
	ObjectID        string          `json:"objectId,omitempty"`
	Status          string          `json:"status"`
	Reason          string          `json:"reason,omitempty"`
	DecisionNote    string          `json:"decisionNote,omitempty"`
	Context         json.RawMessage `json:"context,omitempty"`
}

type ApprovalDecisionRequest struct {
	ApprovalID   string `json:"approvalId"`
	Decision     string `json:"decision"`
	DecisionNote string `json:"decisionNote"`
}
