package main

import (
	"context"
	"log"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/api"
	"github.com/RenDeHuang/opl-console/internal/config"
	"github.com/RenDeHuang/opl-console/internal/store"
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

	router := api.NewRouter(api.Dependencies{
		RuntimeReady: func() api.Readiness {
			return api.Readiness{Ready: true, Checks: map[string]bool{"postgres": true}}
		},
		ProductionReady: func() api.Readiness {
			return api.Readiness{Ready: false, Checks: map[string]bool{"production_config": false}}
		},
	})
	log.Printf("OPL Console API listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, router))
}
