package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/config"
	"github.com/RenDeHuang/opl-console/internal/upstream"
)

func TestProductionReadinessIncludesReachableUpstreams(t *testing.T) {
	fabric := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/fabric/readiness" {
			t.Fatalf("fabric readiness path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer fabric.Close()
	ledger := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Fatalf("ledger health path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ledger.Close()

	readiness := productionReadiness(config.Config{
		DatabaseURL:           "postgres://opl:secret@db.example.com:5432/opl_console",
		PublicURL:             "https://cloud.medopl.cn",
		WorkspaceDomain:       "workspace.medopl.cn",
		KubeconfigPath:        "/var/run/opl-console/kubeconfig",
		KubeNamespace:         "opl-cloud",
		IngressClass:          "qcloud",
		WorkspaceImage:        "ccr.ccs.tencentyun.com/opl/one-person-lab-app:20260704",
		WorkspaceStorageClass: "cbs",
		FabricProvider:        "tke",
		ConsoleUsersJSON:      `[{"id":"usr-admin","email":"admin@medopl.cn","password":"StrongAdminPass2026!","role":"admin"}]`,
		FabricInternalURL:     fabric.URL,
		LedgerInternalURL:     ledger.URL,
		OperatorToken:         "fabric-token",
		LedgerServiceToken:    "ledger-service-token",
		LedgerAdminToken:      "ledger-admin-token",
	}, upstream.New(upstream.ClientConfig{BaseURL: fabric.URL}), upstream.New(upstream.ClientConfig{BaseURL: ledger.URL}))

	if !readiness.Ready {
		t.Fatalf("ready = false, checks = %#v", readiness.Checks)
	}
	if !readiness.Checks["fabric_reachable"] || !readiness.Checks["ledger_reachable"] {
		t.Fatalf("upstream checks = %#v", readiness.Checks)
	}
}

func TestProductionReadinessFailsWhenUpstreamsAreUnavailable(t *testing.T) {
	readiness := productionReadiness(config.Config{
		DatabaseURL:           "postgres://opl:secret@db.example.com:5432/opl_console",
		PublicURL:             "https://cloud.medopl.cn",
		WorkspaceDomain:       "workspace.medopl.cn",
		KubeconfigPath:        "/var/run/opl-console/kubeconfig",
		KubeNamespace:         "opl-cloud",
		IngressClass:          "qcloud",
		WorkspaceImage:        "ccr.ccs.tencentyun.com/opl/one-person-lab-app:20260704",
		WorkspaceStorageClass: "cbs",
		FabricProvider:        "tke",
		ConsoleUsersJSON:      `[{"id":"usr-admin","email":"admin@medopl.cn","password":"StrongAdminPass2026!","role":"admin"}]`,
		FabricInternalURL:     "http://127.0.0.1:1",
		LedgerInternalURL:     "http://127.0.0.1:1",
		OperatorToken:         "fabric-token",
		LedgerServiceToken:    "ledger-service-token",
		LedgerAdminToken:      "ledger-admin-token",
	}, upstream.New(upstream.ClientConfig{BaseURL: "http://127.0.0.1:1"}), upstream.New(upstream.ClientConfig{BaseURL: "http://127.0.0.1:1"}))

	if readiness.Ready {
		t.Fatalf("ready = true, checks = %#v", readiness.Checks)
	}
	if readiness.Checks["fabric_reachable"] || readiness.Checks["ledger_reachable"] {
		t.Fatalf("upstream checks = %#v", readiness.Checks)
	}
}
