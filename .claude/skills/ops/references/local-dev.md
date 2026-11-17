# Local development

Run every tool through `mise exec --` (AGENTS.md explains why). Scripts are in
`.claude/skills/ops/scripts/`.

## First-time setup

### Tools

`.mise.toml` pins the tools. Install them once:

```bash
mise install
```

- **Go** for the API and the scheduler.
- **Node.js** as the runtime for Expo.
- **Bun** as the package manager for the app and for the docs tooling.
- **just** runs the workflows; the recipes are in the root `justfile`.
- **git-cliff** previews release notes (`just release-notes`).
- **kubectl** and **helm** for deploys only (see [deploy-api.md](deploy-api.md)).

### Install dependencies

```bash
mise exec -- just setup          # bun install in apps/mobile
mise exec -- just docs-install   # prettier for markdown, only if you edit docs
```

Go modules download on the first `go run` or `go test`.

### The recipes

Run `just --list` for all of them. The ones you use most:

| Recipe                  | What it does                                                   |
| ----------------------- | -------------------------------------------------------------- |
| `just setup`            | Installs the app's dependencies                                |
| `just clean`            | Deletes `services/api/freebie.db`; the next serve recreates it |
| `just api-serve`        | Runs the API in the foreground on :8080                        |
| `just api-dev`          | Same, rebuilding on change (needs `watchexec`)                 |
| `just api-test`         | Go tests for the API                                           |
| `just api-generate`     | Regenerates sqlc code from `services/api/db/queries.sql`       |
| `just api-deals`        | Creates three test deals (24h, 6h, 2h)                         |
| `just scheduler-test`   | Go tests for the scheduler                                     |
| `just mobile-serve`     | Expo dev server for Expo Go                                    |
| `just mobile-typecheck` | `tsc --noEmit` for the app                                     |
| `just docs-fmt`         | Formats every markdown file to 100 columns                     |

### Environment

| Variable                | Description                             | Default      |
| ----------------------- | --------------------------------------- | ------------ |
| `FREEBIE_DATABASE_PATH` | SQLite path, or a `libsql://` Turso URL | `freebie.db` |
| `FREEBIE_SERVER_HOST`   | Bind address                            | `0.0.0.0`    |
| `FREEBIE_SERVER_PORT`   | Port                                    | `8080`       |

`.envrc` may export production values. Unset `FREEBIE_DATABASE_PATH` before serving locally, or the
local server talks to Turso.

## The API server

### Start

```bash
.claude/skills/ops/scripts/api.sh start     # background, waits for /healthz
mise exec -- just api-serve                       # or foreground, in its own terminal
```

On start the server:

1. Opens `services/api/freebie.db`, creating it if needed.
2. Runs every goose migration in `services/api/internal/db/migrations/` (schema and seed deals).
3. Subscribes existing team followers to any new deals for their teams.
4. Listens on `http://localhost:8080`.

`api.sh` builds the binary to `$TMPDIR/freebies-api/freebie` and runs it from `services/api`, so it
uses the same database as `just api-serve`. The pid and log are in `$TMPDIR/freebies-api/`.

### Check it

```bash
curl -fsS localhost:8080/healthz
curl -s localhost:8080/api/v1/events | head -c 400
.claude/skills/ops/scripts/api.sh status
.claude/skills/ops/scripts/api.sh logs -f
```

### Restart and stop

Restart after any Go or migration change. Migrations only run at startup.

```bash
.claude/skills/ops/scripts/api.sh restart
.claude/skills/ops/scripts/api.sh stop
```

`stop` only kills the server that `api.sh` started. If something else holds :8080 (a
`just api-serve` in another terminal), `status` names the process. Stop that one where it runs, or
`kill <pid>`.

### Reset the database

```bash
.claude/skills/ops/scripts/api.sh stop
mise exec -- just clean          # removes freebie.db and its -wal/-shm/-journal files
.claude/skills/ops/scripts/api.sh start
```

Reset after editing a migration that has already run locally. Never edit one that has shipped; add a
new migration instead.

### Handy CLI commands

Run from `services/api` with `mise exec -- go run . <command>`:

| Command                                                                                          | What it does                                                                                                                                     |
| ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `users list`                                                                                     | Lists users                                                                                                                                      |
| `deals list`                                                                                     | Lists active triggered deals                                                                                                                     |
| `deals create --hours 6`                                                                         | Creates a test deal                                                                                                                              |
| `deals trigger <event-id> [--notify=false] [--hours N] [--unlock-after-days N] [--valid-days N]` | Triggers one event as if its team did it last night, with the deal's own rule window; `--hours` forces a plain window, the other flags build one |
| `notify test` / `notify send <user-id>`                                                          | Sends a test push                                                                                                                                |
| `verify-mlb --team LAD --date 2026-09-01`                                                        | Reads one game from the live MLB source, no DB                                                                                                   |

The full list is in `services/api/docs/cli.md`.

## Tests, typecheck and code generation

### Run everything

```bash
.claude/skills/ops/scripts/check.sh               # api, scheduler and mobile
.claude/skills/ops/scripts/check.sh api mobile    # a subset
```

The same checks one by one:

```bash
mise exec -- just test
(cd apps/mobile && mise exec -- bunx tsc --noEmit)
```

CI runs `go vet` and `go test` for both Go services (`.github/workflows/go.yml`) and the mobile
typecheck (`.github/workflows/mobile.yml`). A pull request must pass both.

### Lint and format

```bash
mise exec -- just docs-fmt       # markdown to 100 columns (just docs-install first)
mise exec -- just md-lint        # markdownlint, as CI runs it
mise exec -- just just-fmt-check # justfile formatting
gofmt -l services/               # Go files that need gofmt
```

`mise exec -- just ready` runs the justfile and Markdown checks, the Go tests and the typecheck
together; run it before committing.

### Code generation

| Changed                                                  | Run                                                                      |
| -------------------------------------------------------- | ------------------------------------------------------------------------ |
| `services/api/db/queries.sql`                            | `mise exec -- just api-generate` (sqlc)                                  |
| Worker OpenAPI spec (`internal/api/worker/gen/api.yaml`) | `cd services/api && mise exec -- go generate ./internal/api/worker/gen/` |
| Scheduler client spec                                    | `mise exec -- just scheduler-generate`                                   |

Commit the generated files with the change that caused them.

### Database changes

1. Add the next numbered migration in `services/api/internal/db/migrations/`. For a deal, see
   [deals.md](deals.md).
2. Add or change queries in `services/api/db/queries.sql`, then `just api-generate`.
3. Restart the API; migrations run on start.
4. Run the tests. `TestOnlyLAEventsActive` lists every active deal, so it fails until you update it
   for a new or retired one.
