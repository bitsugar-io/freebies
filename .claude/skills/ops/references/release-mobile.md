# Releasing the mobile app

The iOS app builds in the cloud with EAS (Expo Application Services) and ships through TestFlight to
the App Store. `apps/mobile/APP-STORE.md` is the canonical listing copy and submission checklist;
this file covers the mechanics. If the two conflict, APP-STORE.md wins.

## Steps

1. Run `scripts/release-preflight.sh`. It checks `app.json`, `eas.json`, the icon and the typecheck.
2. Set the version and update `apps/mobile/CHANGELOG.md` ("Versioning" below).
3. Get the owner's go-ahead: builds use EAS credits, and submits reach Apple.
4. Build with the `preview` profile, install from TestFlight, and test on a phone, push included.
5. Build `production`, submit, and fill in App Store Connect from APP-STORE.md.

Narrate each step to the owner as it runs (EAS builds take 10–20 minutes): say which profile is
building and why, relay the build URL EAS prints, and report the build number and the TestFlight or
submit status when it finishes.

## EAS CLI

Run it as `mise exec -- eas` (installed globally with `npm i -g eas-cli`) or
`mise exec -- bunx eas-cli`, from `apps/mobile`. Log in once with `eas login`; the project belongs
to the Expo account in `app.json` `owner`. Apple certificates and provisioning profiles are managed
by EAS.

## Build profiles (`apps/mobile/eas.json`)

| Profile              | Gives you                                      | Command                                         |
| -------------------- | ---------------------------------------------- | ----------------------------------------------- |
| `development`        | Dev client for the iOS simulator               | `eas build -p ios --profile development`        |
| `development-device` | Dev client for a registered phone              | `eas build -p ios --profile development-device` |
| `preview`            | Internal-distribution build for testers        | `eas build -p ios --profile preview`            |
| `production`         | App Store build, build number auto-incremented | `eas build -p ios --profile production`         |

Install a `development` build on the simulator: download the `.tar.gz` from the build page, then

```bash
tar -xzf ~/Downloads/build-*.tar.gz
xcrun simctl install booted <App>.app
```

Then run `mise exec -- bunx expo start` (no `--go`) and open the app. Starting the dev server is in
[local-dev-mobile.md](local-dev-mobile.md).

## Versioning

| Field                           | Where                               | Rule                                                                           |
| ------------------------------- | ----------------------------------- | ------------------------------------------------------------------------------ |
| Marketing version               | `app.json` `expo.version`           | Semver. Bump it for every App Store release                                    |
| Build number                    | EAS servers                         | `appVersionSource: "remote"`: EAS owns it, and `production` auto-increments it |
| `ios.buildNumber` in `app.json` | `app.json`                          | Ignored by EAS while the source is remote. Don't rely on it                    |
| Runtime version                 | `runtimeVersion.policy: appVersion` | Equals `version`; an OTA update only reaches builds with the same version      |
| Bundle id                       | `ios.bundleIdentifier`              | `io.bitsugar.freebies`. Never change it; the store record depends on it        |

Check or set the remote build number: `eas build:version:get -p ios`,
`eas build:version:set -p ios`.

Every release gets a `CHANGELOG.md` entry (Keep a Changelog format), and the store "What's New" text
in APP-STORE.md.

## TestFlight, then the App Store

```bash
cd apps/mobile
mise exec -- eas build -p ios --profile production
mise exec -- eas submit -p ios --profile production      # uploads to App Store Connect
```

`submit.production` in `eas.json` holds the Apple ID, team and App Store Connect app id, so submit
needs no prompts beyond the Apple login.

1. The upload appears in TestFlight after Apple processes it (minutes to an hour).
2. Add internal testers in App Store Connect and test on a real phone, including a real push. Pushes
   don't work in the simulator or Expo Go.
3. Fill in the version's metadata from APP-STORE.md: name, subtitle, promotional text, keywords,
   description, What's New, category, age rating, URLs, review notes.
4. Go through APP-STORE.md's "Build checklist", then submit for review.

## Screenshots

APP-STORE.md lists the six shots, their order and captions. To capture:

1. Run a `development` or `preview` build on a 6.9" simulator (iPhone Pro Max, 1320×2868).
2. Load real, verified LA deals: the API with current data, and a triggered deal for the unlocked
   and claim shots (`deals trigger`, see [local-dev.md](local-dev.md)).
3. `xcrun simctl io booted screenshot apps/mobile/assets/screenshots/01-unlocked.png`, and so on.
4. Check every shot against `docs/ui.md`: no team logos, and no "NES" or "Nintendo".

## EAS Update (over-the-air)

`app.json` has an `updates.url` and a runtime version policy, but OTA updates aren't usable yet:
`expo-updates` isn't a dependency and no build profile sets a `channel`. To turn it on: install
`expo-updates`, add `"channel": "production"` (and `"preview"`) to the profiles, ship a new build,
then publish JS-only fixes with `eas update --channel production`. Native changes always need a new
build.

## Icon

1024×1024 PNG with no alpha channel (App Store Connect rejects transparency). Check with
`file apps/mobile/assets/icon.png`: it must say `RGB`, not `RGBA`. The preflight script checks it.
