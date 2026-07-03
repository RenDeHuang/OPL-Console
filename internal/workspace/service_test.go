package workspace

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/fabric"
	"github.com/RenDeHuang/opl-console/internal/ledger"
)

func TestCreateWorkspaceUsesFabricPort(t *testing.T) {
	fabricPort := &recordingFabric{
		route: fabric.RuntimeHandle{URL: "https://workspace.example/ws-alpha"},
	}
	service := NewService(fabricPort)

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
	if result.URL != "https://workspace.example/ws-alpha" {
		t.Fatalf("workspace URL = %q", result.URL)
	}

	assertCalls(t, fabricPort.calls, []string{
		"create_storage",
		"create_compute",
		"attach_storage",
		"create_route",
	})
	assertPackage(t, fabricPort.createStorage.Package)
	assertPackage(t, fabricPort.createCompute.Package)
	if fabricPort.createStorage.StorageID != "stg-ws-alpha" {
		t.Fatalf("storage id = %q", fabricPort.createStorage.StorageID)
	}
	if fabricPort.createStorage.BillingAccountID != "acct-owner" {
		t.Fatalf("storage billing account = %q", fabricPort.createStorage.BillingAccountID)
	}
	if fabricPort.createCompute.ComputeID != "cmp-ws-alpha" {
		t.Fatalf("compute id = %q", fabricPort.createCompute.ComputeID)
	}
	if fabricPort.createCompute.BillingAccountID != "acct-owner" {
		t.Fatalf("compute billing account = %q", fabricPort.createCompute.BillingAccountID)
	}
	if fabricPort.attachStorage.AttachmentID != "att-ws-alpha" {
		t.Fatalf("attachment id = %q", fabricPort.attachStorage.AttachmentID)
	}
	if fabricPort.attachStorage.ComputeID != "cmp-ws-alpha" {
		t.Fatalf("attachment compute id = %q", fabricPort.attachStorage.ComputeID)
	}
	if fabricPort.attachStorage.StorageID != "stg-ws-alpha" {
		t.Fatalf("attachment storage id = %q", fabricPort.attachStorage.StorageID)
	}
	if fabricPort.attachStorage.MountPath != "/data" {
		t.Fatalf("mount path = %q", fabricPort.attachStorage.MountPath)
	}
	if fabricPort.createRoute.WorkspaceID != "ws-alpha" {
		t.Fatalf("route workspace id = %q", fabricPort.createRoute.WorkspaceID)
	}
	if fabricPort.createRoute.WorkspaceName != "Alpha Lab" {
		t.Fatalf("route workspace name = %q", fabricPort.createRoute.WorkspaceName)
	}
	if fabricPort.createRoute.ComputeID != "cmp-ws-alpha" {
		t.Fatalf("route compute id = %q", fabricPort.createRoute.ComputeID)
	}
	if fabricPort.createRoute.Token != "share-token" {
		t.Fatalf("route token = %q", fabricPort.createRoute.Token)
	}
}

func TestCreateWorkspaceWrapsComputeErrorAndPreservesStorage(t *testing.T) {
	providerErr := errors.New("provider compute failed")
	fabricPort := &recordingFabric{createComputeErr: providerErr}
	service := NewService(fabricPort)

	_, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("error does not preserve provider error: %v", err)
	}
	if !strings.Contains(err.Error(), "create compute") {
		t.Fatalf("error does not include stage context: %v", err)
	}
	assertCalls(t, fabricPort.calls, []string{
		"create_storage",
		"create_compute",
	})
}

func TestCreateWorkspaceWrapsAttachErrorAndDestroysComputeOnly(t *testing.T) {
	providerErr := errors.New("provider attach failed")
	fabricPort := &recordingFabric{attachStorageErr: providerErr}
	service := NewService(fabricPort)

	_, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("error does not preserve provider error: %v", err)
	}
	if !strings.Contains(err.Error(), "attach storage") {
		t.Fatalf("error does not include stage context: %v", err)
	}
	assertCalls(t, fabricPort.calls, []string{
		"create_storage",
		"create_compute",
		"attach_storage",
		"destroy_compute",
	})
	if fabricPort.destroyCompute.ComputeID != "cmp-ws-alpha" {
		t.Fatalf("destroy compute id = %q", fabricPort.destroyCompute.ComputeID)
	}
}

func TestCreateWorkspaceWrapsRouteErrorAndDestroysComputeOnly(t *testing.T) {
	providerErr := errors.New("provider route failed")
	fabricPort := &recordingFabric{
		route:          fabric.RuntimeHandle{ProviderResourceID: "local-route/ws-alpha"},
		createRouteErr: providerErr,
	}
	service := NewService(fabricPort)

	_, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("error does not preserve provider error: %v", err)
	}
	if !strings.Contains(err.Error(), "create workspace route") {
		t.Fatalf("error does not include stage context: %v", err)
	}
	assertCalls(t, fabricPort.calls, []string{
		"create_storage",
		"create_compute",
		"attach_storage",
		"create_route",
		"destroy_compute",
	})
	if fabricPort.destroyCompute.ComputeID != "cmp-ws-alpha" {
		t.Fatalf("destroy compute id = %q", fabricPort.destroyCompute.ComputeID)
	}
}

func TestCreateWorkspaceDoesNotDestroyRouteOnRouteErrorWithEmptyHandle(t *testing.T) {
	providerErr := errors.New("provider route failed")
	fabricPort := &recordingFabric{createRouteErr: providerErr}
	service := NewService(fabricPort)

	_, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("error does not preserve provider error: %v", err)
	}
	assertCalls(t, fabricPort.calls, []string{
		"create_storage",
		"create_compute",
		"attach_storage",
		"create_route",
		"destroy_compute",
	})
	if fabricPort.destroyCompute.ComputeID != "cmp-ws-alpha" {
		t.Fatalf("destroy compute id = %q", fabricPort.destroyCompute.ComputeID)
	}
}

func TestCreateWorkspaceCreatesApprovalWhenPolicyRequiresReview(t *testing.T) {
	fabricPort := &recordingFabric{}
	repository := &recordingWorkspaceRepository{
		createContext: CreateContext{
			ActorUserID:      "usr-owner",
			OrganizationID:   "org-alpha",
			HoldAmountFen:    100,
			RequiresApproval: true,
			Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
		},
	}
	ledgerPort := &recordingLedger{}
	service := NewService(fabricPort, WithRepository(repository), WithLedger(ledgerPort))

	result, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
		ActorUserID:      "usr-owner",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if result.State != "approval_required" || result.ApprovalID == "" {
		t.Fatalf("result = %#v, want approval_required with approval id", result)
	}
	assertCalls(t, fabricPort.calls, []string{})
	if len(ledgerPort.holds) != 0 {
		t.Fatalf("holds = %#v, want no hold before approval", ledgerPort.holds)
	}
	if len(repository.approvals) != 1 || repository.approvals[0].Action != "workspace.create" {
		t.Fatalf("approvals = %#v", repository.approvals)
	}
	if len(ledgerPort.auditEvents) != 1 || ledgerPort.auditEvents[0].Result != "approval_required" {
		t.Fatalf("audit events = %#v", ledgerPort.auditEvents)
	}
}

func TestCreateWorkspacePersistsFacadeStateAndLedgerEvidence(t *testing.T) {
	fabricPort := &recordingFabric{
		route: fabric.RuntimeHandle{ProviderResourceID: "route/ws-alpha", Status: "ready", URL: "https://workspace.example/ws-alpha"},
	}
	repository := &recordingWorkspaceRepository{
		createContext: CreateContext{
			ActorUserID:    "usr-owner",
			OrganizationID: "org-alpha",
			HoldAmountFen:  100,
			Package:        fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
		},
	}
	ledgerPort := &recordingLedger{}
	service := NewService(fabricPort, WithRepository(repository), WithLedger(ledgerPort))

	result, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
		ActorUserID:      "usr-owner",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if result.State != "ready" || result.URL != "https://workspace.example/ws-alpha" {
		t.Fatalf("result = %#v", result)
	}
	assertCalls(t, fabricPort.calls, []string{
		"create_storage",
		"create_compute",
		"attach_storage",
		"create_route",
	})
	if len(ledgerPort.holds) != 1 || ledgerPort.holds[0].ResourceID != "ws-alpha" {
		t.Fatalf("holds = %#v", ledgerPort.holds)
	}
	if len(repository.created) != 1 {
		t.Fatalf("created records = %#v", repository.created)
	}
	created := repository.created[0]
	if created.WorkspaceID != "ws-alpha" || created.ComputeProviderID == "" || created.StorageProviderID == "" || created.RouteURL == "" {
		t.Fatalf("created record = %#v", created)
	}
	if len(ledgerPort.receipts) != 1 || ledgerPort.receipts[0].ReceiptType != "workspace.governance.created" {
		t.Fatalf("receipts = %#v", ledgerPort.receipts)
	}
	if len(ledgerPort.auditEvents) != 1 || ledgerPort.auditEvents[0].Result != "succeeded" {
		t.Fatalf("audit events = %#v", ledgerPort.auditEvents)
	}
}

type recordingFabric struct {
	calls []string

	createStorage fabric.CreateStorageRequest
	createCompute fabric.CreateComputeRequest
	attachStorage fabric.AttachStorageRequest
	createRoute   fabric.CreateRouteRequest

	destroyCompute fabric.DestroyComputeRequest
	destroyStorage fabric.DestroyStorageRequest
	destroyRoute   fabric.DestroyWorkspaceRouteRequest

	route fabric.RuntimeHandle

	createComputeErr error
	attachStorageErr error
	createRouteErr   error
}

type recordingWorkspaceRepository struct {
	createContext CreateContext
	approvals     []ApprovalRequest
	created       []CreatedWorkspace
}

func (r *recordingWorkspaceRepository) PrepareCreate(ctx context.Context, request CreateWorkspaceRequest) (CreateContext, error) {
	return r.createContext, nil
}

func (r *recordingWorkspaceRepository) CreateApproval(ctx context.Context, request ApprovalRequest) (string, error) {
	r.approvals = append(r.approvals, request)
	return fmt.Sprintf("approval-%d", len(r.approvals)), nil
}

func (r *recordingWorkspaceRepository) SaveCreated(ctx context.Context, record CreatedWorkspace) error {
	r.created = append(r.created, record)
	return nil
}

func (r *recordingWorkspaceRepository) Handoff(ctx context.Context, request HandoffRequest) (HandoffResult, error) {
	return HandoffResult{WorkspaceID: request.WorkspaceID, URL: "https://workspace.example/" + request.WorkspaceID, State: "running"}, nil
}

type recordingLedger struct {
	holds       []ledger.HoldRequest
	auditEvents []ledger.AuditEvent
	receipts    []ledger.Receipt
}

func (r *recordingLedger) GetWallet(ctx context.Context, billingAccountID string) (ledger.Wallet, error) {
	return ledger.Wallet{BillingAccountID: billingAccountID, BalanceFen: 1000, AvailableFen: 900}, nil
}

func (r *recordingLedger) FreezeHold(ctx context.Context, request ledger.HoldRequest) error {
	r.holds = append(r.holds, request)
	return nil
}

func (r *recordingLedger) ReleaseHold(ctx context.Context, holdID string, actorUserID string) error {
	return nil
}

func (r *recordingLedger) DebitHold(ctx context.Context, holdID string, actorUserID string) error {
	return nil
}

func (r *recordingLedger) RecordManualTopUp(ctx context.Context, request ledger.TopUpRequest) error {
	return nil
}

func (r *recordingLedger) RecordAuditEvent(ctx context.Context, event ledger.AuditEvent) error {
	r.auditEvents = append(r.auditEvents, event)
	return nil
}

func (r *recordingLedger) RecordReceipt(ctx context.Context, receipt ledger.Receipt) error {
	r.receipts = append(r.receipts, receipt)
	return nil
}

func (f *recordingFabric) CreateCompute(ctx context.Context, request fabric.CreateComputeRequest) (fabric.RuntimeHandle, error) {
	f.calls = append(f.calls, "create_compute")
	f.createCompute = request
	if f.createComputeErr != nil {
		return fabric.RuntimeHandle{}, f.createComputeErr
	}
	return fabric.RuntimeHandle{ProviderResourceID: "local-compute/" + request.ComputeID, Status: "running"}, nil
}

func (f *recordingFabric) CreateStorage(ctx context.Context, request fabric.CreateStorageRequest) (fabric.RuntimeHandle, error) {
	f.calls = append(f.calls, "create_storage")
	f.createStorage = request
	return fabric.RuntimeHandle{ProviderResourceID: "local-storage/" + request.StorageID, Status: "available"}, nil
}

func (f *recordingFabric) AttachStorage(ctx context.Context, request fabric.AttachStorageRequest) (fabric.RuntimeHandle, error) {
	f.calls = append(f.calls, "attach_storage")
	f.attachStorage = request
	if f.attachStorageErr != nil {
		return fabric.RuntimeHandle{}, f.attachStorageErr
	}
	return fabric.RuntimeHandle{ProviderResourceID: request.ComputeID + ":" + request.StorageID, Status: "attached"}, nil
}

func (f *recordingFabric) CreateWorkspaceRoute(ctx context.Context, request fabric.CreateRouteRequest) (fabric.RuntimeHandle, error) {
	f.calls = append(f.calls, "create_route")
	f.createRoute = request
	if f.createRouteErr != nil {
		return f.route, f.createRouteErr
	}
	return f.route, nil
}

func (f *recordingFabric) DestroyCompute(ctx context.Context, request fabric.DestroyComputeRequest) error {
	f.calls = append(f.calls, "destroy_compute")
	f.destroyCompute = request
	return nil
}

func (f *recordingFabric) DestroyStorage(ctx context.Context, request fabric.DestroyStorageRequest) error {
	f.calls = append(f.calls, "destroy_storage")
	f.destroyStorage = request
	return nil
}

func (f *recordingFabric) DestroyWorkspaceRoute(ctx context.Context, request fabric.DestroyWorkspaceRouteRequest) error {
	f.calls = append(f.calls, "destroy_route")
	f.destroyRoute = request
	return nil
}

func (f *recordingFabric) ResetWorkspaceToken(ctx context.Context, request fabric.ResetWorkspaceTokenRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{}, nil
}

func (f *recordingFabric) DeleteWorkspaceToken(ctx context.Context, request fabric.DeleteWorkspaceTokenRequest) error {
	return nil
}

func assertPackage(t *testing.T, plan fabric.PackagePlan) {
	t.Helper()
	if plan.ID != "basic" {
		t.Fatalf("package id = %q", plan.ID)
	}
}

func assertCalls(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("calls = %v, want %v", got, want)
	}
}
