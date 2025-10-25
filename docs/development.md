# Development Guide

## Prerequisites

- [Task](https://taskfile.dev/) - Task runner
- [mise](https://mise.jdx.dev/) - Version manager (optional but recommended)
- Node.js 20+
- Go 1.23+

## Quick Start

```bash
# First-time setup
task setup

# Start backend (Terminal 1)
task backend:serve

# Start mobile app (Terminal 2)
task mobile:serve
```

Press `i` for iOS simulator, `a` for Android, or scan QR with Expo Go.

## Project Structure

```
freebies/
├── apps/
│   └── mobile/          # React Native (Expo) app
├── services/
│   └── backend/         # Go API server
├── docs/                # Project-wide documentation
└── Taskfile.yml         # Root task definitions
```

## Available Tasks

Run `task --list` to see all available tasks.

| Task                 | Description                             |
| -------------------- | --------------------------------------- |
| `task setup`         | First-time setup (install dependencies) |
| `task clean`         | Wipe database                           |
| `task backend:serve` | Start backend server                    |
| `task backend:build` | Build backend binary                    |
| `task backend:test`  | Run backend tests                       |
| `task mobile:serve`  | Start Expo dev server                   |
| `task docs:fmt`      | Format markdown files                   |

## Development Workflow

### Backend Development

See [Backend Development Guide](../services/backend/docs/development.md) for detailed backend docs.

```bash
# Start server (auto-migrates database)
task backend:serve

# Create test deals
task backend:deals

# List users
./services/backend/bin/freebie users list
```

### Mobile Development

```bash
# Start Expo dev server
task mobile:serve

# Run on iOS simulator
# Press 'i' in the terminal
```

### Database

SQLite database with automatic migrations on startup.

```bash
# Wipe and recreate database
task clean
task backend:serve
```

### Adding New Offers

1. Create migration in `services/backend/internal/db/migrations/`
2. Restart server (migrations run automatically)

Example migration (`007_new_offer.sql`):

```sql
-- +goose Up
INSERT INTO events (id, offer_id, team_id, team_name, league, partner_name, ...)
VALUES (...);

-- +goose Down
DELETE FROM events WHERE id = '...';
```

## Environment Variables

| Variable                | Description          | Default      |
| ----------------------- | -------------------- | ------------ |
| `FREEBIE_DATABASE_PATH` | SQLite database path | `freebie.db` |
| `FREEBIE_SERVER_HOST`   | Server bind address  | `0.0.0.0`    |
| `FREEBIE_SERVER_PORT`   | Server port          | `8080`       |

## Code Style

### Markdown

Format all markdown files to 100 character width:

```bash
task docs:install  # First time only
task docs:fmt
```

### Go

```bash
task backend:lint
task backend:fmt
```

## Deployment

### Backend (Fly.io)

```bash
cd services/backend
fly deploy
```

### Mobile (EAS)

```bash
cd apps/mobile
eas build --platform ios
eas submit --platform ios
```
