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
}
