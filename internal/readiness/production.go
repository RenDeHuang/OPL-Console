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
		"database_url":            strings.TrimSpace(cfg.DatabaseURL) != "",
		"public_url":              strings.HasPrefix(cfg.PublicURL, "https://") && !strings.Contains(cfg.PublicURL, "127.0.0.1"),
		"workspace_domain":        strings.Contains(cfg.WorkspaceDomain, ".") && !strings.Contains(cfg.WorkspaceDomain, "localhost"),
		"kube_config":             strings.TrimSpace(cfg.KubeconfigPath) != "" || inClusterServiceAccount(),
		"kube_namespace":          strings.TrimSpace(cfg.KubeNamespace) != "",
		"ingress_class":           cfg.IngressClass == "qcloud" || cfg.IngressClass == "nginx-production",
		"workspace_image":         productionImage(cfg.WorkspaceImage),
		"workspace_storage_class": strings.TrimSpace(cfg.WorkspaceStorageClass) != "",
		"fabric_provider":         cfg.FabricProvider == "tke",
		"auth_seed":               productionAuthSeed(cfg.ConsoleUsersJSON),
	}
	ready := true
	for _, ok := range checks {
		ready = ready && ok
	}
	return Report{Ready: ready, Checks: checks}
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
