#!/usr/bin/env bash
# Write the next numbered sprite migration (a diff against sprites.lock.json),
# run the sprite tests, then update the lock. On failure the migration is removed.
# Usage: ship.sh <slug>   e.g. ship.sh churro -> NNN_churro_sprites.sql
set -euo pipefail
slug="${1:?usage: ship.sh <slug>}"
[[ "$slug" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || { echo "slug must be kebab-case: $slug" >&2; exit 1; }
here="$(cd "$(dirname "$0")" && pwd)"
root="$(git rev-parse --show-toplevel)"
dir="$root/services/api/internal/db/migrations"
last="$(ls "$dir" | sed -n 's/^\([0-9]\{3\}\)_.*/\1/p' | sort -n | tail -1)"
next="$(printf '%03d' $((10#$last + 1)))"
file="$dir/${next}_${slug//-/_}_sprites.sql"
tmp="$(mktemp)"
python3 -I "$here/sprites.py" sql > "$tmp" || { rm -f "$tmp"; exit 1; }
mv "$tmp" "$file"
echo "wrote: $file"
if ! (cd "$root/services/api" && mise exec -- go test ./internal/sprites/ ./internal/api/handlers/); then
  rm -f "$file"
  echo "tests failed; removed $file and left sprites.lock.json unchanged" >&2
  exit 1
fi
python3 -I "$here/sprites.py" lock
