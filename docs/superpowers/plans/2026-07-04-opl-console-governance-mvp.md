# OPL Console Governance MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the OPL Console governance MVP across API, managed Workspace lifecycle facade, Fabric/Ledger integrations, UI, and production readiness.

**Architecture:** Keep the Go backend as a modular monolith. Console owns governance services and consumes Fabric/Ledger through ports. PostgreSQL stores Console governance state and managed-resource facade state. React + TypeScript consumes the Console governance API.

**Tech Stack:** Go 1.23, chi, pgx, goose, PostgreSQL, Kubernetes client-go, React, TypeScript, Vite, React Router, TanStack Query.

---

## Task 1: Governance Contract And Schema

**Files:**
- Modify: `contracts/opl-console-business-object-contract.json`
- Modify: `contracts/opl-console-route-api-contract.json`
- Modify: `migrations/00001_initial.sql`
- Create: `internal/store/schema_test.go`

- [ ] Add `teams`, `roles`, `policies`, `approvals`, and `managed_resource_views`
  to the migration with foreign keys to existing users and organizations.
- [ ] Add schema tests that parse the migration and assert every governance table
  exists.
- [ ] Run `GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/store`.
- [ ] Commit with `feat: add governance schema`.

## Task 2: Auth And Session API

**Files:**
- Modify: `internal/auth/auth.go`
- Create: `internal/auth/session.go`
- Create: `internal/auth/session_test.go`
- Create: `internal/store/auth.go`
- Create: `internal/api/auth.go`
- Modify: `internal/api/router.go`
- Modify: `cmd/console-api/main.go`

- [ ] Add password hashing and verification using `bcrypt`.
- [ ] Add session token hashing, CSRF token hashing, and secure cookie helpers.
- [ ] Add PostgreSQL-backed user/session repository methods.
- [ ] Add `POST /api/auth/login`, `POST /api/auth/logout`, and
  `GET /api/auth/session`.
- [ ] Tests must cover successful login, disabled user rejection, bad password
  rejection, session lookup, logout, owner/admin role guards, and credential
  leakage prevention.
- [ ] Run `GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/auth ./internal/api ./internal/store`.
- [ ] Commit with `feat: add console auth sessions`.

## Task 3: Governance Read Model API

**Files:**
- Create: `internal/console/model.go`
- Create: `internal/console/service.go`
- Create: `internal/console/service_test.go`
- Create: `internal/store/governance.go`
- Create: `internal/api/governance.go`
- Modify: `internal/api/router.go`

- [ ] Add organization, team, role, policy, approval, and managed-resource view
  repository methods.
- [ ] Add owner routes for `/api/me`, `/api/packages`, `/api/workspaces`,
  `/api/billing/wallet`, `/api/billing/ledger`, and support ticket basics.
- [ ] Add admin routes for users, organizations, teams, policies, approvals,
  managed resources, runtime, Fabric catalog, Ledger, and audit views.
- [ ] Ensure owner routes scope by session user and organization membership.
- [ ] Ensure admin routes require active admin.
- [ ] Run `GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/console ./internal/api ./internal/store`.
- [ ] Commit with `feat: add governance read model api`.

## Task 4: Managed Workspace Lifecycle Facade

**Files:**
- Modify: `internal/workspace/service.go`
- Create: `internal/workspace/lifecycle_test.go`
- Modify: `internal/fabric/port.go`
- Modify: `internal/fabric/local/fake.go`
- Modify: `internal/ledger/port.go`
- Create: `internal/api/workspace.go`
- Modify: `internal/api/router.go`

- [ ] Implement managed Workspace create/configure/suspend/delete/token actions.
- [ ] Create path must evaluate policy, create approval if needed, freeze managed
  holds when policy allows, call Fabric, write facade state, and record Ledger
  governance receipts.
- [ ] Preserve storage after storage creation if a later stage fails.
- [ ] Clean compute/route on attach/route failures when safe.
- [ ] Add tests for success, policy-blocked, approval-required, insufficient
  balance, Fabric failure after storage, and token reset/delete.
- [ ] Run `GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/workspace ./internal/api`.
- [ ] Commit with `feat: add managed workspace lifecycle facade`.

## Task 5: Ledger Governance Integration

**Files:**
- Modify: `internal/ledger/port.go`
- Modify: `internal/ledger/postgres/wallet.go`
- Create: `internal/ledger/postgres/governance_test.go`
- Create: `internal/api/ledger.go`

- [ ] Add audit event and receipt recording methods.
- [ ] Add managed billing evidence ledger entries and list APIs.
- [ ] Keep Ledger calls scoped to governance receipts and managed billing
  evidence.
- [ ] Add tests for append-only receipts, audit filtering, billing evidence list,
  and owner/admin visibility.
- [ ] Run `GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/ledger/postgres ./internal/api`.
- [ ] Commit with `feat: add ledger governance evidence`.

## Task 6: Console Governance UI

**Files:**
- Modify: `apps/web/src/api/client.ts`
- Modify: `apps/web/src/App.tsx`
- Modify: `apps/web/src/pages/LoginPage.tsx`
- Modify: `apps/web/src/pages/OwnerOverviewPage.tsx`
- Modify: `apps/web/src/pages/AdminOverviewPage.tsx`
- Create: `apps/web/src/pages/WorkspacesPage.tsx`
- Create: `apps/web/src/pages/WorkspaceDetailPage.tsx`
- Create: `apps/web/src/pages/PoliciesPage.tsx`
- Create: `apps/web/src/pages/ApprovalsPage.tsx`
- Create: `apps/web/src/pages/BillingPage.tsx`
- Create: `apps/web/src/pages/SupportPage.tsx`

- [ ] Add typed API clients for auth, governance read model, managed Workspace,
  billing views, support, policies, approvals, runtime, and production readiness.
- [ ] Implement login using `POST /api/auth/login`.
- [ ] Implement owner pages for managed Workspaces, billing, support, and
  readiness.
- [ ] Implement admin pages for users, organizations, policies, approvals,
  resources, Fabric, Ledger, audit, and production readiness.
- [ ] Keep UI dense and operational; no marketing landing page.
- [ ] Run `npm --prefix apps/web run typecheck` and `npm --prefix apps/web run build`.
- [ ] Commit with `feat: add console governance ui`.

## Task 7: Production Readiness And Deploy

**Files:**
- Create: `internal/readiness/production.go`
- Create: `internal/readiness/production_test.go`
- Modify: `cmd/console-api/main.go`
- Create: `Dockerfile`
- Create: `deploy/tke/opl-console.k8s.json`
- Create: `deploy/production-manifest.example.json`
- Create: `.github/workflows/test.yml`
- Create: `.github/workflows/deploy-tke-production.yml`
- Create: `tools/production-verifier.go`
- Modify: `README.md`
- Create: `docs/runtime/production-runbook.md`

- [ ] Implement fail-closed production readiness for DB, env, domains, TKE,
  image, storage class, ingress class, secrets, and default credential rejection.
- [ ] Add Dockerfile for Go API plus built React assets.
- [ ] Add TKE manifest with no inline secrets, readiness/liveness probes, and
  explicit Console/Workspace domains.
- [ ] Add GitHub Actions test workflow.
- [ ] Add production verifier that checks readiness and exercises a managed
  Workspace lifecycle only when explicitly run by an operator.
- [ ] Run full verification:

```bash
GOTOOLCHAIN=go1.23.12 go test -count=1 ./...
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
git diff --check
```

- [ ] Commit with `feat: add production readiness and deploy`.

## Completion Gate

- [ ] All tasks above are committed.
- [ ] Full verification passes.
- [ ] `git status --short --branch` is clean.
- [ ] Push `HEAD:main` to `origin`.
- [ ] Update the active goal status.
