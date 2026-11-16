#!/usr/bin/env bash
# Code-only deploy of the API and scheduler to DigitalOcean, narrated step by step:
#   checks -> registry login -> build linux/amd64 -> push :latest -> restart freebie-api
#   -> wait for the new pod -> migrations and /healthz -> summary.
# CronJobs pull :latest on their next run, so the scheduler needs nothing more.
# Changes production: run only with the owner's go-ahead. Chart or value changes need a Helm
# upgrade instead (references/deploy-api.md, step 2b). Noisy docker output goes to a log that is
# printed only on failure.
# Usage: deploy.sh --yes
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
case "${1:-}" in
  --yes) ;;
  -h|--help) echo "usage: deploy.sh --yes"; exit 0 ;;
  *) echo "usage: deploy.sh --yes   (deploys to production; get the owner's go-ahead first)" >&2; exit 2 ;;
esac
cd "$(git rev-parse --show-toplevel)"
ns="${FREEBIE_NAMESPACE:-freebies}"
image="${FREEBIE_IMAGE:-registry.digitalocean.com/freebie/freebie-api:latest}"
log="$(mktemp "${TMPDIR:-/tmp}/freebie-deploy.XXXXXX")"
total=6 n=0 start=$SECONDS
rev="$(git rev-parse --abbrev-ref HEAD) @ $(git rev-parse --short HEAD)"

step() { n=$((n + 1)); printf '\n[%d/%d] %s\n' "$n" "$total" "$1"; [[ -n "${2:-}" ]] && printf '      %s\n' "$2"; return 0; }
say() { printf '      %s\n' "$*"; }
quiet() { # run a noisy command; show its output only if it fails
  if ! "$@" >>"$log" 2>&1; then
    echo "      failed: $*" >&2; tail -n 40 "$log" >&2; echo "      full log: $log" >&2; exit 1
  fi
}

echo "Deploying $rev to production ($ns on DigitalOcean)."
old_pod="$(mise exec -- kubectl get pods -n "$ns" -l app.kubernetes.io/name=freebie-api \
  -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"

step "Checking everything is ready" "Charts, docker, DigitalOcean login, registry space, cluster access."
"$here/deploy-preflight.sh" | sed 's/^/      /'

step "Logging in to the container registry" "So docker can push the new image to DigitalOcean."
quiet doctl registry login

step "Building the image" "Both Go binaries (API and scheduler) for the cluster's linux/amd64 nodes. Takes a minute or two."
quiet docker build --platform linux/amd64 -t "$image" .
say "built $(docker image inspect "$image" --format '{{.Size}}' | awk '{printf "%.0f MB", $1/1e6}')"

step "Pushing the image" "Uploads it as :latest; only changed layers are sent."
quiet docker push "$image"
say "pushed $(docker image inspect "$image" --format '{{index .RepoDigests 0}}' | sed 's/.*@//' | cut -c1-19)"

step "Restarting the API" "Kubernetes starts a new pod on the new image, waits until it's healthy, then stops ${old_pod:-the old one}. The new pod applies any new database migrations as it starts."
mise exec -- kubectl rollout restart deployment/freebie-api -n "$ns" >/dev/null
if ! mise exec -- kubectl rollout status deployment/freebie-api -n "$ns" --timeout=180s >>"$log" 2>&1; then
  echo "      the new pod didn't become ready; the old one keeps serving. Recent API logs:" >&2
  mise exec -- kubectl logs -n "$ns" -l app.kubernetes.io/name=freebie-api --tail=40 >&2 || true
  echo "      see references/deploy-api-troubleshooting.md; roll back per references/deploy-api.md" >&2
  exit 1
fi
new_pod="$(mise exec -- kubectl get pods -n "$ns" -l app.kubernetes.io/name=freebie-api \
  --field-selector=status.phase=Running -o jsonpath='{.items[-1:].metadata.name}')"
say "new pod $new_pod is ready"

step "Checking the new API" "Migrations it ran on startup, then the public health check."
logs="$(mise exec -- kubectl logs -n "$ns" "$new_pod" --tail=300 2>/dev/null || true)"
migrated="$(grep -E 'OK +[0-9]+_' <<<"$logs" | sed -E 's/.*OK +([0-9]+_[^ ]+).*/\1/' || true)"
version="$(grep -oE 'migrated database to version: [0-9]+|no migrations to run. current version: [0-9]+' <<<"$logs" | grep -oE '[0-9]+$' || true)"
if [[ -n "$migrated" ]]; then say "migrations applied: $(paste -sd' ' - <<<"$migrated")"; else say "no new migrations"; fi
[[ -n "$version" ]] && say "database at version $version"
"$here/deploy-status.sh" --health | sed 's/^/      /'

printf '\nDone in %ds. %s is live at %s.\n' $((SECONDS - start)) "$rev" "https://freebie-api.bitsugar.io"
echo "The scheduler's CronJobs pick up the new image on their next run. Build log: $log"
