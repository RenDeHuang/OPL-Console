package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetWallet(ctx context.Context, billingAccountID string) (ledger.Wallet, error) {
	var wallet ledger.Wallet
	err := s.pool.QueryRow(ctx, `
		SELECT id, balance_fen, frozen_fen, balance_fen - frozen_fen
		FROM billing_accounts
		WHERE id = $1
	`, billingAccountID).Scan(&wallet.BillingAccountID, &wallet.BalanceFen, &wallet.FrozenFen, &wallet.AvailableFen)
	return wallet, err
}

func (s *Store) QuoteWorkspace(ctx context.Context, request ledger.WorkspaceQuoteRequest) (ledger.WorkspaceQuote, error) {
	var quote ledger.WorkspaceQuote
	quote.BillingAccountID = request.BillingAccountID
	quote.PackageID = request.PackageID
	quote.Currency = "CNY"
	quote.HoldDays = 7
	quote.Source = "ledger_postgres"
	err := s.pool.QueryRow(ctx, `
		SELECT p.storage_gb, p.compute_hourly_fen, p.storage_gb_month_fen,
		       b.balance_fen, b.frozen_fen, b.balance_fen - b.frozen_fen
		FROM workspace_packages p
		CROSS JOIN billing_accounts b
		WHERE p.id = $1 AND p.available = true AND b.id = $2
	`, request.PackageID, request.BillingAccountID).Scan(
		&quote.StorageGB,
		&quote.ComputeHourlyFen,
		&quote.StorageGBMonthFen,
		&quote.BalanceFen,
		&quote.FrozenFen,
		&quote.AvailableFen,
	)
	if err != nil {
		return ledger.WorkspaceQuote{}, err
	}
	quote.ComputeHoldFen = quote.ComputeHourlyFen * 24 * int64(quote.HoldDays)
	quote.StorageHoldFen = quote.StorageGBMonthFen * int64(quote.StorageGB)
	quote.TotalHoldFen = quote.ComputeHoldFen + quote.StorageHoldFen
	quote.SufficientBalance = quote.AvailableFen >= quote.TotalHoldFen
	return quote, nil
}

func (s *Store) FreezeHold(ctx context.Context, request ledger.HoldRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var available int64
	if err := tx.QueryRow(ctx, `
		SELECT balance_fen - frozen_fen FROM billing_accounts WHERE id = $1 FOR UPDATE
	`, request.BillingAccountID).Scan(&available); err != nil {
		return err
	}
	if available < request.AmountFen {
		return ledger.ErrInsufficientBalance
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_holds (id, billing_account_id, resource_type, resource_id, amount_fen, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
	`, request.HoldID, request.BillingAccountID, request.ResourceType, request.ResourceID, request.AmountFen); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing_accounts SET frozen_fen = frozen_fen + $1, updated_at = now() WHERE id = $2
	`, request.AmountFen, request.BillingAccountID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ReleaseHold(ctx context.Context, holdID string, actorUserID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var billingAccountID string
	var amountFen int64
	if err := tx.QueryRow(ctx, `
		UPDATE wallet_holds SET status = 'released', updated_at = now()
		WHERE id = $1 AND status = 'active'
		RETURNING billing_account_id, amount_fen
	`, holdID).Scan(&billingAccountID, &amountFen); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.ErrHoldNotActive
		}
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing_accounts
		SET frozen_fen = frozen_fen - $1, updated_at = now()
		WHERE id = $2
	`, amountFen, billingAccountID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) DebitHold(ctx context.Context, holdID string, actorUserID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var billingAccountID string
	var amountFen int64
	if err := tx.QueryRow(ctx, `
		UPDATE wallet_holds SET status = 'debited', updated_at = now()
		WHERE id = $1 AND status = 'active'
		RETURNING billing_account_id, amount_fen
	`, holdID).Scan(&billingAccountID, &amountFen); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.ErrHoldNotActive
		}
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing_accounts
		SET frozen_fen = frozen_fen - $1,
		    balance_fen = balance_fen - $1,
		    updated_at = now()
		WHERE id = $2
	`, amountFen, billingAccountID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) RecordManualTopUp(ctx context.Context, request ledger.TopUpRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		INSERT INTO manual_topups (id, billing_account_id, amount_fen, actor_user_id, note)
		VALUES ($1, $2, $3, $4, $5)
	`, request.TopUpID, request.BillingAccountID, request.AmountFen, request.ActorUserID, request.Note); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing_accounts SET balance_fen = balance_fen + $1, updated_at = now() WHERE id = $2
	`, request.AmountFen, request.BillingAccountID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) BillingLedger(ctx context.Context, billingAccountID string) ([]ledger.LedgerEntry, error) {
	return s.billingLedger(ctx, `
		SELECT id, billing_account_id, COALESCE(workspace_id, ''), resource_type, COALESCE(resource_id, ''),
		       amount_fen, kind, description, created_at
		FROM billing_ledger_entries
		WHERE billing_account_id = $1
		ORDER BY created_at DESC
	`, billingAccountID)
}

func (s *Store) AdminBillingLedger(ctx context.Context) ([]ledger.LedgerEntry, error) {
	return s.billingLedger(ctx, `
		SELECT id, billing_account_id, COALESCE(workspace_id, ''), resource_type, COALESCE(resource_id, ''),
		       amount_fen, kind, description, created_at
		FROM billing_ledger_entries
		ORDER BY created_at DESC
		LIMIT 200
	`)
}

func (s *Store) billingLedger(ctx context.Context, query string, args ...any) ([]ledger.LedgerEntry, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []ledger.LedgerEntry
	for rows.Next() {
		var entry ledger.LedgerEntry
		var createdAt time.Time
		if err := rows.Scan(
			&entry.ID,
			&entry.BillingAccountID,
			&entry.WorkspaceID,
			&entry.ResourceType,
			&entry.ResourceID,
			&entry.AmountFen,
			&entry.Kind,
			&entry.Description,
			&createdAt,
		); err != nil {
			return nil, err
		}
		entry.CreatedAt = createdAt.Format(time.RFC3339)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *Store) ReconciliationStatus(ctx context.Context) (ledger.ReconciliationStatus, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM billing_ledger_entries`).Scan(&count); err != nil {
		return ledger.ReconciliationStatus{}, err
	}
	return ledger.ReconciliationStatus{
		Ready:  true,
		Status: "ready",
		Reason: "ledger_entries_available",
	}, nil
}

func (s *Store) RecordAuditEvent(ctx context.Context, event ledger.AuditEvent) error {
	if len(event.Metadata) == 0 {
		event.Metadata = []byte(`{}`)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_events (id, actor_user_id, action, object_type, object_id, request_id, result, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.ID, event.ActorUserID, event.Action, event.ObjectType, event.ObjectID, event.RequestID, event.Result, event.Metadata)
	return err
}

func (s *Store) RecordReceipt(ctx context.Context, receipt ledger.Receipt) error {
	if len(receipt.Payload) == 0 {
		receipt.Payload = []byte(`{}`)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO receipts (id, receipt_type, subject_type, subject_id, operation_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, receipt.ID, receipt.ReceiptType, receipt.SubjectType, receipt.SubjectID, receipt.OperationID, receipt.Payload)
	return err
}

func (s *Store) Receipts(ctx context.Context, subjectID string) ([]ledger.Receipt, error) {
	query := `
		SELECT id, receipt_type, subject_type, subject_id, COALESCE(operation_id, ''), payload
		FROM receipts
	`
	args := []any{}
	if subjectID != "" {
		query += ` WHERE subject_id = $1`
		args = append(args, subjectID)
	}
	query += ` ORDER BY created_at DESC LIMIT 200`
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var receipts []ledger.Receipt
	for rows.Next() {
		var receipt ledger.Receipt
		if err := rows.Scan(
			&receipt.ID,
			&receipt.ReceiptType,
			&receipt.SubjectType,
			&receipt.SubjectID,
			&receipt.OperationID,
			&receipt.Payload,
		); err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, rows.Err()
}
