# AGENTS.md

Test: `task api:test scheduler:test` | Format markdown: `task docs:fmt`

Read [docs/development.md](docs/development.md) first, then
[docs/contributing.md](docs/contributing.md). Between them they cover setup, the tasks, project
structure, code style, branching, commit messages, the pull request process, and how to add game
sources, offers and database changes. All of it applies to agents exactly as it applies to people.
This file carries only what is specific to agents.

## Running tools

Invoke tools through `mise`, not from your path:

```bash
mise exec -- task api:test scheduler:test
```

`mise` is active in a person's shell and supplies the versions `.mise.toml` declares. An agent's
shell has no activation, so a bare `go` resolves to whatever is installed globally, usually an older
version. Task itself is not declared there, so it comes from your path either way; the Go, Node and
Bun it calls are what `mise` pins.

The symptom is a check that fails here and passes in continuous integration, on a file nobody
edited. When that happens, establish which version ran before treating the failure as real.

## Commit trailer

When committing via Claude Code, end the message with:

```
🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
