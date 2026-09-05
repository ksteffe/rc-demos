#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"
use_demo_context

run_probe() {
  local name="$1"
  shift
  kubectl -n applications run "${name}" \
    --quiet \
    --rm \
    --restart=Never \
    --image=curlimages/curl:8.15.0 \
    --labels=backstage.io/kubernetes-id=request-coordinator,network.runtimeconditions.io/plane=data \
    --command -- curl --silent --show-error --fail --max-time 5 "$@"
}

run_probe allowed-inventory \
  http://stock-provider-v2.applications.svc.cluster.local:8080/items/blankets/availability

if run_probe undeclared-admin --request POST \
  http://stock-provider-v2.applications.svc.cluster.local:8080/admin/items; then
  echo "ERROR: undeclared POST /admin/items was reachable." >&2
  exit 1
fi

if run_probe undeclared-service \
  http://supporter-index-v1.applications.svc.cluster.local:8080/donors/example; then
  echo "ERROR: undeclared donor service was reachable." >&2
  exit 1
fi

echo "Declared L7 access works; undeclared path and destination remain blocked."

