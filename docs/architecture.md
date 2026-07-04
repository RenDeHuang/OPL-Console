# OPL Console Architecture

OPL Console is the independent governance-surface repository for OPL Cloud.

The product boundary follows `one-person-lab-cloud`: Console manages accounts,
organizations, permissions, quota, approvals, billing visibility, policy,
lifecycle controls, audit, and operational status for OPL Cloud-hosted or
organization-managed resources. Console is not the OPL Workspace runtime, not
OPL Fabric itself, and not OPL Ledger itself.

The backend is a Go modular monolith. Fabric and Ledger are internal modules in
v1, but their ports are shaped as future OPL Fabric and OPL Ledger API
boundaries. Keeping these modules in-process is an implementation choice, not a
product ownership claim.

The frontend is React + TypeScript. PostgreSQL is the only runtime persistence
store. Kubernetes provisioning uses Go `client-go`.

Console orchestrates managed Workspace lifecycle only through governance-facing
commands: create, configure, suspend, delete, approve, quota, access policy, and
status visibility. The Workspace workbench remains the OPL Workspace product.

Fabric owns the managed runtime resource boundary for compute, storage,
attachment, route, connector, environment, cleanup, and runtime-readiness
operations. The TKE adapter creates managed compute and storage resources with
`client-go`, routes `/w/{workspaceId}` ingress traffic to the Console validator
service when the Workspace is managed by OPL Cloud, and stores token handoff
state in Kubernetes Secrets.

Ledger owns receipt, provenance, audit, retention, and billing-evidence
semantics. Console records governance receipts and exposes managed billing and
ledger views, but usage signals may originate from Gateway, Fabric, Workspace,
or other approved capability callers.

Managed Workspace creation provisions storage before compute. If a later create
step fails after storage exists, storage is preserved for the retention and hold
lifecycle while compute is cleaned up when applicable.

See `docs/boundaries.md` for the fixed ownership and activation rules.
