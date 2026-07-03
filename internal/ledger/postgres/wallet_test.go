package postgres

import (
	"testing"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

func TestStoreSatisfiesLedgerPort(t *testing.T) {
	var _ ledger.Port = (*Store)(nil)
}
