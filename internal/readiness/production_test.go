package readiness

import (
	"testing"

	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/config"
)

func TestProductionReadinessPassesWithRequiredInputs(t *testing.T) {
	report := Production(config.Config{
		DatabaseURL:           "postgres://opl:secret@db.example.com:5432/opl_console",
		PublicURL:             "https://cloud.medopl.cn",
		WorkspaceDomain:       "workspace.medopl.cn",
		KubeconfigPath:        "/var/run/opl-console/kubeconfig",
		KubeNamespace:         "opl-cloud",
		IngressClass:          "qcloud",
		WorkspaceImage:        "ccr.ccs.tencentyun.com/opl/one-person-lab-app:20260704",
		WorkspaceStorageClass: "cbs",
		ConsoleUsersJSON:      `[{"id":"usr-admin","email":"admin@medopl.cn","password":"StrongAdminPass2026!","role":"admin"}]`,
	})

	if !report.Ready {
		t.Fatalf("ready = false, checks = %#v", report.Checks)
	}
}

func TestProductionReadinessFailsClosedForMissingInputsAndDefaultAdmin(t *testing.T) {
	report := Production(config.Config{
		DatabaseURL:           "",
		PublicURL:             "http://127.0.0.1:8787",
		WorkspaceDomain:       "workspace.medopl.cn",
		KubeNamespace:         "opl-cloud",
		IngressClass:          "nginx",
		WorkspaceImage:        "ghcr.io/gaofeng21cn/one-person-lab-app:latest",
		WorkspaceStorageClass: "cbs",
		ConsoleUsersJSON:      `[{"id":"usr-admin-bootstrap","email":"admin@opl.local","password":"` + auth.DefaultAdminPassword + `","role":"admin"}]`,
	})

	if report.Ready {
		t.Fatalf("ready = true, checks = %#v", report.Checks)
	}
	for _, check := range []string{"database_url", "public_url", "kube_config", "ingress_class", "workspace_image", "auth_seed"} {
		if report.Checks[check] {
			t.Fatalf("%s = true, checks = %#v", check, report.Checks)
		}
	}
}
