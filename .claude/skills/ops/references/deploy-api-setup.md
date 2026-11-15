# Deploying: one-time setup and CI

Cluster, registry, database, tunnel and secrets, done once. Day-to-day deploys are in
[deploy-api.md](deploy-api.md).

## One-time setup

Done once per cluster. Everything here creates or changes cloud resources: the owner runs it, or
approves each step.

### Prerequisites

1. Tools: `doctl`, `docker`, `turso`, and `kubectl` and `helm` (from `.mise.toml`).
2. A DigitalOcean account with an API token (`doctl auth init`).
3. A Turso account at [turso.tech](https://turso.tech).
4. A Cloudflare account with the domain on it.

### Deploy variables

Set in `.envrc` (direnv) and read by the Helm commands:

```bash
export FREEBIE_DATABASE_PATH="libsql://freebie-xxx.turso.io?authToken=xxx"
export FREEBIE_WORKER_SECRET="a-long-random-string"
export CF_TUNNEL_TOKEN="the-tunnel-token"
```

`FREEBIE_DATABASE_PATH` also switches a local `freebie serve` to Turso, so `.envrc` keeps it
commented out. Export it only in the shell you deploy from.

### 1. kubectl

```bash
doctl kubernetes cluster list
doctl kubernetes cluster kubeconfig save <cluster-name>
```

### 2. Container registry

```bash
doctl registry create freebie --subscription-tier starter
doctl registry login
doctl kubernetes cluster registry add <cluster-name>     # lets the cluster pull images
```

### 3. Turso database

```bash
turso auth login
turso db create freebie
turso db show freebie --url
turso db tokens create freebie
```

Join them as `libsql://<host>?authToken=<token>` for `FREEBIE_DATABASE_PATH`. The API migrates the
empty database on its first start.

### 4. Cloudflare Tunnel

1. In the [Cloudflare Zero Trust dashboard](https://one.dash.cloudflare.com/), create a tunnel named
   `freebie-api`.
2. Add a public hostname: subdomain `freebie-api`, domain `bitsugar.io`, type `HTTP`, URL
   `freebie-api.freebies.svc.cluster.local:8080`.
3. Copy the tunnel token into `CF_TUNNEL_TOKEN`.

### 5. First install

Build and push the image, then install all three charts: see [deploy-api.md](deploy-api.md).

## CI

### What runs today

| Workflow                                    | Trigger                                                | Does                                             |
| ------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------ |
| `.github/workflows/go.yml`                  | push to `main`, pull requests touching Go, SQL or YAML | `go vet` and `go test` for the api and scheduler |
| `.github/workflows/mobile.yml`              | push to `main`, pull requests                          | `bun install --frozen-lockfile`, `tsc --noEmit`  |
| `.github/workflows/commit-lint.yml`         | pull requests                                          | Conventional Commits check                       |
| `labeler.yml`, `greetings.yml`, `stale.yml` | issues and pull requests                               | Housekeeping                                     |

None of them builds an image or deploys. Merging to `main` changes nothing in production until
someone runs the steps in [deploy-api.md](deploy-api.md).

### If a deploy workflow is added

Older docs describe a `.github/workflows/deploy.yaml` that never landed. A deploy job would need
these repository secrets:

| Secret                      | Purpose                                         |
| --------------------------- | ----------------------------------------------- |
| `DIGITALOCEAN_ACCESS_TOKEN` | doctl and registry login                        |
| `KUBECONFIG_DATA`           | Base64-encoded kubeconfig                       |
| `TURSO_DATABASE_URL`        | `libsql://...?authToken=...` for `databasePath` |
| `CF_TUNNEL_TOKEN`           | Cloudflare Tunnel token                         |
| `WORKER_SECRET`             | Bearer token shared by the api and scheduler    |

Tag images with the commit SHA and pass it as `--set image.tag=<sha>`, so a rollback is a Helm
rollback. Pin every action to a commit SHA, as the existing workflows do.
