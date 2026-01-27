# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
make generate    # Generate proto (buf) + sqlc code
make build       # Compile to bin/server
make run         # Run with go run
make deps        # go mod tidy && download
```

Individual generation:
```bash
make proto       # buf dep update && buf generate
make sqlc        # sqlc generate
```

Requires: Go 1.25, protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway in PATH.

## Environment Variables

```
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
GRPC_PORT=9090   # default
HTTP_PORT=8080   # default
```

## Architecture

gRPC + HTTP Gateway backend with layered architecture:

```
handler/proto/api.proto    → API contract (protobuf)
        ↓ buf generate
compiled/                  → Generated code (gRPC stubs, gateway, sqlc models)
        ↓
handler/                   → RPC implementations
        ↓
service/                   → Business logic
        ↓
database/                  → Migrations & queries (sqlc)
```

**Servers:** gRPC (9090) and HTTP Gateway (8080) run concurrently. Gateway proxies REST to gRPC.

## Key Patterns

**Authentication:**
- Bearer token via gRPC metadata (`Authorization: Bearer <token>`)
- Auth interceptor in `handler/handler.go` with public method whitelist
- Use `UserFromContext(ctx)` to get authenticated user

**Code Generation:**
- All generated code goes to `compiled/` package
- Proto: `handler/proto/api.proto` → `compiled/*.pb.go`
- SQLC: `database/query.sql` → `compiled/query.sql.go`
- Schema source: `database/migration/sql/*.sql`

**Database:**
- Migrations embedded via `database/migration/embed.go`, auto-run on startup
- Soft deletes: check `deleted_at IS NULL` in queries
- Uses pgx/v5 with connection pooling

**Adding new RPC:**
1. Define in `handler/proto/api.proto` with `google.api.http` annotation
2. Run `make proto`
3. Implement method in `handler/`
4. Add to `publicMethods` map if no auth required
