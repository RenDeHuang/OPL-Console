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
	ID       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Policy   string `json:"policy"`
	URL      string `json:"url,omitempty"`
	Provider string `json:"provider,omitempty"`
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

type SupportTicketView struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	Subject     string `json:"subject"`
	Body        string `json:"body,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt,omitempty"`
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
	ID              string `json:"id"`
	OrganizationID  string `json:"organizationId,omitempty"`
	PolicyID        string `json:"policyId,omitempty"`
	RequesterUserID string `json:"requesterUserId,omitempty"`
	ReviewerUserID  string `json:"reviewerUserId,omitempty"`
	Action          string `json:"action,omitempty"`
	ObjectType      string `json:"objectType,omitempty"`
	ObjectID        string `json:"objectId,omitempty"`
	Status          string `json:"status"`
	Reason          string `json:"reason,omitempty"`
	DecisionNote    string `json:"decisionNote,omitempty"`
}

type ApprovalDecisionRequest struct {
	ApprovalID   string `json:"approvalId"`
	Decision     string `json:"decision"`
	DecisionNote string `json:"decisionNote"`
}
