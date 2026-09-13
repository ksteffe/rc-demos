#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
demo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
content_log=${TMPDIR:-/tmp}/runtimeconditions-content-api.log
web_log=${TMPDIR:-/tmp}/runtimeconditions-web-demo.log
content_port=${DEMO_CONTENT_PORT:-18090}
web_port=${DEMO_WEB_PORT:-18080}

cleanup() {
  kill "$web_pid" "$content_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

(
  cd "$demo_root"
  PORT="$content_port" go run ./app/cmd/content-api
) >"$content_log" 2>&1 &
content_pid=$!

(
  cd "$demo_root"
  CONTENT_API_URL="http://127.0.0.1:$content_port" \
  GA_MEASUREMENT_ID=G-DEMO123 \
  PORT="$web_port" \
  go run ./app/cmd/web
) >"$web_log" 2>&1 &
web_pid=$!

attempt=0
until page=$(curl --fail --silent "http://127.0.0.1:$web_port/"); do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo "web demo did not become ready" >&2
    cat "$content_log" >&2
    cat "$web_log" >&2
    exit 1
  fi
  sleep 1
done

printf '%s' "$page" | grep -q 'Profiles can compose without changing the Profile specification.'
printf '%s' "$page" | grep -q 'googletagmanager.com/gtag/js?id=G-DEMO123'
echo "smoke test: web page, content API, and Google Analytics markup are present"
