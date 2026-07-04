package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/RenDeHuang/opl-console/internal/api"
	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/config"
	"github.com/RenDeHuang/opl-console/internal/console"
	"github.com/RenDeHuang/opl-console/internal/fabric"
	"github.com/RenDeHuang/opl-console/internal/fabric/local"
	"github.com/RenDeHuang/opl-console/internal/fabric/tke"
	ledgerpostgres "github.com/RenDeHuang/opl-console/internal/ledger/postgres"
	"github.com/RenDeHuang/opl-console/internal/readiness"
	"github.com/RenDeHuang/opl-console/internal/store"
	"github.com/RenDeHuang/opl-console/internal/workspace"
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
	authStore := store.NewAuthStore(pool)
	bootstrapUsers, err := auth.BootstrapUsersFromJSON(cfg.ConsoleUsersJSON)
	if err != nil {
		log.Fatal(err)
	}
	if err := authStore.SeedBootstrapUsers(context.Background(), bootstrapUsers); err != nil {
		log.Fatal(err)
	}
	authService := auth.NewService(auth.ServiceConfig{
		Users:    authStore,
		Sessions: authStore,
	})
	ledgerPort := ledgerpostgres.New(pool)
	governanceService := console.NewService(store.NewGovernanceStore(pool), console.WithLedger(ledgerPort))
	fabricPort, err := buildFabricPort(cfg)
	if err != nil {
		log.Fatal(err)
	}
	workspaceService := workspace.NewService(
		fabricPort,
		workspace.WithRepository(store.NewWorkspaceStore(pool)),
		workspace.WithLedger(ledgerPort),
	)

	router := api.NewRouter(api.Dependencies{
		Auth:              authService,
		Governance:        governanceService,
		Workspace:         workspaceService,
		SessionCookieName: cfg.SessionCookieName,
		RuntimeReady: func() api.Readiness {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := pool.Ping(ctx)
			ready := err == nil
			return api.Readiness{Ready: ready, Checks: map[string]bool{"postgres": ready}}
		},
		ProductionReady: func() api.Readiness {
			report := readiness.Production(cfg)
			return api.Readiness{Ready: report.Ready, Checks: report.Checks}
		},
	})
	server := http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("OPL Console API listening on %s", cfg.Addr)
	log.Fatal(server.ListenAndServe())
}

func buildFabricPort(cfg config.Config) (fabric.Port, error) {
	switch cfg.FabricProvider {
	case "", "local":
		return local.New(), nil
	case "tke":
		return tke.NewFromKubeConfig(tke.Config{
			Namespace:    cfg.KubeNamespace,
			Image:        cfg.WorkspaceImage,
			StorageClass: cfg.WorkspaceStorageClass,
			IngressClass: cfg.IngressClass,
		}, cfg.KubeconfigPath)
	default:
		return nil, fmt.Errorf("unsupported OPL_FABRIC_PROVIDER %q", cfg.FabricProvider)
	}
}
