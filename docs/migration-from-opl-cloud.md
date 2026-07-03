# Migration From OPL-Cloud

OPL Console keeps the OPL Console and OPL Workspace control-plane product rules from OPL-Cloud.

Kept: product names, Basic and Pro CPU packages, seven-day holds, Workspace URL token lifecycle, owner/admin surface separation, billing ledger, audit, receipts, readiness, and contracts.

Rewritten: Node API to Go, React JS to React + TypeScript, state-object persistence to PostgreSQL relational repositories, `kubectl` shell provider to `client-go`, and JS tests to Go/TypeScript tests.

Discarded for v1: JSON runtime store, JS compatibility aliases, Local Docker as product runtime, and standalone Fabric/Ledger deployment.

The Fabric port now includes cleanup and workspace token methods in addition to create operations. TKE ingress routes `/w/{workspaceId}` to the Console validator service rather than directly to workspace compute. Workspace create preserves storage after storage has been created and a later provisioning step fails.
