---
name: ops
description:
  Run, test, change and ship Freebies (LA sports deals app). Covers first-time setup (mise, just),
  starting, restarting and stopping the Go API on :8080 and the Expo app in the iOS simulator or
  Expo Go, resetting freebie.db, Go tests and the mobile typecheck, sqlc generation, adding,
  updating or retiring an LA deal (events migration, trigger_rule, LA-only rules, verification),
  deploying the API and scheduler to DigitalOcean Kubernetes (Docker image, Helm charts, Turso,
  Cloudflare Tunnel, CronJobs), and releasing the iOS app with EAS build, TestFlight and App Store
  submit. Use when asked to set up, run, test, add a deal or offer, deploy, check production, or
  release the app.
compatibility:
  Designed for Claude Code in the freebies repo. Needs mise (Go, Node, Bun, just, kubectl, helm),
  curl and sqlite3; deploys need docker, doctl and the secrets in .envrc; releases need eas-cli.
metadata:
  author: bitsugar
  version: "1.0"
---

# ops

Every repo workflow that isn't UI design. The API (`services/api`) migrates its SQLite database on
start, the scheduler calls it on a cron, and the Expo app (`apps/mobile`) talks to it. Run tools
through `mise exec --` (AGENTS.md explains why).

## Route

| Need                                                                           | Read                                                                                 |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| Setup, running the API, tests, `just api-generate`                             | [references/local-dev.md](references/local-dev.md)                                   |
| Running the app: simulator, Expo Go, dev client                                | [references/local-dev-mobile.md](references/local-dev-mobile.md)                     |
| Something fails locally                                                        | [references/local-dev-troubleshooting.md](references/local-dev-troubleshooting.md)   |
| Adding, changing or retiring a deal                                            | [references/deals.md](references/deals.md)                                           |
| A deal's `trigger_rule`: metrics, windows, deadlines                           | [references/deals-trigger-rule.md](references/deals-trigger-rule.md)                 |
| Deploying the API and scheduler                                                | [references/deploy-api.md](references/deploy-api.md)                                 |
| One-time cluster setup, secrets, CI                                            | [references/deploy-api-setup.md](references/deploy-api-setup.md)                     |
| Checking production, deploy failures                                           | [references/deploy-api-troubleshooting.md](references/deploy-api-troubleshooting.md) |
| EAS builds, versioning, TestFlight, App Store                                  | [references/release-mobile.md](references/release-mobile.md)                         |
| Sprites, screens and UI copy                                                   | the `design` skill                                                                   |
| Where deal, team, sprite and badge data must live (never hardcoded in the app) | `docs/architecture.md#server-driven-content`                                         |
| Store listing copy                                                             | `apps/mobile/APP-STORE.md` wins                                                      |
| Branch, commit and code-style rules                                            | `CONTRIBUTING.md` wins                                                               |

## Workflow

1. Pick the job from the Route table and read its reference file.
2. Use the job's scripts below rather than retyping the commands.
3. Deploys, EAS builds and submits change production or spend money: get the owner's go-ahead first,
   then tell the owner what each step is doing as it runs.
4. Finish with `scripts/check.sh` and fix every failure.

## Scripts

- `scripts/api.sh start|stop|restart|status|logs [-f]`: runs the API in the background, pid and log
  under `$TMPDIR`.
- `scripts/mobile.sh [expo args]`: starts Expo for Expo Go with a clean Metro cache.
- `scripts/check.sh [api|scheduler|mobile ...]`: Go vet and tests plus the mobile typecheck.
- `scripts/new-migration.sh <slug> [--print]`: next numbered migration with an `events` INSERT
  template.
- `scripts/deploy-preflight.sh [--helm]`: read-only; charts, docker, DigitalOcean login, registry
  space, cluster, kubectl context and pull secret; `--helm` also checks the deploy variables.
- `scripts/deploy.sh --yes`: code-only production deploy: preflight, build, push, restart, wait for
  the rollout, `/healthz`. Owner's go-ahead first.
- `scripts/deploy-status.sh [--health]`: read-only; pods, CronJobs, Jobs, Helm releases and the
  public `/healthz`.
- `scripts/release-preflight.sh [--no-typecheck]`: read-only; checks `app.json`, `eas.json`, the
  icon and the typecheck.
