# Migration From OPL-Cloud

OPL Console keeps the governance-surface rules from OPL-Cloud and
`one-person-lab-cloud`.

Kept: product names, Console owner/admin separation, managed Workspace lifecycle
controls, managed billing and usage visibility, policy and approval concepts,
audit, receipts, readiness, and contracts.

Rewritten: Node API to Go, React JS to React + TypeScript, state-object
persistence to PostgreSQL relational repositories, `kubectl` shell provider to
`client-go`, and JS tests to Go/TypeScript tests.

Reframed: billing, Workspace, Fabric, and Ledger are not Console-owned product
domains. Console presents billing visibility and lifecycle controls only for OPL
Cloud-hosted or organization-managed usage. Workspace remains the workbench,
Fabric remains the resource platform, and Ledger remains the evidence platform.

Discarded for v1: JSON runtime store, JS compatibility aliases, Local Docker as
product runtime, and standalone Fabric/Ledger deployment.

The Fabric port now includes cleanup and workspace token methods in addition to
create operations. TKE ingress routes managed `/w/{workspaceId}` traffic to the
Console validator service rather than directly to workspace compute. Managed
Workspace create preserves storage after storage has been created and a later
provisioning step fails.
