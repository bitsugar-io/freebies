# Deployment

The backend is deployed to [DigitalOcean Kubernetes](https://www.digitalocean.com/products/kubernetes)
(DOKS) with [Turso](https://turso.tech) for the database and
[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)
for external access.

See the [top-level architecture doc](../../../docs/architecture.md) for the full infrastructure
diagram.

## Prerequisites

1. Install tools: `doctl`, `kubectl`, `helm`, `terraform`
2. DigitalOcean account with API token
3. Turso account at [turso.tech](https://turso.tech)
4. Cloudflare account with a tunnel configured

## Initial Setup

### 1. Provision Infrastructure

```bash
cd terraform
export TF_VAR_do_token=$DIGITALOCEAN_TOKEN
terraform init && terraform apply
```

This creates the DOKS cluster and DOCR container registry.

### 2. Configure kubectl

```bash
# Get kubeconfig from Terraform output
terraform output -raw kubeconfig > ~/.kube/config-freebie

# Or use doctl
doctl kubernetes cluster kubeconfig save freebie
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
2. Create a tunnel
3. Configure a public hostname routing to `http://localhost:8080`
4. Copy the tunnel token

### 5. Deploy

```bash
# Build and push image
docker build -t registry.digitalocean.com/freebie/freebie-api:latest services/backend/
docker push registry.digitalocean.com/freebie/freebie-api:latest

# Deploy API
helm upgrade --install freebie-api charts/api/ \
  --namespace freebie --create-namespace \
  --set databasePath="libsql://freebie-xxx.turso.io?authToken=xxx" \
  --set workerSecret="your-secret-here" \
  --set image.tag=latest

# Deploy CronJobs
helm upgrade --install freebie-cronjobs charts/cronjobs/ \
  --namespace freebie \
  --set workerSecret="your-secret-here"

# Deploy Cloudflare Tunnel
helm upgrade --install freebie-cloudflare charts/cloudflare/ \
  --namespace freebie \
  --set tunnel.token="your-tunnel-token"
```

## Ongoing Deployments

CI/CD via GitHub Actions (`.github/workflows/deploy.yaml`) handles deployments automatically on
push to `main`. See the [architecture doc](../../../docs/architecture.md#cicd) for details.

## Environment Variables

| Variable                | Description          | Example                                       |
| ----------------------- | -------------------- | --------------------------------------------- |
| `FREEBIE_DATABASE_PATH` | Turso connection URL | `libsql://freebie-xxx.turso.io?authToken=xxx` |
| `FREEBIE_WORKER_SECRET` | Internal API token   | (any strong random string)                    |
| `FREEBIE_SERVER_HOST`   | Server bind address  | `0.0.0.0`                                     |
| `FREEBIE_SERVER_PORT`   | Server port          | `8080`                                        |

## Troubleshooting

```bash
# Check pod status
kubectl get pods -n freebie

# View API logs
kubectl logs -n freebie -l app.kubernetes.io/name=freebie-api

# View CronJob history
kubectl get jobs -n freebie

# Port-forward to test locally
kubectl port-forward -n freebie svc/freebie-api 8080:8080

# Exec into pod
kubectl exec -it -n freebie deploy/freebie-api -- sh
```

## Cost

| Component | Cost |
|-----------|------|
| DOKS cluster (1 node, s-1vcpu-2gb) | ~$12/mo |
| DOCR (starter tier) | Free |
| Turso (free tier) | Free |
| Cloudflare Tunnel | Free |
| **Total** | **~$12/mo** |
