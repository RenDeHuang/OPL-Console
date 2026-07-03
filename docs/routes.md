# OPL Console Routes

Routes are fixed by `contracts/opl-console-route-api-contract.json`.

Public routes expose health, readiness, and auth. Lab Owner routes expose workspace, billing, and support workflows. Admin routes expose users, readiness, Fabric internals, Ledger internals, audit, and support queue.

The current Go API implementation exposes only:

- `GET /api/healthz`
- `GET /api/runtime/readiness`
- `GET /api/production/readiness`

Owner, admin, auth, workspace, billing, and support API routes are the contract baseline for upcoming implementation tasks.

Workspace routes expose `GET /w/{workspaceId}?token=...` in the contract baseline. In TKE, ingress for `/w/{workspaceId}` routes to the Console validator service, which validates and hands off to the workspace runtime instead of sending ingress traffic directly to the workspace compute service.
