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
