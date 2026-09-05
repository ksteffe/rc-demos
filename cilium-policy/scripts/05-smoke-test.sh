#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"
use_demo_context

expectation="${1:-}"
if [[ "${expectation}" != "blocked" && "${expectation}" != "allowed" ]]; then
  echo "Usage: $0 blocked|allowed" >&2
  exit 2
fi

log_file="$(mktemp)"
kubectl -n applications port-forward svc/intake-gateway-v1 18080:8080 >"${log_file}" 2>&1 &
forward_pid=$!
cleanup() {
  kill "${forward_pid}" >/dev/null 2>&1 || true
  rm -f "${log_file}"
}
trap cleanup EXIT

ready=0
for _ in {1..20}; do
  if curl --silent --fail --max-time 1 http://localhost:18080/healthz >/dev/null; then
    ready=1
    break
  fi
  sleep 0.5
done
if [[ "${ready}" -ne 1 ]]; then
  echo "Port-forward did not become ready:" >&2
  sed -n '1,20p' "${log_file}" >&2
  exit 1
fi

response_file="$(mktemp)"
trap 'rm -f "${response_file}"; cleanup' EXIT
status="$(curl --silent --show-error --max-time 10 \
  --output "${response_file}" --write-out '%{http_code}' \
  --request POST http://localhost:18080/requests \
  --header 'content-type: application/json' \
  --data '{"itemId":"blankets","quantity":2,"destination":"Shelter A"}')"

cat "${response_file}"
if [[ "${expectation}" == "allowed" && "${status}" != "201" ]]; then
  echo "Expected HTTP 201 after policy application, got ${status}." >&2
  exit 1
fi
if [[ "${expectation}" == "blocked" && "${status}" == 2* ]]; then
  echo "Expected policy to block the workflow, but it returned HTTP ${status}." >&2
  exit 1
fi

echo "Observed expected ${expectation} behavior (HTTP ${status})."

