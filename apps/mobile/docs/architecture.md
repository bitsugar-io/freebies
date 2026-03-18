# Mobile Architecture

## Overview

The Freebies mobile app is a React Native app built with Expo, targeting iOS (Android later). Builds
are handled by EAS (Expo Application Services).

## Technology Choices

### Expo (vs Bare React Native)

**Why Expo:**

- **Managed workflow** - No Xcode/Android Studio required for most development
- **OTA updates** - Push JS updates without App Store review (future)
- **EAS Build** - Cloud builds without local iOS toolchain (Fastlane, CocoaPods, etc.)
- **Expo modules** - First-party support for push notifications, secure storage, device info

**Trade-offs accepted:**

- Slightly larger binary size
- Some native modules require custom dev client (haven't needed yet)

### EAS Cloud Builds (vs Local Builds)

**Why EAS:**

- **No local dependencies** - Don't need Fastlane, CocoaPods, Ruby version management
- **Consistent builds** - Same environment every time, no "works on my machine"
- **Credentials management** - EAS handles signing certificates and provisioning profiles
- **CI/CD ready** - Easy to add automated builds on push

**Trade-offs accepted:**

- Build queue times (usually 5-10 min)
- Requires internet connection
- Free tier has limited builds per month

### State Management

**Current approach:**

- **React Context** for global state (user, theme)
- **Custom hooks** for data fetching (`useEvents`, `useActiveDeals`, `useSubscriptions`)
- **No Redux/Zustand** - App is simple enough that Context + hooks works well

**When we'd add a state library:**

- Complex cross-component state updates
- Need for middleware (logging, persistence)
- Performance issues with Context re-renders

## Key Flows

### User Registration

1. App launches, checks SecureStore for auth token
2. If none, generates device ID and calls `POST /users`
3. Backend returns user ID + auth token
4. Token stored in SecureStore for future requests

### Push Notifications

1. App requests notification permissions
2. Gets Expo push token from `expo-notifications`
3. Sends token to backend via `PUT /users/:id/push-token`
4. Backend stores token, uses it to send notifications via Expo Push API

### Subscriptions

1. User toggles subscription on EventCard
2. Calls `POST` or `DELETE /users/:id/subscriptions/:eventId`
3. Local state updates optimistically
4. Backend tracks which events user wants notifications for

## Future Considerations

- **Offline support**: Cache events/deals locally, sync when online
- **Android**: Currently iOS-only, Expo makes Android straightforward
- **OTA updates**: Use EAS Update for JS-only changes without App Store
- **Deep linking**: Handle notification taps to specific deals
