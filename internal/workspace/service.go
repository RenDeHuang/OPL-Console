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

func (s *Service) Handoff(ctx context.Context, request HandoffRequest) (HandoffResult, error) {
	if s.repo == nil {
		return HandoffResult{}, fmt.Errorf("workspace repository not configured")
	}
	return s.repo.Handoff(ctx, request)
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
			approvalID, err := s.repo.CreateApproval(ctx, ApprovalRequest{
				OrganizationID:  createContext.OrganizationID,
				RequesterUserID: createContext.ActorUserID,
				Action:          "workspace.create",
				ObjectType:      "workspace",
				ObjectID:        request.WorkspaceID,
				Reason:          "managed workspace policy requires approval",
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
	computeID := "cmp-" + request.WorkspaceID
	storageID := "stg-" + request.WorkspaceID
	attachmentID := "att-" + request.WorkspaceID
	holdID := ""

	if s.ledger != nil && createContext.HoldAmountFen > 0 {
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
			AmountFen:        createContext.HoldAmountFen,
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
	return CreateWorkspaceResult{WorkspaceID: request.WorkspaceID, URL: route.URL, State: "ready"}, nil
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
