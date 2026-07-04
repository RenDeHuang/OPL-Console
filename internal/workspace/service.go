package workspace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/RenDeHuang/opl-console/internal/fabric"
	"github.com/RenDeHuang/opl-console/internal/ledger"
)

type Service struct {
	fabric fabric.Port
	repo   Repository
	ledger ledger.Port
}

type Option func(*Service)

func WithRepository(repo Repository) Option {
	return func(s *Service) {
		s.repo = repo
	}
}

func WithLedger(ledgerPort ledger.Port) Option {
	return func(s *Service) {
		s.ledger = ledgerPort
	}
}

func NewService(fabricPort fabric.Port, options ...Option) *Service {
	service := &Service{fabric: fabricPort}
	for _, option := range options {
		option(service)
	}
	return service
}

type Repository interface {
	PrepareCreate(ctx context.Context, request CreateWorkspaceRequest) (CreateContext, error)
	CreateApproval(ctx context.Context, request ApprovalRequest) (string, error)
	SaveCreated(ctx context.Context, record CreatedWorkspace) error
	Handoff(ctx context.Context, request HandoffRequest) (HandoffResult, error)
	RuntimeForAction(ctx context.Context, request ActionRequest) (RuntimeRecord, error)
	UpdateWorkspaceState(ctx context.Context, change StateChange) error
	UpdateResourceState(ctx context.Context, change ResourceStateChange) error
	UpsertLifecycleStep(ctx context.Context, step LifecycleStepChange) error
	SaveBackup(ctx context.Context, backup BackupRecord) error
	ReplaceActiveToken(ctx context.Context, change TokenChange) error
	DeleteActiveToken(ctx context.Context, request ActionRequest) error
}

type CreateContext struct {
	ActorUserID      string
	OrganizationID   string
	HoldAmountFen    int64
	RequiresApproval bool
	Package          fabric.PackagePlan
}

type ApprovalRequest struct {
	OrganizationID  string
	PolicyID        string
	RequesterUserID string
	Action          string
	ObjectType      string
	ObjectID        string
	Reason          string
	Context         json.RawMessage
}

type CreatedWorkspace struct {
	WorkspaceID       string
	Name              string
	BillingAccountID  string
	OrganizationID    string
	PackageID         string
	ComputeID         string
	StorageID         string
	AttachmentID      string
	ComputeProviderID string
	StorageProviderID string
	AttachProviderID  string
	RouteProviderID   string
	RouteURL          string
	Token             string
	LifecycleSteps    []LifecycleStepChange
}

type CreateWorkspaceRequest struct {
	WorkspaceID      string `json:"workspaceId"`
	Name             string `json:"name"`
	BillingAccountID string `json:"billingAccountId"`
	PackageID        string `json:"packageId"`
	Token            string `json:"token"`
	ActorUserID      string `json:"-"`
}

type CreateWorkspaceResult struct {
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url,omitempty"`
	State       string `json:"state,omitempty"`
	ApprovalID  string `json:"approvalId,omitempty"`
}

type HandoffRequest struct {
	WorkspaceID string
	Token       string
}

type HandoffResult struct {
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url"`
	State       string `json:"state"`
}

type ActionRequest struct {
	WorkspaceID string `json:"workspaceId"`
	ActorUserID string `json:"-"`
	Confirm     bool   `json:"confirm,omitempty"`
}

type TokenRequest struct {
	WorkspaceID string `json:"workspaceId"`
	ActorUserID string `json:"-"`
	Token       string `json:"token"`
}

type ActionResult struct {
	WorkspaceID string `json:"workspaceId"`
	State       string `json:"state"`
	URL         string `json:"url,omitempty"`
}

type RuntimeRecord struct {
	WorkspaceID      string
	ActorUserID      string
	BillingAccountID string
	ComputeID        string
	StorageID        string
	AttachmentID     string
	State            string
}

type StateChange struct {
	WorkspaceID string
	ActorUserID string
	State       string
}

type ResourceStateChange struct {
	WorkspaceID  string
	ResourceType string
	ResourceID   string
	Status       string
	Metadata     map[string]string
}

type LifecycleStepChange struct {
	WorkspaceID        string
	StepName           string
	DesiredState       string
	ActualState        string
	ProviderResourceID string
	ErrorCode          string
}

type BackupRecord struct {
	BackupID           string
	WorkspaceID        string
	StorageID          string
	ProviderResourceID string
	Status             string
	ActorUserID        string
}

type TokenChange struct {
	WorkspaceID string
	ActorUserID string
	Token       string
	URL         string
}

func (s *Service) Handoff(ctx context.Context, request HandoffRequest) (HandoffResult, error) {
	if s.repo == nil {
		return HandoffResult{}, fmt.Errorf("workspace repository not configured")
	}
	return s.repo.Handoff(ctx, request)
}

func (s *Service) StopCompute(ctx context.Context, request ActionRequest) (ActionResult, error) {
	runtime, err := s.runtimeForAction(ctx, request)
	if err != nil {
		return ActionResult{}, err
	}
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "stopping_server"}); err != nil {
		return ActionResult{}, fmt.Errorf("mark stopping server: %w", err)
	}
	handle, err := s.fabric.StopCompute(ctx, fabric.StopComputeRequest{ComputeID: runtime.ComputeID})
	if err != nil {
		return ActionResult{}, fmt.Errorf("stop compute: %w", err)
	}
	if err := s.repo.UpdateResourceState(ctx, ResourceStateChange{WorkspaceID: request.WorkspaceID, ResourceType: "compute", ResourceID: runtime.ComputeID, Status: handle.Status}); err != nil {
		return ActionResult{}, err
	}
	state := "stopped_server_disk_retained"
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: state}); err != nil {
		return ActionResult{}, fmt.Errorf("stop workspace state: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.compute.stop", request.WorkspaceID, "succeeded", map[string]string{"state": state}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: state}, nil
}

func (s *Service) RestartCompute(ctx context.Context, request ActionRequest) (ActionResult, error) {
	runtime, err := s.runtimeForAction(ctx, request)
	if err != nil {
		return ActionResult{}, err
	}
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "restarting_server"}); err != nil {
		return ActionResult{}, fmt.Errorf("mark restarting server: %w", err)
	}
	handle, err := s.fabric.RestartCompute(ctx, fabric.RestartComputeRequest{ComputeID: runtime.ComputeID})
	if err != nil {
		return ActionResult{}, fmt.Errorf("restart compute: %w", err)
	}
	if err := s.repo.UpdateResourceState(ctx, ResourceStateChange{WorkspaceID: request.WorkspaceID, ResourceType: "compute", ResourceID: runtime.ComputeID, Status: handle.Status}); err != nil {
		return ActionResult{}, err
	}
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "running"}); err != nil {
		return ActionResult{}, fmt.Errorf("restart workspace state: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.compute.restart", request.WorkspaceID, "succeeded", map[string]string{"state": "running"}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: "running"}, nil
}

func (s *Service) DestroyCompute(ctx context.Context, request ActionRequest) (ActionResult, error) {
	runtime, err := s.runtimeForAction(ctx, request)
	if err != nil {
		return ActionResult{}, err
	}
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "destroying_server"}); err != nil {
		return ActionResult{}, fmt.Errorf("mark destroying server: %w", err)
	}
	if runtime.AttachmentID != "" {
		if _, err := s.fabric.DetachStorage(ctx, fabric.DetachStorageRequest{AttachmentID: runtime.AttachmentID, ComputeID: runtime.ComputeID, StorageID: runtime.StorageID}); err != nil {
			return ActionResult{}, fmt.Errorf("detach storage before compute destroy: %w", err)
		}
		_ = s.repo.UpdateResourceState(ctx, ResourceStateChange{WorkspaceID: request.WorkspaceID, ResourceType: "attachment", ResourceID: runtime.AttachmentID, Status: "detached_retained"})
	}
	if err := s.fabric.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: runtime.ComputeID}); err != nil {
		return ActionResult{}, fmt.Errorf("destroy compute: %w", err)
	}
	_ = s.repo.UpdateResourceState(ctx, ResourceStateChange{WorkspaceID: request.WorkspaceID, ResourceType: "compute", ResourceID: runtime.ComputeID, Status: "destroyed"})
	state := "server_destroyed_disk_retained"
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: state}); err != nil {
		return ActionResult{}, fmt.Errorf("destroy compute state: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.compute.destroy", request.WorkspaceID, "succeeded", map[string]string{"state": state}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: state}, nil
}

func (s *Service) DestroyStorage(ctx context.Context, request ActionRequest) (ActionResult, error) {
	if !request.Confirm {
		return ActionResult{}, fmt.Errorf("storage_destroy_confirmation_required")
	}
	runtime, err := s.runtimeForAction(ctx, request)
	if err != nil {
		return ActionResult{}, err
	}
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "destroying_disk"}); err != nil {
		return ActionResult{}, fmt.Errorf("mark destroying disk: %w", err)
	}
	if err := s.fabric.DestroyStorage(ctx, fabric.DestroyStorageRequest{StorageID: runtime.StorageID}); err != nil {
		return ActionResult{}, fmt.Errorf("destroy storage: %w", err)
	}
	_ = s.repo.UpdateResourceState(ctx, ResourceStateChange{WorkspaceID: request.WorkspaceID, ResourceType: "storage", ResourceID: runtime.StorageID, Status: "destroyed"})
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "destroyed"}); err != nil {
		return ActionResult{}, fmt.Errorf("destroy storage state: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.storage.destroy", request.WorkspaceID, "succeeded", map[string]string{"state": "destroyed"}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: "destroyed"}, nil
}

func (s *Service) CreateStorageBackup(ctx context.Context, request ActionRequest) (ActionResult, error) {
	runtime, err := s.runtimeForAction(ctx, request)
	if err != nil {
		return ActionResult{}, err
	}
	backupID, err := randomID("backup")
	if err != nil {
		return ActionResult{}, err
	}
	handle, err := s.fabric.CreateStorageBackup(ctx, fabric.BackupStorageRequest{BackupID: backupID, WorkspaceID: request.WorkspaceID, StorageID: runtime.StorageID})
	if err != nil {
		return ActionResult{}, fmt.Errorf("create storage backup: %w", err)
	}
	if err := s.repo.SaveBackup(ctx, BackupRecord{BackupID: backupID, WorkspaceID: request.WorkspaceID, StorageID: runtime.StorageID, ProviderResourceID: handle.ProviderResourceID, Status: handle.Status, ActorUserID: request.ActorUserID}); err != nil {
		return ActionResult{}, fmt.Errorf("save backup: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.storage.backup", request.WorkspaceID, "succeeded", map[string]string{"backupId": backupID}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: "creating_storage_backup"}, nil
}

func (s *Service) RestoreStorageBackup(ctx context.Context, request ActionRequest) (ActionResult, error) {
	if _, err := s.runtimeForAction(ctx, request); err != nil {
		return ActionResult{}, err
	}
	if _, err := s.fabric.RestoreStorageBackup(ctx, fabric.RestoreStorageRequest{WorkspaceID: request.WorkspaceID}); err != nil {
		return ActionResult{}, fmt.Errorf("restore storage backup: %w", err)
	}
	if err := s.repo.UpdateWorkspaceState(ctx, StateChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, State: "restoring_storage_backup"}); err != nil {
		return ActionResult{}, fmt.Errorf("restore backup state: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.storage.restore", request.WorkspaceID, "succeeded", map[string]string{"state": "restoring_storage_backup"}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: "restoring_storage_backup"}, nil
}

func (s *Service) ResetWorkspaceToken(ctx context.Context, request TokenRequest) (ActionResult, error) {
	if _, err := s.runtimeForAction(ctx, ActionRequest{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID}); err != nil {
		return ActionResult{}, err
	}
	route, err := s.fabric.ResetWorkspaceToken(ctx, fabric.ResetWorkspaceTokenRequest{WorkspaceID: request.WorkspaceID, Token: request.Token})
	if err != nil {
		return ActionResult{}, fmt.Errorf("reset workspace token: %w", err)
	}
	if err := s.repo.ReplaceActiveToken(ctx, TokenChange{WorkspaceID: request.WorkspaceID, ActorUserID: request.ActorUserID, Token: request.Token, URL: route.URL}); err != nil {
		return ActionResult{}, fmt.Errorf("replace workspace token: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.token.reset", request.WorkspaceID, "succeeded", map[string]string{"url": route.URL}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: "ready", URL: route.URL}, nil
}

func (s *Service) DeleteWorkspaceToken(ctx context.Context, request ActionRequest) (ActionResult, error) {
	if _, err := s.runtimeForAction(ctx, request); err != nil {
		return ActionResult{}, err
	}
	if err := s.fabric.DeleteWorkspaceToken(ctx, fabric.DeleteWorkspaceTokenRequest{WorkspaceID: request.WorkspaceID}); err != nil {
		return ActionResult{}, fmt.Errorf("delete workspace token: %w", err)
	}
	if err := s.repo.DeleteActiveToken(ctx, request); err != nil {
		return ActionResult{}, fmt.Errorf("delete active token: %w", err)
	}
	if err := s.recordAudit(ctx, request.ActorUserID, "workspace.token.delete", request.WorkspaceID, "succeeded", map[string]string{"state": "token_deleted"}); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{WorkspaceID: request.WorkspaceID, State: "token_deleted"}, nil
}

func (s *Service) runtimeForAction(ctx context.Context, request ActionRequest) (RuntimeRecord, error) {
	if s.repo == nil {
		return RuntimeRecord{}, fmt.Errorf("workspace repository not configured")
	}
	runtime, err := s.repo.RuntimeForAction(ctx, request)
	if err != nil {
		return RuntimeRecord{}, fmt.Errorf("load workspace runtime: %w", err)
	}
	return runtime, nil
}

func (s *Service) CreateWorkspace(ctx context.Context, request CreateWorkspaceRequest) (CreateWorkspaceResult, error) {
	createContext := CreateContext{
		ActorUserID: request.ActorUserID,
		Package:     fabric.PackagePlan{ID: request.PackageID, CPU: 2, MemoryGB: 4, StorageGB: 10},
	}
	if s.repo != nil {
		var err error
		createContext, err = s.repo.PrepareCreate(ctx, request)
		if err != nil {
			return CreateWorkspaceResult{}, fmt.Errorf("prepare create: %w", err)
		}
		if createContext.RequiresApproval {
			approvalContext, err := json.Marshal(map[string]any{
				"requesterUserId":     createContext.ActorUserID,
				"organizationId":      createContext.OrganizationID,
				"packageId":           request.PackageID,
				"estimatedHoldFen":    createContext.HoldAmountFen,
				"policyRuleTriggered": "workspace_lifecycle.requiresApproval",
				"postApprovalActions": []string{"freeze_hold", "create_storage", "create_compute", "attach_storage", "create_route"},
			})
			if err != nil {
				return CreateWorkspaceResult{}, err
			}
			approvalID, err := s.repo.CreateApproval(ctx, ApprovalRequest{
				OrganizationID:  createContext.OrganizationID,
				RequesterUserID: createContext.ActorUserID,
				Action:          "workspace.create",
				ObjectType:      "workspace",
				ObjectID:        request.WorkspaceID,
				Reason:          "managed workspace policy requires approval",
				Context:         approvalContext,
			})
			if err != nil {
				return CreateWorkspaceResult{}, fmt.Errorf("create approval: %w", err)
			}
			_ = s.recordAudit(ctx, createContext.ActorUserID, "workspace.create", request.WorkspaceID, "approval_required", map[string]string{"approvalId": approvalID})
			return CreateWorkspaceResult{WorkspaceID: request.WorkspaceID, State: "approval_required", ApprovalID: approvalID}, nil
		}
	}
	plan := createContext.Package
	if plan.ID == "" {
		plan = fabric.PackagePlan{ID: request.PackageID, CPU: 2, MemoryGB: 4, StorageGB: 10}
	}
	holdAmountFen := createContext.HoldAmountFen
	if s.ledger != nil {
		quote, err := s.ledger.QuoteWorkspace(ctx, ledger.WorkspaceQuoteRequest{
			BillingAccountID: request.BillingAccountID,
			PackageID:        request.PackageID,
		})
		if err != nil {
			return CreateWorkspaceResult{}, fmt.Errorf("quote workspace: %w", err)
		}
		holdAmountFen = quote.TotalHoldFen
	}
	computeID := "cmp-" + request.WorkspaceID
	storageID := "stg-" + request.WorkspaceID
	attachmentID := "att-" + request.WorkspaceID
	holdID := ""

	if s.ledger != nil && holdAmountFen > 0 {
		generatedHoldID, err := randomID("hold")
		if err != nil {
			return CreateWorkspaceResult{}, err
		}
		holdID = generatedHoldID
		if err := s.ledger.FreezeHold(ctx, ledger.HoldRequest{
			HoldID:           holdID,
			BillingAccountID: request.BillingAccountID,
			ResourceType:     "workspace",
			ResourceID:       request.WorkspaceID,
			AmountFen:        holdAmountFen,
			ActorUserID:      createContext.ActorUserID,
		}); err != nil {
			_ = s.recordAudit(ctx, createContext.ActorUserID, "workspace.create", request.WorkspaceID, "billing_failed", map[string]string{"error": err.Error()})
			return CreateWorkspaceResult{}, fmt.Errorf("freeze workspace hold: %w", err)
		}
	}

	storage, err := s.fabric.CreateStorage(ctx, fabric.CreateStorageRequest{
		StorageID: storageID, BillingAccountID: request.BillingAccountID, Package: plan,
	})
	if err != nil {
		s.releaseHold(ctx, holdID, createContext.ActorUserID)
		return CreateWorkspaceResult{}, fmt.Errorf("create storage: %w", err)
	}
	compute, err := s.fabric.CreateCompute(ctx, fabric.CreateComputeRequest{
		ComputeID: computeID, BillingAccountID: request.BillingAccountID, Package: plan,
	})
	if err != nil {
		s.releaseHold(ctx, holdID, createContext.ActorUserID)
		if s.repo != nil {
			_ = s.repo.UpsertLifecycleStep(ctx, LifecycleStepChange{WorkspaceID: request.WorkspaceID, StepName: "create_compute", DesiredState: "running", ActualState: "failed", ErrorCode: "create_compute_failed"})
		}
		return CreateWorkspaceResult{}, fmt.Errorf("create compute: %w", err)
	}
	attachment, err := s.fabric.AttachStorage(ctx, fabric.AttachStorageRequest{
		AttachmentID: attachmentID, ComputeID: computeID, StorageID: storageID, MountPath: "/data",
	})
	if err != nil {
		_ = s.fabric.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: computeID})
		s.releaseHold(ctx, holdID, createContext.ActorUserID)
		return CreateWorkspaceResult{}, fmt.Errorf("attach storage: %w", err)
	}
	route, err := s.fabric.CreateWorkspaceRoute(ctx, fabric.CreateRouteRequest{
		WorkspaceID: request.WorkspaceID, WorkspaceName: request.Name, ComputeID: computeID, Token: request.Token,
	})
	if err != nil {
		_ = s.fabric.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: computeID})
		s.releaseHold(ctx, holdID, createContext.ActorUserID)
		return CreateWorkspaceResult{}, fmt.Errorf("create workspace route: %w", err)
	}
	if s.repo != nil {
		if err := s.repo.SaveCreated(ctx, CreatedWorkspace{
			WorkspaceID:       request.WorkspaceID,
			Name:              request.Name,
			BillingAccountID:  request.BillingAccountID,
			OrganizationID:    createContext.OrganizationID,
			PackageID:         plan.ID,
			ComputeID:         computeID,
			StorageID:         storageID,
			AttachmentID:      attachmentID,
			ComputeProviderID: compute.ProviderResourceID,
			StorageProviderID: storage.ProviderResourceID,
			AttachProviderID:  attachment.ProviderResourceID,
			RouteProviderID:   route.ProviderResourceID,
			RouteURL:          route.URL,
			Token:             request.Token,
			LifecycleSteps: []LifecycleStepChange{
				{WorkspaceID: request.WorkspaceID, StepName: "create_storage", DesiredState: "available", ActualState: storage.Status, ProviderResourceID: storage.ProviderResourceID},
				{WorkspaceID: request.WorkspaceID, StepName: "create_compute", DesiredState: "running", ActualState: compute.Status, ProviderResourceID: compute.ProviderResourceID},
				{WorkspaceID: request.WorkspaceID, StepName: "attach_storage", DesiredState: "attached", ActualState: attachment.Status, ProviderResourceID: attachment.ProviderResourceID},
				{WorkspaceID: request.WorkspaceID, StepName: "create_route", DesiredState: "ready", ActualState: route.Status, ProviderResourceID: route.ProviderResourceID},
			},
		}); err != nil {
			return CreateWorkspaceResult{}, fmt.Errorf("save workspace facade state: %w", err)
		}
	}
	if err := s.recordReceipt(ctx, request.WorkspaceID, holdID, route.URL); err != nil {
		return CreateWorkspaceResult{}, err
	}
	if err := s.recordAudit(ctx, createContext.ActorUserID, "workspace.create", request.WorkspaceID, "succeeded", map[string]string{"url": route.URL}); err != nil {
		return CreateWorkspaceResult{}, err
	}
	return CreateWorkspaceResult{WorkspaceID: request.WorkspaceID, URL: route.URL, State: "running"}, nil
}

func (s *Service) releaseHold(ctx context.Context, holdID string, actorUserID string) {
	if s.ledger == nil || holdID == "" {
		return
	}
	_ = s.ledger.ReleaseHold(ctx, holdID, actorUserID)
}

func (s *Service) recordAudit(ctx context.Context, actorUserID string, action string, workspaceID string, result string, metadata map[string]string) error {
	if s.ledger == nil {
		return nil
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	id, err := randomID("audit")
	if err != nil {
		return err
	}
	return s.ledger.RecordAuditEvent(ctx, ledger.AuditEvent{
		ID:          id,
		ActorUserID: actorUserID,
		Action:      action,
		ObjectType:  "workspace",
		ObjectID:    workspaceID,
		RequestID:   id,
		Result:      result,
		Metadata:    payload,
	})
}

func (s *Service) recordReceipt(ctx context.Context, workspaceID string, holdID string, url string) error {
	if s.ledger == nil {
		return nil
	}
	id, err := randomID("receipt")
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{"holdId": holdID, "url": url})
	if err != nil {
		return err
	}
	return s.ledger.RecordReceipt(ctx, ledger.Receipt{
		ID:          id,
		ReceiptType: "workspace.governance.created",
		SubjectType: "workspace",
		SubjectID:   workspaceID,
		OperationID: id,
		Payload:     payload,
	})
}

func randomID(prefix string) (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(buf[:]), nil
}
