#!/usr/bin/env bash
# Download one sprite as PNG from the local API.
# Usage: fetch.sh <id> [scale] [out.png]   (API: $FREEBIE_API, default http://localhost:8080)
set -euo pipefail
id="${1:?usage: fetch.sh <id> [scale] [out.png]}"
scale="${2:-12}"
out="${3:-${TMPDIR:-/tmp}/$id.png}"
api="${FREEBIE_API:-http://localhost:8080}"
code="$(curl -s -o "$out" -w '%{http_code}' "$api/api/v1/sprites/$id.png?scale=$scale")"
[[ "$code" == 200 ]] || { echo "GET $id.png returned $code (is just api-serve running?)" >&2; exit 1; }
echo "png: $out"
