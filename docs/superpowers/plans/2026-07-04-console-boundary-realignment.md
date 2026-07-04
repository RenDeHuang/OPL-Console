# OPL Console Boundary Realignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 OPL Console 边界重新对齐到 `one-person-lab-cloud` 的 Console 职责：Lab Owner 商业控制面 + Admin 治理运维面，而不是 Workspace workbench、Fabric 平台真相或 Ledger 平台真相。

**Architecture:** Console 保留用户、组织、成员、策略、审批、钱包视图、工作空间生命周期控制、URL 访问控制、支持工单、审计和 readiness。Fabric 负责真实 compute/storage/attachment/route/backup/restore/runtime status 执行；Ledger 负责 wallet、hold、debit、topup、reconcile、receipt/evidence 真相；Workspace 负责运行时 workbench。Console 通过端口/API 编排和展示，不吞掉平台职责。

**Tech Stack:** React + TypeScript frontend, Go backend, PostgreSQL, Kubernetes `client-go`, future HTTP clients for OPL Fabric and OPL Ledger.

---

## Reference Alignment

Use these source files as the boundary baseline:

- `/home/dev/medopl-3/docs/product/console-workspace-v1.md`
- `/home/dev/medopl-3/packages/console/api/routes/index.js`
- `/home/dev/medopl-3/packages/contracts/opl-cloud-workspace-lifecycle-contract.json`
- `/home/dev/medopl-3/packages/contracts/opl-cloud-billing-ledger-contract.json`
- `/home/dev/medopl-3/packages/contracts/opl-cloud-evidence-ledger-contract.json`
- `/home/dev/medopl-3/packages/fabric/src/resource-catalog.js`
- `/home/dev/medopl-3/tools/production-verifier.js`

Current OPL Console files to realign:

- Modify: `docs/boundaries.md`
- Modify: `docs/architecture.md`
- Modify: `docs/routes.md`
- Modify: `contracts/opl-console-route-api-contract.json`
- Modify: `internal/workspace/port.go`
- Modify: `internal/fabric/port.go`
- Modify: `internal/ledger/port.go`
- Modify: `internal/workspace/service.go`
- Modify: `internal/api/routes.go`
- Modify: `apps/web/src/pages/WorkspacesPage.tsx`
- Create: `apps/web/src/pages/WorkspaceDetailPage.tsx`
- Modify: `apps/web/src/pages/BillingPage.tsx`
- Modify: `apps/web/src/pages/SupportPage.tsx`
- Modify: `apps/web/src/pages/ApprovalsPage.tsx`
- Modify: `apps/web/src/pages/AdminOverviewPage.tsx`

---

### Task 1: Correct Product Boundary Docs

**Boundary correction:** Console is not only a governance shell. It owns the Lab Owner control path:

`sign in -> create Workspace -> confirm cost and hold -> copy URL -> share URL with members`

It also owns Admin views for readiness, Fabric catalog internals, Ledger evidence, support queue, users, wallets, manual top-ups, governance policies, and audit.

- [ ] Update `docs/boundaries.md` to split Console into two surfaces:
  - Lab Owner Console: overview, workspaces, create workspace, URL access, billing wallet, receipts, account/lab, support, alerts.
  - Admin Console: users, roles, wallet top-ups, policies, audit, runtime readiness, Fabric catalog internals, Ledger events/receipts, support queue.
- [ ] Add explicit non-goals:
  - Workspace task/file/artifact workbench.
  - Fabric resource execution truth.
  - Ledger billing/evidence truth.
  - Gateway usage generation.
  - MAS/domain quality judgment.
- [ ] Run:
  ```bash
  rg -n "OPL Console|OPL Fabric|OPL Ledger|OPL Workspace|Billing" docs/boundaries.md
  ```
- [ ] Commit:
  ```bash
  git add docs/boundaries.md docs/architecture.md docs/routes.md
  git commit -m "docs: realign console product boundary"
  ```

### Task 2: Replace Destructive Workspace Boundary

**Boundary correction:** `delete workspace` must not mean destroy compute + route + storage. OPL-Cloud contract says:

- `server_destroy_does_not_destroy_disk`
- `disk_destroy_requires_explicit_confirmation`
- `disk_destroy_is_the_only_action_that_stops_storage_billing`
- `stable_url_should_survive_restart_or_server_recreation_when_possible`

- [ ] Update `contracts/opl-console-route-api-contract.json` owner routes:
  - Keep `POST /api/workspaces`.
  - Add `GET /api/workspaces/{id}`.
  - Replace generic delete with:
    - `POST /api/workspaces/{id}/stop-compute`
    - `POST /api/workspaces/{id}/restart-compute`
    - `POST /api/workspaces/{id}/destroy-compute`
    - `POST /api/workspaces/{id}/create-backup`
    - `POST /api/workspaces/{id}/restore-backup`
    - `POST /api/workspaces/{id}/destroy-storage`
    - `POST /api/workspaces/{id}/tokens/reset`
    - `POST /api/workspaces/{id}/tokens/delete`
- [ ] Update `internal/workspace/port.go` with service methods matching those actions.
- [ ] Update `internal/api/routes.go` to expose the action routes.
- [ ] Add tests in `internal/api/workspace_test.go` proving generic delete is not the only lifecycle route.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/api ./internal/workspace
  ```
- [ ] Commit:
  ```bash
  git add contracts/opl-console-route-api-contract.json internal/workspace internal/api
  git commit -m "feat: split workspace lifecycle actions"
  ```

### Task 3: Add Workspace Lifecycle State Machine

**Boundary correction:** Console owns lifecycle visibility and policy gating, not raw workbench. A Workspace must expose desired state, actual Fabric state, billing state, token state, and last error.

- [ ] Add lifecycle states in `internal/workspace/model.go`:
  - `draft`
  - `freezing_storage_balance`
  - `creating_server`
  - `creating_disk`
  - `deploying_runtime`
  - `configuring_url`
  - `running`
  - `stopping_server`
  - `stopped_server_disk_retained`
  - `restarting_server`
  - `destroying_server`
  - `server_destroyed_disk_retained`
  - `creating_storage_backup`
  - `restoring_storage_backup`
  - `destroying_disk`
  - `destroyed`
  - `failed`
  - `cleanup_required`
- [ ] Store per-step status in PostgreSQL using a migration:
  - workspace ID
  - step name
  - desired state
  - actual state
  - provider resource ID
  - error code
  - last checked time
- [ ] Add tests proving:
  - compute destroy retains storage.
  - storage destroy requires explicit confirmation.
  - failed compute create after storage does not mark storage destroyed.
- [ ] Run:
  ```bash
  TEST_DATABASE_URL=postgres://opl:secret@127.0.0.1:5432/opl_console?sslmode=disable GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/workspace ./internal/store
  ```
- [ ] Commit:
  ```bash
  git add migrations internal/workspace internal/store
  git commit -m "feat: add workspace lifecycle state machine"
  ```

### Task 4: Make Fabric a Contracted Port, Not Console Logic

**Boundary correction:** Console asks Fabric to create/destroy/inspect resources; Fabric owns the execution semantics.

- [ ] Expand `internal/fabric/port.go` to include:
  - quote/read catalog
  - create compute
  - stop/restart/destroy compute
  - create storage
  - destroy storage
  - attach/detach storage
  - create route
  - reset/delete route token handoff
  - create/prune/restore storage backup
  - runtime status
  - runtime readiness
- [ ] Keep local/TKE adapters behind the same interface.
- [ ] Add an HTTP client implementation placeholder behind env:
  - `OPL_FABRIC_URL`
  - `OPL_FABRIC_TOKEN`
- [ ] Update production readiness to check Fabric URL/version when external mode is configured.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/fabric ./internal/readiness
  ```
- [ ] Commit:
  ```bash
  git add internal/fabric internal/readiness internal/config
  git commit -m "feat: define fabric integration boundary"
  ```

### Task 5: Make Ledger the Pricing, Hold, Debit, Receipt Truth

**Boundary correction:** Console may show quote/hold/debit/receipt, but Ledger owns the money and evidence truth.

- [ ] Expand `internal/ledger/port.go` to include:
  - quote workspace package
  - get wallet
  - freeze compute hold
  - freeze storage hold
  - release hold
  - debit usage
  - manual top-up
  - billing ledger query
  - reconciliation status
  - governance receipt record/query
  - task evidence receipt record/query
- [ ] Add an owner route:
  - `POST /api/billing/workspace-quote`
- [ ] Update create workspace flow to call Ledger quote/hold before Fabric provisioning.
- [ ] Add tests proving UI-visible price comes from backend quote, not frontend arithmetic.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/ledger ./internal/workspace ./internal/api
  ```
- [ ] Commit:
  ```bash
  git add internal/ledger internal/workspace internal/api apps/web/src
  git commit -m "feat: make ledger quote and hold authoritative"
  ```

### Task 6: Add Workspace Detail UI

**Boundary correction:** A production Console cannot only show a list. It must show real control-plane state.

- [ ] Create `apps/web/src/pages/WorkspaceDetailPage.tsx`.
- [ ] Add route `/workspaces/:id`.
- [ ] Show:
  - Workspace name and state.
  - URL copy/open/reset/delete token.
  - compute state and billing state.
  - storage state, size, mount, billing state.
  - attachment state.
  - route state.
  - package and seven-day hold.
  - latest Fabric runtime status.
  - latest Ledger receipts.
  - latest support tickets.
  - audit trail.
- [ ] Replace list-row destructive delete with detail-page guarded actions.
- [ ] Run:
  ```bash
  npm --prefix apps/web run typecheck
  npm --prefix apps/web run build
  ```
- [ ] Commit:
  ```bash
  git add apps/web/src
  git commit -m "feat: add workspace detail control surface"
  ```

### Task 7: Strengthen Approvals With Context

**Boundary correction:** Approval is not a yes/no queue. It must show cost, quota, policy, resource impact, and post-approval actions.

- [ ] Extend approval read model to include:
  - requester
  - organization/team
  - requested package
  - quote
  - balance sufficiency
  - policy rule triggered
  - actions that will run after approval
- [ ] Update `apps/web/src/pages/ApprovalsPage.tsx` to show this context.
- [ ] Add backend tests for approval context serialization.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/console ./internal/api
  npm --prefix apps/web run typecheck
  ```
- [ ] Commit:
  ```bash
  git add internal/console internal/api apps/web/src/pages/ApprovalsPage.tsx
  git commit -m "feat: add approval decision context"
  ```

### Task 8: Split Owner Ledger View From Admin Ledger View

**Boundary correction:** Lab Owner sees human-readable wallet, holds, charges, receipts. Admin sees raw ledger events, reconciliation, evidence receipts, and manual top-up audit.

- [ ] Update owner billing page to show:
  - balance
  - frozen
  - available
  - workspace holds
  - recent debits
  - top-ups
  - readable receipts
- [ ] Update admin pages to show:
  - raw Ledger events
  - task evidence receipts
  - reconciliation reports
  - manual top-up audit
- [ ] Add permission tests:
  - owner cannot access raw admin ledger route.
  - admin can access raw ledger route.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/api ./internal/auth
  npm --prefix apps/web run build
  ```
- [ ] Commit:
  ```bash
  git add internal/api apps/web/src/pages
  git commit -m "feat: separate owner and admin ledger views"
  ```

### Task 9: Connect Support Tickets To Workspace Incidents

**Boundary correction:** Support is part of Console because it handles managed resource operations. It must attach operational context instead of being a generic text box.

- [ ] Extend support ticket model with:
  - workspace ID
  - failed lifecycle step
  - Fabric error code
  - runtime status snapshot
  - Ledger hold/debit summary
  - priority
  - assignee
  - comments
  - status transitions
- [ ] Update owner UI to create ticket from Workspace detail.
- [ ] Update admin support queue to inspect and resolve tickets.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/console ./internal/api
  npm --prefix apps/web run typecheck
  ```
- [ ] Commit:
  ```bash
  git add migrations internal/console internal/api apps/web/src/pages
  git commit -m "feat: attach support tickets to workspace incidents"
  ```

### Task 10: Expand Production Readiness Into Dependency Readiness

**Boundary correction:** `/api/production/readiness` must not be a coarse flag. It must prove Console can safely operate against Fabric, Ledger, PostgreSQL, K8s/TKE, DNS, TLS, image registry, storage class, secrets, and ingress.

- [ ] Update readiness output with named checks:
  - database
  - auth seed/default password disabled
  - public URL
  - workspace domain
  - workspace image
  - Kubernetes config
  - ingress class
  - storage class
  - Fabric mode/API/version
  - Ledger mode/API/version
  - TCR/TKE image readiness
  - DNS/TLS
  - secret presence
- [ ] Update Admin UI to show every check, not only ready/false.
- [ ] Update production verifier to create a Workspace, validate URL handoff, verify persistence, then clean up without destroying storage unless explicitly requested.
- [ ] Run:
  ```bash
  GOTOOLCHAIN=go1.23.12 go test -count=1 ./internal/readiness ./internal/api
  npm --prefix apps/web run build
  ```
- [ ] Commit:
  ```bash
  git add internal/readiness tools apps/web/src/pages/AdminOverviewPage.tsx
  git commit -m "feat: expand production readiness checks"
  ```

### Task 11: Final Boundary Verification

- [ ] Verify no UI labels imply Console owns Fabric or Ledger truth.
- [ ] Verify no owner UI exposes raw request fingerprints, dedup rows, raw runtime evidence, production readiness, manual settlement, or raw Ledger events.
- [ ] Verify admin UI exposes Fabric catalog internals and raw Ledger evidence only behind admin auth.
- [ ] Run full verification:
  ```bash
  npm --prefix apps/web run typecheck
  npm --prefix apps/web run build
  TEST_DATABASE_URL=postgres://opl:secret@127.0.0.1:5432/opl_console?sslmode=disable GOTOOLCHAIN=go1.23.12 go test -count=1 ./...
  git diff --check
  ```
- [ ] Commit docs and route contract updates if any final wording changed:
  ```bash
  git add docs contracts
  git commit -m "docs: document console boundary verification"
  ```
