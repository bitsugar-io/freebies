# Service Separation Design

## Summary

Split `services/backend` into two independent services: `services/api` (HTTP server) and
`services/scheduler` (CronJob client). Rename `charts/cronjobs` to `charts/scheduler`.

## Current State

- Single Go binary (`freebie`) in `services/backend/` contains everything: API server, worker logic,
  worker remote client, admin CLI commands
- One Docker image shared by API Deployment and CronJobs
- CronJobs run `freebie worker remote check-triggers|send-reminders` which makes HTTP calls to the
  API — no direct database access

## Target State

### services/api (renamed from services/backend)

The HTTP API server. Owns all business logic, database access, and internal worker endpoints.

```
services/api/
├── cmd/
│   ├── root.go           # Cobra root command, viper config
│   ├── serve.go          # HTTP server
│   ├── users.go          # Admin: list users
│   ├── deals.go          # Admin: create test deals
│   └── notify.go         # Admin: send test notifications
├── internal/
│   ├── api/              # HTTP handlers, middleware, routes
│   │   ├── handlers/     # Request handlers
│   │   ├── middleware/    # Auth, internal auth
│   │   └── worker/gen/   # Generated OpenAPI server code
│   ├── config/           # Viper config structs
│   ├── db/               # sqlc queries, migrations, models
│   ├── notify/           # Expo push notification client
│   ├── sources/          # Game data sources (mlb, etc.)
│   ├── triggers/         # Trigger evaluation logic
│   └── worker/           # Worker service (CheckTriggers, SendReminders)
├── docs/                 # API-specific docs
├── Dockerfile            # CGO enabled (SQLite), distroless
├── go.mod                # module github.com/retr0h/freebie/services/api
└── go.sum
```

### services/scheduler (new)

Lightweight CLI that calls the API's internal worker endpoints via HTTP. No database, no business
logic.

```
services/scheduler/
├── cmd/
│   ├── root.go           # Cobra root, viper config
│   ├── check_triggers.go # POST /internal/worker/check-triggers
│   └── send_reminders.go # POST /internal/worker/send-reminders
├── internal/
│   ├── client/gen/       # Generated OpenAPI HTTP client
│   └── config/           # Viper config (api URL, worker secret)
├── Dockerfile            # Pure Go, no CGO, distroless
├── go.mod                # module github.com/retr0h/freebie/services/scheduler
└── go.sum
```

### Shared Patterns

Both services follow identical conventions:

- **CLI**: Cobra for commands, Viper for config with `mapstructure` env tags
- **Logging**: `slog` with structured key-value pairs
- **Layout**: `cmd/` for CLI commands, `internal/` for packages
- **Config**: env vars prefixed with `FREEBIE_` (api) / `SCHEDULER_` (scheduler)
- **Dockerfile**: multi-stage build, distroless runtime, nonroot user
- **Error handling**: `fmt.Errorf("context: %w", err)` wrapping

### Helm Charts

```
charts/
├── api/          # No changes (already correct)
├── scheduler/    # Renamed from charts/cronjobs, uses scheduler image
└── cloudflare/   # No changes
```

`charts/scheduler/` changes:

- Image reference updated to `freebie-scheduler` (separate, smaller image)
- CronJob commands change from `./freebie worker remote check-triggers` to
  `./scheduler check-triggers`

### CI/CD

`.github/workflows/deploy.yaml` updated to:

- Build and push two images: `freebie-api` and `freebie-scheduler`
- Deploy three Helm charts: api, scheduler, cloudflare

### What Moves Where

| Current location                        | Destination                           |
| --------------------------------------- | ------------------------------------- |
| `services/backend/cmd/serve.go`         | `services/api/cmd/serve.go`           |
| `services/backend/cmd/users.go`         | `services/api/cmd/users.go`           |
| `services/backend/cmd/deals.go`         | `services/api/cmd/deals.go`           |
| `services/backend/cmd/notify.go`        | `services/api/cmd/notify.go`          |
| `services/backend/cmd/root.go`          | `services/api/cmd/root.go`            |
| `services/backend/cmd/worker.go`        | Deleted (local worker mode removed)   |
| `services/backend/cmd/worker_remote.go` | Split into `services/scheduler/cmd/`  |
| `services/backend/internal/*`           | `services/api/internal/*`             |
| `services/backend/internal/client/`     | `services/scheduler/internal/client/` |
| `charts/cronjobs/`                      | `charts/scheduler/`                   |

### What Gets Deleted

- `cmd/worker.go` — local worker mode (direct DB access from CLI). The scheduler uses HTTP, and the
  worker service logic stays in the API.
- `cmd/worker_remote.go` — replaced by scheduler commands.

## Config

### services/api

| Variable                | Description                         | Default      |
| ----------------------- | ----------------------------------- | ------------ |
| `FREEBIE_DATABASE_PATH` | Turso connection URL                | `freebie.db` |
| `FREEBIE_SERVER_HOST`   | Bind address                        | `0.0.0.0`    |
| `FREEBIE_SERVER_PORT`   | Port                                | `8080`       |
| `FREEBIE_WORKER_SECRET` | Bearer token for internal endpoints | (required)   |

### services/scheduler

| Variable                  | Description    | Default                   |
| ------------------------- | -------------- | ------------------------- |
| `SCHEDULER_API_URL`       | API server URL | `http://freebie-api:8080` |
| `SCHEDULER_WORKER_SECRET` | Bearer token   | (required)                |
