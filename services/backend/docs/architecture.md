# Backend Architecture

## Overview

The Freebie backend is a Go HTTP API server with SQLite for persistence, deployed on Fly.io.

## Technology Choices

### Fly.io (vs AWS Lambda/SQS)

**Why Fly.io:**

- **Persistent disk storage** - SQLite lives on disk, no external database needed. Lambda is
  stateless and would require RDS/DynamoDB ($20-50+/month minimum).
- **No cold starts** - Always-on VM means consistent response times. Lambda cold starts add
  100-500ms latency.
- **Simpler architecture** - Single binary serves HTTP, runs migrations, sends notifications. Lambda
  would need API Gateway + Lambda + SQS + separate notification workers.
- **Predictable pricing** - $5/month for smallest VM with 1GB persistent disk. Lambda pricing is
  usage-based and harder to predict with SQS/API Gateway costs.
- **Easy deployment** - `fly deploy` from CLI. No CloudFormation/Terraform/SAM templates.

**Trade-offs accepted:**

- Single region (LAX) for now - acceptable for MVP targeting LA-area users
- Would need [LiteFS](https://fly.io/docs/litefs/) for multi-region SQLite replication if we expand
- Less "infinite scale" than Lambda - but a single Fly VM handles thousands of concurrent users

### SQLite + Turso

**Development:** Local SQLite file (`freebie.db`)

- Zero setup, fast iteration

**Production:** [Turso](https://turso.tech) (hosted SQLite)

- SQLite-compatible, same queries work
- Allows Fly.io to scale to zero (database is external)
- Free tier: 9GB storage, 500M reads/month
- No volume mounting needed for scheduled jobs

```bash
# Dev
./bin/freebie serve  # uses local freebie.db

# Prod
FREEBIE_DATABASE_PATH="libsql://xxx.turso.io?authToken=xxx" ./bin/freebie serve
```

### Go (vs Node.js/Python)

**Why Go:**

- **Single binary deployment** - No runtime dependencies, simple Dockerfile
- **Low memory footprint** - Runs well on smallest Fly VM (256MB)
- **Strong typing** - Catches errors at compile time
- **sqlc** - Type-safe SQL queries generated from schema

## Components

```
┌─────────────────────────────────────────────────────────┐
│                        Fly.io                           │
│  ┌─────────────────────┐   ┌─────────────────────────┐  │
│  │    App Machine      │   │   Scheduled Machine     │  │
│  │  ┌───────────────┐  │   │  ┌───────────────────┐  │  │
│  │  │  HTTP API     │  │   │  │  worker run       │  │  │
│  │  │  (serve)      │  │   │  │  (hourly)         │  │  │
│  │  └───────┬───────┘  │   │  └─────────┬─────────┘  │  │
│  │          │          │   │            │            │  │
│  │  auto-start/stop    │   │  6am: check-triggers    │  │
│  │  on HTTP traffic    │   │  6pm: send-reminders    │  │
│  └──────────┼──────────┘   └────────────┼────────────┘  │
└─────────────┼──────────────────────────┼────────────────┘
              │                          │
              └────────────┬─────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
   ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
   │    Turso     │ │  MLB Stats   │ │  Expo Push   │
   │   (SQLite)   │ │     API      │ │ Notification │
   └──────────────┘ └──────────────┘ └──────────────┘
```

## Trigger System

The trigger system checks sports API data and creates deals when conditions are met.

### Data Flow

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│     events      │     │     sources      │     │    triggers     │
│   (database)    │────▶│  (API clients)   │────▶│   (evaluator)   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                        │
        │  team_id: "LAD"        │  statsapi.mlb.com     │  9 >= 7?
        │  league: "mlb"         │  → schedule           │  → TRIGGERED
        │  trigger_rule: {...}   │  → boxscore           │  → create deal
        │                        │  → GameStats          │
```

### Event Record (Source of Truth)

Each event in the database drives the entire trigger flow:

```json
{
  "team_id": "LAD",
  "league": "mlb",
  "trigger_rule": {
    "metric": "strikeouts",
    "operator": ">=",
    "value": 7,
    "redemption_window": "next_day"
  }
}
```

- `league` → determines which Source to use
- `team_id` → passed to Source to fetch correct team data
- `trigger_rule` → evaluated against game metrics

### Sources (`internal/sources/`)

Interface-based design for fetching game data from external APIs:

```go
type Source interface {
    League() string
    GetYesterdaysGame(ctx, teamID) (*GameStats, error)
    GetGameByDate(ctx, teamID, date) (*GameStats, error)
}
```

Sources auto-register via `init()`:

- `sources/mlb/` → MLB Stats API (statsapi.mlb.com)
- Future: `sources/nba/`, `sources/nfl/`

### Triggers (`internal/triggers/`)

Orchestrates checking all events:

1. Load active events from database
2. For each event, get Source by league
3. Fetch game stats for the date
4. Evaluate rule against metrics
5. Create `triggered_event` if conditions met (idempotent by game_id)

## API Design

- RESTful JSON API at `/api/v1/`
- Bearer token authentication (stored in `users.token`)
- Stateless requests - all state in SQLite

## Background Jobs

Background jobs run on a separate Fly.io scheduled machine (not in the HTTP server process). This
allows the app machine to scale to zero when idle.

### Scheduled Machine

A Fly.io machine runs `worker run` hourly. The command checks Pacific Time and runs the appropriate
job:

- **6am PT**: `check-triggers` - Check yesterday's game results, create triggered events, notify
  subscribers
- **6pm PT**: `send-reminders` - Send reminder notifications for deals expiring soon

```bash
# Create the scheduled machine (first time only)
# See deployment.md for the full command with correct syntax
fly machine run registry.fly.io/freebie-api:<image-tag> worker run \
  --schedule hourly -a freebie-api --region sjc
```

### Worker Commands

```bash
# Run scheduled jobs (checks PT hour, runs appropriate job)
./freebie worker run

# Manually check triggers for yesterday
./freebie worker check-triggers

# Check triggers for a specific date
./freebie worker check-triggers --date 2024-04-15

# Manually send reminders
./freebie worker send-reminders
```

See [Deployment Guide](deployment.md) for full setup instructions.

## Future Considerations

- **Multi-region**: Turso supports edge replicas for lower latency
- **More leagues**: Add `sources/nba/`, `sources/nfl/` implementations
- **Caching**: Add Redis on Fly if needed, but Turso is fast enough for now
