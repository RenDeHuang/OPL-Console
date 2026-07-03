# OPL Console Boundaries

This repository implements the OPL Console governance surface.

The product boundary follows `one-person-lab-cloud`: OPL Console manages accounts,
organizations, permissions, billing visibility, quota, approvals, policy, and
operational status for OPL Cloud-hosted or organization-managed resources.

OPL Console is not the OPL Workspace runtime, not OPL Fabric itself, and not OPL
Ledger itself. Console is also not the only entry point for Fabric or Ledger.
Local OPL App, cloud OPL Workspace, MAS, and approved domain agents may call
Fabric and Ledger capabilities directly when their capability profile allows it.
Console becomes active when a resource, workspace, connector, receipt policy,
quota, skill pack, credential, or billing rule is hosted by OPL Cloud or managed
by an organization.

## Product Ownership

| Area | Product owner | Console responsibility |
| --- | --- | --- |
| OPL App | Local OPL product | No ownership. Console may govern organization-managed cloud capabilities used by App. |
| OPL Workspace | Cloud workbench product | Create, configure, suspend, delete, quota, access policy, and lifecycle visibility for managed Workspaces. Workspace task/file/artifact UX stays in Workspace. |
| OPL Gateway | AI access product | Billing and usage visibility, quota policy, key/provider governance when managed by OPL Cloud. Gateway usage generation stays in Gateway. |
| OPL Fabric | Platform resource layer | Select packages, request managed compute/storage/connectors/environments, approve policy, and show admin status. Fabric owns resource execution. |
| OPL Ledger | Platform evidence layer | Define receipt/audit/retention policy and record Console-visible governance receipts. Ledger owns receipt/provenance storage semantics. |
| Billing | Cross-product commercial layer | Show wallet, usage, quota, holds, invoices, and managed-resource billing state. Usage signals come from Gateway, Fabric, Workspace, and managed policies. |

## In-Process Module Rule

For v1, Fabric and Ledger adapters may live inside this Go backend as internal
modules. That is an implementation shortcut, not a product ownership claim.

Internal modules must keep service-shaped ports:

- Console calls Fabric through a Fabric port for managed compute, storage,
  attachment, route, connector, environment, and runtime-readiness operations.
- Console calls Ledger through a Ledger port for governance receipts, audit
  events, billing evidence, retention policy, and account ledger views.
- Console-owned services may orchestrate a managed Workspace lifecycle, but the
  Workspace runtime and workbench experience remain outside Console.

When OPL Fabric or OPL Ledger become standalone services, Console should replace
the in-process adapters with API clients without changing Console product
responsibilities.

## Console Activation Rule

Console participates only when at least one of these is true:

- the resource is hosted by OPL Cloud;
- the resource is organization-managed;
- team policy requires approval, quota, audit, retention, or billing controls;
- a credential, connector, environment, skill pack, or agent package is shared
  or governed by the organization;
- an administrator needs operational visibility for managed resources.

If a local App, Workspace, MAS flow, user-provided SSH/HPC resource, or local
connector uses Fabric/Ledger outside organization policy, it does not become a
Console-managed resource by default.

## Current Implementation Target

The current repository should build the governed OPL Console slice:

- account, organization, team, role, and session management;
- policy and approval controls for managed Workspaces and resources;
- managed Workspace lifecycle facade;
- billing and usage visibility for managed Cloud usage;
- Fabric port/client integration for managed runtime resources;
- Ledger port/client integration for audit, receipts, and retention;
- production readiness and deployment handoff for the Console service.

It should not implement the Workspace workbench, Gateway provider runtime,
Fabric's full resource platform, Ledger's full evidence platform, or MAS/domain
quality judgment as Console-owned domains.
