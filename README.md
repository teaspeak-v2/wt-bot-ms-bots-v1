# WT-Bot Bots Microservice

`wt-bot-ms-bots-v1` manages bot configuration and lifecycle for the WT-Bot platform.

- CRUD for bot records
- Runtime configuration endpoint for bot instances
- Status updates from bot instances

## Architecture

- **HTTP API**: chi router, JSON-only responses, JWT middleware
- **Service-to-service**: `X-Service-Key` header for bot-facing endpoints
- **Persistence**: Postgres via `pgxpool`
- **Caching**: Redis for bot configuration caching
- **Migrations**: `golang-migrate` with embedded SQL files
- **Docs**: Swagger/OpenAPI via swag annotations
- **Logging**: structured JSON logs with `slog`

## Quickstart

```bash
make docker-up
make run
```

The API listens on `:8080` by default.

## Environment

Copy `.env.example` to `.env` and adjust values as needed.

## Commands

- `make build` — compile all packages
- `make test` — run unit tests
- `make vet` — run Go vet
- `make lint` — run golangci-lint
- `make swag` — regenerate Swagger docs
- `make migrate-up` / `make migrate-down` — run DB migrations

## API Summary

- `GET /api/v1/bots`
- `POST /api/v1/bots`
- `GET /api/v1/bots/{id}`
- `PATCH /api/v1/bots/{id}`
- `DELETE /api/v1/bots/{id}`
- `GET /api/v1/bots/{id}/config` (service key)
- `POST /api/v1/bots/{id}/status` (service key)
- `GET /healthz`
- `GET /readyz`

Swagger UI: `http://localhost:8080/swagger/index.html`
