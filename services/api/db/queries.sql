-- name: ListLeagues :many
SELECT * FROM leagues ORDER BY display_order ASC;

-- name: GetLeague :one
SELECT * FROM leagues WHERE id = ? LIMIT 1;

-- name: CreateLeague :one
INSERT INTO leagues (id, name, icon, display_order)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByDeviceID :one
SELECT * FROM users WHERE device_id = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, device_id, push_token, platform, token)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByToken :one
SELECT * FROM users WHERE token = ? LIMIT 1;

-- name: UpdateUserPushToken :exec
UPDATE users SET push_token = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ListUsersWithPushTokens :many
SELECT * FROM users WHERE push_token IS NOT NULL AND push_token != '';

-- name: ListAllUsers :many
SELECT * FROM users;

-- name: ListEvents :many
SELECT * FROM events ORDER BY created_at DESC;

-- name: ListActiveEvents :many
SELECT * FROM events WHERE is_active = 1 ORDER BY created_at DESC;

-- name: GetEvent :one
SELECT * FROM events WHERE id = ? LIMIT 1;

-- name: CreateEvent :one
INSERT INTO events (id, offer_id, team_id, team_name, league, team_color, icon, partner_name, offer_name, offer_description, trigger_condition, trigger_rule, region_code, offer_url, affiliate_url, affiliate_tagline, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateEvent :exec
UPDATE events
SET offer_name = ?, offer_description = ?, trigger_condition = ?, region_code = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = ?;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE user_id = ? AND event_id = ? LIMIT 1;

-- name: ListUserSubscriptions :many
SELECT s.*, e.*
FROM subscriptions s
JOIN events e ON s.event_id = e.id
WHERE s.user_id = ?
ORDER BY s.created_at DESC;

-- name: ListEventSubscribers :many
SELECT s.*, u.*
FROM subscriptions s
JOIN users u ON s.user_id = u.id
WHERE s.event_id = ?;

-- name: CreateSubscription :one
INSERT INTO subscriptions (id, user_id, event_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE user_id = ? AND event_id = ?;

-- name: CountUserSubscriptions :one
SELECT COUNT(*) as count FROM subscriptions WHERE user_id = ?;

-- name: CreateTriggeredEvent :one
INSERT INTO triggered_events (id, event_id, game_id, expires_at, payload)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTriggeredEvents :many
SELECT te.*, e.*
FROM triggered_events te
JOIN events e ON te.event_id = e.id
ORDER BY te.triggered_at DESC
LIMIT ?;

-- name: CreateNotification :one
INSERT INTO notifications (id, user_id, triggered_event_id, status)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateNotificationStatus :exec
UPDATE notifications SET status = ? WHERE id = ?;

-- name: ListPendingNotifications :many
SELECT n.*, u.push_token, u.platform
FROM notifications n
JOIN users u ON n.user_id = u.id
WHERE n.status = 'pending';

-- name: ListActiveTriggeredEvents :many
SELECT te.*, e.*
FROM triggered_events te
JOIN events e ON te.event_id = e.id
WHERE datetime(te.expires_at) > datetime('now')
ORDER BY te.expires_at ASC;

-- name: ListActiveTriggeredEventsForUser :many
SELECT te.*, e.*,
    CASE WHEN d.id IS NOT NULL THEN 1 ELSE 0 END as is_dismissed,
    d.type as dismissal_type
FROM triggered_events te
JOIN events e ON te.event_id = e.id
JOIN subscriptions s ON s.event_id = e.id AND s.user_id = ?
LEFT JOIN dismissals d ON d.triggered_event_id = te.id AND d.user_id = ?
WHERE datetime(te.expires_at) > datetime('now')
ORDER BY te.expires_at ASC;

-- name: GetTriggeredEvent :one
SELECT te.*, e.*
FROM triggered_events te
JOIN events e ON te.event_id = e.id
WHERE te.id = ?;

-- name: CreateDismissal :one
INSERT INTO dismissals (id, user_id, triggered_event_id, type)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetDismissal :one
SELECT * FROM dismissals WHERE user_id = ? AND triggered_event_id = ? LIMIT 1;

-- name: DeleteDismissal :exec
DELETE FROM dismissals WHERE user_id = ? AND triggered_event_id = ?;

-- name: CountUserClaimedDeals :one
SELECT COUNT(*) as count FROM dismissals WHERE user_id = ? AND type = 'got_it';

-- name: ListExpiringTriggeredEvents :many
SELECT te.*, e.*
FROM triggered_events te
JOIN events e ON te.event_id = e.id
WHERE datetime(te.expires_at) > datetime('now')
  AND datetime(te.expires_at) <= datetime('now', '+' || ? || ' hours')
ORDER BY te.expires_at ASC;

-- name: ListUsersForReminder :many
SELECT DISTINCT u.id, u.push_token, u.platform
FROM users u
JOIN subscriptions s ON s.user_id = u.id
JOIN triggered_events te ON te.event_id = s.event_id
LEFT JOIN dismissals d ON d.triggered_event_id = te.id AND d.user_id = u.id
LEFT JOIN notifications n ON n.user_id = u.id AND n.triggered_event_id = te.id
WHERE te.id = ?
  AND u.push_token IS NOT NULL
  AND u.push_token != ''
  AND (d.id IS NULL OR d.type != 'stop_reminding');

-- name: GetTriggeredEventByGameID :one
SELECT * FROM triggered_events
WHERE event_id = ? AND game_id = ?
LIMIT 1;

-- name: ListFeatureFlags :many
SELECT * FROM feature_flags ORDER BY key;

-- name: ListAllEnabledScreenBlocks :many
SELECT * FROM screen_blocks
WHERE enabled = 1
ORDER BY screen, position ASC;
