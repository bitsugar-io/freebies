# AGENTS.md

Test: `just test` | Before committing: `just ready`

Read [CONTRIBUTING.md](CONTRIBUTING.md) first. It covers setup, configuration, project structure,
code standards, testing, branching, commit messages, pull requests and releases, and it links the
[AI policy](AI_POLICY.md). All of it applies to agents exactly as it applies to people. This file
carries only what is specific to agents.

## Running tools

Invoke tools through `mise`, not from your path:

```bash
mise exec -- just test
```

`mise` is active in a person's shell and supplies the versions `.mise.toml` declares. An agent's
shell has no activation, so a bare `just` or `go` resolves to whatever is installed globally,
usually an older version.

The symptom is a check that fails here and passes in continuous integration, on a file nobody
edited. When that happens, establish which version ran before treating the failure as real.

## Skills

Repo workflows live as skills in `.claude/skills/`: `ops` (setup, running the API and app, tests,
deals, deploys, iOS releases) and `design` (screens, UI copy and sprites). Use the matching skill
instead of improvising the steps.

## Server-driven content

**Content is server-driven.** New deals, teams, leagues, seasons, sprites and sport badges ship
through the API (database rows and endpoints), never through an app release. Don't hardcode team,
league or sprite lists in `apps/mobile`; bundled art is an offline fallback only. See
[Server-driven content](docs/architecture.md#server-driven-content).

## UI rules

Before any change a user can see (screens, sprites, copy, pushes, store metadata), read
[docs/ui.md](docs/ui.md). It is binding; the `design` skill applies it.

## The running app

The API on :8080 and the Metro bundler are usually the ones the user is running. Do not start, stop
or restart them unless the user asks, even where a skill says to restart after a change. Check with
the ops skill's `scripts/api.sh status` first; it names a server it did not start, and its `stop`
only kills the one it started. Never run `just clean` against a database the user is using.

## Reporting what you checked

Say what changed, which checks ran, and what you could not check. Report Go tests, the typecheck and
what you saw in the simulator separately. A push arriving on a device or a screen matching
[docs/ui.md](docs/ui.md) needs a person to look, so list it as outstanding rather than inferring it
from a passing test.

## Commit trailer

When committing via Claude Code, end the message with:

```text
🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
