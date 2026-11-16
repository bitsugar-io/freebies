# Local development: troubleshooting

Fixes for common local failures. Setup and daily use are in [local-dev.md](local-dev.md).

## Troubleshooting

### A Go check fails here but passes in CI

A bare `go` is the global install, usually older than the version `.mise.toml` pins. Run through
`mise exec --` and confirm with `mise exec -- go version`.

### `api.sh start` times out

Read the log: `scripts/api.sh logs`. Usual causes:

- **Port in use.** `scripts/api.sh status` names the process on :8080. Stop it, or set
  `FREEBIE_SERVER_PORT`.
- **Migration error.** The log names the file. Fix the new migration, `just clean`, start again.
- **Turso URL from `.envrc`.** `unset FREEBIE_DATABASE_PATH` and start again.

### The app shows nothing or a network error

- Is the API up? `curl localhost:8080/healthz`.
- Expo Go on a phone or the Android emulator can't reach `localhost`. Use the iOS simulator.
- After an API restart, pull to refresh or press `r` in Expo.

### `expo install --fix` or `bun add` fails for a new version

`apps/mobile/bunfig.toml` sets `minimumReleaseAge = 604800`: bun refuses packages published less
than 7 days ago, matching the Dependabot cooldown. A version Expo asks for on release day fails to
resolve. Wait out the week, or pin the newest version older than 7 days. Don't remove the hold.

### Red screen after adding a dependency

Metro serves a stale cache. Restart with a clean cache: `scripts/mobile.sh` (it passes `-c`), or
`mise exec -- bunx expo start --go -c`. If that isn't enough: `mise exec -- just mobile-reset`.

### "Project is incompatible with this version of Expo Go"

Expo Go must match the SDK in `apps/mobile/package.json`. Let the simulator install the matching
Expo Go, or use a development build (see [local-dev-mobile.md](local-dev-mobile.md)).

### The simulator shows the Expo icon, not the app icon

That's Expo Go. Only a development or release build shows the real icon.

### Stale data after editing a migration

Migrations that already ran don't run again. `just clean`, then restart the API.
