# Architecture

## Overview

Freebie runs on a DigitalOcean Kubernetes (DOKS) cluster. The system monitors live sports games,
detects when trigger conditions are met (e.g., 7+ strikeouts), and sends push notifications to
subscribed users.

## Infrastructure

```
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
│  │  (daily 6am PT)                            │          │
│  │                                            │          │
│  │  send-reminders ──── same pattern          │          │
│  │  (daily 6pm PT)                            │          │
│  └────────────────────────────────────────────┘          │
└──────────────────────────────────────────────────────────┘
                       │
                       ▼
         ┌──────────────────────────┐
         │  Turso (hosted SQLite)   │
         │  libsql://xxx.turso.io   │
         └──────────────────────────┘
```

**Deployed by**: Helm charts (`charts/`) + GitHub Actions (`.github/workflows/deploy.yaml`)

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

External traffic reaches the API through a [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/),
eliminating the need for a Kubernetes LoadBalancer or Ingress controller.

### How it works

1. A `cloudflared` pod runs inside the cluster as a Deployment
2. On startup, it establishes an **outbound** connection to Cloudflare's edge network
3. Cloudflare routes traffic for your domain through this tunnel to the pod
4. The pod forwards requests to the `freebie-api` ClusterIP Service on port 8080

```
User request → your-domain.com
       → Cloudflare edge (TLS termination, DDoS protection, caching)
       → Cloudflare Tunnel (encrypted, outbound-initiated)
       → cloudflared pod in cluster
       → freebie-api:8080 (ClusterIP)
```

### Why not a LoadBalancer or Ingress?

| Approach | Cost | Complexity |
|----------|------|------------|
| DO Load Balancer | ~$12/mo | Need cert-manager for TLS |
| Nginx Ingress | $0 (NodePort) | Ingress controller + cert-manager |
| Cloudflare Tunnel | $0 | Single deployment, TLS handled by CF |

### Setup

1. Create a tunnel in the [Cloudflare Zero Trust dashboard](https://one.dash.cloudflare.com/)
2. Configure the tunnel to route your domain to `http://localhost:8080`
3. Copy the tunnel token
4. The Helm chart (`charts/cloudflare/`) deploys `cloudflared` with this token:

```yaml
# charts/cloudflare/values.yaml
tunnel:
  token: ""  # Set via --set or GitHub Secrets
```

The tunnel token is stored as a Kubernetes Secret and injected as the `TUNNEL_TOKEN` environment
variable. The `cloudflared` container runs with `tunnel --no-autoupdate run`.

### Security benefits

- **No open inbound ports** — the tunnel is outbound-only from the cluster
- **No public IP needed** — the API service is ClusterIP (cluster-internal only)
- **DDoS protection** — Cloudflare's edge absorbs attacks before they reach the cluster
- **TLS everywhere** — terminated at Cloudflare edge, no cert management needed in-cluster

## Worker Flow

### Check Triggers (6am PT daily)

1. K8s CronJob starts a pod running `./scheduler check-triggers`
2. The scheduler uses a generated HTTP client to `POST /internal/worker/check-triggers`
3. The API authenticates via bearer token (`FREEBIE_WORKER_SECRET`)
4. The worker service queries Turso for active events
5. For each event, it fetches game data from the source API (e.g., MLB)
6. If a trigger condition is met, it creates a `triggered_event` record
7. It sends push notifications to subscribed users via Expo
8. Returns a JSON summary to the CronJob pod

### Send Reminders (6pm PT daily)

1. K8s CronJob starts a pod running `./scheduler send-reminders`
2. The scheduler calls `POST /internal/worker/send-reminders`
3. The API queries Turso for deals expiring within 6 hours
4. For each expiring deal, it sends reminder notifications to eligible users
5. Returns a JSON summary

## API Layers

### Public API (`/api/v1/`)

User-facing endpoints for the mobile app. Protected by per-user bearer tokens.

- Events, leagues, subscriptions, active deals, dismissals

### Internal Worker API (`/internal/worker/`)

Protected by a shared bearer token (`FREEBIE_WORKER_SECRET`). Called by CronJobs within the
cluster.

- `POST /internal/worker/check-triggers` — check game results
- `POST /internal/worker/send-reminders` — send expiring deal reminders

The internal API is defined by an **OpenAPI 3.0 spec**
(`internal/api/worker/gen/api.yaml`). Server handlers and HTTP client code are
**generated** from this spec using `oapi-codegen`:

```
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

| Chart | Purpose |
|-------|---------|
| `charts/api/` | API Deployment + ClusterIP Service + Secret |
| `charts/scheduler/` | Scheduler CronJobs (check-triggers, send-reminders) |
| `charts/cloudflare/` | Cloudflare Tunnel Deployment |

Charts are installed separately so they can be upgraded independently:

```bash
helm upgrade --install freebie-api charts/api/ --namespace freebies
helm upgrade --install freebie-scheduler charts/scheduler/ --namespace freebies
helm upgrade --install freebie-cloudflare charts/cloudflare/ --namespace freebies
```

## Networking

```
Internet → Cloudflare Edge → cloudflared pod → ClusterIP Service → API pod
```

- No Load Balancer or NodePort needed
- TLS terminated at Cloudflare edge
- Tunnel configured in Cloudflare dashboard to route domain → `localhost:8080`
- Cost: $0

## CI/CD

GitHub Actions (`.github/workflows/deploy.yaml`) runs on push to `main`:

1. Build two Docker images:
   - `freebie-api` from `services/api/Dockerfile`
   - `freebie-scheduler` from `services/scheduler/Dockerfile`
2. Push both to DOCR with SHA-based tags
3. `helm upgrade` all three charts with the new image tags

Required GitHub Secrets:

| Secret | Purpose |
|--------|---------|
| `DIGITALOCEAN_ACCESS_TOKEN` | DO API token for doctl + DOCR login |
| `KUBECONFIG_DATA` | Base64-encoded kubeconfig |
| `TURSO_DATABASE_URL` | Turso connection URL (`libsql://...?authToken=...`) |
| `CF_TUNNEL_TOKEN` | Cloudflare Tunnel token |
| `WORKER_SECRET` | Bearer token for internal worker API |
| `SCHEDULER_API_URL` | API URL for the scheduler to call |

## Cost

| Component | Cost |
|-----------|------|
| DOKS cluster (1 node, s-1vcpu-2gb) | ~$12/mo |
| DOCR (starter tier) | Free |
| Turso (free tier) | Free |
| Cloudflare Tunnel | Free |
| **Total** | **~$12/mo** |
