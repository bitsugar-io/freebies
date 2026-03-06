# Backend Architecture

## Overview

The Freebie backend is a Go HTTP API server with [Turso](https://turso.tech) (hosted SQLite) for
persistence, deployed on DigitalOcean Kubernetes (DOKS).

See the [top-level architecture doc](../../../docs/architecture.md) for the full infrastructure
diagram including Kubernetes, Cloudflare Tunnel, and CI/CD.

## Technology Choices

### DigitalOcean Kubernetes (DOKS)

**Why DOKS:**

- **Managed Kubernetes** — DO handles control plane, upgrades, and node provisioning
- **Cheap for small projects** — single node at ~$12/mo runs everything
- **Helm charts** — declarative, version-controlled deployments
- **CronJobs** — native K8s primitive for scheduled worker tasks

### SQLite + Turso

**Development:** Local SQLite file (`freebie.db`)

- Zero setup, fast iteration

**Production:** [Turso](https://turso.tech) (hosted SQLite)

- SQLite-compatible, same queries work locally and in production
- No PVC or persistent storage needed in the cluster
- Free tier: 9GB storage, 500M reads/month

```bash
# Dev
./bin/freebie serve  # uses local freebie.db

# Prod
FREEBIE_DATABASE_PATH="libsql://xxx.turso.io?authToken=xxx" ./bin/freebie serve
```

### Go (vs Node.js/Python)

**Why Go:**

- **Single binary deployment** — no runtime dependencies, distroless container image
- **Low memory footprint** — runs well on small K8s nodes
- **Strong typing** — catches errors at compile time
- **sqlc** — type-safe SQL queries generated from schema

## Components

```
┌──────────────────────────────────────────────────────────────┐
│                     DOKS Cluster                             │
│  ┌─────────────────────┐   ┌──────────────────────────────┐  │
│  │   API Deployment    │   │   Scheduler (ephemeral pods) │  │
│  │  ┌───────────────┐  │   │  ┌────────────────────────┐  │  │
│  │  │  HTTP API     │  │   │  │  ./scheduler            │  │  │
│  │  │  (serve)      │◄─┼───┼──│  check-triggers        │  │  │
│  │  └───────┬───────┘  │   │  │  send-reminders        │  │  │
│  │          │          │   │  └────────────────────────┘  │  │
│  └──────────┼──────────┘   └──────────────────────────────┘  │
└─────────────┼────────────────────────────────────────────────┘
              │
  ┌───────────┼───────────────────────────┐
  ▼           ▼               ▼           │
┌────────────┐ ┌────────────┐ ┌──────────┐│
│   Turso    │ │ MLB Stats  │ │ Expo Push││
│  (SQLite)  │ │    API     │ │  Notif.  ││
└────────────┘ └────────────┘ └──────────┘│
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
- Stateless requests — all state in Turso

## Background Jobs

Background jobs run as Kubernetes CronJobs using the `services/scheduler/` container. Each CronJob
pod calls the API's internal worker endpoints via HTTP — it does not access the database directly.

- **6am PT** (`check-triggers`): Check yesterday's game results, create triggered events, notify
  subscribers
- **6pm PT** (`send-reminders`): Send reminder notifications for deals expiring soon

See the [top-level architecture doc](../../../docs/architecture.md#worker-flow) for the full
worker flow.

## Future Considerations

- **More leagues**: Add `sources/nba/`, `sources/nfl/` implementations
- **Multi-region**: Turso supports edge replicas for lower latency
