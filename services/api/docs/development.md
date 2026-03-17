# Backend Development

## Prerequisites

- Go 1.23+
- [Task](https://taskfile.dev/)
- [sqlc](https://sqlc.dev/) (installed via `go tool`)

## Quick Start

```bash
# From repo root
task api:serve
```

The server will:

1. Create the SQLite database if it doesn't exist
2. Run all migrations (schema + seed data)
3. Start listening on `http://localhost:8080`

## Project Structure

```
services/api/
├── cmd/                    # CLI commands (serve, users, deals)
├── internal/
│   ├── api/                # HTTP handlers and middleware
│   ├── config/             # Configuration structs
│   ├── db/                 # Database (sqlc generated + migrations)
│   ├── notify/             # Push notification service
│   ├── rules/              # Deal trigger evaluation
│   └── sources/            # External data sources (MLB, etc.)
├── db/
│   ├── queries.sql         # sqlc query definitions
│   └── sqlc.yaml           # sqlc configuration
└── docs/                   # Documentation
```

## Database

SQLite database with [goose](https://github.com/pressly/goose) migrations.

### Migrations

Located in `internal/db/migrations/`. Run automatically on server start.

```bash
# Wipe database and start fresh
task api:clean
task api:serve
```

### Adding New Migrations

1. Create a new file: `internal/db/migrations/NNN_description.sql`
2. Use goose format:

```sql
-- +goose Up
CREATE TABLE ...;

-- +goose Down
DROP TABLE ...;
```

### Updating Queries

1. Edit `db/queries.sql`
2. Regenerate: `task api:sqlc`

## CLI Commands

```bash
./bin/freebie serve              # Start API server
./bin/freebie users list         # List all users
./bin/freebie deals list         # List active deals
./bin/freebie deals create       # Create test deal
./bin/freebie deals trigger ID   # Trigger specific event
./bin/freebie notify test        # Send test notification to all users
./bin/freebie notify send ID     # Send notification to specific user
```

## Configuration

Configuration via flags, environment variables, or config file.

| Flag      | Env Var                 | Default      |
| --------- | ----------------------- | ------------ |
| `--db`    | `FREEBIE_DATABASE_PATH` | `freebie.db` |
| `--host`  | `FREEBIE_SERVER_HOST`   | `0.0.0.0`    |
| `--port`  | `FREEBIE_SERVER_PORT`   | `8080`       |
| `--debug` | -                       | `false`      |

## Testing

```bash
task api:test
```

## Deployment

Deployed to DigitalOcean Kubernetes (DOKS). See [deployment guide](deployment.md) for full
instructions.
