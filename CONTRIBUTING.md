# Contributing

Thanks for contributing to Freebies: LA Sports Deals.

## Before you start

- Read the [Code of Conduct](CODE_OF_CONDUCT.md) and [AI policy](AI_POLICY.md).
- Check existing issues and pull requests before starting work.
- Keep the API backwards compatible. Changes to endpoints, the database schema or push payloads
  reach phones that have not updated yet; discuss breaking behavior first.
- Before any change a user can see (screens, sprites, copy, pushes, store metadata), read
  [docs/ui.md](docs/ui.md). It is binding.
- Keep changes focused, and describe what you actually verified.

## Development setup

Start with the [quickstart](README.md#quickstart). [mise](https://mise.jdx.dev) installs the tools
pinned in [.mise.toml](.mise.toml) (Go, Node, Bun, just, git-cliff, kubectl and helm):

```sh
mise install
just setup
```

Run `just --list` for development commands. The [justfile](justfile) defines their behavior.

The how-to lives in two [Claude skills](https://agentskills.io) under `.claude/skills/`. People can
read their reference files directly:

| Job                                       | Read                                                                                       |
| ----------------------------------------- | ------------------------------------------------------------------------------------------ |
| Setup, running the API, tests, generation | [local-dev.md](.claude/skills/ops/references/local-dev.md)                                 |
| Running the app                           | [local-dev-mobile.md](.claude/skills/ops/references/local-dev-mobile.md)                   |
| Something fails locally                   | [local-dev-troubleshooting.md](.claude/skills/ops/references/local-dev-troubleshooting.md) |
| Adding, changing or retiring a deal       | [deals.md](.claude/skills/ops/references/deals.md)                                         |
| Deploying the API and scheduler           | [deploy-api.md](.claude/skills/ops/references/deploy-api.md)                               |
| Releasing the iOS app                     | [release-mobile.md](.claude/skills/ops/references/release-mobile.md)                       |
| Screens, UI copy and sprites              | [design skill](.claude/skills/design/SKILL.md)                                             |

## Configuration

The API reads `FREEBIE_*` environment variables:

| Variable                | Description                             | Default      |
| ----------------------- | --------------------------------------- | ------------ |
| `FREEBIE_DATABASE_PATH` | SQLite path, or a `libsql://` Turso URL | `freebie.db` |
| `FREEBIE_SERVER_HOST`   | Bind address                            | `0.0.0.0`    |
| `FREEBIE_SERVER_PORT`   | Port                                    | `8080`       |

`just api-serve` runs from `services/api`, so the local database is `services/api/freebie.db`. A
local `.envrc` may export production values; unset `FREEBIE_DATABASE_PATH` before serving locally,
or the local server talks to Turso. Feature flags and production data are in
[docs/operations.md](docs/operations.md).

## Project structure

```text
apps/mobile/                  React Native (Expo) app
services/api/                 Go API server
  db/queries.sql              sqlc queries
  internal/db/migrations/     goose migrations, run on server start
  internal/api/               HTTP handlers and routes
  internal/sources/           live game data sources (Source interface)
services/scheduler/           scheduler that calls the API from Kubernetes CronJobs
charts/                       Helm charts: api, scheduler, cloudflare
docs/                         architecture, UI rules and operations
.claude/skills/               ops and design workflows
asset/                        README artwork
```

[docs/architecture.md](docs/architecture.md) describes how the pieces fit.

## Code standards

**Content is server-driven.** Deals, teams, leagues, seasons, sprites and sport badges ship through
the API as database rows, never through an app release. Do not hardcode team, league or sprite lists
in `apps/mobile`; bundled art is an offline fallback only. See
[Server-driven content](docs/architecture.md#server-driven-content).

- **Go**: format with `gofmt` and keep `go vet` clean. Database access goes through sqlc: edit
  `services/api/db/queries.sql`, then run `just api-generate`. Commit generated files with the
  change that caused them.
- **New game sources** implement the `Source` interface (including `Metrics()`) in
  `services/api/internal/sources/` and are registered in the worker. Pick the provider with
  [data-sources.md](.claude/skills/ops/references/data-sources.md); run `go run . deals audit`
  before shipping a deal.
- **Deals** are LA only (`region_code` `us-ca-la` or `us-ca-oc`, an LA or Orange County team).
  Verify a deal on the sponsor's or team's own page first; anything unverified ships with
  `is_active = 0`. [deals.md](.claude/skills/ops/references/deals.md) covers the migration and
  trigger rule.
- **Database changes** add the next numbered goose migration. Never edit a migration that has
  shipped; add a new one.
- **UI** follows [docs/ui.md](docs/ui.md): NES palette tokens only, integer pixel scale, sprite
  rules, motion limits, accessibility and the legal guardrails (no team logos or colors, no
  "Nintendo" or "NES" in user-facing copy). Change that file in the same pull request as code that
  breaks one of its rules.
- **Sprites** are drawn, previewed and shipped with the design skill. Never edit the generated SQL.
  Every sprite must be original: generic food, no restaurant packaging or logos, nothing traced from
  a game.
- **TypeScript**: follow the existing patterns in `apps/mobile/`.

## Testing

```sh
just test               # Go tests for the API and the scheduler
just mobile-typecheck   # tsc --noEmit for the app
just md-lint            # markdownlint, as CI runs it
just just-fmt-check     # justfile formatting
```

The ops skill's `scripts/check.sh` runs `go vet`, the Go tests and the typecheck together. CI runs
`go vet` and `go test` for both services and the mobile typecheck. Add tests for new behavior.
`TestOnlyLAEventsActive` lists every active deal, so it fails until you update it for a new or
retired one.

A passing test does not prove a push arrives or a screen looks right. Say which checks ran and which
you did in the simulator or on a device.

## Before committing

Run `just ready`. It checks the justfile and Markdown, then runs the Go tests and the mobile
typecheck. `just docs-fmt` formats Markdown to 100 columns with prettier (`just docs-install`
first). Update documentation when behavior or commands change.

## Branching

Create a branch from `main` before committing, named `type/short-description`:

| Prefix      | Use for                              |
| ----------- | ------------------------------------ |
| `feat/`     | New features                         |
| `fix/`      | Bug fixes                            |
| `docs/`     | Documentation changes                |
| `refactor/` | Code restructuring (no new behavior) |
| `chore/`    | Tooling, dependencies, CI            |

Never commit directly to `main`. Agents must not commit unless the user explicitly asks, and pushing
requires separate instruction.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org) with the 50/72 rule: a subject of at
most 50 characters in the imperative mood, capitalized and without a period, and a body wrapped at
72 that says what changed and why.

- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
- **Scopes**: `api`, `scheduler`, `mobile`, `docs`, `db`

```text
feat(api): Add NBA game source
fix(mobile): Prevent duplicate push registrations
docs: Add branching and commit conventions
```

Most pull requests should be a single commit; squash and rebase before review.

## Submitting a pull request

Describe the problem, the resulting behavior and how you validated it. Include screenshots for UI
changes, and name any rule in [docs/ui.md](docs/ui.md) you changed. Open a draft for early feedback.
Keep one logical change per pull request. Follow [AI_POLICY.md](AI_POLICY.md) for AI assistance.

## Releases

A `vMAJOR.MINOR.PATCH` tag triggers the [release workflow](.github/workflows/release.yml). It runs
the Go tests and publishes a GitHub Release whose notes come from Git history.
[cliff.toml](cliff.toml) defines the Conventional Commit groups and exclusions. Preview unreleased
notes with `just release-notes`, which uses git-cliff from `mise install`.

The release carries notes only. The API and scheduler ship as a container image to Kubernetes
([deploy-api.md](.claude/skills/ops/references/deploy-api.md)), and the iOS app ships through EAS
and TestFlight ([release-mobile.md](.claude/skills/ops/references/release-mobile.md)). A maintainer
must explicitly authorize tagging, deploying or submitting a build.
