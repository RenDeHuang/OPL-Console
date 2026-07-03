# OPL Console Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first independent OPL Console repository foundation from the approved split design.

**Architecture:** OPL Console starts as a Go modular monolith with React + TypeScript frontend, PostgreSQL persistence, and a `client-go` Fabric adapter boundary. Internal Fabric and Ledger modules are shaped as future external API ports while remaining in-process for v1.

**Tech Stack:** React, TypeScript, Vite, Go, chi, pgx, goose, PostgreSQL, Kubernetes client-go, React Router, TanStack Query.

---

## File Structure Map

Create these top-level paths:

```text
apps/web/                         React + TypeScript UI
cmd/console-api/                  Go API process
cmd/migrate/                      Goose migration runner
internal/api/                     HTTP router, handlers, middleware, DTOs
internal/auth/                    sessions, password hash, RBAC
internal/billing/                 pricing and hold policy
internal/config/                  env parsing
internal/contracts/               JSON contract loader
internal/fabric/                  future OPL Fabric port
internal/fabric/local/            fake/dev Fabric implementation
internal/fabric/tke/              client-go TKE implementation
internal/ledger/                  future OPL Ledger port
internal/ledger/postgres/         PostgreSQL Ledger implementation
internal/store/                   pgx pool, transactions, repositories
internal/workspace/               Workspace lifecycle orchestration
contracts/                        machine-readable Console contracts
migrations/                       PostgreSQL migrations
deploy/k8s/                       Console deployment manifests
docs/                             architecture and migration docs
```

The implementation should use the design spec as source of truth:

```text
docs/superpowers/specs/2026-07-03-opl-console-split-design.md
```

---

### Task 1: Repository Scaffold

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `Makefile`
- Create: `package.json`
- Create: `apps/web/package.json`
- Create: `apps/web/index.html`
- Create: `apps/web/src/main.tsx`
- Create: `apps/web/src/App.tsx`
- Create: `apps/web/src/styles.css`
- Create: `apps/web/tsconfig.json`
- Create: `apps/web/vite.config.ts`
- Create: `docker-compose.dev.yml`

- [ ] **Step 1: Write scaffold files**

Create `go.mod`:

```go
module github.com/RenDeHuang/opl-console

go 1.23

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.2
	github.com/pressly/goose/v3 v3.24.1
	k8s.io/api v0.31.4
	k8s.io/apimachinery v0.31.4
	k8s.io/client-go v0.31.4
)
```

Create `.gitignore`:

```gitignore
.env
.env.local
.runtime/
dist/
node_modules/
apps/web/node_modules/
apps/web/dist/
coverage/
*.log
```

Create `Makefile`:

```makefile
.PHONY: test build migrate-up dev-api dev-web

test:
	go test ./...
	npm --prefix apps/web run typecheck
	npm --prefix apps/web run build

build:
	go build ./cmd/console-api
	npm --prefix apps/web run build

migrate-up:
	go run ./cmd/migrate up

dev-api:
	go run ./cmd/console-api

dev-web:
	npm --prefix apps/web run dev -- --host 127.0.0.1
```

Create root `package.json`:

```json
{
  "private": true,
  "scripts": {
    "test": "make test",
    "build": "make build",
    "dev:web": "npm --prefix apps/web run dev",
    "dev:api": "go run ./cmd/console-api"
  }
}
```

Create `apps/web/package.json`:

```json
{
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "typecheck": "tsc -b"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.83.0",
    "@vitejs/plugin-react": "^4.6.0",
    "lucide-react": "^0.468.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-router": "^7.6.0",
    "vite": "^6.0.7"
  },
  "devDependencies": {
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "typescript": "^5.7.2"
  }
}
```

Create `apps/web/index.html`:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>OPL Console</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

Create `apps/web/src/main.tsx`:

```tsx
import React from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App } from "./App";
import "./styles.css";

const queryClient = new QueryClient();

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </React.StrictMode>
);
```

Create `apps/web/src/App.tsx`:

```tsx
import { BrowserRouter, Route, Routes } from "react-router";

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<main className="shell"><h1>OPL Console</h1></main>} />
      </Routes>
    </BrowserRouter>
  );
}
```

Create `apps/web/src/styles.css`:

```css
:root {
  color: #172033;
  background: #f7f8fb;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

body {
  margin: 0;
}

.shell {
  max-width: 1120px;
  margin: 0 auto;
  padding: 32px;
}
```

Create `apps/web/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "useDefineForClassFields": true,
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "allowJs": false,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx"
  },
  "include": ["src"]
}
```

Create `apps/web/vite.config.ts`:

```ts
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "127.0.0.1",
    port: 5173,
    proxy: {
      "/api": "http://127.0.0.1:8787",
      "/w": "http://127.0.0.1:8787"
    }
  }
});
```

Create `docker-compose.dev.yml`:

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: opl_console
      POSTGRES_USER: opl
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U opl -d opl_console"]
      interval: 5s
      timeout: 5s
      retries: 20
```

- [ ] **Step 2: Download dependencies**

Run:

```bash
go mod tidy
npm install --prefix apps/web
```

Expected: both commands exit `0`, `go.sum` and `apps/web/package-lock.json` are created.

- [ ] **Step 3: Verify scaffold**

Run:

```bash
go test ./...
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
```

Expected: Go reports no packages or passing packages; TypeScript and Vite build pass.

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "chore: scaffold opl console repository"
```

---

### Task 2: PostgreSQL Migration Foundation

**Files:**
- Create: `migrations/00001_initial.sql`
- Create: `internal/config/config.go`
- Create: `internal/store/db.go`
- Create: `cmd/migrate/main.go`
- Create: `internal/store/db_test.go`

- [ ] **Step 1: Create initial migration**

Create `migrations/00001_initial.sql`:

```sql
-- +goose Up
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin')),
  status TEXT NOT NULL CHECK (status IN ('active', 'disabled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organizations (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  billing_account_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memberships (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL REFERENCES organizations(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  role TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, user_id)
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  csrf_token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE billing_accounts (
  id TEXT PRIMARY KEY,
  owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'organization')),
  owner_id TEXT NOT NULL,
  balance_fen BIGINT NOT NULL DEFAULT 0,
  frozen_fen BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wallet_transactions (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  amount_fen BIGINT NOT NULL,
  kind TEXT NOT NULL,
  reason TEXT NOT NULL,
  actor_user_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wallet_holds (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  amount_fen BIGINT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'released', 'debited')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE billing_ledger_entries (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  workspace_id TEXT,
  resource_type TEXT NOT NULL,
  resource_id TEXT,
  amount_fen BIGINT NOT NULL,
  kind TEXT NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE manual_topups (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  amount_fen BIGINT NOT NULL,
  actor_user_id TEXT NOT NULL,
  note TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspace_packages (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  cpu INTEGER NOT NULL,
  memory_gb INTEGER NOT NULL,
  storage_gb INTEGER NOT NULL,
  compute_hourly_fen BIGINT NOT NULL,
  storage_gb_month_fen BIGINT NOT NULL,
  available BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE compute_resources (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  package_id TEXT NOT NULL REFERENCES workspace_packages(id),
  provider_resource_id TEXT,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE storage_volumes (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  package_id TEXT NOT NULL REFERENCES workspace_packages(id),
  provider_resource_id TEXT,
  size_gb INTEGER NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE storage_attachments (
  id TEXT PRIMARY KEY,
  compute_id TEXT NOT NULL REFERENCES compute_resources(id),
  storage_id TEXT NOT NULL REFERENCES storage_volumes(id),
  mount_path TEXT NOT NULL DEFAULT '/data',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspaces (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  name TEXT NOT NULL,
  package_id TEXT NOT NULL REFERENCES workspace_packages(id),
  compute_id TEXT REFERENCES compute_resources(id),
  storage_id TEXT REFERENCES storage_volumes(id),
  attachment_id TEXT REFERENCES storage_attachments(id),
  slug TEXT NOT NULL UNIQUE,
  state TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspace_tokens (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id),
  token_hash TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'deleted')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE runtime_operations (
  id TEXT PRIMARY KEY,
  operation_type TEXT NOT NULL,
  actor_user_id TEXT NOT NULL,
  billing_account_id TEXT,
  workspace_id TEXT,
  stage TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
  error_code TEXT,
  error_message TEXT,
  request_id TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at TIMESTAMPTZ
);

CREATE TABLE support_tickets (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  workspace_id TEXT,
  subject TEXT NOT NULL,
  body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  user_id TEXT,
  level TEXT NOT NULL,
  message TEXT NOT NULL,
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_events (
  id TEXT PRIMARY KEY,
  actor_user_id TEXT NOT NULL,
  action TEXT NOT NULL,
  object_type TEXT NOT NULL,
  object_id TEXT NOT NULL,
  request_id TEXT NOT NULL,
  result TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE receipts (
  id TEXT PRIMARY KEY,
  receipt_type TEXT NOT NULL,
  subject_type TEXT NOT NULL,
  subject_id TEXT NOT NULL,
  operation_id TEXT,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO workspace_packages (id, name, cpu, memory_gb, storage_gb, compute_hourly_fen, storage_gb_month_fen, available)
VALUES
  ('basic', 'Basic Workspace', 2, 4, 10, 39, 36, true),
  ('pro', 'Pro Workspace', 8, 16, 100, 309, 36, true);

-- +goose Down
DROP TABLE receipts;
DROP TABLE audit_events;
DROP TABLE notifications;
DROP TABLE support_tickets;
DROP TABLE runtime_operations;
DROP TABLE workspace_tokens;
DROP TABLE workspaces;
DROP TABLE storage_attachments;
DROP TABLE storage_volumes;
DROP TABLE compute_resources;
DROP TABLE workspace_packages;
DROP TABLE manual_topups;
DROP TABLE billing_ledger_entries;
DROP TABLE wallet_holds;
DROP TABLE wallet_transactions;
DROP TABLE billing_accounts;
DROP TABLE sessions;
DROP TABLE memberships;
DROP TABLE organizations;
DROP TABLE users;
```

- [ ] **Step 2: Implement config**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr                  string
	DatabaseURL           string
	PublicURL             string
	WorkspaceDomain       string
	KubeconfigPath         string
	KubeNamespace          string
	IngressClass           string
	WorkspaceImage         string
	WorkspaceStorageClass  string
	SessionCookieName      string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:                 env("OPL_CONSOLE_ADDR", "127.0.0.1:8787"),
		DatabaseURL:          env("DATABASE_URL", "postgres://opl:secret@127.0.0.1:5432/opl_console?sslmode=disable"),
		PublicURL:            env("OPL_PUBLIC_URL", "http://127.0.0.1:8787"),
		WorkspaceDomain:      env("OPL_WORKSPACE_DOMAIN", "workspace.medopl.cn"),
		KubeconfigPath:        env("KUBECONFIG", ""),
		KubeNamespace:         env("OPL_K8S_NAMESPACE", "opl-cloud"),
		IngressClass:          env("OPL_INGRESS_CLASS", "nginx"),
		WorkspaceImage:        env("OPL_WORKSPACE_IMAGE", "ghcr.io/gaofeng21cn/one-person-lab-app:latest"),
		WorkspaceStorageClass: env("OPL_WORKSPACE_STORAGE_CLASS", "cbs"),
		SessionCookieName:     env("OPL_SESSION_COOKIE_NAME", "opl_console_session"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func EnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
```

- [ ] **Step 3: Implement database connector**

Create `internal/store/db.go`:

```go
package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
```

- [ ] **Step 4: Implement migration runner**

Create `cmd/migrate/main.go`:

```go
package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/RenDeHuang/opl-console/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Run(command, db, "migrations"); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 5: Add database smoke test**

Create `internal/store/db_test.go`:

```go
package store

import (
	"context"
	"os"
	"testing"
)

func TestOpenRequiresReachableDatabase(t *testing.T) {
	t.Setenv("PGCONNECT_TIMEOUT", "1")
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := Open(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()
}
```

- [ ] **Step 6: Run migration against local PostgreSQL**

Run:

```bash
docker compose -f docker-compose.dev.yml up -d postgres
go run ./cmd/migrate up
```

Expected: migration completes and creates all initial tables.

- [ ] **Step 7: Verify**

Run:

```bash
go test ./...
```

Expected: tests pass, with `TestOpenRequiresReachableDatabase` skipped unless `TEST_DATABASE_URL` is set.

- [ ] **Step 8: Commit**

```bash
git add .
git commit -m "feat: add postgres migration foundation"
```

---

### Task 3: Go API Shell and Readiness

**Files:**
- Create: `cmd/console-api/main.go`
- Create: `internal/api/router.go`
- Create: `internal/api/readiness.go`
- Create: `internal/api/router_test.go`

- [ ] **Step 1: Write failing API tests**

Create `internal/api/router_test.go`:

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	handler := NewRouter(Dependencies{})
	request := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if body := response.Body.String(); body != "{\"ok\":true}\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestRuntimeReadiness(t *testing.T) {
	handler := NewRouter(Dependencies{RuntimeReady: func() Readiness {
		return Readiness{Ready: true, Checks: map[string]bool{"postgres": true}}
	}})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/readiness", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
go test ./internal/api
```

Expected: fail because `NewRouter`, `Dependencies`, and `Readiness` are undefined.

- [ ] **Step 3: Implement router and readiness DTO**

Create `internal/api/readiness.go`:

```go
package api

type Readiness struct {
	Ready  bool            `json:"ready"`
	Checks map[string]bool `json:"checks"`
}
```

Create `internal/api/router.go`:

```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Dependencies struct {
	RuntimeReady    func() Readiness
	ProductionReady func() Readiness
}

func NewRouter(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Get("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	router.Get("/api/runtime/readiness", func(w http.ResponseWriter, r *http.Request) {
		check := deps.RuntimeReady
		if check == nil {
			check = func() Readiness { return Readiness{Ready: false, Checks: map[string]bool{"configured": false}} }
		}
		writeJSON(w, http.StatusOK, check())
	})
	router.Get("/api/production/readiness", func(w http.ResponseWriter, r *http.Request) {
		check := deps.ProductionReady
		if check == nil {
			check = func() Readiness { return Readiness{Ready: false, Checks: map[string]bool{"configured": false}} }
		}
		writeJSON(w, http.StatusOK, check())
	})
	return router
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
```

Create `cmd/console-api/main.go`:

```go
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
```

- [ ] **Step 4: Verify tests pass**

Run:

```bash
go test ./...
```

Expected: all Go tests pass.

- [ ] **Step 5: Smoke test server**

Run:

```bash
docker compose -f docker-compose.dev.yml up -d postgres
go run ./cmd/migrate up
go run ./cmd/console-api
```

In a second terminal:

```bash
curl -s http://127.0.0.1:8787/api/healthz
```

Expected:

```json
{"ok":true}
```

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "feat: add console api health and readiness"
```

---

### Task 4: Contracts Migration Baseline

**Files:**
- Create: `contracts/opl-console-route-api-contract.json`
- Create: `contracts/opl-console-business-object-contract.json`
- Create: `contracts/opl-console-fabric-api-boundary-contract.json`
- Create: `contracts/opl-console-ledger-api-boundary-contract.json`
- Create: `internal/contracts/loader.go`
- Create: `internal/contracts/loader_test.go`

- [ ] **Step 1: Add route contract**

Create `contracts/opl-console-route-api-contract.json`:

```json
{
  "schemaVersion": 1,
  "owner": "OPL Console",
  "purpose": "Current public, owner, admin, and workspace routes owned by the OPL Console repository.",
  "lifecycle": "current",
  "routes": {
    "public": [
      "GET /api/healthz",
      "GET /api/runtime/readiness",
      "GET /api/production/readiness",
      "POST /api/auth/login",
      "POST /api/auth/logout",
      "GET /api/auth/session"
    ],
    "owner": [
      "GET /api/me",
      "GET /api/packages",
      "GET /api/workspaces",
      "POST /api/workspaces",
      "GET /api/workspaces/{id}",
      "POST /api/workspaces/{id}/stop",
      "POST /api/workspaces/{id}/restart",
      "POST /api/workspaces/{id}/destroy-compute",
      "POST /api/workspaces/{id}/destroy-storage",
      "POST /api/workspaces/{id}/tokens/reset",
      "POST /api/workspaces/{id}/tokens/delete",
      "GET /api/billing/wallet",
      "GET /api/billing/ledger",
      "GET /api/support/tickets",
      "POST /api/support/tickets"
    ],
    "admin": [
      "GET /api/admin/users",
      "GET /api/admin/workspaces",
      "GET /api/admin/runtime",
      "GET /api/admin/fabric/catalog",
      "GET /api/admin/ledger",
      "GET /api/admin/audit",
      "POST /api/admin/users/{id}/topups",
      "GET /api/admin/support/tickets"
    ],
    "workspace": [
      "GET /w/{workspaceId}?token=..."
    ]
  }
}
```

Create `contracts/opl-console-business-object-contract.json`:

```json
{
  "schemaVersion": 1,
  "owner": "OPL Console",
  "purpose": "Business objects persisted and presented by OPL Console.",
  "lifecycle": "current",
  "objects": [
    "User",
    "Organization",
    "Membership",
    "BillingAccount",
    "WalletHold",
    "BillingLedgerEntry",
    "ComputeResource",
    "StorageVolume",
    "StorageAttachment",
    "Workspace",
    "WorkspaceToken",
    "RuntimeOperation",
    "SupportTicket",
    "AuditEvent",
    "Receipt"
  ]
}
```

Create `contracts/opl-console-fabric-api-boundary-contract.json`:

```json
{
  "schemaVersion": 1,
  "owner": "OPL Fabric boundary inside OPL Console",
  "purpose": "Stable internal API shape for the v1 in-process Fabric module and future standalone OPL Fabric service.",
  "lifecycle": "current",
  "operations": [
    "ListCatalog",
    "RuntimeReadiness",
    "CreateCompute",
    "StopCompute",
    "RestartCompute",
    "DestroyCompute",
    "CreateStorage",
    "AttachStorage",
    "DetachStorage",
    "DestroyStorage",
    "CreateWorkspaceRoute",
    "ResetWorkspaceToken",
    "DeleteWorkspaceToken",
    "InspectRuntime"
  ]
}
```

Create `contracts/opl-console-ledger-api-boundary-contract.json`:

```json
{
  "schemaVersion": 1,
  "owner": "OPL Ledger boundary inside OPL Console",
  "purpose": "Stable internal API shape for the v1 in-process Ledger module and future standalone OPL Ledger service.",
  "lifecycle": "current",
  "operations": [
    "GetWallet",
    "FreezeHold",
    "ReleaseHold",
    "DebitHold",
    "RecordWalletTransaction",
    "RecordManualTopUp",
    "RecordBillingLedgerEntry",
    "RecordAuditEvent",
    "RecordReceipt",
    "ListAccountLedger",
    "ListAdminLedger"
  ]
}
```

- [ ] **Step 2: Write failing contract loader test**

Create `internal/contracts/loader_test.go`:

```go
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
```

- [ ] **Step 3: Run test to verify failure**

Run:

```bash
go test ./internal/contracts
```

Expected: fail because `LoadDirectory` is undefined.

- [ ] **Step 4: Implement contract loader**

Create `internal/contracts/loader.go`:

```go
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
```

- [ ] **Step 5: Verify**

Run:

```bash
go test ./internal/contracts
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "feat: add opl console contract baseline"
```

---

### Task 5: Auth and RBAC Skeleton

**Files:**
- Create: `internal/auth/auth.go`
- Create: `internal/auth/auth_test.go`

- [ ] **Step 1: Write failing auth test**

Create `internal/auth/auth_test.go`:

```go
package auth

import "testing"

func TestCanAccessAdmin(t *testing.T) {
	admin := User{ID: "usr-admin", Role: RoleAdmin, Status: StatusActive}
	owner := User{ID: "usr-owner", Role: RoleOwner, Status: StatusActive}

	if !CanAccessAdmin(admin) {
		t.Fatal("admin should access admin routes")
	}
	if CanAccessAdmin(owner) {
		t.Fatal("owner should not access admin routes")
	}
}

func TestDisabledUserCannotAccessOwner(t *testing.T) {
	user := User{ID: "usr-disabled", Role: RoleOwner, Status: StatusDisabled}
	if CanAccessOwner(user) {
		t.Fatal("disabled user should not access owner routes")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
go test ./internal/auth
```

Expected: fail because auth types/functions are undefined.

- [ ] **Step 3: Implement auth roles**

Create `internal/auth/auth.go`:

```go
package auth

type Role string
type Status string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"

	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type User struct {
	ID     string
	Email  string
	Role   Role
	Status Status
}

func CanAccessOwner(user User) bool {
	return user.Status == StatusActive && (user.Role == RoleOwner || user.Role == RoleAdmin)
}

func CanAccessAdmin(user User) bool {
	return user.Status == StatusActive && user.Role == RoleAdmin
}
```

- [ ] **Step 4: Verify**

Run:

```bash
go test ./...
```

Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add auth role boundary"
```

---

### Task 6: Ledger Port and PostgreSQL Wallet Foundation

**Files:**
- Create: `internal/ledger/port.go`
- Create: `internal/ledger/postgres/wallet.go`
- Create: `internal/ledger/postgres/wallet_test.go`

- [ ] **Step 1: Write failing Ledger interface test**

Create `internal/ledger/postgres/wallet_test.go`:

```go
package postgres

import (
	"testing"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

func TestStoreSatisfiesLedgerPort(t *testing.T) {
	var _ ledger.Port = (*Store)(nil)
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
go test ./internal/ledger/...
```

Expected: fail because `ledger.Port` and `Store` are undefined.

- [ ] **Step 3: Define Ledger port**

Create `internal/ledger/port.go`:

```go
package ledger

import "context"

type Wallet struct {
	BillingAccountID string `json:"billingAccountId"`
	BalanceFen      int64  `json:"balanceFen"`
	FrozenFen       int64  `json:"frozenFen"`
	AvailableFen    int64  `json:"availableFen"`
}

type HoldRequest struct {
	HoldID           string
	BillingAccountID string
	ResourceType     string
	ResourceID       string
	AmountFen        int64
	ActorUserID      string
}

type TopUpRequest struct {
	TopUpID          string
	BillingAccountID string
	AmountFen        int64
	ActorUserID      string
	Note             string
}

type Port interface {
	GetWallet(ctx context.Context, billingAccountID string) (Wallet, error)
	FreezeHold(ctx context.Context, request HoldRequest) error
	ReleaseHold(ctx context.Context, holdID string, actorUserID string) error
	DebitHold(ctx context.Context, holdID string, actorUserID string) error
	RecordManualTopUp(ctx context.Context, request TopUpRequest) error
}
```

- [ ] **Step 4: Implement minimal PostgreSQL store**

Create `internal/ledger/postgres/wallet.go`:

```go
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetWallet(ctx context.Context, billingAccountID string) (ledger.Wallet, error) {
	var wallet ledger.Wallet
	err := s.pool.QueryRow(ctx, `
		SELECT id, balance_fen, frozen_fen, balance_fen - frozen_fen
		FROM billing_accounts
		WHERE id = $1
	`, billingAccountID).Scan(&wallet.BillingAccountID, &wallet.BalanceFen, &wallet.FrozenFen, &wallet.AvailableFen)
	return wallet, err
}

func (s *Store) FreezeHold(ctx context.Context, request ledger.HoldRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var available int64
	if err := tx.QueryRow(ctx, `
		SELECT balance_fen - frozen_fen FROM billing_accounts WHERE id = $1 FOR UPDATE
	`, request.BillingAccountID).Scan(&available); err != nil {
		return err
	}
	if available < request.AmountFen {
		return fmt.Errorf("insufficient_balance")
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_holds (id, billing_account_id, resource_type, resource_id, amount_fen, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
	`, request.HoldID, request.BillingAccountID, request.ResourceType, request.ResourceID, request.AmountFen); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing_accounts SET frozen_fen = frozen_fen + $1, updated_at = now() WHERE id = $2
	`, request.AmountFen, request.BillingAccountID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ReleaseHold(ctx context.Context, holdID string, actorUserID string) error {
	_, err := s.pool.Exec(ctx, `
		WITH hold AS (
			UPDATE wallet_holds SET status = 'released', updated_at = now()
			WHERE id = $1 AND status = 'active'
			RETURNING billing_account_id, amount_fen
		)
		UPDATE billing_accounts
		SET frozen_fen = frozen_fen - hold.amount_fen, updated_at = now()
		FROM hold
		WHERE billing_accounts.id = hold.billing_account_id
	`, holdID)
	return err
}

func (s *Store) DebitHold(ctx context.Context, holdID string, actorUserID string) error {
	_, err := s.pool.Exec(ctx, `
		WITH hold AS (
			UPDATE wallet_holds SET status = 'debited', updated_at = now()
			WHERE id = $1 AND status = 'active'
			RETURNING billing_account_id, amount_fen
		)
		UPDATE billing_accounts
		SET frozen_fen = frozen_fen - hold.amount_fen,
		    balance_fen = balance_fen - hold.amount_fen,
		    updated_at = now()
		FROM hold
		WHERE billing_accounts.id = hold.billing_account_id
	`, holdID)
	return err
}

func (s *Store) RecordManualTopUp(ctx context.Context, request ledger.TopUpRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		INSERT INTO manual_topups (id, billing_account_id, amount_fen, actor_user_id, note)
		VALUES ($1, $2, $3, $4, $5)
	`, request.TopUpID, request.BillingAccountID, request.AmountFen, request.ActorUserID, request.Note); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing_accounts SET balance_fen = balance_fen + $1, updated_at = now() WHERE id = $2
	`, request.AmountFen, request.BillingAccountID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
```

- [ ] **Step 5: Verify**

Run:

```bash
go test ./internal/ledger/...
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "feat: add internal ledger wallet port"
```

---

### Task 7: Workspace Domain and Fabric Port

**Files:**
- Create: `internal/fabric/port.go`
- Create: `internal/fabric/local/fake.go`
- Create: `internal/workspace/service.go`
- Create: `internal/workspace/service_test.go`

- [ ] **Step 1: Write failing workspace lifecycle test**

Create `internal/workspace/service_test.go`:

```go
package workspace

import (
	"context"
	"testing"

	"github.com/RenDeHuang/opl-console/internal/fabric/local"
)

func TestCreateWorkspaceUsesFabricPort(t *testing.T) {
	fabric := local.New()
	service := NewService(fabric)

	result, err := service.CreateWorkspace(context.Background(), CreateWorkspaceRequest{
		WorkspaceID:      "ws-alpha",
		Name:             "Alpha Lab",
		BillingAccountID: "acct-owner",
		PackageID:        "basic",
		Token:            "share-token",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if result.WorkspaceID != "ws-alpha" {
		t.Fatalf("workspace id = %q", result.WorkspaceID)
	}
	if result.URL == "" {
		t.Fatal("workspace URL is empty")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
go test ./internal/workspace
```

Expected: fail because workspace and fabric packages are undefined.

- [ ] **Step 3: Define Fabric port**

Create `internal/fabric/port.go`:

```go
package fabric

import "context"

type PackagePlan struct {
	ID        string
	CPU       int
	MemoryGB  int
	StorageGB int
}

type CreateComputeRequest struct {
	ComputeID        string
	BillingAccountID string
	Package          PackagePlan
}

type CreateStorageRequest struct {
	StorageID        string
	BillingAccountID string
	Package          PackagePlan
}

type AttachStorageRequest struct {
	AttachmentID string
	ComputeID    string
	StorageID    string
	MountPath    string
}

type CreateRouteRequest struct {
	WorkspaceID string
	WorkspaceName string
	ComputeID    string
	Token        string
}

type RuntimeHandle struct {
	ProviderResourceID string
	Status             string
	URL                string
}

type Port interface {
	CreateCompute(ctx context.Context, request CreateComputeRequest) (RuntimeHandle, error)
	CreateStorage(ctx context.Context, request CreateStorageRequest) (RuntimeHandle, error)
	AttachStorage(ctx context.Context, request AttachStorageRequest) (RuntimeHandle, error)
	CreateWorkspaceRoute(ctx context.Context, request CreateRouteRequest) (RuntimeHandle, error)
}
```

- [ ] **Step 4: Implement local fake Fabric**

Create `internal/fabric/local/fake.go`:

```go
package local

import (
	"context"
	"fmt"
	"strings"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

type Fake struct{}

func New() *Fake {
	return &Fake{}
}

func (f *Fake) CreateCompute(ctx context.Context, request fabric.CreateComputeRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: "local-compute/" + request.ComputeID, Status: "running"}, nil
}

func (f *Fake) CreateStorage(ctx context.Context, request fabric.CreateStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: "local-storage/" + request.StorageID, Status: "available"}, nil
}

func (f *Fake) AttachStorage(ctx context.Context, request fabric.AttachStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: fmt.Sprintf("%s:%s", request.ComputeID, request.StorageID), Status: "attached"}, nil
}

func (f *Fake) CreateWorkspaceRoute(ctx context.Context, request fabric.CreateRouteRequest) (fabric.RuntimeHandle, error) {
	slug := strings.ToLower(strings.ReplaceAll(request.WorkspaceName, " ", "-"))
	return fabric.RuntimeHandle{
		ProviderResourceID: "local-route/" + request.WorkspaceID,
		Status:             "ready",
		URL:                "http://127.0.0.1:8787/w/" + slug + "?token=" + request.Token,
	}, nil
}
```

- [ ] **Step 5: Implement workspace service skeleton**

Create `internal/workspace/service.go`:

```go
package workspace

import (
	"context"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

type Service struct {
	fabric fabric.Port
}

func NewService(fabricPort fabric.Port) *Service {
	return &Service{fabric: fabricPort}
}

type CreateWorkspaceRequest struct {
	WorkspaceID      string
	Name             string
	BillingAccountID string
	PackageID        string
	Token            string
}

type CreateWorkspaceResult struct {
	WorkspaceID string
	URL         string
}

func (s *Service) CreateWorkspace(ctx context.Context, request CreateWorkspaceRequest) (CreateWorkspaceResult, error) {
	plan := fabric.PackagePlan{ID: request.PackageID, CPU: 2, MemoryGB: 4, StorageGB: 10}
	computeID := "cmp-" + request.WorkspaceID
	storageID := "stg-" + request.WorkspaceID
	attachmentID := "att-" + request.WorkspaceID

	if _, err := s.fabric.CreateStorage(ctx, fabric.CreateStorageRequest{
		StorageID: storageID, BillingAccountID: request.BillingAccountID, Package: plan,
	}); err != nil {
		return CreateWorkspaceResult{}, err
	}
	if _, err := s.fabric.CreateCompute(ctx, fabric.CreateComputeRequest{
		ComputeID: computeID, BillingAccountID: request.BillingAccountID, Package: plan,
	}); err != nil {
		return CreateWorkspaceResult{}, err
	}
	if _, err := s.fabric.AttachStorage(ctx, fabric.AttachStorageRequest{
		AttachmentID: attachmentID, ComputeID: computeID, StorageID: storageID, MountPath: "/data",
	}); err != nil {
		return CreateWorkspaceResult{}, err
	}
	route, err := s.fabric.CreateWorkspaceRoute(ctx, fabric.CreateRouteRequest{
		WorkspaceID: request.WorkspaceID, WorkspaceName: request.Name, ComputeID: computeID, Token: request.Token,
	})
	if err != nil {
		return CreateWorkspaceResult{}, err
	}
	return CreateWorkspaceResult{WorkspaceID: request.WorkspaceID, URL: route.URL}, nil
}
```

- [ ] **Step 6: Verify**

Run:

```bash
go test ./internal/fabric/... ./internal/workspace
```

Expected: pass.

- [ ] **Step 7: Commit**

```bash
git add .
git commit -m "feat: add workspace lifecycle and fabric port"
```

---

### Task 8: TKE client-go Adapter Skeleton

**Files:**
- Create: `internal/fabric/tke/client.go`
- Create: `internal/fabric/tke/client_test.go`

- [ ] **Step 1: Write failing client-go test**

Create `internal/fabric/tke/client_test.go`:

```go
package tke

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

func TestCreateComputeCreatesDeploymentAndService(t *testing.T) {
	client := New(Config{
		Namespace:    "opl-cloud",
		Image:        "ghcr.io/gaofeng21cn/one-person-lab-app:latest",
		StorageClass: "cbs",
		IngressClass: "nginx",
	}, fake.NewSimpleClientset())

	_, err := client.CreateCompute(context.Background(), fabric.CreateComputeRequest{
		ComputeID:        "cmp-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	})
	if err != nil {
		t.Fatalf("create compute: %v", err)
	}

	deployments, err := client.client.AppsV1().Deployments("opl-cloud").List(context.Background(), metav1ListOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments.Items) != 1 {
		t.Fatalf("deployments = %d", len(deployments.Items))
	}
	if deployments.Items[0].TypeMeta.Kind != "" {
		var _ appsv1.Deployment
	}

	services, err := client.client.CoreV1().Services("opl-cloud").List(context.Background(), metav1ListOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(services.Items) != 1 {
		t.Fatalf("services = %d", len(services.Items))
	}
	if services.Items[0].Spec.Type != corev1.ServiceTypeClusterIP {
		t.Fatalf("service type = %s", services.Items[0].Spec.Type)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
go test ./internal/fabric/tke
```

Expected: fail because `New`, `Config`, and `metav1ListOptions` are undefined.

- [ ] **Step 3: Implement TKE adapter skeleton**

Create `internal/fabric/tke/client.go`:

```go
package tke

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

type Config struct {
	Namespace    string
	Image        string
	StorageClass string
	IngressClass string
}

type Client struct {
	cfg    Config
	client kubernetes.Interface
}

func New(cfg Config, client kubernetes.Interface) *Client {
	return &Client{cfg: cfg, client: client}
}

func (c *Client) CreateCompute(ctx context.Context, request fabric.CreateComputeRequest) (fabric.RuntimeHandle, error) {
	name := request.ComputeID
	labels := map[string]string{"app": "opl-workspace", "opl-workspace-compute": request.ComputeID}
	replicas := int32(1)
	_, err := c.client.AppsV1().Deployments(c.cfg.Namespace).Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "workspace",
						Image: c.cfg.Image,
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					}},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, err
	}
	_, err = c.client.CoreV1().Services(c.cfg.Namespace).Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{{Name: "http", Port: 3000}},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, err
	}
	return fabric.RuntimeHandle{ProviderResourceID: "deployment/" + name, Status: "running"}, nil
}

func (c *Client) CreateStorage(ctx context.Context, request fabric.CreateStorageRequest) (fabric.RuntimeHandle, error) {
	storageClass := c.cfg.StorageClass
	_, err := c.client.CoreV1().PersistentVolumeClaims(c.cfg.Namespace).Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: request.StorageID},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &storageClass,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, err
	}
	return fabric.RuntimeHandle{ProviderResourceID: "pvc/" + request.StorageID, Status: "available"}, nil
}

func (c *Client) AttachStorage(ctx context.Context, request fabric.AttachStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: request.AttachmentID, Status: "attached"}, nil
}

func (c *Client) CreateWorkspaceRoute(ctx context.Context, request fabric.CreateRouteRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{ProviderResourceID: "ingress/" + request.WorkspaceID, Status: "ready"}, nil
}
```

Create `internal/fabric/tke/client_test_helpers.go` if needed:

```go
package tke

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func metav1ListOptions() metav1.ListOptions {
	return metav1.ListOptions{}
}
```

- [ ] **Step 4: Verify**

Run:

```bash
go test ./internal/fabric/tke
go test ./internal/fabric/...
```

Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add tke client-go fabric adapter skeleton"
```

---

### Task 9: Frontend Console Shell

**Files:**
- Create: `apps/web/src/api/client.ts`
- Create: `apps/web/src/pages/LoginPage.tsx`
- Create: `apps/web/src/pages/OwnerOverviewPage.tsx`
- Create: `apps/web/src/pages/AdminOverviewPage.tsx`
- Modify: `apps/web/src/App.tsx`
- Modify: `apps/web/src/styles.css`

- [ ] **Step 1: Create API client**

Create `apps/web/src/api/client.ts`:

```ts
export type Healthz = { ok: boolean };
export type Readiness = { ready: boolean; checks: Record<string, boolean> };

async function request<T>(path: string): Promise<T> {
  const response = await fetch(path, { credentials: "include" });
  if (!response.ok) {
    throw new Error(`request_failed:${response.status}`);
  }
  return response.json() as Promise<T>;
}

export const api = {
  healthz: () => request<Healthz>("/api/healthz"),
  runtimeReadiness: () => request<Readiness>("/api/runtime/readiness")
};
```

- [ ] **Step 2: Create pages**

Create `apps/web/src/pages/LoginPage.tsx`:

```tsx
export function LoginPage() {
  return (
    <main className="shell narrow">
      <h1>OPL Console</h1>
      <form className="panel">
        <label>
          Email
          <input type="email" name="email" autoComplete="email" />
        </label>
        <label>
          Password
          <input type="password" name="password" autoComplete="current-password" />
        </label>
        <button type="submit">Sign in</button>
      </form>
    </main>
  );
}
```

Create `apps/web/src/pages/OwnerOverviewPage.tsx`:

```tsx
import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

export function OwnerOverviewPage() {
  const readiness = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });

  return (
    <main className="shell">
      <h1>Workspaces</h1>
      <section className="panel">
        <h2>Runtime</h2>
        <p>{readiness.data?.ready ? "Ready" : "Not ready"}</p>
      </section>
    </main>
  );
}
```

Create `apps/web/src/pages/AdminOverviewPage.tsx`:

```tsx
export function AdminOverviewPage() {
  return (
    <main className="shell">
      <h1>Admin</h1>
      <section className="panel">
        <h2>Runtime Readiness</h2>
      </section>
    </main>
  );
}
```

- [ ] **Step 3: Wire routes**

Modify `apps/web/src/App.tsx`:

```tsx
import { BrowserRouter, Link, Route, Routes } from "react-router";
import { AdminOverviewPage } from "./pages/AdminOverviewPage";
import { LoginPage } from "./pages/LoginPage";
import { OwnerOverviewPage } from "./pages/OwnerOverviewPage";

export function App() {
  return (
    <BrowserRouter>
      <nav className="topbar">
        <Link to="/">OPL Console</Link>
        <Link to="/login">Login</Link>
        <Link to="/admin">Admin</Link>
      </nav>
      <Routes>
        <Route path="/" element={<OwnerOverviewPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<AdminOverviewPage />} />
      </Routes>
    </BrowserRouter>
  );
}
```

Modify `apps/web/src/styles.css`:

```css
:root {
  color: #172033;
  background: #f7f8fb;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

body {
  margin: 0;
}

input,
button {
  font: inherit;
}

button {
  border: 0;
  background: #1769aa;
  color: #fff;
  padding: 10px 14px;
  border-radius: 6px;
  cursor: pointer;
}

.topbar {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 0 32px;
  border-bottom: 1px solid #d9dee7;
  background: #fff;
}

.topbar a {
  color: #172033;
  text-decoration: none;
}

.shell {
  max-width: 1120px;
  margin: 0 auto;
  padding: 32px;
}

.narrow {
  max-width: 420px;
}

.panel {
  background: #fff;
  border: 1px solid #d9dee7;
  border-radius: 8px;
  padding: 20px;
}

.panel label {
  display: grid;
  gap: 6px;
  margin-bottom: 14px;
}

.panel input {
  border: 1px solid #c7ceda;
  border-radius: 6px;
  padding: 10px 12px;
}
```

- [ ] **Step 4: Verify**

Run:

```bash
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
```

Expected: TypeScript and Vite build pass.

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add react console shell"
```

---

### Task 10: Documentation and Verification Contract

**Files:**
- Create: `docs/architecture.md`
- Create: `docs/routes.md`
- Create: `docs/migration-from-opl-cloud.md`
- Create: `README.md`

- [ ] **Step 1: Write architecture doc**

Create `docs/architecture.md`:

```markdown
# OPL Console Architecture

OPL Console is the independent control-plane repository for OPL Console and OPL Workspace lifecycle management.

The backend is a Go modular monolith. Fabric and Ledger are internal modules in v1, but their ports are shaped as future OPL Fabric and OPL Ledger API boundaries.

The frontend is React + TypeScript. PostgreSQL is the only runtime persistence store. Kubernetes provisioning uses Go `client-go`.
```

- [ ] **Step 2: Write routes doc**

Create `docs/routes.md`:

```markdown
# OPL Console Routes

Routes are fixed by `contracts/opl-console-route-api-contract.json`.

Public routes expose health, readiness, and auth. Lab Owner routes expose workspace, billing, and support workflows. Admin routes expose users, readiness, Fabric internals, Ledger internals, audit, and support queue.
```

- [ ] **Step 3: Write migration doc**

Create `docs/migration-from-opl-cloud.md`:

```markdown
# Migration From OPL-Cloud

OPL Console keeps the OPL Console and OPL Workspace control-plane product rules from OPL-Cloud.

Kept: product names, Basic and Pro CPU packages, seven-day holds, Workspace URL token lifecycle, owner/admin surface separation, billing ledger, audit, receipts, readiness, and contracts.

Rewritten: Node API to Go, React JS to React + TypeScript, state-object persistence to PostgreSQL relational repositories, `kubectl` shell provider to `client-go`, and JS tests to Go/TypeScript tests.

Discarded for v1: JSON runtime store, JS compatibility aliases, Local Docker as product runtime, and standalone Fabric/Ledger deployment.
```

- [ ] **Step 4: Write README**

Create `README.md`:

```markdown
# OPL Console

OPL Console is the independent control-plane repository for OPL Console and OPL Workspace.

## Stack

- Frontend: React + TypeScript
- Backend: Go
- DB: PostgreSQL
- K8s: Go client-go

## Development

```bash
docker compose -f docker-compose.dev.yml up -d postgres
go run ./cmd/migrate up
go run ./cmd/console-api
npm --prefix apps/web run dev
```

## Verification

```bash
make test
```
```

- [ ] **Step 5: Verify full repository**

Run:

```bash
gofmt -w cmd internal
go mod tidy
go test ./...
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
git diff --check
```

Expected: all commands pass.

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "docs: document opl console split"
```

---

## Plan Self-Review Checklist

- Spec coverage: this plan covers repository scaffold, PostgreSQL, API shell, contracts, auth boundary, Ledger port, Workspace/Fabric port, TKE `client-go` skeleton, React shell, and docs.
- Scope boundary: this plan does not implement OPL Gateway, `one-person-lab-app`, standalone OPL Fabric service, standalone OPL Ledger service, GPU packages, marketplaces, or external payment settlement.
- Type consistency: money is CNY fen `BIGINT`; Go fields use `AmountFen`, `BalanceFen`, `FrozenFen`, and `AvailableFen`.
- Recovery foundation: `runtime_operations` exists in the migration; detailed compensation behavior should be implemented in the next lifecycle plan after this foundation passes.
- Verification: every task has at least one concrete command and expected result.
