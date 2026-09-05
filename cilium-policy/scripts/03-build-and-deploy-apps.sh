#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"
use_demo_context

require_container_engine
build_and_load_image localhost/rc-resource-demo:local "${DEMO_DIR}/apps/resource-demo"

kubectl apply -f "${DEMO_DIR}/manifests/applications.yaml"

for deployment_name in inventory-service fulfillment-service donor-directory request-coordinator notification-worker; do
  kubectl -n applications rollout status "deployment/${deployment_name}" --timeout=5m
done

echo "The community resource fulfillment application is ready."

