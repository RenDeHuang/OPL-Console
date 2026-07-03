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

type DestroyStorageRequest struct {
	StorageID string
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

type Port interface {
	CreateCompute(ctx context.Context, request CreateComputeRequest) (RuntimeHandle, error)
	CreateStorage(ctx context.Context, request CreateStorageRequest) (RuntimeHandle, error)
	AttachStorage(ctx context.Context, request AttachStorageRequest) (RuntimeHandle, error)
	CreateWorkspaceRoute(ctx context.Context, request CreateRouteRequest) (RuntimeHandle, error)
	DestroyCompute(ctx context.Context, request DestroyComputeRequest) error
	DestroyStorage(ctx context.Context, request DestroyStorageRequest) error
	DestroyWorkspaceRoute(ctx context.Context, request DestroyWorkspaceRouteRequest) error
	ResetWorkspaceToken(ctx context.Context, request ResetWorkspaceTokenRequest) (RuntimeHandle, error)
	DeleteWorkspaceToken(ctx context.Context, request DeleteWorkspaceTokenRequest) error
}
