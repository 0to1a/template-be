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

Requires: Go 1.26, protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway in PATH.

## Environment Variables

```
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
GRPC_PORT=9090   # default
HTTP_PORT=8080   # default
```

## Architecture

gRPC + HTTP Gateway backend with layered architecture. Module name is `project`.

```
handler/proto/api.proto    → API contract (protobuf)
        ↓ buf generate
compiled/                  → Generated code (gRPC stubs, gateway, sqlc models)
        ↓
handler/                   → RPC implementations (one file per domain)
        ↓
service/                   → Business logic (one service per domain)
        ↓
database/                  → Migrations & queries (sqlc)
```

**Servers:** gRPC (9090) and HTTP Gateway (8080) run concurrently. Gateway proxies REST to gRPC.

**Wiring:** `main.go` initializes `compiled.Queries` → services → `Handler`. Services receive `*compiled.Queries`; Handler receives services + queries. New services must be added to both `NewHandler()` and the `Handler` struct in `handler/handler.go`, then constructed in `cmd/server/main.go`.

## Key Patterns

**Authentication:**
- Bearer token via gRPC metadata (`Authorization: Bearer <token>`)
- Auth interceptor in `handler/handler.go` with `publicMethods` whitelist
- Use `UserFromContext(ctx)` to get `*AuthenticatedUser` (includes `ID`, `Email`, `Name`, `SelectedCompanyID`)
- Login uses hardcoded OTP (`123456`) — tokens are random hex strings stored in the `users.token` column

**Code Generation — never edit `compiled/` directly:**
- Proto: `handler/proto/api.proto` → `compiled/*.pb.go` (buf)
- SQLC: `database/query.sql` → `compiled/query.sql.go` (sqlc with pgx/v5)
- Schema source: `database/migration/sql/*.sql`

**Database:**
- Migrations embedded via `database/migration/embed.go`, auto-run on startup
- Soft deletes: all queries must include `deleted_at IS NULL`
- Uses pgx/v5 with connection pooling (`pgxpool`)
- Migration files follow `000NNN_description.{up,down}.sql` naming
- Roles in `company_users`: `admin` or `member` (enforced by CHECK constraint)

**Adding a new RPC:**
1. Define in `handler/proto/api.proto` with `google.api.http` annotation
2. Run `make proto`
3. Implement method in `handler/` (one file per domain, e.g. `handler/company.go`)
4. Add to `publicMethods` map in `handler/handler.go` if no auth required

**Adding a new query:**
1. Add SQL with `-- name: QueryName :one|:many|:exec` annotation to `database/query.sql`
2. Run `make sqlc`
3. Use generated method via `queries.QueryName(ctx, params)` in service layer

**Error handling in handlers:** Map service-layer sentinel errors (e.g. `service.ErrNotAdmin`) to gRPC status codes. Return `status.Error(codes.X, "message")`.

**API Collection (`hoppscotch.json`):**
- When adding or updating HTTP endpoints, always update `hoppscotch.json` to reflect the changes
- Requests are organized by domain folders (Auth, User, Company, etc.)
- Use `<<base_url>>` variable for the endpoint base URL
- Use `<<token>>` variable for Bearer auth tokens
- Include `testScript` to auto-set env vars when useful (e.g. saving token from login response)
- Each request should have a `description` explaining its purpose and requirements
