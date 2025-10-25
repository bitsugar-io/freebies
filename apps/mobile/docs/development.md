# Mobile Development

## Prerequisites

- Node.js 20+
- [EAS CLI](https://docs.expo.dev/eas/) - Expo build service

### Install Dependencies

```bash
# EAS CLI (global)
npm install -g eas-cli

# Project dependencies
cd apps/mobile
npm install
```

## Quick Start

```bash
# From repo root
task mobile:serve
```

Press `i` for iOS simulator, `a` for Android, or scan QR with Expo Go.

## Project Structure

```
apps/mobile/
├── src/
│   ├── components/     # Reusable UI components
│   ├── screens/        # Screen components
│   ├── services/       # API and notification services
│   └── types/          # TypeScript types
├── assets/             # Images, icons, screenshots
├── App.tsx             # App entry point
├── app.json            # Expo configuration
└── eas.json            # EAS Build configuration
```

## Building

All builds run on EAS cloud servers.

### Development Build (Simulator)

Creates a standalone app with your icon on the simulator:

```bash
eas build --platform ios --profile development
```

After build completes, download the `.tar.gz` from the build URL, then:

```bash
tar -xzf ~/Downloads/build-*.tar.gz
open -a Simulator
xcrun simctl install booted Freebie.app
```

Or drag the `.app` file onto the simulator window.

### Preview Build (TestFlight)

```bash
eas build --platform ios --profile preview
eas submit --platform ios
```

### Production Build (App Store)

```bash
eas build --platform ios --profile production
eas submit --platform ios
```

## Configuration

### app.json

| Field                  | Description                  |
| ---------------------- | ---------------------------- |
| `version`              | Public version (1.0.0)       |
| `ios.buildNumber`      | Increment for each build     |
| `ios.bundleIdentifier` | App ID (io.bitsugar.freebie) |

### eas.json

| Profile              | Purpose                        |
| -------------------- | ------------------------------ |
| `development`        | Simulator build with dev tools |
| `development-device` | Physical device with dev tools |
| `preview`            | TestFlight distribution        |
| `production`         | App Store submission           |

## Troubleshooting

### Icon not showing in simulator

You're running in Expo Go, which uses its own icon. Build a development build:

```bash
eas build --platform ios --profile development
```

### Clear cache

```bash
npx expo start --clear
```

### Push notifications not working

Push notifications only work on physical devices. Use `preview` or `production` builds on a real
device.
