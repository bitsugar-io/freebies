#!/usr/bin/env bash
# Run the Go API server in the background for local development.
# Usage: api.sh start|stop|restart|status|logs [-f]
#   Port: $FREEBIE_SERVER_PORT (default 8080). State: $TMPDIR/freebies-api/{freebie,api.pid,api.log}
set -euo pipefail
usage="usage: api.sh start|stop|restart|status|logs [-f]"
root="$(git rev-parse --show-toplevel)"
port="${FREEBIE_SERVER_PORT:-8080}"
state="${TMPDIR:-/tmp}/freebies-api"
bin="$state/freebie" pidf="$state/api.pid" log="$state/api.log"
mkdir -p "$state"

healthy() { curl -fsS -m 2 "http://localhost:$port/healthz" >/dev/null 2>&1; }
our_pid() { [[ -f "$pidf" ]] && kill -0 "$(cat "$pidf")" 2>/dev/null && cat "$pidf"; }
listener() { lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null | head -1 || true; }

# detach DIR CMD...: run CMD in its own session with stdin/stdout/stderr on the
# log file, so nothing the caller pipes (api.sh restart | tail) stays open and
# killing the caller does not kill the server. Writes the pid to $pidf.
detach() {
  local dir="$1"; shift
  : >"$log"
  if command -v setsid >/dev/null 2>&1; then
    (cd "$dir" && exec setsid nohup "$@" </dev/null >>"$log" 2>&1 & echo $! >"$pidf")
  else
    # macOS has no setsid(1); start a new session through perl, which ships with it.
    (cd "$dir" && exec perl -MPOSIX -e 'POSIX::setsid(); exec @ARGV or die "exec: $!"' -- nohup "$@" \
      </dev/null >>"$log" 2>&1 & echo $! >"$pidf")
  fi
  disown 2>/dev/null || true
}

start() {
  if pid="$(our_pid)"; then echo "already running: pid $pid on :$port"; return 0; fi
  if [[ -n "$(listener)" ]]; then
    p="$(listener)"
    echo ":$port is taken by pid $p ($(ps -o comm= -p "$p" 2>/dev/null)). Stop it or set FREEBIE_SERVER_PORT." >&2
    exit 1
  fi
  if [[ "${FREEBIE_DATABASE_PATH:-}" == libsql://* ]]; then
    echo "FREEBIE_DATABASE_PATH points at Turso (production). Run: unset FREEBIE_DATABASE_PATH" >&2
    exit 1
  fi
  echo "building..."
  (cd "$root/services/api" && mise exec -- go build -o "$bin" .) || { echo "build failed" >&2; exit 1; }
  detach "$root/services/api" "$bin" serve --port "$port"
  for _ in $(seq 1 60); do
    if healthy; then
      # Record the process that owns the port; nohup can leave a wrapper pid in $!.
      p="$(listener)"; [[ -n "$p" ]] && echo "$p" >"$pidf"
      echo "up: http://localhost:$port (pid $(cat "$pidf"), log $log, healthz ok)"; return 0
    fi
    kill -0 "$(cat "$pidf")" 2>/dev/null || { echo "server exited; last log lines:" >&2; tail -n 20 "$log" >&2; rm -f "$pidf"; exit 1; }
    sleep 1
  done
  echo "no /healthz after 60s; see $log" >&2
  exit 1
}

stop() {
  if pid="$(our_pid)"; then
    kill "$pid"
    for _ in $(seq 1 10); do kill -0 "$pid" 2>/dev/null || break; sleep 0.5; done
    kill -0 "$pid" 2>/dev/null && kill -9 "$pid"
    rm -f "$pidf"
    echo "stopped pid $pid"
  else
    rm -f "$pidf"
    p="$(listener)"
    if [[ -n "$p" ]]; then
      echo "not started by api.sh, but pid $p ($(ps -o comm= -p "$p" 2>/dev/null)) holds :$port. Stop it where it runs, or: kill $p" >&2
      exit 1
    fi
    echo "not running"
  fi
}

status() {
  if pid="$(our_pid)"; then echo "pid $pid (api.sh), log $log"; else echo "not started by api.sh"; fi
  p="$(listener)"
  [[ -n "$p" ]] && echo ":$port held by pid $p ($(ps -o comm= -p "$p" 2>/dev/null))"
  if healthy; then echo "healthz: ok"; else echo "healthz: no answer on :$port"; return 1; fi
}

case "${1:-}" in
  start) start ;;
  stop) stop ;;
  restart) stop || true; start ;;
  status) status ;;
  logs)
    [[ -f "$log" ]] || { echo "no log yet at $log" >&2; exit 1; }
    if [[ "${2:-}" == "-f" ]]; then tail -n 50 -f "$log"; else tail -n 50 "$log"; fi ;;
  -h|--help) echo "$usage" ;;
  *) echo "$usage" >&2; exit 2 ;;
esac
