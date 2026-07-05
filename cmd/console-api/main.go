package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/RenDeHuang/opl-console/internal/api"
	"github.com/RenDeHuang/opl-console/internal/config"
	"github.com/RenDeHuang/opl-console/internal/readiness"
	"github.com/RenDeHuang/opl-console/internal/store"
	"github.com/RenDeHuang/opl-console/internal/upstream"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := store.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	fabricClient := upstream.New(upstream.ClientConfig{BaseURL: cfg.FabricInternalURL, BearerToken: cfg.OperatorToken})
	ledgerClient := upstream.New(upstream.ClientConfig{BaseURL: cfg.LedgerInternalURL, BearerToken: cfg.LedgerServiceToken})
	ledgerAdminClient := upstream.New(upstream.ClientConfig{BaseURL: cfg.LedgerInternalURL, BearerToken: cfg.LedgerAdminToken})

	router := api.NewRouter(api.Dependencies{
		RuntimeReady: func() api.Readiness {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := pool.Ping(ctx)
			ready := err == nil
			return api.Readiness{Ready: ready, Checks: map[string]bool{"postgres": ready}}
		},
		ProductionReady: func() api.Readiness {
			return productionReadiness(cfg, fabricClient, ledgerClient)
		},
		Fabric:      fabricClient,
		Ledger:      ledgerClient,
		LedgerAdmin: ledgerAdminClient,
	})
	server := http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("OPL Console API listening on %s", cfg.Addr)
	log.Fatal(server.ListenAndServe())
}

func productionReadiness(cfg config.Config, fabricClient *upstream.Client, ledgerClient *upstream.Client) api.Readiness {
	report := readiness.Production(cfg)
	checks := map[string]bool{}
	for key, value := range report.Checks {
		checks[key] = value
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	checks["fabric_reachable"] = fabricClient != nil && fabricClient.Check(ctx, "/api/fabric/readiness")
	checks["ledger_reachable"] = ledgerClient != nil && ledgerClient.Check(ctx, "/healthz")
	ready := report.Ready && checks["fabric_reachable"] && checks["ledger_reachable"]
	return api.Readiness{Ready: ready, Checks: checks}
}
