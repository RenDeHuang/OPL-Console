# OPL Console

OPL Console is the independent control-plane repository for OPL Console and OPL Workspace.

## Stack

- Frontend: React + TypeScript
- Backend: Go
- DB: PostgreSQL
- K8s: Go client-go

## Development

Go commands require Go 1.23. In environments where the default Go toolchain download times out, use `GOTOOLCHAIN=go1.23.12`.

Fresh checkout setup:

```bash
docker compose -f docker-compose.dev.yml up -d postgres
npm install --prefix apps/web
GOTOOLCHAIN=go1.23.12 go run ./cmd/migrate up
```

Terminal 1, API:

```bash
GOTOOLCHAIN=go1.23.12 go run ./cmd/console-api
```

Terminal 2, web dev server:

```bash
npm --prefix apps/web run dev
```

Integration environment variables used by the Go API:

```bash
OPL_FABRIC_INTERNAL_URL=http://opl-fabric-api:8787
OPL_LEDGER_INTERNAL_URL=http://opl-ledger-api:8788
OPL_OPERATOR_TOKEN=replace-with-fabric-service-token
OPL_LEDGER_SERVICE_TOKEN=replace-with-ledger-service-token
OPL_LEDGER_ADMIN_TOKEN=replace-with-ledger-admin-token
```

The browser never reads these values. OPL Console reads them server-side and
delegates Fabric and Ledger calls over the internal network.

## Verification

```bash
gofmt -w cmd internal
GOTOOLCHAIN=go1.23.12 go mod tidy
GOTOOLCHAIN=go1.23.12 go test ./...
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
git diff --check
```
