# Deployment

The backend is deployed to [Fly.io](https://fly.io) with [Turso](https://turso.tech) for the
database.

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   Fly.io                        │
│  ┌─────────────────┐  ┌─────────────────────┐   │
│  │   App Machine   │  │  Scheduled Machine  │   │
│  │  (HTTP server)  │  │   (hourly worker)   │   │
│  │                 │  │                     │   │
│  │  ./freebie      │  │  ./freebie          │   │
│  │  serve          │  │  worker run         │   │
│  └────────┬────────┘  └──────────┬──────────┘   │
│           │                      │              │
└───────────┼──────────────────────┼──────────────┘
            │                      │
            └──────────┬───────────┘
                       │
                       ▼
              ┌────────────────┐
              │     Turso      │
              │   (SQLite)     │
              └────────────────┘
```

- **App Machine**: Serves HTTP API, auto-starts on traffic, scales to zero
- **Scheduled Machine**: Runs hourly, checks time and runs appropriate job
  - 6am PT: Check triggers for yesterday's games
  - 6pm PT: Send reminder notifications

## Prerequisites

1. Install Fly CLI: `brew install flyctl`
2. Login: `fly auth login`
3. Create Turso account at [turso.tech](https://turso.tech)

## Initial Setup

### 1. Create Turso Database

```bash
# Install Turso CLI
brew install tursodatabase/tap/turso

# Login
turso auth login

# Create database
turso db create freebie

# Get connection URL
turso db show freebie --url

# Create auth token
turso db tokens create freebie
```

### 2. Create Fly App

```bash
cd services/backend

# Create app (first time only)
fly launch --no-deploy

# Set Turso secrets
fly secrets set FREEBIE_DATABASE_PATH="libsql://your-db.turso.io?authToken=your-token"
```

### 3. Deploy

```bash
# Deploy app machines
fly deploy

# Get the deployed image tag from fly status output:
#   Image = freebie-api:deployment-01XXXXXXXXX
fly status -a freebie-api

# Create scheduled worker (first time only)
# Use the image tag from above
fly machine run registry.fly.io/freebie-api:deployment-XXXX worker run \
  --schedule hourly -a freebie-api --region sjc
```

**Important**: The scheduled machine command (`worker run`) must be passed as positional arguments
after the image, not with `--command`. Do not use `--rm` as the machine needs to persist for future
scheduled runs.

## Ongoing Deployments

After initial setup, just run:

```bash
fly deploy
```

This updates the app machines. The scheduled machine uses the same image and will pick up changes on
its next run.

To update the scheduled machine's image immediately:

```bash
fly machine update <machine-id> --image registry.fly.io/freebie-api:deployment-XXXX -a freebie-api
```

## Managing Machines

```bash
# List all machines
fly machine list -a freebie-api

# Check machine status (shows command and schedule)
fly machine status <machine-id> -a freebie-api

# View logs
fly logs -a freebie-api

# View logs for specific machine
fly logs -a freebie-api --machine <machine-id>

# Delete a machine
fly machine destroy <machine-id> -a freebie-api --force
```

### Identifying Machine Types

- **App machines**: Have `Process Group = app`, run `serve` command
- **Scheduled machine**: Has `Command = ["worker","run"]`, no process group, state cycles between
  `stopped` and `started`

Check with: `fly machine status <machine-id> -a freebie-api`

## Environment Variables

| Variable                | Description                 | Example                                       |
| ----------------------- | --------------------------- | --------------------------------------------- |
| `FREEBIE_DATABASE_PATH` | Turso connection URL        | `libsql://freebie-xxx.turso.io?authToken=xxx` |
| `PORT`                  | HTTP port (set in fly.toml) | `8080`                                        |

## Secrets Management

```bash
# Set a secret
fly secrets set FREEBIE_DATABASE_PATH="libsql://..."

# List secrets
fly secrets list

# Remove a secret
fly secrets unset FREEBIE_DATABASE_PATH
```

## Troubleshooting

### Check logs

```bash
fly logs -a freebie-api
```

### SSH into machine

```bash
fly ssh console -a freebie-api
```

### Verify database connection

```bash
fly ssh console -a freebie-api -C "./freebie version"
```

### Restart machines

```bash
fly machine restart <machine-id> -a freebie-api
```

## Cost

- **Fly.io**: ~$5/month for smallest VM (256MB)
- **Turso**: Free tier (9GB storage, 500M reads/month)
