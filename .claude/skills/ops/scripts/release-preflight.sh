#!/usr/bin/env bash
# Read-only checks before an EAS build: app.json fields, icon size and alpha, eas.json profiles,
# changelog entry, and the TypeScript typecheck. Builds and submits nothing.
# Usage: release-preflight.sh [--no-typecheck]
set -euo pipefail
case "${1:-}" in
  -h|--help) echo "usage: release-preflight.sh [--no-typecheck]"; exit 0 ;;
  ""|--no-typecheck) ;;
  *) echo "usage: release-preflight.sh [--no-typecheck]" >&2; exit 2 ;;
esac
root="$(git rev-parse --show-toplevel)"
app="$root/apps/mobile"
[[ -f "$app/app.json" ]] || { echo "missing $app/app.json" >&2; exit 1; }
[[ -f "$app/eas.json" ]] || { echo "missing $app/eas.json" >&2; exit 1; }
fail=0

# app.json and eas.json facts (python3 -I: no imports from the repo).
facts="$(python3 -I - "$app" <<'PY'
import json, sys, os
app = sys.argv[1]
e = json.load(open(os.path.join(app, "app.json")))["expo"]
eas = json.load(open(os.path.join(app, "eas.json")))
ios = e.get("ios", {})
print(f"name={e.get('name','')}")
print(f"version={e.get('version','')}")
print(f"bundle={ios.get('bundleIdentifier','')}")
print(f"icon={e.get('icon','')}")
print(f"source={eas.get('cli',{}).get('appVersionSource','')}")
print(f"profiles={','.join(eas.get('build',{}))}")
print(f"submit={'yes' if eas.get('submit',{}).get('production',{}).get('ios') else 'no'}")
PY
)"
get() { sed -n "s/^$1=//p" <<<"$facts"; }
name="$(get name)" version="$(get version)" bundle="$(get bundle)" icon="$(get icon)"
echo "app: \"$name\" $version ($bundle), build number source: $(get source)"

[[ "$bundle" == "io.bitsugar.freebies" ]] || { echo "FAIL: bundleIdentifier is $bundle, want io.bitsugar.freebies" >&2; fail=1; }
[[ "$name" == "Freebies" ]] || echo "warning: display name is \"$name\"; APP-STORE.md's checklist wants \"Freebies\"" >&2
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "FAIL: version \"$version\" is not x.y.z" >&2; fail=1; }

for p in preview production; do
  [[ ",$(get profiles)," == *",$p,"* ]] || { echo "FAIL: eas.json has no build profile $p" >&2; fail=1; }
done
[[ "$(get submit)" == yes ]] || { echo "FAIL: eas.json has no submit.production.ios" >&2; fail=1; }

iconpath="$app/${icon#./}"
if [[ -f "$iconpath" ]]; then
  info="$(file -b "$iconpath")"
  echo "icon: $info"
  [[ "$info" == *"1024 x 1024"* ]] || { echo "FAIL: icon must be 1024 x 1024" >&2; fail=1; }
  [[ "$info" == *RGBA* || "$info" == *"gray+alpha"* ]] && { echo "FAIL: icon has an alpha channel; App Store Connect rejects it" >&2; fail=1; }
else
  echo "FAIL: icon not found: $iconpath" >&2; fail=1
fi

grep -q "^## \[$version\]" "$app/CHANGELOG.md" 2>/dev/null \
  || echo "warning: CHANGELOG.md has no \"## [$version]\" entry" >&2

if [[ "${1:-}" != "--no-typecheck" ]]; then
  echo "typecheck..."
  if (cd "$app" && mise exec -- bunx tsc --noEmit); then echo "ok: typecheck"; else echo "FAIL: typecheck" >&2; fail=1; fi
fi

[[ $fail -eq 0 ]] && echo "preflight ok" || { echo "preflight failed" >&2; exit 1; }
