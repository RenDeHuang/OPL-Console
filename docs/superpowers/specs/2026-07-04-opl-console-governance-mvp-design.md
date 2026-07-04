# OPL Console Governance MVP Design

Date: 2026-07-04

## Goal

Build the first usable OPL Console governance surface in the independent
`OPL-Console` repository.

The fixed stack remains:

```text
Frontend: React + TypeScript
Backend: Go
DB: PostgreSQL
K8s: Go client-go
```

## Boundary

OPL Console is a governance surface. It owns organization, account, team, role,
policy, approval, managed-resource visibility, support, operational readiness,
and lifecycle controls for OPL Cloud-hosted or organization-managed resources.

OPL Console does not own:

- OPL Workspace workbench runtime;
- OPL Fabric platform truth or unmanaged resource execution;
- OPL Ledger platform truth or all evidence generation;
- OPL Gateway usage generation;
- MAS/domain judgment or App-local flows.

For v1, Fabric and Ledger adapters can remain in-process inside the Go backend.
They must still be consumed through service-shaped ports so they can later move
behind APIs without changing Console product ownership.

## MVP Flow

1. Admin or Lab Owner signs in to Console.
2. Console loads session, organization, role, and managed-resource visibility.
3. Owner requests a managed Workspace lifecycle action.
4. Console evaluates policy and approval state.
5. Console writes governance audit/receipt records through Ledger.
6. Console calls Fabric for managed compute, storage, attachment, route, or
   runtime-readiness operations.
7. Console records lifecycle state and exposes a managed Workspace URL only for
   OPL Cloud-hosted or organization-managed Workspaces.
8. Console UI shows governance status, not the Workspace workbench itself.

## Backend Modules

```text
internal/api              HTTP routes, DTOs, auth context, error mapping
internal/auth             password hashing, sessions, CSRF, role checks
internal/console          governance services and read model
internal/workspace        managed Workspace lifecycle facade
internal/fabric           Fabric port and client-go adapter
internal/ledger           Ledger port and PostgreSQL adapter
internal/store            PostgreSQL repositories and transactions
internal/readiness        runtime and production readiness checks
```

## Data Model Additions

Existing tables are reused where possible. The MVP adds governance-specific
objects:

- `teams`
- `roles`
- `policies`
- `approvals`
- `managed_resource_views`

Existing resource tables remain facade state for managed resources:

- `billing_accounts`
- `workspace_packages`
- `compute_resources`
- `storage_volumes`
- `storage_attachments`
- `workspaces`
- `workspace_tokens`
- `wallet_holds`
- `billing_ledger_entries`
- `runtime_operations`
- `audit_events`
- `receipts`

## API Target

Public:

```text
GET  /api/healthz
GET  /api/runtime/readiness
GET  /api/production/readiness
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/session
```

Owner governance:

```text
GET  /api/me
GET  /api/packages
GET  /api/workspaces
POST /api/workspaces
GET  /api/workspaces/{id}
POST /api/workspaces/{id}/stop-compute
POST /api/workspaces/{id}/restart-compute
POST /api/workspaces/{id}/destroy-compute
POST /api/workspaces/{id}/create-backup
POST /api/workspaces/{id}/restore-backup
POST /api/workspaces/{id}/destroy-storage
POST /api/workspaces/{id}/tokens/reset
POST /api/workspaces/{id}/tokens/delete
GET  /api/billing/wallet
POST /api/billing/workspace-quote
GET  /api/billing/ledger
GET  /api/support/tickets
POST /api/support/tickets
```

Admin governance:

```text
GET  /api/admin/users
GET  /api/admin/organizations
GET  /api/admin/teams
GET  /api/admin/policies
POST /api/admin/policies
GET  /api/admin/approvals
POST /api/admin/approvals/{id}/approve
POST /api/admin/approvals/{id}/reject
GET  /api/admin/workspaces
GET  /api/admin/resources
GET  /api/admin/runtime
GET  /api/admin/fabric/catalog
GET  /api/admin/ledger
GET  /api/admin/audit
POST /api/admin/users/{id}/topups
GET  /api/admin/support/tickets
```

Managed Workspace entry:

```text
GET /w/{workspaceId}?token=...
```

The `/w` route applies only to managed Workspaces.

## UI Target

The Console UI should be operational rather than marketing-led:

- login;
- owner overview;
- managed Workspace list/detail/create;
- policy and approval state;
- managed billing visibility;
- support tickets;
- admin users/orgs/teams/policies/approvals;
- admin runtime, Fabric, Ledger, audit, and production readiness.

## Production Target

`GET /api/production/readiness` must be fail-closed and check:

- PostgreSQL connectivity and migration state;
- production URL/domain settings;
- workspace domain settings;
- TKE namespace;
- storage class;
- ingress class;
- workspace image;
- kubeconfig or in-cluster client configuration;
- required secret references;
- default credential rejection;
- deployment readiness inputs.

Deployment adds:

- Dockerfile;
- Kubernetes/TKE manifests;
- GitHub Actions build/test workflow;
- production manifest example;
- operator production verifier.

## Success Criteria

- Console boundary docs and contracts stay aligned with `docs/boundaries.md`.
- Governance API supports real login/session and read model.
- Managed Workspace create path can call Fabric and Ledger ports without
  treating Workspace/Fabric/Ledger as Console-owned domains.
- UI can exercise the governance API.
- Production readiness fails closed until required production inputs exist.
- Full local verification passes:

```bash
GOTOOLCHAIN=go1.23.12 go test -count=1 ./...
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
git diff --check
```
