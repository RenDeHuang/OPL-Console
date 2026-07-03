package console

import (
	"context"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/auth"
)

type fakeRepository struct {
	user         UserView
	organization OrganizationView
	packages     []Package
	workspaces   []ManagedWorkspace
	adminUsers   []UserView
}

func (f fakeRepository) UserByID(ctx context.Context, userID string) (UserView, error) {
	return f.user, nil
}

func (f fakeRepository) PrimaryOrganizationForUser(ctx context.Context, userID string) (OrganizationView, error) {
	return f.organization, nil
}

func (f fakeRepository) Packages(ctx context.Context) ([]Package, error) {
	return f.packages, nil
}

func (f fakeRepository) WorkspacesForUser(ctx context.Context, userID string) ([]ManagedWorkspace, error) {
	return f.workspaces, nil
}

func (f fakeRepository) AdminUsers(ctx context.Context) ([]UserView, error) {
	return f.adminUsers, nil
}

func TestServiceMeCombinesUserAndOrganization(t *testing.T) {
	service := NewService(fakeRepository{
		user:         UserView{ID: "usr-owner", Email: "owner@opl.local", Role: "owner", Status: "active"},
		organization: OrganizationView{ID: "org-alpha", Name: "Alpha Lab", Status: "active"},
	})

	me, err := service.Me(context.Background(), auth.User{ID: "usr-owner"})
	if err != nil {
		t.Fatalf("Me: %v", err)
	}

	if me.User.ID != "usr-owner" || me.Organization.ID != "org-alpha" {
		t.Fatalf("me = %#v", me)
	}
}

func TestServiceExposesReadModelLists(t *testing.T) {
	service := NewService(fakeRepository{
		packages:   []Package{{ID: "basic", Name: "Basic Workspace"}},
		workspaces: []ManagedWorkspace{{ID: "ws-alpha", Name: "Alpha Workspace"}},
		adminUsers: []UserView{{ID: "usr-admin", Email: "admin@opl.local"}},
	})

	packages, err := service.Packages(context.Background())
	if err != nil {
		t.Fatalf("Packages: %v", err)
	}
	workspaces, err := service.Workspaces(context.Background(), auth.User{ID: "usr-owner"})
	if err != nil {
		t.Fatalf("Workspaces: %v", err)
	}
	users, err := service.AdminUsers(context.Background())
	if err != nil {
		t.Fatalf("AdminUsers: %v", err)
	}

	if len(packages) != 1 || len(workspaces) != 1 || len(users) != 1 {
		t.Fatalf("packages=%#v workspaces=%#v users=%#v", packages, workspaces, users)
	}
}
