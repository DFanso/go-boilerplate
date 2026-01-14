# go-boilerplate

A reference Go microservice workspace inspired by the architecture of the `allinonepos` services. It ships two services that illustrate common patterns:

- **identity-api** – user registration, login, JWT issuance and validation
- **dummy-api** – CRUD over a simple `items` resource that requires identity tokens and talks to `identity-api` via gRPC

Both services rely on the same tooling stack: [goa](https://goa.design/), `sqlc`, `golang-migrate`, `direnv`, [`air`](https://github.com/air-verse/air), Cobra-based CLIs, and PostgreSQL via `pgx`.

## Repository layout

```
go-boilerplate/
├── docker-compose.yml       # Postgres for each service
├── go.work                  # workspace tying the modules together
├── identity-api/            # identity microservice
├── dummy-api/               # CRUD microservice
├── Makefile                 # automation helpers
└── README.md
```

Each service mirrors the layout used in `/allinonepos/pos-*-api`:

- `cmd/<service>/` – Cobra CLI entry points (`serve`, `migrate`, …)
- `design/` – goa DSL definitions (HTTP + gRPC transports)
- `gen/` – generated goa transport code
- `internal/` – config, services, security helpers, SQLC generated query layer
- `migrations/` – PostgreSQL migrations managed via `golang-migrate`
- `sqlc.yaml` – SQL to Go code generation config
- `.air.toml` – hot reload config for local dev

## Tooling & prerequisites

- Go 1.25+
- direnv (`brew install direnv`) – automatically sources `.envrc`
- Docker (for the bundled Postgres instances)
- goa/sqlc/golang-migrate/air CLIs (install via `make tools`)

Enable the environment locally:

```bash
direnv allow
```

This loads defaults such as `IDENTITY_DATABASE_URL`, `DUMMY_IDENTITY_GRPC_TARGET`, etc. Override them by creating `.envrc.local`.

## Running locally

1. **Start databases**
   ```bash
   docker compose up -d
   ```

2. **Install CLIs (once)**
   ```bash
   make tools
   ```

3. **Apply migrations**
   ```bash
   make migrate-identity
   make migrate-dummy
   ```

4. **Run the services**
   ```bash
   make identity-run   # starts HTTP :8081 / gRPC :9081
   make dummy-run      # starts HTTP :8082 / gRPC :9082
   ```

   For live reload, run `air` inside each service (configuration already provided).

## Services in detail

### identity-api
- goa design exposes HTTP & gRPC endpoints for `register`, `login`, `validate_token`
- Stores users via SQLC generated queries (`internal/db/sqlc`)
- Passwords hashed with bcrypt, tokens issued via JWT (HS256)
- Provides a Go + gRPC client (exported from `gen/grpc/identity`) for inter-service calls

Useful commands:
```bash
cd identity-api
# hot reload
air
# run HTTP+gRPC server once
go run ./cmd/identity-api serve
# migrations (up/down/drop)
go run ./cmd/identity-api migrate --action up
```

### dummy-api
- Implements CRUD for `items` with PostgreSQL persistence
- Every request requires a Bearer token; service validates it by calling `identity-api` over gRPC before hitting the DB
- Provides both HTTP and gRPC transports via the generated goa server
- Serves OpenAPI spec at `/openapi.json`

Auth flow:
1. Register & log in using `identity-api`
2. Pass `Authorization: Bearer <token>` to any dummy endpoint (HTTP) or populate the `token` field for gRPC methods
3. `dummy-api` trims the bearer prefix, calls `identity-api.ValidateToken`, and uses the returned `user_id` as the `owner_id` for all CRUD operations

Example HTTP session:
```bash
# Register a user
curl -X POST http://localhost:8081/v1/identity/register \
  -d '{"email":"demo@example.com","password":"changeme123","display_name":"Demo"}'

# Login and capture the token
TOKEN=$(curl -s -X POST http://localhost:8081/v1/identity/login \
  -d '{"email":"demo@example.com","password":"changeme123"}' | jq -r .access_token)

# Create an item via dummy-api
curl -X POST http://localhost:8082/v1/dummy/items \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"First Item","description":"Protected resource"}'
```

## Code generation & maintenance

Re-run goa + sqlc for both services after editing their designs or SQL:
```bash
make generate
```

Format & tidy modules:
```bash
make fmt
make tidy
```

## Testing & next steps

- Extend the services with additional methods, e.g., refresh tokens or paginated list endpoints
- Wire the generated gRPC clients inside other services (dummy already shows how to consume identity)
- Adopt the provided templates when creating new microservices under this workspace (`cp -R identity-api new-api && rename`)

Happy hacking!
