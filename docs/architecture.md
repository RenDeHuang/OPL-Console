# OPL Console Architecture

OPL Console is the independent governance-surface repository for OPL Cloud.

The product boundary follows `one-person-lab-cloud`: Console is the Lab Owner
commercial control surface and the Admin governance/operations surface for OPL
Cloud-hosted or organization-managed resources. Console is not the OPL Workspace
runtime, not OPL Fabric itself, and not OPL Ledger itself.

The backend is a Go modular monolith. Fabric and Ledger are internal modules in
v1, but their ports are shaped as future OPL Fabric and OPL Ledger API
boundaries. Keeping these modules in-process is an implementation choice, not a
product ownership claim.

The frontend is React + TypeScript. PostgreSQL is the only runtime persistence
store. Kubernetes provisioning uses Go `client-go`.

Console orchestrates managed Workspace lifecycle only through control-plane
commands: create, stop compute, restart compute, destroy compute, create/restore
storage backup, destroy storage, reset/delete URL token, approve, quota, access
policy, and status visibility. There is no generic Workspace delete action:
destroying compute never destroys storage, and destroying storage requires
explicit confirmation. The Workspace workbench remains the OPL Workspace
product.

Fabric owns the managed runtime resource boundary for compute, storage,
attachment, route, runtime status, backup, restore, connector, environment,
cleanup, and runtime-readiness operations. The TKE adapter creates managed
compute and storage resources with `client-go`, routes `/w/{workspaceId}`
ingress traffic to the Console validator service when the Workspace is managed
by OPL Cloud, and stores token handoff state in Kubernetes Secrets.

Ledger owns quote, wallet, hold, debit, top-up, reconciliation, receipt,
provenance, audit, retention, and billing-evidence semantics. Console asks
Ledger for authoritative quotes and displays owner/admin ledger views, but usage
signals may originate from Gateway, Fabric, Workspace, or other approved
capability callers.

Managed Workspace creation provisions storage before compute. If a later create
step fails after storage exists, storage is preserved for the retention and hold
lifecycle while compute is cleaned up when applicable.

See `docs/boundaries.md` for the fixed ownership and activation rules.
