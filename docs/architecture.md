# Architecture

## Overview

Freebies: LA Sports Deals is an LA-only rewards app: when something happens in LA (today, an LA team
result such as a Ducks home win or 7+ Dodgers strikeouts), subscribed users get a push for the food
deal it unlocks. The backend runs on a DigitalOcean Kubernetes (DOKS) cluster. It checks finished
games, detects when trigger conditions are met, and sends push notifications to subscribed users.

## Infrastructure

```text
                    Internet
                       │
                       ▼
              Cloudflare Edge (TLS)
                       │
                       ▼ Cloudflare Tunnel
┌──────────────────────────────────────────────────────────┐
│  DOKS Cluster (1 node, s-1vcpu-2gb)                      │
│                                                          │
│  ┌────────────────────────────────────────────┐          │
│  │  Deployment: cloudflared                   │          │
│  │  Maintains outbound tunnel to CF edge      │          │
│  │  Routes traffic → freebie-api:8080         │          │
│  └────────────────────┬───────────────────────┘          │
│                       │                                  │
│                       ▼                                  │
│  ┌────────────────────────────────────────────┐          │
│  │  Deployment: freebie-api (1 replica)       │          │
│  │  Container: freebie serve                  │          │
│  │  ClusterIP Service :8080                   │          │
│  │                                            │          │
│  │  Public:   /api/v1/*    (mobile app)       │          │
│  │  Internal: /internal/*  (worker jobs)      │          │
│  │  Health:   /healthz                        │          │
│  └────────────────────┬───────────────────────┘          │
│                       │                                  │
│  ┌────────────────────┼───────────────────────┐          │
│  │  CronJobs (ephemeral pods)                 │          │
│  │                    │ HTTP POST              │          │
│  │  check-triggers ───┘ Bearer token auth     │          │
│  │  (7am PT, 9am retry)                       │          │
│  │                                            │          │
│  │  send-reminders ──── same pattern          │          │
│  │  (11am, 3pm, 5pm PT)                       │          │
│  └────────────────────────────────────────────┘          │
└──────────────────────────────────────────────────────────┘
                       │
                       ▼
         ┌──────────────────────────┐
         │  Turso (hosted SQLite)   │
         │  libsql://xxx.turso.io   │
         └──────────────────────────┘
```

**Deployed by**: Helm charts (`charts/`), run by hand. How to deploy:
[deploy-api.md](../.claude/skills/ops/references/deploy-api.md) in the `ops` skill.

## Database: Turso (hosted SQLite)

[Turso](https://turso.tech) provides hosted SQLite via the libsql protocol. The same Go code works
with both local SQLite files (development) and Turso (production).

```bash
# Development — local SQLite file
./bin/freebie serve  # uses freebie.db

# Production — Turso
FREEBIE_DATABASE_PATH="libsql://freebie-xxx.turso.io?authToken=xxx" ./bin/freebie serve
```

The connection logic in `internal/db/conn.go` detects the scheme:

- `libsql://` prefix → opens a Turso connection via `libsql` driver
- Anything else → opens a local SQLite file via `sqlite3` driver

**Why Turso over local SQLite on a PVC:**

- No `ReadWriteOnce` PVC constraint — multiple pods can connect
- No risk of data loss from pod eviction or node failure
- CronJob pods and the API pod share the same database without HTTP indirection
- Free tier: 9GB storage, 500M reads/month
- Migrations still run on API startup via goose (same as local)

**Note:** The scheduler CronJob pods call the API via HTTP rather than connecting to Turso directly.
This keeps all business logic (trigger evaluation, notification sending) in the API process and
avoids duplicating dependencies in the scheduler container.

## Cloudflare Tunnel

External traffic reaches the API through a
[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/),
eliminating the need for a Kubernetes LoadBalancer or Ingress controller.

### How it works

1. A `cloudflared` pod runs inside the cluster as a Deployment
2. On startup, it establishes an **outbound** connection to Cloudflare's edge network
3. Cloudflare routes traffic for your domain through this tunnel to the pod
4. The pod forwards requests to the `freebie-api` ClusterIP Service on port 8080

```text
User request → your-domain.com
       → Cloudflare edge (TLS termination, DDoS protection, caching)
       → Cloudflare Tunnel (encrypted, outbound-initiated)
       → cloudflared pod in cluster
       → freebie-api:8080 (ClusterIP)
```

### Why not a LoadBalancer or Ingress?

| Approach          | Cost          | Complexity                           |
| ----------------- | ------------- | ------------------------------------ |
| DO Load Balancer  | ~$12/mo       | Need cert-manager for TLS            |
| Nginx Ingress     | $0 (NodePort) | Ingress controller + cert-manager    |
| Cloudflare Tunnel | $0            | Single deployment, TLS handled by CF |

Setup steps are in [deploy-api-setup.md](../.claude/skills/ops/references/deploy-api-setup.md).

### Security benefits

- **No open inbound ports** — the tunnel is outbound-only from the cluster
- **No public IP needed** — the API service is ClusterIP (cluster-internal only)
- **DDoS protection** — Cloudflare's edge absorbs attacks before they reach the cluster
- **TLS everywhere** — terminated at Cloudflare edge, no cert management needed in-cluster

## Worker Flow

### Check Triggers (7am PT, retried 9am)

1. K8s CronJob starts a pod running `./scheduler check-triggers`
2. The scheduler uses a generated HTTP client to `POST /internal/worker/check-triggers`
3. The API authenticates via bearer token (`FREEBIE_WORKER_SECRET`)
4. The worker service queries Turso for active events
5. For each event, it fetches game data from that league's source (one `Source` per league, matched
   case-insensitively; see [Game data sources](#game-data-sources)). An event whose league has no
   source, or whose rule metric the source never emits, is logged as a warning and skipped
6. If a trigger condition is met, it creates a `triggered_event` record (with `available_at` when
   the deal's code unlocks later)
7. It sends push notifications to subscribed users via Expo
8. Returns a JSON summary to the CronJob pod

### Game data sources

Each league has a `Source` (`services/api/internal/sources/`) that returns a team's finished game
for a Pacific date, with a flat map of metrics that `trigger_rule`s compare against. Sources are
chosen per league for reliability on past dates, then terms and simplicity; the full decision table
and fallbacks are in [data-sources.md](../.claude/skills/ops/references/data-sources.md).

| League           | Source                           | Package                      |
| ---------------- | -------------------------------- | ---------------------------- |
| MLB              | MLB Stats API through `mlb-sdk`  | `sources/mlb`                |
| NHL              | NHL web API (`api-web.nhle.com`) | `sources/nhl`                |
| NBA, WNBA, NFL   | ESPN public site API             | `sources/nba`, `wnba`, `nfl` |
| MLS, NWSL        | ESPN public site API             | `sources/mls`, `nwsl`        |
| NCAA (USC, UCLA) | ESPN public site API             | `sources/ncaa`               |

The six ESPN leagues share one client (`sources/espn`): retry with backoff, a per-date cache and
Pacific-date filtering, with thin per-league packages for the team-id map and metrics. The NBA and
WNBA's own feeds (`cdn.nba.com`) answer 403 to server traffic, and MLS, NWSL and the NCAA have no
public official feed, so ESPN is the only workable free choice. ESPN's API is unofficial and its
terms do not license commercial use, so budget a licensed feed before monetizing. Every source
declares the metrics it emits (`Metrics()`); `go run . deals audit` lists each deal as OK,
NO_SOURCE, UNKNOWN_METRIC, UNKNOWN_TEAM, BAD_RULE or INACTIVE and exits non-zero if an active deal
can never fire, and a Go test runs the same audit over every migration.

### Send Reminders (11am, 3pm and 5pm PT)

1. K8s CronJob starts a pod running `./scheduler send-reminders`
2. The scheduler calls `POST /internal/worker/send-reminders`
3. The API queries Turso for deals expiring within 6 hours
4. For each expiring deal, it sends reminder notifications to eligible users
5. Returns a JSON summary

## Server-driven content

New deals and teams must reach users without an App Store update, so everything about them lives in
the database and is served by the API. The app renders whatever it receives.

| Content                                  | Source of truth                                  | Served by                             |
| ---------------------------------------- | ------------------------------------------------ | ------------------------------------- |
| Deals: offer, trigger, deadline, sponsor | `events` table (goose migrations)                | `GET /api/v1/events`, active deals    |
| Deal pictures                            | `sprites` table, generated by the `design` skill | `event.sprite`, `GET /api/v1/sprites` |
| Teams, leagues, season windows, colors   | `teams` table                                    | `GET /api/v1/teams`                   |
| Sport badges                             | `sprites` table (`sport-*` ids)                  | `team.badge`                          |
| Push text and deadlines                  | built by the worker from the deal                | Expo push                             |
| Feature flags, banners                   | `feature_flags`, `screen_blocks`                 | `GET /api/v1/config`                  |

Rules:

- The app never hardcodes team, league or sprite lists. An unknown team, league or sport must render
  with no code change.
- Bundled art in the app is an offline fallback only.
- Only new screens, new kinds of UI and fonts need an app release.

## API Layers

### Public API (`/api/v1/`)

User-facing endpoints for the mobile app. Protected by per-user bearer tokens.

- Events, leagues, subscriptions, active deals, dismissals
- Sprites: `GET /api/v1/sprites` (all), `/sprites/{id}` (JSON), `/sprites/{id}.png?scale=N` (PNG).
  Pixel-art data for each deal, drawn natively by the app (see [UI rules](ui.md#sprites))

### Internal Worker API (`/internal/worker/`)

Protected by a shared bearer token (`FREEBIE_WORKER_SECRET`). Called by CronJobs within the cluster.

- `POST /internal/worker/check-triggers` — check game results
- `POST /internal/worker/send-reminders` — send expiring deal reminders

The internal API is defined by an **OpenAPI 3.0 spec** (`internal/api/worker/gen/api.yaml`). Server
handlers and HTTP client code are **generated** from this spec using `oapi-codegen`:

```text
internal/api/worker/gen/
├── api.yaml          # OpenAPI spec
├── cfg.yaml          # oapi-codegen config (chi-server + strict-server)
├── generate.go       # //go:generate directive
└── worker.gen.go     # Generated server code

internal/client/gen/
├── api.yaml          # Same OpenAPI spec
├── cfg.yaml          # oapi-codegen config (client only)
├── generate.go       # //go:generate directive
└── client.gen.go     # Generated HTTP client
```

To regenerate after spec changes:

```bash
go generate ./internal/api/worker/gen/ ./internal/client/gen/
```

## Helm Charts

Each component is deployed as an independent Helm chart:

| Chart                | Purpose                                             |
| -------------------- | --------------------------------------------------- |
| `charts/api/`        | API Deployment + ClusterIP Service + Secret         |
| `charts/scheduler/`  | Scheduler CronJobs (check-triggers, send-reminders) |
| `charts/cloudflare/` | Cloudflare Tunnel Deployment                        |

Charts are installed separately so they can be upgraded independently. Commands are in
[deploy-api.md](../.claude/skills/ops/references/deploy-api.md).

## Networking

```text
Internet → Cloudflare Edge → cloudflared pod → ClusterIP Service → API pod
```

- No Load Balancer or NodePort needed
- TLS terminated at Cloudflare edge
- Tunnel configured in Cloudflare dashboard to route domain → `localhost:8080`
- Cost: $0

## Deploys and CI

CI tests the Go services and typechecks the app; it doesn't deploy. Deploys are manual. Steps,
secrets and cost: [deploy-api.md](../.claude/skills/ops/references/deploy-api.md).
