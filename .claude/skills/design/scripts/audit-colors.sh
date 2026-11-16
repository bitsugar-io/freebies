#!/usr/bin/env bash
# List hex colors in app source that are not NES palette tokens.
# Usage: audit-colors.sh [path]   (default: apps/mobile/src)
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
target="${1:-$root/apps/mobile/src}"
allowed='000000|004058|0000BC|FCFCFC|BCBCBC|7C7C7C|F8B800|F8D878|58D854|007800|F83800|A81000|3CBCFC|FCA044|A4E4FC|F0D0B0'
# Logo art only: the palm sunset logo (src/components/pixel/logoArt.ts) uses four extra NES
# master-palette entries (purple, violet, magenta, blue). They are allowed in that one file, never in UI.
logo='4428BC|6844FC|940084|0058F8'
hits="$(grep -rnoiE "#[0-9a-f]{6}\b|#[0-9a-f]{3}\b" "$target" --include='*.ts' --include='*.tsx' \
  | grep -viE "#($allowed)\b" \
  | grep -viE "/logoArt\.ts:[0-9]+:#($logo)\b" || true)"
if [[ -z "$hits" ]]; then echo "ok: only NES tokens"; exit 0; fi
echo "$hits"
echo "$(echo "$hits" | wc -l | tr -d ' ') non-token color(s): move them to theme tokens (useTheme.tsx is the one allowed place for raw hex)." >&2
exit 1
