# OPL Console Production Runbook

## Readiness

Runtime readiness:

```text
GET /api/runtime/readiness
```

Production readiness:

```text
GET /api/production/readiness
```

Production readiness is fail-closed. It rejects local URLs, missing auth seed,
the built-in admin bootstrap credential, `:latest` Workspace images, missing
Kubernetes config, and missing production storage/ingress inputs.

## Required Inputs

- `DATABASE_URL`
- `OPL_PUBLIC_URL`
- `OPL_WORKSPACE_DOMAIN`
- `OPL_WORKSPACE_IMAGE`
- `OPL_K8S_NAMESPACE`
- `OPL_INGRESS_CLASS`
- `OPL_WORKSPACE_STORAGE_CLASS`
- `OPL_CONSOLE_USERS_JSON`
- `KUBECONFIG` or equivalent mounted production kubeconfig path

## Human Gates

Explicit operator approval is required before:

- applying production manifests;
- injecting or rotating production secrets;
- creating real managed Workspace resources;
- destroying retained managed storage;
- running production verifier against public domains.
