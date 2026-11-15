#!/usr/bin/env bash
# Read-only view of production: pods, CronJobs, recent Jobs and Helm releases, then /healthz.
# Usage: deploy-status.sh [--health]   (--health: only curl the public /healthz)
#   Namespace: $FREEBIE_NAMESPACE (default freebies). URL: $FREEBIE_PUBLIC_URL (default https://freebie-api.bitsugar.io)
set -euo pipefail
ns="${FREEBIE_NAMESPACE:-freebies}"
url="${FREEBIE_PUBLIC_URL:-https://freebie-api.bitsugar.io}"
case "${1:-}" in
  -h|--help) echo "usage: deploy-status.sh [--health]"; exit 0 ;;
  ""|--health) ;;
  *) echo "usage: deploy-status.sh [--health]" >&2; exit 2 ;;
esac
cd "$(git rev-parse --show-toplevel)"

health() {
  echo "== $url/healthz"
  if curl -fsS -m 10 "$url/healthz"; then echo; else echo "no healthy answer from $url/healthz" >&2; return 1; fi
}
[[ "${1:-}" == "--health" ]] && { health; exit $?; }

mise exec -- kubectl version --client >/dev/null 2>&1 || { echo "kubectl not available: run mise install" >&2; exit 1; }
ctx="$(mise exec -- kubectl config current-context 2>/dev/null)" || {
  echo "kubectl not configured: no current context. Run: doctl kubernetes cluster kubeconfig save <cluster-name>" >&2; exit 1; }
echo "context: $ctx  namespace: $ns"
mise exec -- kubectl get namespace "$ns" --request-timeout=10s >/dev/null 2>&1 || {
  echo "cannot read namespace $ns in context $ctx (wrong cluster, expired credentials, or not installed yet)" >&2; exit 1; }

echo "== deployments and pods"
mise exec -- kubectl get deploy,pods -n "$ns" -o wide --request-timeout=10s
echo "== cronjobs"
mise exec -- kubectl get cronjobs -n "$ns" --request-timeout=10s
echo "== recent jobs"
mise exec -- kubectl get jobs -n "$ns" --sort-by=.metadata.creationTimestamp --request-timeout=10s | tail -n 6
echo "== helm releases"
mise exec -- helm list -n "$ns"
health
