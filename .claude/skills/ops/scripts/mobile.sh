#!/usr/bin/env bash
# Start the Expo dev server for Expo Go with a clean Metro cache (foreground).
# Usage: mobile.sh [extra expo start args]   e.g. mobile.sh --ios
set -euo pipefail
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  echo "usage: mobile.sh [extra expo start args]"; exit 0
fi
root="$(git rev-parse --show-toplevel)"
cd "$root/apps/mobile"
if [[ ! -d node_modules ]]; then
  echo "installing app dependencies (first run)..."
  mise exec -- bun install
fi
port="${FREEBIE_SERVER_PORT:-8080}"
if ! curl -fsS -m 2 "http://localhost:$port/healthz" >/dev/null 2>&1; then
  echo "warning: no API on :$port. The app needs it; start it with .claude/skills/ops/scripts/api.sh start" >&2
fi
echo "press i for the iOS simulator; Ctrl-C to stop"
exec mise exec -- bunx expo start --go -c "$@"
