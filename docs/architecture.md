# Architecture

## Overview

Freebie runs on a DigitalOcean Kubernetes (DOKS) cluster. The system monitors live sports games,
detects when trigger conditions are met (e.g., 7+ strikeouts), and sends push notifications to
subscribed users.

## Infrastructure

```
DOKS Cluster (1 node, s-1vcpu-2gb)
│
├── Deployment: freebie-api (1 replica)
│   ├── PVC: 1Gi DO Block Storage (ReadWriteOnce)
│   ├── Container: freebie serve
│   ├── Mounts: /data/freebie.db
│   └── Exposes: ClusterIP Service :8080
│
├── CronJob: check-triggers (daily 14:00 UTC / 6am PT)
│   └── Container: freebie worker remote check-triggers
│       → POST http://freebie-api:8080/internal/worker/check-triggers
│
├── CronJob: send-reminders (daily 02:00 UTC / 6pm PT)
│   └── Container: freebie worker remote send-reminders
│       → POST http://freebie-api:8080/internal/worker/send-reminders
│
└── Deployment: cloudflared (Cloudflare Tunnel)
    └── Routes external traffic → freebie-api:8080
```

**Provisioned by**: Terraform (`terraform/`) — DOKS cluster + DOCR registry
**Deployed by**: Helm charts (`charts/`) + GitHub Actions (`.github/workflows/deploy.yaml`)
**Cost**: ~$12/mo (1 node) + $0.10/mo (1Gi PV)

## Database Access Pattern

SQLite is an embedded, file-based database. It does not support concurrent access from multiple
processes. The DO Block Storage PVC is `ReadWriteOnce` — only one pod can mount it.

**The API pod is the single gateway to SQLite.** All database reads and writes go through the API's
HTTP endpoints.

```
┌──────────────────┐      HTTP POST       ┌──────────────────────┐
│  CronJob pod     │ ──────────────────→  │  API pod             │
│  (freebie worker │  /internal/worker/   │  (freebie serve)     │
│   remote ...)    │  check-triggers      │                      │
│                  │                      │  ┌────────────────┐  │
│  NO db mount     │                      │  │ SQLite on PVC  │  │
│  NO db access    │                      │  │ /data/freebie  │  │
└──────────────────┘                      │  └────────────────┘  │
                                          └──────────────────────┘
```

CronJob pods **never** mount the database volume. They run the `freebie worker remote` CLI command,
which uses a generated HTTP client to call the API's internal worker endpoints. The worker logic
(checking triggers, querying subscribers, sending notifications) executes **inside the API
process**.

## Worker Flow

### Check Triggers (6am PT daily)

1. K8s CronJob starts a pod running `freebie worker remote check-triggers`
2. The CLI uses the generated HTTP client to `POST /internal/worker/check-triggers`
3. The API receives the request, authenticates via bearer token
4. The worker service queries SQLite for active events
5. For each event, it fetches game data from the source API (e.g., MLB)
6. If a trigger condition is met, it creates a `triggered_event` record in SQLite
7. It sends push notifications to subscribed users via Expo
8. Returns a JSON summary to the CronJob pod

### Send Reminders (6pm PT daily)

1. K8s CronJob starts a pod running `freebie worker remote send-reminders`
2. The CLI calls `POST /internal/worker/send-reminders`
3. The API queries SQLite for deals expiring within 6 hours
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
| `charts/api/` | API Deployment + ClusterIP Service + PVC + Secret |
| `charts/cronjobs/` | check-triggers and send-reminders CronJobs |
| `charts/cloudflare/` | Cloudflare Tunnel Deployment |

Charts are installed separately so they can be upgraded independently:

```bash
helm upgrade --install freebie-api charts/api/ --namespace freebie
helm upgrade --install freebie-cronjobs charts/cronjobs/ --namespace freebie
helm upgrade --install freebie-cloudflare charts/cloudflare/ --namespace freebie
```

## Networking

External traffic reaches the API through a **Cloudflare Tunnel**:

```
Internet → Cloudflare Edge → cloudflared pod → ClusterIP Service → API pod
```

- No Load Balancer or NodePort needed
- TLS terminated at Cloudflare edge
- Tunnel configured in Cloudflare dashboard to route domain → `localhost:8080`
- Cost: $0

## Terraform

Infrastructure is provisioned with Terraform (`terraform/`):

| File | Resource |
|------|----------|
| `doks.tf` | DOKS Kubernetes cluster (1 node) |
| `docr.tf` | Container Registry (starter tier, free) |
| `variables.tf` | Region, node size, K8s version |
| `outputs.tf` | Cluster endpoint, kubeconfig, registry URL |

```bash
cd terraform
export TF_VAR_do_token=$DIGITALOCEAN_TOKEN
terraform init && terraform apply
```

## CI/CD

GitHub Actions (`.github/workflows/deploy.yaml`) runs on push to `main`:

1. Build Docker image from `services/backend/Dockerfile`
2. Push to DOCR with SHA-based tag
3. `helm upgrade` all three charts with the new image tag

Required GitHub Secrets:

| Secret | Purpose |
|--------|---------|
| `DIGITALOCEAN_ACCESS_TOKEN` | DO API token for doctl + DOCR login |
| `KUBECONFIG_DATA` | Base64-encoded kubeconfig from Terraform |
| `CF_TUNNEL_TOKEN` | Cloudflare Tunnel token |
| `WORKER_SECRET` | Bearer token for internal worker API |
