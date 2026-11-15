#!/usr/bin/env bash
# Validate every sprite in sprites.py and render an 8x contact sheet.
# Usage: preview.sh [out.png]   (default: $TMPDIR/sprite-sheet.png)
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
out="${1:-${TMPDIR:-/tmp}/sprite-sheet.png}"
python3 -I "$(dirname "$0")/sprites.py" sheet "$out"
echo "sheet: $out"
