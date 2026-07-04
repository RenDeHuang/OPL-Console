package fabric

import "context"

type PackagePlan struct {
	ID        string
	CPU       int
	MemoryGB  int
	StorageGB int
}

type CreateComputeRequest struct {
	ComputeID        string
	BillingAccountID string
	Package          PackagePlan
}

type CreateStorageRequest struct {
	StorageID        string
	BillingAccountID string
	Package          PackagePlan
}

type AttachStorageRequest struct {
	AttachmentID string
	ComputeID    string
	StorageID    string
	MountPath    string
}

type DestroyComputeRequest struct {
	ComputeID string
}

type StopComputeRequest struct {
	ComputeID string
}

type RestartComputeRequest struct {
	ComputeID string
}

type DestroyStorageRequest struct {
	StorageID string
}

type DetachStorageRequest struct {
	AttachmentID string
	ComputeID    string
	StorageID    string
}

type BackupStorageRequest struct {
	BackupID    string
	WorkspaceID string
	StorageID   string
}

type RestoreStorageRequest struct {
	WorkspaceID string
	BackupID    string
}

type RuntimeStatusRequest struct {
	WorkspaceID string
	ComputeID   string
	StorageID   string
}

type DestroyWorkspaceRouteRequest struct {
	WorkspaceID string
}

type ResetWorkspaceTokenRequest struct {
	WorkspaceID string
	Token       string
}

type DeleteWorkspaceTokenRequest struct {
	WorkspaceID string
}

type CreateRouteRequest struct {
	WorkspaceID   string
	WorkspaceName string
	ComputeID     string
	Token         string
}

type RuntimeHandle struct {
	ProviderResourceID string
	Status             string
	URL                string
}

type RuntimeStatus struct {
	Ready        bool              `json:"ready"`
	WorkspaceID  string            `json:"workspaceId"`
	ComputeState string            `json:"computeState"`
	StorageState string            `json:"storageState"`
	RouteState   string            `json:"routeState"`
	Message      string            `json:"message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Port interface {
	CreateCompute(ctx context.Context, request CreateComputeRequest) (RuntimeHandle, error)
	StopCompute(ctx context.Context, request StopComputeRequest) (RuntimeHandle, error)
	RestartCompute(ctx context.Context, request RestartComputeRequest) (RuntimeHandle, error)
	CreateStorage(ctx context.Context, request CreateStorageRequest) (RuntimeHandle, error)
	AttachStorage(ctx context.Context, request AttachStorageRequest) (RuntimeHandle, error)
	DetachStorage(ctx context.Context, request DetachStorageRequest) (RuntimeHandle, error)
	CreateWorkspaceRoute(ctx context.Context, request CreateRouteRequest) (RuntimeHandle, error)
	CreateStorageBackup(ctx context.Context, request BackupStorageRequest) (RuntimeHandle, error)
	RestoreStorageBackup(ctx context.Context, request RestoreStorageRequest) (RuntimeHandle, error)
	DestroyCompute(ctx context.Context, request DestroyComputeRequest) error
	DestroyStorage(ctx context.Context, request DestroyStorageRequest) error
	DestroyWorkspaceRoute(ctx context.Context, request DestroyWorkspaceRouteRequest) error
	ResetWorkspaceToken(ctx context.Context, request ResetWorkspaceTokenRequest) (RuntimeHandle, error)
	DeleteWorkspaceToken(ctx context.Context, request DeleteWorkspaceTokenRequest) error
	RuntimeStatus(ctx context.Context, request RuntimeStatusRequest) (RuntimeStatus, error)
}
