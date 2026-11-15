#!/usr/bin/env bash
# Read-only checks before a deploy: lint and render the Helm charts, check docker and the
# deploy variables (names only; values are never printed). Touches no cluster.
# Usage: deploy-preflight.sh
set -euo pipefail
[[ "${1:-}" == "-h" || "${1:-}" == "--help" ]] && { echo "usage: deploy-preflight.sh"; exit 0; }
root="$(git rev-parse --show-toplevel)"
cd "$root"
fail=0
bad() { echo "FAIL: $*" >&2; fail=1; }

mise exec -- helm version --short >/dev/null 2>&1 || { echo "helm not available: run mise install" >&2; exit 1; }
for chart in api scheduler cloudflare; do
  if mise exec -- helm lint "charts/$chart" --quiet \
       --set databasePath=x --set workerSecret=x --set tunnelToken=x >/dev/null \
     && mise exec -- helm template "freebie-$chart" "charts/$chart" --namespace freebies \
       --set databasePath=x --set workerSecret=x --set tunnelToken=x >/dev/null; then
    echo "ok: charts/$chart lints and renders"
  else
    bad "charts/$chart: run mise exec -- helm lint charts/$chart"
  fi
done

if command -v docker >/dev/null && docker info >/dev/null 2>&1; then
  echo "ok: docker is running"
else
  bad "docker is not running (needed to build the image)"
fi

for v in FREEBIE_DATABASE_PATH FREEBIE_WORKER_SECRET CF_TUNNEL_TOKEN; do
  if [[ -n "${!v:-}" ]]; then echo "ok: $v is set"; else bad "$v is not set (see .claude/skills/ops/references/deploy-api-setup.md)"; fi
done
if [[ -n "${FREEBIE_DATABASE_PATH:-}" && "${FREEBIE_DATABASE_PATH}" != libsql://* ]]; then
  bad "FREEBIE_DATABASE_PATH is not a libsql:// Turso URL"
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "warning: uncommitted changes; the image is built from the working tree" >&2
fi
echo "branch: $(git rev-parse --abbrev-ref HEAD) @ $(git rev-parse --short HEAD)"
[[ $fail -eq 0 ]] && echo "preflight ok" || { echo "preflight failed" >&2; exit 1; }
