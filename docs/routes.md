# OPL Console Routes

Routes are fixed by `contracts/opl-console-route-api-contract.json`.

Public routes expose health, readiness, and auth. Lab Owner routes expose
governance views for managed workspaces, managed billing visibility, and support
workflows. Admin routes expose users, readiness, policy, Fabric status for
managed resources, Ledger views for governed receipts, audit, and support queue.

Route ownership does not imply product ownership of Workspace, Fabric, Ledger,
Gateway, or billing usage generation. Console routes are facades for resources
hosted by OPL Cloud or managed by an organization.

The current Go API implementation exposes only:

- `GET /api/healthz`
- `GET /api/runtime/readiness`
- `GET /api/production/readiness`

Owner, admin, auth, managed workspace, managed billing, and support API routes
are the contract baseline for upcoming implementation tasks.

Workspace routes expose `GET /w/{workspaceId}?token=...` in the contract
baseline only for OPL Cloud-hosted or organization-managed Workspaces. In TKE,
ingress for managed `/w/{workspaceId}` routes to the Console validator service,
which validates policy and token state before handing off to the Workspace
runtime instead of sending ingress traffic directly to the workspace compute
service.
