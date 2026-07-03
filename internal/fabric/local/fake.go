package local

import (
	"context"
	"fmt"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

type Fake struct{}

var _ fabric.Port = (*Fake)(nil)

func New() *Fake {
	return &Fake{}
}

func (f *Fake) CreateCompute(ctx context.Context, request fabric.CreateComputeRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: "local-compute/" + request.ComputeID, Status: "running"}, nil
}

func (f *Fake) CreateStorage(ctx context.Context, request fabric.CreateStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: "local-storage/" + request.StorageID, Status: "available"}, nil
}

func (f *Fake) AttachStorage(ctx context.Context, request fabric.AttachStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: fmt.Sprintf("%s:%s", request.ComputeID, request.StorageID), Status: "attached"}, nil
}

func (f *Fake) DestroyCompute(ctx context.Context, request fabric.DestroyComputeRequest) error {
	return nil
}

func (f *Fake) DestroyStorage(ctx context.Context, request fabric.DestroyStorageRequest) error {
	return nil
}

func (f *Fake) DestroyWorkspaceRoute(ctx context.Context, request fabric.DestroyWorkspaceRouteRequest) error {
	return nil
}

func (f *Fake) ResetWorkspaceToken(ctx context.Context, request fabric.ResetWorkspaceTokenRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{
		ProviderResourceID: "local-route/" + request.WorkspaceID,
		Status:             "ready",
		URL:                workspaceRouteURL(request.WorkspaceID, request.Token),
	}, nil
}

func (f *Fake) DeleteWorkspaceToken(ctx context.Context, request fabric.DeleteWorkspaceTokenRequest) error {
	return nil
}

func (f *Fake) CreateWorkspaceRoute(ctx context.Context, request fabric.CreateRouteRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{
		ProviderResourceID: "local-route/" + request.WorkspaceID,
		Status:             "ready",
		URL:                workspaceRouteURL(request.WorkspaceID, request.Token),
	}, nil
}

func workspaceRouteURL(workspaceID, token string) string {
	return "http://127.0.0.1:8787/w/" + workspaceID + "?token=" + token
}
