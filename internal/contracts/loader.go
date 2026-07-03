package contracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Contract struct {
	Path          string `json:"-"`
	SchemaVersion int    `json:"schemaVersion"`
	Owner         string `json:"owner"`
	Purpose       string `json:"purpose"`
	Lifecycle     string `json:"lifecycle"`
}

func LoadDirectory(dir string) ([]Contract, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	contracts := make([]Contract, 0, len(matches))
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var contract Contract
		if err := json.Unmarshal(raw, &contract); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		contract.Path = path
		if contract.SchemaVersion == 0 || contract.Owner == "" || contract.Lifecycle == "" {
			return nil, fmt.Errorf("%s: missing required contract metadata", path)
		}
		contracts = append(contracts, contract)
	}
	return contracts, nil
}
