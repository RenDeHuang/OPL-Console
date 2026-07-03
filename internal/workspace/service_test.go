package workspace

import (
	"context"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/fabric/local"
)

func TestCreateWorkspaceUsesFabricPort(t *testing.T) {
	fabric := local.New()
	service := NewService(fabric)

	result, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if result.WorkspaceID != "ws-alpha" {
		t.Fatalf("workspace id = %q", result.WorkspaceID)
	}
	if result.URL == "" {
		t.Fatal("workspace URL is empty")
	}
}
