#!/bin/sh
# Development consumer of the canonical request-logger-http Profile.
#
#   1. evaluate every Condition and bind it to a local resource (dev-bind)
#   2. start the local resources those bindings name
#   3. run request-logger-http with the generated environment
#   4. check that /demo reaches both dependencies
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
demo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
repo_root=$(CDPATH= cd -- "$demo_root/.." && pwd)
profile=${PROFILE:-"$repo_root/artifacts/request-logger-http.profile.yaml"}
todos_port=${DEV_TODOS_API_PORT:-18081}
redis_port=${DEV_REDIS_PORT:-16379}
app_port=${DEV_APP_PORT:-18080}
work_dir=$(mktemp -d)
env_file="$work_dir/request-logger.dev.env"

redis_pid=""
redis_container=""
todos_pid=""
app_pid=""
cleanup() {
  for pid in $app_pid $todos_pid $redis_pid; do
    kill "$pid" 2>/dev/null || true
  done
  if [ -n "$redis_container" ]; then
    docker rm -f "$redis_container" >/dev/null 2>&1 || true
  fi
  rm -rf "$work_dir"
}
trap cleanup EXIT INT TERM

log() { printf '[dev] %s\n' "$*"; }

log "consuming $profile"
(cd "$demo_root" && go run ./cmd/dev-bind \
  -profile "$profile" \
  -todos-api-url "http://127.0.0.1:$todos_port" \
  -redis-addr "127.0.0.1:$redis_port" \
  -env-out "$env_file")

log "building workloads from their module roots"
(cd "$repo_root/apps/todos-api" && go build -o "$work_dir/todos-api" .)
(cd "$repo_root/apps/request-logger-http" && go build -o "$work_dir/request-logger" .)

if command -v redis-server >/dev/null 2>&1; then
  log "starting local redis-server on 127.0.0.1:$redis_port"
  redis-server --port "$redis_port" --bind 127.0.0.1 --save '' --appendonly no >"$work_dir/redis.log" 2>&1 &
  redis_pid=$!
else
  log "starting redis:7-alpine container on 127.0.0.1:$redis_port"
  redis_container=$(docker run -d --rm -p "127.0.0.1:$redis_port:6379" redis:7-alpine)
fi

log "starting local todos-api on 127.0.0.1:$todos_port"
PORT="$todos_port" "$work_dir/todos-api" >"$work_dir/todos-api.log" 2>&1 &
todos_pid=$!

log "starting request-logger-http on 127.0.0.1:$app_port with:"
sed 's/^/[dev]   /' "$env_file"
(
  set -a
  . "$env_file"
  set +a
  PORT="$app_port" exec "$work_dir/request-logger"
) >"$work_dir/request-logger.log" 2>&1 &
app_pid=$!

attempt=0
until curl -fsS "http://127.0.0.1:$app_port/ready" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    log "request-logger-http did not become ready" >&2
    cat "$work_dir"/*.log >&2
    exit 1
  fi
  sleep 1
done

response=$(curl -sS "http://127.0.0.1:$app_port/demo")
log "/demo -> $response"
printf '%s' "$response" | grep -q '"todosApi":"ok"' || { log "todos API check failed" >&2; exit 1; }
printf '%s' "$response" | grep -q '"cache":"ok"' || { log "cache check failed" >&2; exit 1; }
log "development demo passed: same Profile, local fulfillment"
