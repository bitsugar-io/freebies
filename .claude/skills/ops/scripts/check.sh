#!/usr/bin/env bash
# Run the repo's checks: Go tests for the API and scheduler, and the mobile typecheck.
# Usage: check.sh [api|scheduler|mobile ...]   (default: all three)
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
targets=("$@")
[[ ${#targets[@]} -eq 0 ]] && targets=(api scheduler mobile)
failed=()
for t in "${targets[@]}"; do
  case "$t" in
    api|scheduler)
      echo "== $t: go vet + go test"
      if ! (cd "$root/services/$t" && mise exec -- go vet ./... && mise exec -- go test ./...); then failed+=("$t"); fi ;;
    mobile)
      echo "== mobile: tsc --noEmit"
      [[ -d "$root/apps/mobile/node_modules" ]] || { echo "apps/mobile/node_modules missing: run mise exec -- just setup" >&2; failed+=(mobile); continue; }
      if ! (cd "$root/apps/mobile" && mise exec -- bunx tsc --noEmit); then failed+=(mobile); fi ;;
    -h|--help) echo "usage: check.sh [api|scheduler|mobile ...]"; exit 0 ;;
    *) echo "unknown target: $t (want api, scheduler or mobile)" >&2; exit 2 ;;
  esac
done
if [[ ${#failed[@]} -gt 0 ]]; then echo "FAILED: ${failed[*]}" >&2; exit 1; fi
echo "ok: ${targets[*]}"
