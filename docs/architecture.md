# OPL Console Architecture

OPL Console is the independent control-plane repository for OPL Console and OPL Workspace lifecycle management.

The backend is a Go modular monolith. Fabric and Ledger are internal modules in v1, but their ports are shaped as future OPL Fabric and OPL Ledger API boundaries.

The frontend is React + TypeScript. PostgreSQL is the only runtime persistence store. Kubernetes provisioning uses Go `client-go`.

Fabric owns the workspace runtime boundary for compute, storage, attachment, workspace route, cleanup, and workspace URL token operations. The TKE adapter creates workspace compute and storage resources with `client-go`, routes `/w/{workspaceId}` ingress traffic to the Console validator service, and stores workspace token handoff state in Kubernetes Secrets.

Workspace creation provisions storage before compute. If a later create step fails after storage exists, storage is preserved for the seven-day hold lifecycle while compute is cleaned up when applicable.
