# Deploying the API and scheduler

Every command that changes production needs the owner's go-ahead. Never deploy on your own
initiative.

## Steps

1. Run `scripts/deploy-preflight.sh` (add `--helm` for a chart or value change). It's read-only and
   checks the charts, docker, the DigitalOcean login, the registry (and that it has room for a
   push), the cluster, the kubectl context and the pull secret.
2. If the deploy adds migrations, back up Turso first (they run on start and can't be rolled back):
   `turso db shell freebie .dump > freebie-$(date +%F).sql`, kept outside the repo.
3. Get the owner's go-ahead.
4. Code-only change: `scripts/deploy.sh --yes`. It narrates six numbered steps (checks, registry
   login, build, push, restart, verify) and ends with a summary of the migrations applied and the
   health check; the docker output goes to a log printed only on failure. Relay each step to the
   owner as it runs, then report the summary. Chart or value change: build and push, then Helm
   upgrade ("Deploy" below, step 2b).
5. If it's broken, follow [deploy-api-troubleshooting.md](deploy-api-troubleshooting.md) or roll
   back.

Production data fixes and feature flags are in `docs/operations.md`; the design rationale is in
`docs/architecture.md`.

## Overview

```text
Internet → Cloudflare edge (TLS) → Cloudflare Tunnel → cloudflared pod
        → freebie-api ClusterIP :8080 → API pod → Turso (libsql://...)
CronJobs (scheduler) → HTTP POST http://freebie-api:8080/internal/worker/* (bearer token)
```

### What runs

| Piece     | Chart                | Kubernetes object                                               | Image                                                                          |
| --------- | -------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| API       | `charts/api/`        | Deployment `freebie-api` (1 replica), ClusterIP Service, Secret | `registry.digitalocean.com/freebie/freebie-api:latest`, runs `./freebie serve` |
| Scheduler | `charts/scheduler/`  | CronJobs `check-triggers` and `send-reminders`, Secret          | Same image, runs `./scheduler <job>`                                           |
| Tunnel    | `charts/cloudflare/` | Deployment `cloudflared`, Secret                                | `cloudflare/cloudflared`                                                       |

- Namespace: `freebies`. Cluster: DOKS, one `s-1vcpu-2gb` node, AMD64.
- The root `Dockerfile` builds both binaries into one distroless image (no shell).
- Charts use tag `latest` with `pullPolicy: Always`, so a restart picks up a new push.
- The API runs goose migrations against Turso on startup, so a deploy applies new migrations.
- Liveness and readiness probes hit `/healthz`.

### Schedules

Set in `charts/scheduler/values.yaml` as Pacific wall-clock times (`timeZone: America/Los_Angeles`
on the CronJobs), so they stay put when daylight saving time starts or ends:

| Job              | Cron (Pacific)     | Runs           | Calls                                                     |
| ---------------- | ------------------ | -------------- | --------------------------------------------------------- |
| `check-triggers` | `0 7,9 * * *`      | 7am, 9am retry | `POST /internal/worker/check-triggers` (dedupes per game) |
| `send-reminders` | `0 11,15,17 * * *` | 11am, 3pm, 5pm | `POST /internal/worker/send-reminders`                    |

7am lands before breakfast-only deadlines (McGriddles until 10:30) and after last night's late
finals. At-the-game deals (`game_night`) are never pushed; deals with a later start
(`available_from_local`) show as locked until then. Changing a schedule is a chart change: deploy it
with the Helm upgrade (step 2b), not `deploy.sh`.

### Secrets

| Secret                    | Used by                                                                 | Source                                             |
| ------------------------- | ----------------------------------------------------------------------- | -------------------------------------------------- |
| Turso URL with auth token | API (`FREEBIE_DATABASE_PATH`)                                           | `turso db show --url` and `turso db tokens create` |
| Worker secret             | API (`FREEBIE_WORKER_SECRET`) and scheduler (`SCHEDULER_WORKER_SECRET`) | Any long random string, the same in both           |
| Tunnel token              | cloudflared (`TUNNEL_TOKEN`)                                            | Cloudflare Zero Trust dashboard                    |

Helm stores them as Kubernetes Secrets from `--set` values. They live in the owner's `.envrc`, never
in git.

### Public URL

`https://freebie-api.bitsugar.io`. Release builds of the app call `/api/v1` there.

### Cost

| Component                          | Cost        |
| ---------------------------------- | ----------- |
| DOKS cluster (1 node, s-1vcpu-2gb) | ~$12/mo     |
| DOCR (starter tier)                | Free        |
| Turso (free tier)                  | Free        |
| Cloudflare Tunnel                  | Free        |
| **Total**                          | **~$12/mo** |

## Deploy

Every command here changes production. Run them only with the owner's go-ahead, from the repo root,
with the deploy variables exported (see [deploy-api-setup.md](deploy-api-setup.md)).

### 1. Build and push the image

DOKS nodes are AMD64, so always build for `linux/amd64`, even on an ARM Mac.

```bash
doctl registry login
docker build --platform linux/amd64 -t registry.digitalocean.com/freebie/freebie-api:latest .
docker push registry.digitalocean.com/freebie/freebie-api:latest
```

Same as `just docker-build docker-push`.

### 2a. Code change only: restart the API

```bash
kubectl rollout restart deployment/freebie-api -n freebies
kubectl rollout status deployment/freebie-api -n freebies
```

`scripts/deploy.sh --yes` runs steps 1 and 2a with the checks around them; `just deploy` runs the
bare commands. CronJob pods pull `latest` on every run, so the scheduler needs nothing more.

### 2b. Chart or value change: Helm upgrade

Upgrade only the chart that changed:

```bash
helm upgrade --install freebie-api charts/api/ \
  --namespace freebies --create-namespace \
  --set databasePath="$FREEBIE_DATABASE_PATH" \
  --set workerSecret="$FREEBIE_WORKER_SECRET"

helm upgrade --install freebie-scheduler charts/scheduler/ \
  --namespace freebies \
  --set workerSecret="$FREEBIE_WORKER_SECRET"

helm upgrade --install freebie-cloudflare charts/cloudflare/ \
  --namespace freebies \
  --set tunnelToken="$CF_TUNNEL_TOKEN"
```

Every upgrade must pass the secrets again; Helm doesn't keep `--set` values from the last release
unless you add `--reuse-values`.

A first install runs all three, API first.

### 3. Check it

```bash
.claude/skills/ops/scripts/deploy-status.sh
kubectl logs -n freebies -l app.kubernetes.io/name=freebie-api --tail=50
```

The new pod's log shows the migrations, then `starting server`. See
[deploy-api-troubleshooting.md](deploy-api-troubleshooting.md) if it doesn't.

### Roll back

The image tag is always `latest`, so `kubectl rollout undo` brings back the old pod spec but still
pulls the current image. To roll back code:

1. Check out the last good commit.
2. Build and push it as `latest` (step 1), then restart (step 2a).

For a bad chart change: `helm history <release> -n freebies`, then
`helm rollback <release> <revision> -n freebies`.

A migration that already ran on Turso isn't undone by a rollback. Fix it forward with a new
migration.

### Uninstall

```bash
helm uninstall <freebie-api|freebie-scheduler|freebie-cloudflare> -n freebies
```
