# Duplicate Notifications & Retry Logic

**Date**: 2026-03-18
**Status**: Approved

## Problem Statement

Three related issues affecting notification reliability:

1. **Duplicate notifications** — Users receive 2+ identical push notifications for every deal trigger
   and reminder. Root cause: the mobile app creates a new user record (with a fresh random device_id)
   whenever server verification fails on startup, but the Expo push token stays the same. Multiple
   user records sharing the same push token each have subscriptions, so `notifySubscribers()` and
   `SendReminders()` send multiple pushes to the same device.

2. **No retry on Expo push** — `SendBatch()` in `expo.go` has no retry logic. If Expo returns an
   error or non-200 status, the batch is counted as failed and dropped silently.

3. **No retry on source fetching** — `mlb/client.go` makes raw HTTP calls to `statsapi.mlb.com`
   with no retry. If the API is temporarily down, the check fails and the event is skipped.

## Design

### 1. Fix Duplicate Notifications (Root Cause)

Three layers of defense:

#### 1a. Backend: Clear stale push tokens on registration and creation

In both the `UpdateUserPushToken` handler (PUT `/users/{id}/push-token`) and the `CreateUser`
handler (POST `/users`, when a push token is provided), clear that push token from all other user
records before setting it on the current user. This ensures only one user "owns" a given Expo push
token at any time.

**Changes:**
- Add new sqlc query `ClearPushToken` — sets `push_token = NULL` on all users with a given token
  except the specified user ID
- Call `ClearPushToken` before `UpdateUserPushToken` in both handlers
- Handle unique constraint violations gracefully (return success, not 500) in case of race between
  concurrent requests — the partial unique index from the migration acts as the final safety net

#### 1b. Backend: Deduplicate by push token when sending

In `notifySubscribers()` and `SendReminders()` in `worker/service.go`, deduplicate the message list
by `push_token` before calling `SendBatchConcurrent()`. This is belt-and-suspenders protection
against any remaining duplicate token scenarios.

**Changes:**
- Add deduplication helper that filters `[]ExpoPushMessage` to unique `To` values
- Apply before each `SendBatchConcurrent()` call

#### 1c. Mobile: Use a stable device identifier

Replace `Math.random()` device ID generation with `expo-application`:
- iOS: `Application.getIosIdForVendorAsync()` — persists across app launches, resets on
  uninstall+reinstall
- Android: `Application.androidId` — persists across app launches

**Note:** On uninstall+reinstall, a new user record will still be created. The `ClearPushToken`
logic in `UpdatePushToken` handles this by clearing the token from the orphaned record when the new
user registers their push token.

**Changes:**
- `apps/mobile/src/hooks/useUser.tsx` — replace `generateDeviceId()` with stable platform IDs

### 2. Shared HTTP Retry Helper

Create a reusable `httputil` package used by both Expo push sending and source fetching.

**Package**: `services/api/internal/httputil/retry.go`

**Behavior:**
- Retry up to 3 times with exponential backoff: 1s, 2s, 4s (plus random jitter of +/-25% to
  prevent thundering herd when multiple workers retry simultaneously)
- Retry on network errors, 5xx responses, and 429 (Too Many Requests — Expo rate limiting)
- Do not retry on other 4xx (client errors won't self-heal)
- Respect context cancellation
- Log each retry attempt

**Interface:**
```go
type RetryOptions struct {
    MaxRetries int           // default 3
    BaseDelay  time.Duration // default 1s
}

// Do executes an HTTP request with retry logic. The newReq function is called
// for each attempt to produce a fresh *http.Request (avoids consumed-body issues
// on retry). For GET requests the function can return the same request; for POST
// requests it should rebuild the request with a fresh body reader.
func Do(ctx context.Context, client *http.Client, newReq func() (*http.Request, error), opts *RetryOptions) (*http.Response, error)
```

The `newReq` function pattern ensures request bodies are not consumed across retries. Callers
construct the request fresh on each attempt.

### 3. Notification Retry Logic (Expo Push)

Update `notify/expo.go`:
- Add `context.Context` parameter to `SendBatch()` signature (currently missing)
- Use `httputil.Do()` with a `newReq` function that rebuilds the POST request from the buffered
  JSON payload
- `SendBatchConcurrent()` already has a `ctx` and passes it through

**Additional per-ticket handling:**
- When Expo returns `"error"` status on individual tickets (e.g., `DeviceNotRegistered`), log and
  count as failed — do not retry since these are permanent errors

### 4. Source Fetching Retry Logic

Update `sources/mlb/client.go` `GetSchedule()` and `GetBoxScore()` to use `httputil.Do()` instead
of raw `httpClient.Do()`. Both are GET requests with no body, so the `newReq` function simply
returns a new `http.NewRequestWithContext()` call.

Any future source implementations (NBA, NFL, etc.) that use the same pattern will automatically get
retry behavior.

### 5. Data Cleanup Migration

Goose migration to fix existing duplicate data:

1. For each group of users sharing a push token, keep the most recently created user and set
   `push_token = NULL` on the older records
2. Transfer subscriptions from orphaned users to the kept user via `INSERT OR IGNORE` (respects
   the existing `UNIQUE(user_id, event_id)` constraint), then delete the orphaned subscriptions
3. Add a partial unique index:
   `CREATE UNIQUE INDEX idx_users_push_token ON users(push_token) WHERE push_token IS NOT NULL AND push_token != ''`

This fixes existing data and prevents recurrence at the DB constraint level.

## Files Changed

| File | Change |
|------|--------|
| `services/api/internal/httputil/retry.go` | New — shared retry helper |
| `services/api/internal/httputil/retry_test.go` | New — tests for retry logic |
| `services/api/internal/notify/expo.go` | Add `ctx` to `SendBatch()`, use `httputil.Do()` |
| `services/api/internal/sources/mlb/client.go` | Use `httputil.Do()` in both methods |
| `services/api/internal/worker/service.go` | Deduplicate messages by push token |
| `services/api/db/queries.sql` | Add `ClearPushToken` query |
| `services/api/internal/db/*.go` | Regenerated sqlc |
| `services/api/internal/api/handlers/handlers.go` | Call `ClearPushToken` in both `CreateUser` and `UpdatePushToken` |
| `services/api/internal/db/migrations/013_*.sql` | Data cleanup + unique index migration |
| `apps/mobile/src/hooks/useUser.tsx` | Stable device ID via `expo-application` |
| `apps/mobile/package.json` | Add `expo-application` dependency |
