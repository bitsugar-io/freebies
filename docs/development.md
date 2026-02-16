# Development Guide

## Prerequisites

Install tools using [mise](https://mise.jdx.dev/):

```bash
mise install
```

- **[Go](https://go.dev)** — Backend server and worker.
- **[Node.js](https://nodejs.org)** — Required as a runtime for React Native / Expo.
- **[Bun](https://bun.sh)** — JavaScript package manager used for mobile app and docs tooling.
- **[Task](https://taskfile.dev/)** — Task runner used for building, testing, formatting, and other
  development workflows.

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

## Branching

All work happens on feature branches created from `main`.

### Branch Naming

Use the pattern `type/short-description`:

| Prefix      | Use for                              |
| ----------- | ------------------------------------ |
| `feat/`     | New features                         |
| `fix/`      | Bug fixes                            |
| `docs/`     | Documentation changes                |
| `refactor/` | Code restructuring (no new behavior) |
| `chore/`    | Tooling, deps, CI changes            |

Examples: `feat/add-nba-source`, `fix/notification-retry`, `docs/update-readme`

### Auto-branching

When using Claude Code's `/commit` command while on `main`, a branch is automatically created using
the convention above. You don't need to create branches manually.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) with the 50/72 rule:

- **Subject line**: max 50 characters, imperative mood, capitalized, no period
- **Body**: wrap at 72 characters, separated from subject by a blank line
- **Format**: `type(scope): description`
- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
- **Scopes**: `backend`, `mobile`, `docs`, `db`
- Summarize the "what" and "why", not the "how"

Try to write meaningful commit messages and avoid having too many commits on a PR. Most PRs should
likely have a single commit (although for bigger PRs it may be reasonable to split it in a few). Git
squash and rebase is your friend!

### Examples

```
feat(backend): Add NBA game source

Implements the Source interface for NBA live game data
using the balldontlie API.
```

```
fix(mobile): Prevent duplicate push registrations
```

```
docs: Add branching and commit conventions
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
