# OPL Console Routes

Routes are fixed by `contracts/opl-console-route-api-contract.json`.

Public routes expose health, readiness, and auth. Lab Owner routes expose
governance views for managed workspaces, managed billing visibility, and support
workflows. Admin routes expose users, readiness, policy, Fabric status for
managed resources, Ledger views for governed receipts, audit, and support queue.

Route ownership does not imply product ownership of Workspace, Fabric, Ledger,
Gateway, or billing usage generation. Console routes are facades for resources
hosted by OPL Cloud or managed by an organization.

The Go API implementation exposes the current Console contract:

- `GET /api/healthz`
- `GET /api/runtime/readiness`
- `GET /api/production/readiness`
- auth session routes;
- owner routes for packages, workspaces, Workspace detail, Workspace quote,
  wallet, owner ledger, and support tickets;
- explicit Workspace lifecycle action routes for stop/restart/destroy compute,
  backup/restore/destroy storage, reset/delete token;
- admin routes for users, organizations, teams, roles, policies, approvals,
  managed resources, raw Ledger view, support queue, and readiness.

The old generic `configure`, `suspend`, and `delete` Workspace routes are not
part of the boundary. Compute and storage have separate lifecycle actions.
Storage destruction requires explicit confirmation.

Workspace routes expose `GET /w/{workspaceId}?token=...` in the contract
baseline only for OPL Cloud-hosted or organization-managed Workspaces. In TKE,
ingress for managed `/w/{workspaceId}` routes to the Console validator service,
which validates policy and token state before handing off to the Workspace
runtime instead of sending ingress traffic directly to the workspace compute
service.
