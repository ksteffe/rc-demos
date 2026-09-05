#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"
use_demo_context

require_container_engine
"${CONTAINER_ENGINE}" build -t localhost/rc-resource-demo:local "${DEMO_DIR}/apps/resource-demo"

# kind load docker-image shells out to the docker CLI, so load via an archive
# to stay engine-agnostic.
image_archive="$(mktemp -t rc-resource-demo)"
trap 'rm -f "${image_archive}"' EXIT
"${CONTAINER_ENGINE}" save localhost/rc-resource-demo:local -o "${image_archive}"
kind load image-archive "${image_archive}" --name "${KIND_CLUSTER_NAME}"

kubectl apply -f "${DEMO_DIR}/manifests/applications.yaml"

for deployment_name in inventory-service fulfillment-service donor-directory request-coordinator notification-worker; do
  kubectl -n applications rollout status "deployment/${deployment_name}" --timeout=5m
done

echo "The community resource fulfillment application is ready."

