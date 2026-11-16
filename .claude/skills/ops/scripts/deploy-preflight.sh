#!/usr/bin/env bash
# Read-only checks before a deploy. Changes nothing.
#   Local: Helm charts lint and render, docker is running.
#   DigitalOcean: doctl is logged in and the account is active, the registry exists and has room
#   for a push, the cluster is running, kubectl points at it and can read the namespace, and the
#   namespace has the registry pull secret.
#   With --helm (chart or value change): the deploy variables are set (names only, never values).
# Usage: deploy-preflight.sh [--helm]
set -euo pipefail
case "${1:-}" in
  -h|--help) echo "usage: deploy-preflight.sh [--helm]"; exit 0 ;;
  ""|--helm) ;;
  *) echo "usage: deploy-preflight.sh [--helm]" >&2; exit 2 ;;
esac
helm_mode="${1:-}"
root="$(git rev-parse --show-toplevel)"
cd "$root"
ns="${FREEBIE_NAMESPACE:-freebies}"
registry="${FREEBIE_REGISTRY:-freebie}"
min_free_mb="${FREEBIE_REGISTRY_MIN_FREE_MB:-100}"
fail=0
ok() { echo "ok: $*"; }
bad() { echo "FAIL: $*" >&2; fail=1; }
warn() { echo "warning: $*" >&2; }

# --- local
mise exec -- helm version --short >/dev/null 2>&1 || { echo "helm not available: run mise install" >&2; exit 1; }
for chart in api scheduler cloudflare; do
  if mise exec -- helm lint "charts/$chart" --quiet \
       --set databasePath=x --set workerSecret=x --set tunnelToken=x >/dev/null \
     && mise exec -- helm template "freebie-$chart" "charts/$chart" --namespace "$ns" \
       --set databasePath=x --set workerSecret=x --set tunnelToken=x >/dev/null; then
    ok "charts/$chart lints and renders"
  else
    bad "charts/$chart: run mise exec -- helm lint charts/$chart"
  fi
done

if command -v docker >/dev/null && docker info >/dev/null 2>&1; then
  ok "docker is running"
else
  bad "docker is not running (needed to build the image)"
fi

# --- DigitalOcean
if ! command -v doctl >/dev/null; then
  bad "doctl not installed (brew install doctl, then doctl auth init)"
else
  status="$(doctl account get --format Status --no-header 2>/dev/null || true)"
  if [[ "$status" == "active" ]]; then ok "doctl logged in, account active"
  else bad "doctl account is '${status:-unreachable}': run doctl auth init"; fi

  reg_json="$(doctl registry get "$registry" -o json 2>/dev/null || true)"
  if [[ -z "$reg_json" ]]; then
    bad "registry '$registry' not found (see references/deploy-api-setup.md, step 2)"
  else
    used="$(printf '%s' "$reg_json" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)[0]["storage_usage_bytes"])')"
    # The account's tier isn't exposed by doctl; assume Starter (the documented tier) unless told otherwise.
    tier="${FREEBIE_REGISTRY_TIER:-starter}"
    tier_line="$(doctl registry options subscription-tiers -o json 2>/dev/null | python3 -I -c '
import json, sys
for t in json.load(sys.stdin)["SubscriptionTiers"]:
    if t["slug"] == sys.argv[1]:
        print(t["slug"], t["included_storage_bytes"], str(t["allow_storage_overage"]).lower())
' "$tier" || true)"
    cap="$(awk '{print $2}' <<<"$tier_line")"
    overage="$(awk '{print $3}' <<<"$tier_line")"
    used_mb=$(( used / 1000000 ))
    if [[ -z "$cap" ]]; then
      warn "registry '$registry' uses ${used_mb} MB; could not read the '$tier' tier's cap"
    else
      cap_mb=$(( cap / 1000000 )); free_mb=$(( cap_mb - used_mb ))
      if [[ "$overage" == "true" ]]; then
        ok "registry '$registry' ${used_mb}/${cap_mb} MB ($tier tier, overage billed)"
      elif (( free_mb < min_free_mb )); then
        bad "registry '$registry' ${used_mb}/${cap_mb} MB, only ${free_mb} MB free: a push can fail. Run garbage collection (references/deploy-api-troubleshooting.md)"
      else
        ok "registry '$registry' ${used_mb}/${cap_mb} MB, ${free_mb} MB free ($tier tier)"
      fi
    fi
  fi

  cluster="$(doctl kubernetes cluster list --format Name,Status --no-header 2>/dev/null | awk '$2=="running"{print $1; exit}')"
  if [[ -z "$cluster" ]]; then
    bad "no running DOKS cluster on this account"
  else
    ok "cluster $cluster is running"
    ctx="$(mise exec -- kubectl config current-context 2>/dev/null || true)"
    if [[ "$ctx" != *"$cluster" ]]; then
      bad "kubectl context is '${ctx:-none}', not $cluster: run doctl kubernetes cluster kubeconfig save $cluster"
    elif ! mise exec -- kubectl get namespace "$ns" --request-timeout=10s >/dev/null 2>&1; then
      bad "cannot read namespace $ns (expired credentials? rerun doctl kubernetes cluster kubeconfig save $cluster)"
    else
      ok "kubectl reaches namespace $ns"
      if mise exec -- kubectl get secret "$registry" -n "$ns" --request-timeout=10s >/dev/null 2>&1; then
        ok "pull secret '$registry' present in $ns"
      else
        bad "no pull secret '$registry' in $ns: run doctl kubernetes cluster registry add $cluster"
      fi
    fi
  fi
fi

# --- deploy variables, only needed for a Helm upgrade
if [[ "$helm_mode" == "--helm" ]]; then
  for v in FREEBIE_DATABASE_PATH FREEBIE_WORKER_SECRET CF_TUNNEL_TOKEN; do
    if [[ -n "${!v:-}" ]]; then ok "$v is set"; else bad "$v is not set (see references/deploy-api-setup.md)"; fi
  done
  if [[ -n "${FREEBIE_DATABASE_PATH:-}" && "${FREEBIE_DATABASE_PATH}" != libsql://* ]]; then
    bad "FREEBIE_DATABASE_PATH is not a libsql:// Turso URL"
  fi
fi

if [[ -n "$(git status --porcelain)" ]]; then
  warn "uncommitted changes; the image is built from the working tree"
fi
echo "branch: $(git rev-parse --abbrev-ref HEAD) @ $(git rev-parse --short HEAD)"
[[ $fail -eq 0 ]] && echo "preflight ok" || { echo "preflight failed" >&2; exit 1; }
