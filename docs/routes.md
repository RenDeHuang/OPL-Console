# OPL Console Routes

Routes are fixed by `contracts/opl-console-route-api-contract.json`.

The source truth is the Console slice in `OPL-Cloud`:

- `packages/console/api/routes/index.js`
- `packages/contracts/opl-cloud-route-api-contract.json`

`OPL-Console` is the Go/React/PostgreSQL/client-go rewrite of that Console
surface. It must not expose compatibility routes that are not in the OPL-Cloud
Console manifest.

## Boundary

- Console owns auth, session, state, management state, support entry, resource
  command routes, Workspace command facade routes, readiness, and operator
  summary.
- Fabric owns compute, storage, attachment, route, runtime, and backup
  execution behind Console routes.
- Ledger owns top-up, usage, reconciliation, and task receipt evidence behind
  Console routes.
- Workspace workbench behavior is outside Console.

## API Manifest

```text
GET  /api/healthz
POST /api/auth/login
POST /api/auth/operator-login
POST /api/auth/logout
GET  /api/auth/me
GET  /api/state
GET  /api/operator/summary
GET  /api/management/state
POST /api/billing/topups
POST /api/organizations
POST /api/users
POST /api/organizations/members
POST /api/compute-resources
POST /api/compute-resources/destroy
POST /api/storage-volumes
POST /api/storage-volumes/destroy
POST /api/storage-attachments
POST /api/storage-attachments/detach
POST /api/workspaces
POST /api/workspaces/storage-backups
POST /api/workspaces/restore-storage-backup
POST /api/workspaces/prune-storage-backups
POST /api/workspaces/reset-token
POST /api/workspaces/delete-token
POST /api/billing/request-usage
POST /api/billing/reconciliation
GET  /api/ledger/task-receipts
POST /api/ledger/task-receipts
GET  /api/runtime/readiness
GET  /api/production/readiness
POST /api/workspaces/runtime-status
GET  /api/support/tickets
POST /api/support/tickets
```
