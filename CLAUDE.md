# Freebie

Sports rewards notification platform. Monitors live games and sends push notifications when free food/merch deals are triggered (e.g., Dodgers pitchers get 7+ strikeouts → free Jumbo Jack).

## Tech Stack

- **Backend**: Go 1.25, SQLite (Turso in prod), deployed on Fly.io
- **Mobile**: React Native / Expo (TypeScript)
- **Notifications**: Expo Push Notifications
- **Task runner**: [Task](https://taskfile.dev) (Taskfile.yml)
- **DB migrations**: goose (embedded, auto-run on `task serve`)
- **Code generation**: sqlc (SQL → Go)

## Directory Layout

```
services/backend/   Go API server + worker
apps/mobile/        React Native / Expo app
docs/               Documentation site
```

## Local Development

### Backend
```bash
cd services/backend && task serve    # starts API on localhost:8080
task dev                              # auto-rebuild on changes (requires watchexec)
```

### Mobile
```bash
cd apps/mobile && npx expo start
```

## Key Commands (from services/backend/)

| Command            | Description                          |
|--------------------|--------------------------------------|
| `task serve`       | Start API server                     |
| `task dev`         | Auto-rebuild on file changes         |
| `task clean`       | Wipe local DB (re-migrates on serve) |
| `task generate`    | Regenerate sqlc code                 |
| `task test`        | Run Go tests                         |
| `task build`       | Build production binary              |
| `task deals`       | Create test deals (24h, 6h, 2h)     |
| `task deals:list`  | List active deals                    |
| `task notify`      | Send test notification to all users  |

## Code Generation (sqlc)

1. Edit `services/backend/internal/db/queries.sql`
2. Run `task generate` (from `services/backend/`)
3. Generated Go code appears in `services/backend/internal/db/`

## Database Migrations

Migrations live in `services/backend/internal/db/migrations/` and use goose format:

```sql
-- +goose Up
<SQL statements>

-- +goose Down
<SQL statements>
```

Migrations are embedded in the Go binary and run automatically on server start.

## Adding New Events

Insert a new row into `events` via a migration file (see `003_mlb_data.sql` for reference).

## Adding New Leagues/Sources

Implement the `Source` interface in `services/backend/internal/sources/` and register it in the worker.

## Environment Variables

| Variable                | Default          | Description             |
|-------------------------|------------------|-------------------------|
| `FREEBIE_DATABASE_PATH` | `freebie.db`     | SQLite database path    |
| `FREEBIE_SERVER_HOST`   | `0.0.0.0`       | Server listen address   |
| `FREEBIE_SERVER_PORT`   | `8080`           | Server listen port      |
