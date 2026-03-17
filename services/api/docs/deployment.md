# Deployment

The backend is deployed to [DigitalOcean Kubernetes](https://www.digitalocean.com/products/kubernetes)
(DOKS) with [Turso](https://turso.tech) for the database and
[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)
for external access.

See the [top-level architecture doc](../../../docs/architecture.md) for the full infrastructure
diagram.

## Prerequisites

1. Install tools: `doctl`, `kubectl`, `helm`
2. DigitalOcean account with API token
3. Turso account at [turso.tech](https://turso.tech)
4. Cloudflare account with a tunnel configured
5. Environment variables set in `.envrc` (see below)

### Required Environment Variables

These are set in `.envrc` and used by Helm deploy commands:

```bash
export FREEBIE_DATABASE_PATH="libsql://freebie-xxx.turso.io?authToken=xxx"
export FREEBIE_WORKER_SECRET="your-secret-here"
export CF_TUNNEL_TOKEN="your-tunnel-token"
export SCHEDULER_API_URL="http://freebie-api.freebies.svc.cluster.local:8080"
export SCHEDULER_WORKER_SECRET="$FREEBIE_WORKER_SECRET"
```

## Initial Setup

### 1. Configure kubectl

```bash
doctl kubernetes cluster kubeconfig save <cluster-name>
```

### 2. Create Container Registry

```bash
# Create registry (starter tier is free)
doctl registry create freebie --subscription-tier starter

# Login to registry
doctl registry login

# Integrate registry with K8s cluster
doctl kubernetes cluster list                          # get cluster name
doctl kubernetes cluster registry add <cluster-name>   # allow cluster to pull images
```

### 3. Create Turso Database

```bash
# Install Turso CLI
brew install tursodatabase/tap/turso

# Login and create database
turso auth login
turso db create freebie

# Get connection URL and token
turso db show freebie --url
turso db tokens create freebie
```

### 4. Create Cloudflare Tunnel

1. Go to [Cloudflare Zero Trust dashboard](https://one.dash.cloudflare.com/)
2. Create a tunnel named `freebie-api`
3. Add a public hostname route:
   - Subdomain: `freebie-api` (or whatever you want)
   - Domain: your domain (e.g. `bitsugar.io`)
   - Type: `HTTP`
   - URL: `freebie-api.freebies.svc.cluster.local:8080`
4. Copy the tunnel token for `.envrc`

## Build and Push Docker Images

Build for linux/amd64 (required — DOKS nodes are AMD64, not ARM). A single image contains both
the API and scheduler binaries:

```bash
docker build --platform linux/amd64 \
  -t registry.digitalocean.com/freebie/freebie-api:latest .

docker push registry.digitalocean.com/freebie/freebie-api:latest
```

Note: The Dockerfile is at the repo root and builds both `services/api/` and `services/scheduler/`
into one image. The scheduler Helm chart uses the same image but runs `./scheduler` instead of
`./freebie serve`.

## Deploy

### Deploy API

```bash
helm upgrade --install freebie-api charts/api/ \
  --namespace freebies --create-namespace \
  --set databasePath="$FREEBIE_DATABASE_PATH" \
  --set workerSecret="$FREEBIE_WORKER_SECRET"
```

### Deploy Cloudflare Tunnel

```bash
helm upgrade --install freebie-cloudflare charts/cloudflare/ \
  --namespace freebies \
  --set tunnelToken="$CF_TUNNEL_TOKEN"
```

### Deploy Scheduler (CronJobs)

```bash
helm upgrade --install freebie-scheduler charts/scheduler/ \
  --namespace freebies \
  --set workerSecret="$FREEBIE_WORKER_SECRET"
```

### Deploy Everything

```bash
# Build and push image (contains both API and scheduler binaries)
docker build --platform linux/amd64 \
  -t registry.digitalocean.com/freebie/freebie-api:latest .
docker push registry.digitalocean.com/freebie/freebie-api:latest

# Deploy all charts
helm upgrade --install freebie-api charts/api/ \
  --namespace freebies --create-namespace \
  --set databasePath="$FREEBIE_DATABASE_PATH" \
  --set workerSecret="$FREEBIE_WORKER_SECRET"

helm upgrade --install freebie-cloudflare charts/cloudflare/ \
  --namespace freebies \
  --set tunnelToken="$CF_TUNNEL_TOKEN"

helm upgrade --install freebie-scheduler charts/scheduler/ \
  --namespace freebies \
  --set workerSecret="$FREEBIE_WORKER_SECRET"
```

## Upgrade

After code changes, rebuild and redeploy:

```bash
# Rebuild and push image
docker build --platform linux/amd64 \
  -t registry.digitalocean.com/freebie/freebie-api:latest .
docker push registry.digitalocean.com/freebie/freebie-api:latest

# Restart API pod to pull new image
kubectl rollout restart deployment/freebie-api -n freebies

# Or force Helm to update (e.g. if values changed)
helm upgrade freebie-api charts/api/ \
  --namespace freebies \
  --set databasePath="$FREEBIE_DATABASE_PATH" \
  --set workerSecret="$FREEBIE_WORKER_SECRET"
```

CronJob pods always pull the latest image on each run, so they pick up changes automatically.

## Verify

```bash
# Check pod status
kubectl get pods -n freebies

# Watch pods
kubectl get pods -n freebies -w

# View API logs
kubectl logs -n freebies -l app.kubernetes.io/name=freebie-api

# View CronJob history
kubectl get jobs -n freebies

# Test API locally via port-forward
kubectl port-forward -n freebies svc/freebie-api 8080:8080
curl http://localhost:8080/healthz

# Test API via Cloudflare Tunnel
curl https://freebie-api.bitsugar.io/healthz
```

## Troubleshooting

```bash
# Describe pod (shows events, errors, image pull issues)
kubectl describe pod -n freebies <pod-name>

# Check image pull errors
kubectl get events -n freebies --sort-by=.lastTimestamp

# View API logs (follow)
kubectl logs -n freebies -l app.kubernetes.io/name=freebie-api -f

# Exec into pod (note: distroless has no shell, use debug container)
kubectl debug -n freebies <pod-name> --image=busybox -it

# Restart a deployment
kubectl rollout restart deployment/freebie-api -n freebies

# Check Helm releases
helm list -n freebies

# Uninstall a chart
helm uninstall freebie-api -n freebies
```

### Common Issues

**ErrImagePull / `no match for platform in manifest`**
You built on ARM (Mac) but the cluster is AMD64. Rebuild with `--platform linux/amd64`.

**ErrImagePull / `repository not found`**
The registry isn't linked to the cluster. Run:
```bash
doctl kubernetes cluster registry add <cluster-name>
```

**Pod stuck in CrashLoopBackOff**
Check logs: `kubectl logs -n freebies <pod-name>`
Usually a bad database URL or missing env var.

## CI/CD

GitHub Actions (`.github/workflows/deploy.yaml`) handles deployments automatically on push to
`main`. See the [architecture doc](../../../docs/architecture.md#cicd) for details.

Required GitHub Secrets:

| Secret | Purpose |
|--------|---------|
| `DIGITALOCEAN_ACCESS_TOKEN` | DO API token for doctl + DOCR login |
| `KUBECONFIG_DATA` | Base64-encoded kubeconfig |
| `TURSO_DATABASE_URL` | Turso connection URL with auth token |
| `CF_TUNNEL_TOKEN` | Cloudflare Tunnel token |
| `WORKER_SECRET` | Bearer token for internal worker API |

## Cost

| Component | Cost |
|-----------|------|
| DOKS cluster (1 node, s-1vcpu-2gb) | ~$12/mo |
| DOCR (starter tier) | Free |
| Turso (free tier) | Free |
| Cloudflare Tunnel | Free |
| **Total** | **~$12/mo** |
