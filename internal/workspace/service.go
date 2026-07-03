package workspace

import (
	"context"
	"fmt"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

type Service struct {
	fabric fabric.Port
}

func NewService(fabricPort fabric.Port) *Service {
	return &Service{fabric: fabricPort}
}

type CreateWorkspaceRequest struct {
	WorkspaceID      string
	Name             string
	BillingAccountID string
	PackageID        string
	Token            string
}

type CreateWorkspaceResult struct {
	WorkspaceID string
	URL         string
}

func (s *Service) CreateWorkspace(ctx context.Context, request CreateWorkspaceRequest) (CreateWorkspaceResult, error) {
	plan := fabric.PackagePlan{ID: request.PackageID, CPU: 2, MemoryGB: 4, StorageGB: 10}
	computeID := "cmp-" + request.WorkspaceID
	storageID := "stg-" + request.WorkspaceID
	attachmentID := "att-" + request.WorkspaceID

	if _, err := s.fabric.CreateStorage(ctx, fabric.CreateStorageRequest{
		StorageID: storageID, BillingAccountID: request.BillingAccountID, Package: plan,
	}); err != nil {
		return CreateWorkspaceResult{}, fmt.Errorf("create storage: %w", err)
	}
	if _, err := s.fabric.CreateCompute(ctx, fabric.CreateComputeRequest{
		ComputeID: computeID, BillingAccountID: request.BillingAccountID, Package: plan,
	}); err != nil {
		return CreateWorkspaceResult{}, fmt.Errorf("create compute: %w", err)
	}
	if _, err := s.fabric.AttachStorage(ctx, fabric.AttachStorageRequest{
		AttachmentID: attachmentID, ComputeID: computeID, StorageID: storageID, MountPath: "/data",
	}); err != nil {
		_ = s.fabric.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: computeID})
		return CreateWorkspaceResult{}, fmt.Errorf("attach storage: %w", err)
	}
	route, err := s.fabric.CreateWorkspaceRoute(ctx, fabric.CreateRouteRequest{
		WorkspaceID: request.WorkspaceID, WorkspaceName: request.Name, ComputeID: computeID, Token: request.Token,
	})
	if err != nil {
		if route.ProviderResourceID != "" {
			_ = s.fabric.DestroyWorkspaceRoute(ctx, fabric.DestroyWorkspaceRouteRequest{WorkspaceID: request.WorkspaceID})
		}
		_ = s.fabric.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: computeID})
		return CreateWorkspaceResult{}, fmt.Errorf("create workspace route: %w", err)
	}
	return CreateWorkspaceResult{WorkspaceID: request.WorkspaceID, URL: route.URL}, nil
}
