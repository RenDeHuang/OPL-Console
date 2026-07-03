# OPL Console Split Design

Date: 2026-07-03

## Goal

Split the OPL Console control-plane slice out of `OPL-Cloud` into an independent repository.

The independent repository is **OPL Console**, not the full OPL Cloud product. It owns the console, workspace control plane, wallet/billing surface, support/admin surface, and the minimum internal Fabric/Ledger implementation required for a controlled commercial pilot.

The fixed technology stack is:

```text
Frontend: React + TypeScript
Backend: Go
DB: PostgreSQL
K8s: Go client-go
```

## Source Context

The current `OPL-Cloud` repository is a JavaScript prototype that mixes:

- OPL Console UI
- Node HTTP API
- JSON/PostgreSQL state store
- Workspace lifecycle orchestration
- Local Docker provider
- Tencent TKE provider through `kubectl`
- Wallet, holds, billing ledger, audit, receipts
- Production readiness checks
- Machine-readable contracts

The split keeps product truth and contracts, but rewrites the implementation in the fixed stack above.

The `one-person-lab` framework is a development and lifecycle constraint source. OPL Console does not reimplement the framework internals, but it must follow its stage-led operation, readiness, receipt/evidence, human gate, recovery, audit/replay, and contract-light rules.

References:

- `OPL-Cloud`: https://github.com/RenDeHuang/OPL-Cloud
- `one-person-lab`: https://github.com/gaofeng21cn/one-person-lab

## Repository Boundary

OPL Console owns:

- Lab Owner and Admin login/session flows.
- Organizations, users, memberships, roles, and ownership.
- Workspace package listing and Workspace lifecycle controls.
- Compute, storage, attachment, Workspace URL, and token control-plane state.
- Wallet, seven-day holds, billing ledger, manual top-up, audit, and receipts.
- Support tickets, notifications, runtime readiness, and production readiness.
- Internal Fabric module for TKE provisioning through `client-go`.
- Internal Ledger module for wallet and billing records through PostgreSQL.
- Contracts that fix Console routes, business objects, lifecycle rules, and future Fabric/Ledger API boundaries.

OPL Console does not own:

- `one-person-lab-app` WebUI behavior.
- OPL Gateway internals.
- Standalone OPL Fabric service deployment.
- Standalone OPL Ledger service deployment.
- GPU Workspace productization.
- Connector, environment, and agent marketplaces.
- External payment settlement.

## Architecture Decision

Use a modular monolith.

One Go backend process contains Console, Workspace, Fabric, Ledger, Billing, Auth, Store, API, and Contracts modules. Fabric and Ledger are internal modules for v1, but their ports must be shaped as future external API boundaries.

```text
React UI
  -> Go console-api
    -> internal console/workspace service
      -> internal fabric.Port
        -> TKE client-go implementation
      -> internal ledger.Port
        -> PostgreSQL implementation
      -> PostgreSQL repositories
```

This keeps the commercial pilot deployable while avoiding a hard dependency on future standalone Fabric/Ledger services.

## Proposed Repository Structure

```text
opl-console/
  apps/
    web/                  # React + TypeScript frontend
  cmd/
    console-api/          # Go HTTP API entrypoint
    migrate/              # DB migration entrypoint
  internal/
    auth/                 # sessions, CSRF, role guard
    console/              # Console application service
    workspace/            # Workspace lifecycle orchestration
    fabric/               # future OPL Fabric API shape
      port.go             # Fabric interface
      tke/                # client-go TKE implementation
      local/              # local/dev fake implementation
    ledger/               # future OPL Ledger API shape
      port.go             # Ledger interface
      postgres/           # internal ledger implementation
    billing/              # pricing, hold, settlement policy
    store/                # PostgreSQL repository layer
    api/                  # route handlers and DTOs
    config/               # env config
    contracts/            # contract loading and validation
  contracts/              # migrated and rewritten OPL Console contracts
  migrations/             # SQL migrations
  deploy/
    k8s/                  # Console deployment manifests
  docs/
    architecture.md
    routes.md
    migration-from-opl-cloud.md
```

## Module Rules

- `api` only handles HTTP, auth context, request/response DTOs, and error mapping.
- `workspace` orchestrates lifecycle operations and owns recovery/compensation logic.
- `billing` owns pricing, seven-day hold rules, and billing policy.
- `ledger.Port` owns wallet and ledger operations. Current implementation writes PostgreSQL; future implementation may call OPL Ledger over API.
- `fabric.Port` owns compute, storage, attachment, route, token secret, and readiness operations. Current implementation uses `client-go`; future implementation may call OPL Fabric over API.
- `store` owns PostgreSQL persistence and transaction boundaries.
- Cross-module calls use interfaces, not concrete package imports.
- Runtime state is PostgreSQL only. No JSON file store in the new repository.
- Node API, JS store, `kubectl` shell provider, and JS tests are not compatibility layers. They are source material only.

## one-person-lab Framework Mapping

### Goal / Attempt / Stage

Every high-risk or multi-step action is a lifecycle operation with:

```text
operation_id
operation_type
actor_user_id
account_id
workspace_id
stage
status
started_at
finished_at
error_code
error_message
request_id
```

Workspace creation stages:

```text
validate_package
check_wallet
freeze_hold
create_storage
create_compute
attach_storage
create_workspace_route
mark_ready
```

### Readiness

Readiness is layered:

```text
GET /api/healthz
```

Checks whether the Go process can answer.

```text
GET /api/runtime/readiness
```

Checks PostgreSQL, K8s API connectivity, namespace, storage class, workspace image configuration, ingress class, and required runtime permissions.

```text
GET /api/production/readiness
```

Checks production secret/env presence, domain configuration, image references, RBAC, database migration status, and deployment readiness.

### Receipt / Evidence

The backend records receipts/audit evidence for:

- wallet top-up
- hold freeze, release, and debit
- compute create, stop, restart, destroy
- storage create, attach, detach, destroy
- Workspace URL create, reset, delete
- support ticket creation and admin response
- admin user and wallet operations
- runtime readiness and production readiness checks

### Human Gate

These actions require explicit confirmation input:

- destroy compute
- destroy storage
- reset Workspace token
- delete Workspace token
- restore backup to a new Workspace
- admin manual top-up

### Recovery

Workspace creation is not one opaque transaction. It is a resumable operation with stage-specific compensation.

Examples:

- Failure before hold freeze: record failed operation only.
- Failure after hold freeze and before runtime creation: release hold.
- Failure after storage creation but before compute creation: keep or destroy storage according to operation policy and record a receipt.
- Failure after compute creation: cleanup compute/route when safe, preserve storage when product policy says data must be retained.

### Contract-Light

Contracts preserve product and safety boundaries, not all implementation details.

Required contracts:

- route API contract
- business object contract
- workspace lifecycle contract
- billing ledger contract
- Fabric API boundary contract
- Ledger API boundary contract
- package boundary contract

### Docs Lifecycle

Docs describe current truth only. Migration docs must state owner, status, and removal condition.

## PostgreSQL Model

Core tables:

```text
users
organizations
memberships
sessions

billing_accounts
wallet_transactions
wallet_holds
billing_ledger_entries
manual_topups

workspace_packages
compute_resources
storage_volumes
storage_attachments
workspaces
workspace_tokens
runtime_operations

support_tickets
notifications
audit_events
receipts
```

Persistence rules:

- Store money as integer minor units: CNY fen in `BIGINT`. Never use floating point for money.
- Store Workspace token hashes, not plaintext tokens.
- Use PostgreSQL transactions for wallet/hold/ledger mutations.
- Every user-visible lifecycle mutation writes an audit event.
- `runtime_operations` is the recovery entrypoint.
- `receipts` stores replayable evidence for framework-style review.

## API Surface

### Public

```text
GET  /api/healthz
GET  /api/runtime/readiness
GET  /api/production/readiness
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/session
```

### Lab Owner

```text
GET  /api/me
GET  /api/packages
GET  /api/workspaces
POST /api/workspaces
GET  /api/workspaces/{id}
POST /api/workspaces/{id}/stop
POST /api/workspaces/{id}/restart
POST /api/workspaces/{id}/destroy-compute
POST /api/workspaces/{id}/destroy-storage
POST /api/workspaces/{id}/tokens/reset
POST /api/workspaces/{id}/tokens/delete

GET  /api/billing/wallet
GET  /api/billing/ledger
GET  /api/support/tickets
POST /api/support/tickets
```

### Admin

```text
GET  /api/admin/users
GET  /api/admin/workspaces
GET  /api/admin/runtime
GET  /api/admin/fabric/catalog
GET  /api/admin/ledger
GET  /api/admin/audit
POST /api/admin/users/{id}/topups
GET  /api/admin/support/tickets
```

### Workspace URL

```text
GET /w/{workspaceId}?token=...
```

This validates the token hash and routes to the Workspace runtime when ready.

## Frontend Surface

### Public

- Login

### Lab Owner

- Overview
- Workspaces
- Create Workspace
- Workspace Detail
- Billing Wallet
- Support
- Account

### Admin

- Admin Overview
- Users
- User Wallet
- Runtime Readiness
- Fabric Catalog
- Ledger
- Audit
- Support Queue

Lab Owner UI must not expose raw K8s evidence, request fingerprints, dedup rows, raw Ledger internals, or production readiness internals.

Admin UI may expose Fabric/Ledger internals because those are operator surfaces.

## Fabric Module

`internal/fabric/port.go` defines the stable internal boundary that can later become OPL Fabric API calls.

Required capabilities:

- list catalog/readiness
- create compute
- stop/restart/destroy compute
- create storage
- attach/detach storage
- destroy storage
- create Workspace route/token secret
- reset/delete Workspace route token
- inspect runtime status

The TKE implementation uses `client-go` directly, not shelling out to `kubectl`.

TKE resources:

- Deployment for compute runtime
- Service for runtime traffic
- PVC for persistent storage
- Secret for Workspace token/config
- `networking.k8s.io/v1` Ingress for `/w/{workspaceId}`, with ingress class configured by environment

## Ledger Module

`internal/ledger/port.go` defines the stable internal boundary that can later become OPL Ledger API calls.

Required capabilities:

- get wallet state
- freeze hold
- release hold
- debit hold
- record wallet transaction
- record manual top-up
- record billing ledger entry
- record audit event
- record receipt
- list account ledger
- list admin ledger

The v1 implementation writes PostgreSQL in the same database as Console.

## Migration From OPL-Cloud

Keep:

- product names and ownership definitions
- Workspace lifecycle product rules
- Basic and Pro CPU packages
- seven-day compute/storage hold rules
- permanent Workspace URL token until reset/delete
- Lab Owner vs Admin surface separation
- billing ledger, audit, receipt, and readiness concepts
- machine-readable contracts, rewritten for OPL Console

Rewrite:

- Node HTTP API to Go API
- React JS frontend to React + TypeScript
- JSON/PostgreSQL state object store to relational PostgreSQL repositories
- `kubectl` shell TKE provider to `client-go`
- JS service classes to Go packages with ports and repositories
- JS tests to Go tests and frontend TypeScript tests

Discard:

- JSON file store as runtime persistence
- compatibility route aliases from the JS prototype
- Local Docker as a product runtime provider
- external Fabric/Ledger service deployment in v1

## First Implementation Task Split

1. Repository scaffold
   - Create Go module, React + TypeScript app, migration runner, dev compose, and CI verification commands.

2. Contracts migration
   - Copy relevant `OPL-Cloud/packages/contracts` content.
   - Rewrite as current Console truth.
   - Add Fabric and Ledger API boundary contracts.

3. Auth and RBAC
   - Implement users, sessions, roles, CSRF, owner/admin guards, and seed admin path.

4. Billing and internal Ledger
   - Implement wallet, holds, manual top-up, ledger entries, audit events, and receipts.

5. Workspace domain
   - Implement packages, compute resources, storage volumes, attachments, workspaces, tokens, and runtime operations.

6. Fabric TKE adapter
   - Implement `client-go` creation and reconciliation for Deployment, Service, PVC, Secret, and route.
   - Add local fake adapter for tests.

7. Console API
   - Implement route handlers, DTOs, error codes, request id, and audit middleware.

8. React Console
   - Implement Lab Owner path first: login, create Workspace, show URL, billing wallet, support.
   - Add Admin surfaces after the owner path is stable.

9. Readiness and production contract
   - Implement health, runtime readiness, production readiness, migration status, and K8s permission checks.

10. Migration documentation
    - Document what moved from OPL-Cloud, what was rewritten, what was discarded, and what remains future service work.

## Task Ownership

This work belongs to:

```text
OPL Console / OPL Workspace control-plane slice
```

It does not belong to:

- OPL Gateway
- `one-person-lab-app`
- standalone OPL Fabric service
- standalone OPL Ledger service

Fabric and Ledger are included only as internal Console backend modules for v1. Their route/API boundaries are fixed now so they can be replaced by standalone services later.

## Fixed Implementation Choices

Use these choices for the initial implementation plan:

- Go HTTP router: `github.com/go-chi/chi/v5`.
- PostgreSQL driver: `github.com/jackc/pgx/v5`.
- Migration library: `github.com/pressly/goose/v3`.
- Frontend routing: `react-router`.
- Frontend server-state/data fetching: `@tanstack/react-query`.
- Money representation: CNY fen in signed `BIGINT`.
- TKE route implementation: Kubernetes `networking.k8s.io/v1` Ingress with env-configured ingress class.
- Workspace creation partial failure policy: after storage is successfully created, preserve storage by default and record a recovery receipt; cleanup compute and route when they were created during the failed operation.
