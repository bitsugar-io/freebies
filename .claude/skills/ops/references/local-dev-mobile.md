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
- **Development build** has the real icon, splash and native modules. Build it locally with Xcode
  (free, about 10 minutes the first time), in `apps/mobile`:

  ```bash
  CI=1 mise exec -- bunx expo prebuild --clean --platform ios
  CI=1 mise exec -- bunx expo run:ios --device "iPhone 16e" --port 8083
  ```

  It installs "Freebies" on the simulator and serves it from its own Metro on :8083, leaving the
  usual :8081/:8082 alone. Prebuild rewrites the `ios` and `android` scripts in `package.json` to
  `expo run:*`; undo that with `git restore apps/mobile/package.json` (they stay `expo start`). The
  generated `ios/` folder is gitignored. Rebuild whenever the icon, splash, app name or a native
  module changes, and delete stale installs (`xcrun simctl uninstall booted <bundle id>`). EAS's
  `development` profile builds the same thing in the cloud ([release-mobile.md](release-mobile.md)).

- **Push notifications** need a physical device and a `preview` or `production` build.

## Before calling a UI change done

- `mise exec -- bunx tsc --noEmit` in `apps/mobile` (or `scripts/check.sh mobile`).
- Look at the screen in the simulator.
- UI rules are in the `design` skill and `docs/ui.md`.
