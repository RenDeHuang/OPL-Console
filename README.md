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

## Verification

```bash
gofmt -w cmd internal
GOTOOLCHAIN=go1.23.12 go mod tidy
GOTOOLCHAIN=go1.23.12 go test ./...
npm --prefix apps/web run typecheck
npm --prefix apps/web run build
git diff --check
```
