package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

func TestStoreSatisfiesLedgerPort(t *testing.T) {
	var _ ledger.Port = (*Store)(nil)
}

func TestRecordManualTopUpChangesBalance(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)
	accountID := createBillingAccount(ctx, t, pool, 100)

	err := store.RecordManualTopUp(ctx, ledger.TopUpRequest{
		TopUpID:          testID(t, "topup"),
		BillingAccountID: accountID,
		AmountFen:        250,
		ActorUserID:      testID(t, "actor"),
		Note:             "test top up",
	})
	if err != nil {
		t.Fatalf("RecordManualTopUp returned error: %v", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 350 || wallet.FrozenFen != 0 || wallet.AvailableFen != 350 {
		t.Fatalf("wallet = %+v, want balance 350 frozen 0 available 350", wallet)
	}
}

func TestFreezeHoldIncreasesFrozenAndReducesAvailable(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)
	accountID := createBillingAccount(ctx, t, pool, 1000)
	holdID := testID(t, "hold")

	err := store.FreezeHold(ctx, ledger.HoldRequest{
		HoldID:           holdID,
		BillingAccountID: accountID,
		ResourceType:     "workspace",
		ResourceID:       testID(t, "workspace"),
		AmountFen:        300,
		ActorUserID:      testID(t, "actor"),
	})
	if err != nil {
		t.Fatalf("FreezeHold returned error: %v", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 1000 || wallet.FrozenFen != 300 || wallet.AvailableFen != 700 {
		t.Fatalf("wallet = %+v, want balance 1000 frozen 300 available 700", wallet)
	}
	assertHoldStatus(ctx, t, pool, holdID, "active")
}

func TestFreezeHoldInsufficientBalanceReturnsError(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)
	accountID := createBillingAccount(ctx, t, pool, 100)

	err := store.FreezeHold(ctx, ledger.HoldRequest{
		HoldID:           testID(t, "hold"),
		BillingAccountID: accountID,
		ResourceType:     "workspace",
		ResourceID:       testID(t, "workspace"),
		AmountFen:        101,
		ActorUserID:      testID(t, "actor"),
	})
	if !errors.Is(err, ledger.ErrInsufficientBalance) {
		t.Fatalf("FreezeHold error = %v, want ErrInsufficientBalance", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 100 || wallet.FrozenFen != 0 || wallet.AvailableFen != 100 {
		t.Fatalf("wallet = %+v, want balance 100 frozen 0 available 100", wallet)
	}
}

func TestReleaseHoldReducesFrozenAndMarksHoldReleased(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)
	accountID := createBillingAccount(ctx, t, pool, 1000)
	holdID := freezeHold(ctx, t, store, accountID, 300)

	err := store.ReleaseHold(ctx, holdID, testID(t, "actor"))
	if err != nil {
		t.Fatalf("ReleaseHold returned error: %v", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 1000 || wallet.FrozenFen != 0 || wallet.AvailableFen != 1000 {
		t.Fatalf("wallet = %+v, want balance 1000 frozen 0 available 1000", wallet)
	}
	assertHoldStatus(ctx, t, pool, holdID, "released")
}

func TestDebitHoldReducesFrozenAndBalanceAndMarksHoldDebited(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)
	accountID := createBillingAccount(ctx, t, pool, 1000)
	holdID := freezeHold(ctx, t, store, accountID, 300)

	err := store.DebitHold(ctx, holdID, testID(t, "actor"))
	if err != nil {
		t.Fatalf("DebitHold returned error: %v", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 700 || wallet.FrozenFen != 0 || wallet.AvailableFen != 700 {
		t.Fatalf("wallet = %+v, want balance 700 frozen 0 available 700", wallet)
	}
	assertHoldStatus(ctx, t, pool, holdID, "debited")
}

func TestReleaseHoldMissingOrInactiveReturnsErrHoldNotActive(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)

	if err := store.ReleaseHold(ctx, testID(t, "missing"), testID(t, "actor")); !errors.Is(err, ledger.ErrHoldNotActive) {
		t.Fatalf("ReleaseHold missing error = %v, want ErrHoldNotActive", err)
	}

	accountID := createBillingAccount(ctx, t, pool, 1000)
	holdID := freezeHold(ctx, t, store, accountID, 300)
	if err := store.ReleaseHold(ctx, holdID, testID(t, "actor")); err != nil {
		t.Fatalf("ReleaseHold active returned error: %v", err)
	}
	if err := store.ReleaseHold(ctx, holdID, testID(t, "actor")); !errors.Is(err, ledger.ErrHoldNotActive) {
		t.Fatalf("ReleaseHold inactive error = %v, want ErrHoldNotActive", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 1000 || wallet.FrozenFen != 0 || wallet.AvailableFen != 1000 {
		t.Fatalf("wallet = %+v, want balance 1000 frozen 0 available 1000", wallet)
	}
	assertHoldStatus(ctx, t, pool, holdID, "released")
}

func TestDebitHoldMissingOrInactiveReturnsErrHoldNotActive(t *testing.T) {
	ctx := context.Background()
	pool := testPool(ctx, t)
	store := New(pool)

	if err := store.DebitHold(ctx, testID(t, "missing"), testID(t, "actor")); !errors.Is(err, ledger.ErrHoldNotActive) {
		t.Fatalf("DebitHold missing error = %v, want ErrHoldNotActive", err)
	}

	accountID := createBillingAccount(ctx, t, pool, 1000)
	holdID := freezeHold(ctx, t, store, accountID, 300)
	if err := store.DebitHold(ctx, holdID, testID(t, "actor")); err != nil {
		t.Fatalf("DebitHold active returned error: %v", err)
	}
	if err := store.DebitHold(ctx, holdID, testID(t, "actor")); !errors.Is(err, ledger.ErrHoldNotActive) {
		t.Fatalf("DebitHold inactive error = %v, want ErrHoldNotActive", err)
	}

	wallet := getWallet(ctx, t, store, accountID)
	if wallet.BalanceFen != 700 || wallet.FrozenFen != 0 || wallet.AvailableFen != 700 {
		t.Fatalf("wallet = %+v, want balance 700 frozen 0 available 700", wallet)
	}
	assertHoldStatus(ctx, t, pool, holdID, "debited")
}

func testPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func createBillingAccount(ctx context.Context, t *testing.T, pool *pgxpool.Pool, balanceFen int64) string {
	t.Helper()
	accountID := testID(t, "billing")
	_, err := pool.Exec(ctx, `
		INSERT INTO billing_accounts (id, owner_type, owner_id, balance_fen, frozen_fen, status)
		VALUES ($1, 'user', $2, $3, 0, 'active')
	`, accountID, testID(t, "owner"), balanceFen)
	if err != nil {
		t.Fatalf("create billing account: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		if _, err := pool.Exec(cleanupCtx, `DELETE FROM manual_topups WHERE billing_account_id = $1`, accountID); err != nil {
			t.Errorf("cleanup manual_topups: %v", err)
		}
		if _, err := pool.Exec(cleanupCtx, `DELETE FROM wallet_holds WHERE billing_account_id = $1`, accountID); err != nil {
			t.Errorf("cleanup wallet_holds: %v", err)
		}
		if _, err := pool.Exec(cleanupCtx, `DELETE FROM billing_accounts WHERE id = $1`, accountID); err != nil {
			t.Errorf("cleanup billing_accounts: %v", err)
		}
	})
	return accountID
}

func freezeHold(ctx context.Context, t *testing.T, store *Store, accountID string, amountFen int64) string {
	t.Helper()
	holdID := testID(t, "hold")
	err := store.FreezeHold(ctx, ledger.HoldRequest{
		HoldID:           holdID,
		BillingAccountID: accountID,
		ResourceType:     "workspace",
		ResourceID:       testID(t, "workspace"),
		AmountFen:        amountFen,
		ActorUserID:      testID(t, "actor"),
	})
	if err != nil {
		t.Fatalf("FreezeHold returned error: %v", err)
	}
	return holdID
}

func getWallet(ctx context.Context, t *testing.T, store *Store, accountID string) ledger.Wallet {
	t.Helper()
	wallet, err := store.GetWallet(ctx, accountID)
	if err != nil {
		t.Fatalf("GetWallet returned error: %v", err)
	}
	return wallet
}

func assertHoldStatus(ctx context.Context, t *testing.T, pool *pgxpool.Pool, holdID string, want string) {
	t.Helper()
	var got string
	if err := pool.QueryRow(ctx, `SELECT status FROM wallet_holds WHERE id = $1`, holdID).Scan(&got); err != nil {
		t.Fatalf("query hold status: %v", err)
	}
	if got != want {
		t.Fatalf("hold status = %q, want %q", got, want)
	}
}

func testID(t *testing.T, prefix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-", "_", "-").Replace(strings.ToLower(t.Name()))
	return fmt.Sprintf("%s-%s-%d", prefix, name, time.Now().UnixNano())
}
