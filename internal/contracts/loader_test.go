package contracts

import "testing"

func TestLoadDirectoryRequiresSchemaVersionOwnerAndLifecycle(t *testing.T) {
	loaded, err := LoadDirectory("../../contracts")
	if err != nil {
		t.Fatalf("load contracts: %v", err)
	}
	if len(loaded) != 4 {
		t.Fatalf("loaded %d contracts", len(loaded))
	}
	for _, contract := range loaded {
		if contract.SchemaVersion != 1 {
			t.Fatalf("%s schemaVersion = %d", contract.Path, contract.SchemaVersion)
		}
		if contract.Owner == "" {
			t.Fatalf("%s owner is empty", contract.Path)
		}
		if contract.Lifecycle != "current" {
			t.Fatalf("%s lifecycle = %q", contract.Path, contract.Lifecycle)
		}
	}
}
