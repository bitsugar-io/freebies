# CLI Reference

## Server

```bash
./bin/freebie serve           # Start API server
./bin/freebie serve --debug   # With debug logging
```

## Worker

```bash
./bin/freebie worker check-triggers             # Check yesterday's games
./bin/freebie worker check-triggers --date 2024-10-30  # Check specific date
./bin/freebie worker send-reminders             # Notify expiring deals
```

## Users

```bash
./bin/freebie users list      # List all users and their push tokens
```

## Deals

Manage triggered events (active deals). Useful for testing the full notification-to-deep-link flow.

```bash
./bin/freebie deals list                     # List all active deals
./bin/freebie deals create                   # Create a deal for a random event
./bin/freebie deals create --hours 48        # Set custom expiration
./bin/freebie deals trigger <event-id>       # Trigger a specific event (creates deal + notifies)
./bin/freebie deals trigger <event-id> --no-notify   # Create deal without sending notifications
./bin/freebie deals trigger <event-id> --hours 12    # Custom expiration
```

### Triggering a deal end-to-end

This is the closest simulation to what happens in production when a game result fires a rule. It
creates a `triggered_event`, auto-subscribes users, and sends push notifications with deep link data
so tapping the notification opens the deal in the app.

```bash
# 1. Find the event ID — use deals list (shows Event ID for existing deals)
./bin/freebie deals list

# Or query the API for all events
curl -s https://freebie-api.bitsugar.io/api/v1/events | python3 -m json.tool

# 2. Trigger it — subscribers get a push notification
./bin/freebie deals trigger <event-id>

# 3. Tap the notification on your device — it opens the deal detail modal
```

## Notifications

Send push notifications directly to users. Useful for testing notification delivery and deep linking
independently from the deal trigger flow.

```bash
# Send a test notification to all users with push tokens
./bin/freebie notify test
./bin/freebie notify test --title "Custom Title" --body "Custom body"

# Send to a specific user
./bin/freebie notify send <user-id>
./bin/freebie notify send <user-id> --title "🍔 Deal Unlocked!" --body "Free Jumbo Jack"

# Send with deep link data (tapping opens the deal in the app)
# Get event-id and triggered-event-id from: ./bin/freebie deals list
./bin/freebie notify send <user-id> \
  --title "🍔 Deal Unlocked!" \
  --body "Jack in the Box: Free Jumbo Jack" \
  --event-id <event-id> \
  --triggered-event-id <triggered-event-id>
```

### Notification flags

| Flag                   | Description                                   | Default              |
| ---------------------- | --------------------------------------------- | -------------------- |
| `--title`              | Notification title                            | 🎉 Test Notification |
| `--body`               | Notification body                             | This is a test...    |
| `--event-id`           | Event ID for deep linking (opens deal on tap) | (none)               |
| `--triggered-event-id` | Triggered event ID for deep linking           | (none)               |

### How deep linking works

When a notification includes `eventId` in its data payload, tapping it triggers this flow:

1. Mobile app receives the tap via Expo notification handler
2. `onNotificationTap(eventId)` is called
3. `openDealModal(eventId)` looks up the deal in active deals, then falls back to events
4. The deal detail modal opens

Both `deals trigger` and `notify send --event-id` include the correct data payload for this to work.

## Global Flags

| Flag       | Description                               |
| ---------- | ----------------------------------------- |
| `--config` | Config file path (default: ./config.yaml) |
| `--db`     | Database file path                        |
| `--debug`  | Enable debug logging                      |
| `--json`   | JSON log output                           |
