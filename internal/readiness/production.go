package readiness

import (
	"os"
	"strings"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/config"
)

type Report struct {
	Ready  bool
	Checks map[string]bool
}

func Production(cfg config.Config) Report {
	checks := map[string]bool{
		"database.postgres_url":      strings.TrimSpace(cfg.DatabaseURL) != "",
		"auth.seed_without_defaults": productionAuthSeed(cfg.ConsoleUsersJSON),
		"console.public_https_url":   strings.HasPrefix(cfg.PublicURL, "https://") && !strings.Contains(cfg.PublicURL, "127.0.0.1"),
		"workspace.domain":           strings.Contains(cfg.WorkspaceDomain, ".") && !strings.Contains(cfg.WorkspaceDomain, "localhost"),
		"kubernetes.config":          strings.TrimSpace(cfg.KubeconfigPath) != "" || inClusterServiceAccount(),
		"kubernetes.namespace":       strings.TrimSpace(cfg.KubeNamespace) != "",
		"kubernetes.ingress_class":   cfg.IngressClass == "qcloud" || cfg.IngressClass == "nginx-production",
		"kubernetes.storage_class":   strings.TrimSpace(cfg.WorkspaceStorageClass) != "",
		"registry.workspace_image":   productionImage(cfg.WorkspaceImage),
		"fabric.provider":            cfg.FabricProvider == "tke" || externalDependency(cfg.FabricURL, cfg.FabricToken),
		"fabric.external_contract":   cfg.FabricProvider != "http" || externalDependency(cfg.FabricURL, cfg.FabricToken),
		"ledger.external_contract":   cfg.LedgerURL == "" || externalDependency(cfg.LedgerURL, cfg.LedgerToken),
		"secrets.fabric_token":       cfg.FabricURL == "" || strings.TrimSpace(cfg.FabricToken) != "",
		"secrets.ledger_token":       cfg.LedgerURL == "" || strings.TrimSpace(cfg.LedgerToken) != "",
	}
	ready := true
	for _, ok := range checks {
		ready = ready && ok
	}
	return Report{Ready: ready, Checks: checks}
}

func externalDependency(url string, token string) bool {
	return strings.HasPrefix(strings.TrimSpace(url), "https://") && strings.TrimSpace(token) != ""
}

func inClusterServiceAccount() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return err == nil
}

func productionImage(image string) bool {
	image = strings.TrimSpace(image)
	return image != "" &&
		!strings.HasSuffix(image, ":latest") &&
		!strings.Contains(image, "ghcr.io/gaofeng21cn/one-person-lab-app:latest")
}

func productionAuthSeed(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return false
	}
	if strings.Contains(raw, auth.DefaultAdminPassword) || strings.Contains(raw, auth.DefaultAdminEmail) {
		return false
	}
	return true
}
