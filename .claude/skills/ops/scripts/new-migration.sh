#!/usr/bin/env bash
# Write the next numbered goose migration with an events INSERT template to fill in.
# Usage: new-migration.sh <slug> [--print]   e.g. new-migration.sh dodgers-habit -> NNN_dodgers_habit.sql
#   --print writes the template to stdout instead of creating the file.
set -euo pipefail
usage="usage: new-migration.sh <slug> [--print]"
slug="${1:-}"
[[ "$slug" == "-h" || "$slug" == "--help" ]] && { echo "$usage"; exit 0; }
[[ -n "$slug" ]] || { echo "$usage" >&2; exit 2; }
[[ "$slug" =~ ^[a-z0-9]+([-_][a-z0-9]+)*$ ]] || { echo "slug must be lowercase words joined by - or _: $slug" >&2; exit 2; }
root="$(git rev-parse --show-toplevel)"
dir="$root/services/api/internal/db/migrations"
[[ -d "$dir" ]] || { echo "migrations dir not found: $dir" >&2; exit 1; }
last="$(ls "$dir" | sed -n 's/^\([0-9]\{3\}\)_.*/\1/p' | sort -n | tail -1)"
next="$(printf '%03d' $((10#${last:-0} + 1)))"
file="$dir/${next}_${slug//-/_}.sql"

template() {
  cat <<'SQL'
-- +goose Up
-- TODO Team - Partner offer. Verified for TODO season:
-- TODO https://sponsor-or-team-page
-- (Unverified? Say so here and set is_active = 0.)
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-TODO_LEAGUE-TODO_TEAM-TODO_PARTNER', 'TODO_TEAM-TODO_PARTNER', 'TODO_TEAM_ID', 'TODO Team Name', 'TODO_LEAGUE', '', '',
    'TODO Partner', 'TODO reward',
    'TODO reward, trigger, how to claim, deadline, limits.',
    'TODO Team win at home',
    '{"metric": "TODO", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "next_day"}',
    'us-ca-la', 'TODO https://sponsor-or-team-page', '', '', 1, 'gift'
);

-- +goose Down
DELETE FROM events WHERE id = 'event-TODO_LEAGUE-TODO_TEAM-TODO_PARTNER';
SQL
}

if [[ "${2:-}" == "--print" ]]; then template; exit 0; fi
[[ -e "$file" ]] && { echo "already exists: $file" >&2; exit 1; }
template >"$file"
echo "wrote: $file"
echo "next: replace every TODO (see .claude/skills/ops/references/deals.md), then: grep -n TODO $file"
