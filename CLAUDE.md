# Freebie

Sports rewards notification platform. Monitors live games and sends push notifications when free
food/merch deals are triggered (e.g., Dodgers pitchers get 7+ strikeouts = free Jumbo Jack).

<!-- prettier-ignore -->
@docs/development.md — prerequisites, setup, tasks, code style, testing
@docs/contributing.md — PR workflow and contribution guidelines

## Quick Reference

```bash
task setup              # first-time setup
task api:serve      # start API server (auto-migrates DB)
task api:test       # run Go tests
task mobile:serve       # start Expo dev server
task docs:fmt           # format markdown files
task clean              # wipe database
```

## Directory Structure

```
apps/mobile/                    React Native (Expo) app
services/api/               Go API server + worker
  internal/
    db/                         sqlc queries, generated code, migrations
    server/                     HTTP handlers and routes
    sources/                    Live game data sources (Source interface)
    worker/                     Background job runner
docs/                           Project documentation
```

## Code Standards (Mandatory)

### Branching

See @docs/development.md#branching for full conventions.

When committing changes via `/commit`, create a feature branch first if currently on `main`. Branch
names use the pattern `type/short-description` (e.g., `feat/add-nba-source`,
`fix/notification-retry`, `docs/update-readme`).

### Commit Messages

See @docs/development.md#commit-messages for full conventions.

Follow [Conventional Commits](https://www.conventionalcommits.org/) with the 50/72 rule. Format:
`type(scope): description`.

When committing via Claude Code, end with:

- `🤖 Generated with [Claude Code](https://claude.ai/code)`
- `Co-Authored-By: Claude <noreply@anthropic.com>`

### Go Patterns

- `Source` interface in `internal/sources/` for new game data providers
- sqlc for all database queries — edit `queries.sql`, run `task generate`
- goose migrations in `internal/db/migrations/` — auto-run on server start
- Environment config via `FREEBIE_*` variables (see docs/development.md)
