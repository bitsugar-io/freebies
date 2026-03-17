# Contributing to Freebie

Thanks for your interest in contributing! This guide covers the process for submitting changes.

## Before You Start

- **Check existing work** — search open issues and PRs to avoid duplicating effort.
- **Think about backwards compatibility** — changes to the API, database schema, or notification
  payloads may affect the mobile app. Consider migration paths.
- **Start small** — smaller, focused PRs are easier to review and merge than large sweeping changes.

## Making Changes

### Code Style

- **Go**: run `task api:lint` and `task api:fmt` before committing.
- **Markdown**: run `task docs:fmt` to format to 100-character width.
- **TypeScript**: follow existing patterns in `apps/mobile/`.

### Documentation

- Update docs if your change affects setup, configuration, or usage.
- Add migration notes for database schema changes.

### Testing

- Run `task api:test` and verify your changes don't break existing tests.
- Add tests for new functionality when possible.

## Submitting a PR

1. **Create a feature branch** from `main` using the `type/short-description` convention (see
   [development guide](development.md#branching) for details).
2. **Write clear commit messages** following
   [Conventional Commits](https://www.conventionalcommits.org/) (see
   [commit conventions](development.md#commit-messages)).
3. **Describe your changes** in the PR description:
   - What changed and why
   - Link to related issues if applicable
   - Include screenshots for UI changes
4. **Open as draft** if you want early feedback before the PR is ready for final review.
5. **Keep PRs focused** — one logical change per PR. Split unrelated changes into separate PRs.

## Adding New Features

### New Game Sources

Implement the `Source` interface in `services/api/internal/sources/` and register it in the worker.
See existing sources for reference.

### New Offers/Events

Create a goose migration in `services/api/internal/db/migrations/`. See `003_mlb_data.sql` for the
expected format.

### Database Changes

1. Add a new migration file (next sequence number).
2. Update `services/api/internal/db/queries.sql` if new queries are needed.
3. Run `task generate` to regenerate sqlc code.
4. Migrations run automatically on server start — no manual steps needed.
