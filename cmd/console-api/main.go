package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/RenDeHuang/opl-console/internal/api"
	"github.com/RenDeHuang/opl-console/internal/auth"
	"github.com/RenDeHuang/opl-console/internal/config"
	"github.com/RenDeHuang/opl-console/internal/console"
	"github.com/RenDeHuang/opl-console/internal/fabric/local"
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
	governanceService := console.NewService(store.NewGovernanceStore(pool))
	workspaceService := workspace.NewService(local.New())

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
			return api.Readiness{Ready: false, Checks: map[string]bool{"production_config": false}}
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
