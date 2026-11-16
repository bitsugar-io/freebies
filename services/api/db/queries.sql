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

-- name: ClearPushToken :exec
UPDATE users SET push_token = NULL, updated_at = CURRENT_TIMESTAMP
WHERE push_token = ? AND id != ?;

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
INSERT INTO triggered_events (id, event_id, game_id, expires_at, payload, available_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTriggeredEvents :many
SELECT te.*, e.*
FROM triggered_events te
JOIN events e ON te.event_id = e.id
ORDER BY te.triggered_at DESC
LIMIT ?;

-- name: CreateNotification :one
INSERT INTO notifications (id, user_id, triggered_event_id, status, kind, sent_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListUsersRemindedSince :many
-- Users who already got a reminder for this deal at or after :since (a UTC
-- "YYYY-MM-DD HH:MM:SS" string). 'pending' (written just before the send) and
-- 'unknown' (Expo may have it) count as sent, so a push is never doubled;
-- only 'failed' sends retry.
SELECT DISTINCT user_id FROM notifications
WHERE triggered_event_id = sqlc.arg(triggered_event_id)
  AND kind = 'reminder'
  AND status IN ('sent', 'pending', 'unknown')
  AND datetime(sent_at) >= datetime(sqlc.arg(since));

-- name: ListTriggerPushDeals :many
-- Live deals whose "you won" push the worker still owes followers: unexpired
-- at :now, and recorded by the checker (no payload) or by
-- `deals trigger --notify` (payload "notify": true). Test deals from
-- `deals create` and `deals trigger --notify=false` are never pushed.
SELECT te.id, te.event_id, te.available_at, te.expires_at,
    e.partner_name, e.offer_name, e.trigger_condition
FROM triggered_events te
JOIN events e ON te.event_id = e.id
WHERE datetime(te.expires_at) > datetime(sqlc.arg(now))
  AND (te.payload IS NULL
    OR (json_valid(te.payload) AND json_extract(te.payload, '$.notify') = 1))
ORDER BY te.triggered_at ASC;

-- name: GetTriggerPushDeal :one
-- One deal in the shape ListTriggerPushDeals returns.
SELECT te.id, te.event_id, te.available_at, te.expires_at,
    e.partner_name, e.offer_name, e.trigger_condition
FROM triggered_events te
JOIN events e ON te.event_id = e.id
WHERE te.id = ?;

-- name: ListUsersOwedTriggerPush :many
-- Followers of a deal's event who should get its "you won" push and have not:
-- they followed before it triggered (a later follower sees it in the app but
-- is not told they won), have a push token, have not claimed or silenced it,
-- and have no trigger notification that is 'sent', 'pending' (written just
-- before a send) or 'unknown' (Expo may have it). So a user gets the push at
-- most once per deal, and a failed send is retried by the next run.
SELECT u.id, u.push_token, u.alerts_paused_until
FROM subscriptions s
JOIN users u ON u.id = s.user_id
JOIN triggered_events te ON te.event_id = s.event_id
LEFT JOIN dismissals d ON d.triggered_event_id = te.id AND d.user_id = u.id
WHERE te.id = sqlc.arg(triggered_event_id)
  AND u.push_token IS NOT NULL
  AND u.push_token != ''
  AND d.id IS NULL
  AND datetime(s.created_at) <= datetime(te.triggered_at)
  AND NOT EXISTS (
    SELECT 1 FROM notifications n
    WHERE n.triggered_event_id = te.id
      AND n.user_id = u.id
      AND n.kind = 'trigger'
      AND n.status IN ('sent', 'pending', 'unknown')
  );

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
-- A user's unexpired won deals. A deal is the user's when any of these holds:
--   1. they follow its event now (subscriptions);
--   2. they were pushed about it (a notifications row, any kind or status);
--   3. it was live in their list when they unfollowed (kept_deals, written by
--      KeepWonDeals just before the subscription is deleted);
--   4. they claimed or silenced it (dismissals).
-- So unfollowing a team never takes away a deal already won.
SELECT te.*, e.*,
    CASE WHEN d.id IS NOT NULL THEN 1 ELSE 0 END as is_dismissed,
    d.type as dismissal_type
FROM triggered_events te
JOIN events e ON te.event_id = e.id
LEFT JOIN dismissals d ON d.triggered_event_id = te.id AND d.user_id = sqlc.arg(user_id)
WHERE datetime(te.expires_at) > datetime('now')
  AND (
    d.id IS NOT NULL
    OR EXISTS (SELECT 1 FROM subscriptions s WHERE s.event_id = te.event_id AND s.user_id = sqlc.arg(user_id))
    OR EXISTS (SELECT 1 FROM notifications n WHERE n.triggered_event_id = te.id AND n.user_id = sqlc.arg(user_id))
    OR EXISTS (SELECT 1 FROM kept_deals k WHERE k.triggered_event_id = te.id AND k.user_id = sqlc.arg(user_id))
  )
ORDER BY te.expires_at ASC;

-- name: KeepWonDeals :exec
-- Before an unfollow deletes the subscription: keep the event's live deals the
-- user can see now, so they stay in WON (see ListActiveTriggeredEventsForUser).
INSERT OR IGNORE INTO kept_deals (user_id, triggered_event_id)
SELECT s.user_id, te.id
FROM subscriptions s
JOIN triggered_events te ON te.event_id = s.event_id
WHERE s.user_id = sqlc.arg(user_id)
  AND s.event_id = sqlc.arg(event_id)
  AND datetime(te.expires_at) > datetime('now');

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

-- name: UpdateDismissalType :one
-- Upgrades a silenced deal (stop_reminding) to claimed (got_it).
UPDATE dismissals SET type = ?, dismissed_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND triggered_event_id = ?
RETURNING *;

-- name: DeleteDismissal :exec
DELETE FROM dismissals WHERE user_id = ? AND triggered_event_id = ?;

-- name: ListClaimedDealsForUser :many
-- Every deal the user claimed ('got_it'), including ones whose window has closed.
SELECT te.id AS triggered_event_id, te.triggered_at, te.available_at, d.dismissed_at, sqlc.embed(e)
FROM dismissals d
JOIN triggered_events te ON te.id = d.triggered_event_id
JOIN events e ON e.id = te.event_id
WHERE d.user_id = ? AND d.type = 'got_it'
ORDER BY d.dismissed_at DESC, d.id DESC
LIMIT ?;

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
SELECT DISTINCT u.id, u.push_token, u.platform, u.alerts_paused_until
FROM users u
JOIN subscriptions s ON s.user_id = u.id
JOIN triggered_events te ON te.event_id = s.event_id
LEFT JOIN dismissals d ON d.triggered_event_id = te.id AND d.user_id = u.id
LEFT JOIN notifications n ON n.user_id = u.id AND n.triggered_event_id = te.id
WHERE te.id = ?
  AND u.push_token IS NOT NULL
  AND u.push_token != ''
  -- Anyone who dismissed it, by claiming it (got_it) or silencing it
  -- (stop_reminding), gets no more reminders for it.
  AND d.id IS NULL;

-- name: GetTriggeredEventByGameID :one
SELECT * FROM triggered_events
WHERE event_id = ? AND game_id = ?
LIMIT 1;

-- name: BackfillTeamSubscriptions :exec
INSERT OR IGNORE INTO subscriptions (id, user_id, event_id)
SELECT
    lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-' || hex(randomblob(2)) || '-' || hex(randomblob(2)) || '-' || hex(randomblob(6))) as id,
    s.user_id,
    e.id as event_id
FROM events e
JOIN (
    SELECT DISTINCT sub.user_id, ev.team_id
    FROM subscriptions sub
    JOIN events ev ON sub.event_id = ev.id
) s ON s.team_id = e.team_id
LEFT JOIN subscriptions existing ON existing.user_id = s.user_id AND existing.event_id = e.id
WHERE e.is_active = 1
  AND existing.id IS NULL;

-- name: ListFeatureFlags :many
SELECT * FROM feature_flags ORDER BY key;

-- name: ListAllEnabledScreenBlocks :many
SELECT * FROM screen_blocks
WHERE enabled = 1
ORDER BY screen, position ASC;

-- name: ListSprites :many
SELECT * FROM sprites ORDER BY id;

-- name: GetSprite :one
SELECT * FROM sprites WHERE id = ? LIMIT 1;

-- name: ListTeams :many
SELECT t.id, t.name, t.full_name, t.league, t.sport, t.season_start, t.season_end,
       t.return_label, t.display_order, t.badge_sprite_id, t.primary_color, t.secondary_color,
       COALESCE(l.name, UPPER(t.league)) AS league_name,
       (SELECT COUNT(*) FROM events e WHERE e.team_id = t.id AND e.is_active = 1) AS active_deals
FROM teams t
LEFT JOIN leagues l ON l.id = t.league
ORDER BY t.display_order;

-- name: UpdateUserAlertsPause :exec
UPDATE users SET alerts_paused_until = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;
