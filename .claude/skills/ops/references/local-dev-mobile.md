# Local development: the mobile app

Starting the Expo app. API setup and tests are in [local-dev.md](local-dev.md).

## Start

Start the API first; the development build calls `http://localhost:8080/api/v1` (set in
`apps/mobile/src/api/client.ts`). Release builds call `https://freebie-api.bitsugar.io`.

```bash
.claude/skills/ops/scripts/mobile.sh         # bunx expo start --go -c, in the foreground
mise exec -- just mobile-serve                     # same, without clearing the Metro cache
```

In the Expo terminal:

- `i` opens the iOS simulator. `localhost` reaches your laptop from there.
- `a` opens the Android emulator. It can't see `localhost`; use the iOS simulator.
- `r` reloads, `j` opens the debugger.
- Scan the QR code to run in Expo Go on a phone. The phone can't reach `localhost` either, so use
  the simulator for anything that calls the API.

## Expo Go or a development build

`expo-dev-client` is installed, so a plain `expo start` expects a development build. The `--go` flag
switches to Expo Go, which is the quick path for UI work.

- **Expo Go** must match the app's Expo SDK (`expo` in `apps/mobile/package.json`). The simulator
  offers to install the right version; on a phone, update Expo Go from the store. A mismatch shows
  "Project is incompatible with this version of Expo Go".
- **Development build** has the real icon, splash and native modules. Build it once with EAS (the
  `development` profile; see [release-mobile.md](release-mobile.md)), install it on the simulator,
  then run `mise exec -- bunx expo start` without `--go`.
- **Push notifications** need a physical device and a `preview` or `production` build.

## Before calling a UI change done

- `mise exec -- bunx tsc --noEmit` in `apps/mobile` (or `scripts/check.sh mobile`).
- Look at the screen in the simulator.
- UI rules are in the `design` skill and `docs/ui.md`.
