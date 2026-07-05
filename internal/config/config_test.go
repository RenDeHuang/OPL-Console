package config

import "testing"

func TestLoadFabricProviderDefaultsLocalAndReadsEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load default: %v", err)
	}
	if cfg.FabricProvider != "local" {
		t.Fatalf("FabricProvider default = %q, want local", cfg.FabricProvider)
	}

	t.Setenv("OPL_FABRIC_PROVIDER", "tke")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load tke: %v", err)
	}
	if cfg.FabricProvider != "tke" {
		t.Fatalf("FabricProvider = %q, want tke", cfg.FabricProvider)
	}
}

func TestLoadReadsInternalUpstreamURLsAndTokens(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("OPL_FABRIC_INTERNAL_URL", "http://opl-fabric-api:8787")
	t.Setenv("OPL_LEDGER_INTERNAL_URL", "http://opl-ledger-api:8788")
	t.Setenv("OPL_OPERATOR_TOKEN", "fabric-token")
	t.Setenv("OPL_LEDGER_SERVICE_TOKEN", "ledger-service-token")
	t.Setenv("OPL_LEDGER_ADMIN_TOKEN", "ledger-admin-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.FabricInternalURL != "http://opl-fabric-api:8787" {
		t.Fatalf("FabricInternalURL = %q", cfg.FabricInternalURL)
	}
	if cfg.LedgerInternalURL != "http://opl-ledger-api:8788" {
		t.Fatalf("LedgerInternalURL = %q", cfg.LedgerInternalURL)
	}
	if cfg.OperatorToken != "fabric-token" {
		t.Fatalf("OperatorToken = %q", cfg.OperatorToken)
	}
	if cfg.LedgerServiceToken != "ledger-service-token" {
		t.Fatalf("LedgerServiceToken = %q", cfg.LedgerServiceToken)
	}
	if cfg.LedgerAdminToken != "ledger-admin-token" {
		t.Fatalf("LedgerAdminToken = %q", cfg.LedgerAdminToken)
	}
}
