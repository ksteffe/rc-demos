#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"

require_container_engine
for command_name in kind helm kubectl; do
  require_command "${command_name}"
done

if kind get clusters | grep -Fxq "${KIND_CLUSTER_NAME}"; then
  echo "KinD cluster ${KIND_CLUSTER_NAME} already exists; leaving it unchanged."
else
  kind create cluster \
    --name "${KIND_CLUSTER_NAME}" \
    --image "${KIND_NODE_IMAGE}" \
    --config "${DEMO_DIR}/cluster/kind.yaml"
fi
use_demo_context

helm repo add cilium https://helm.cilium.io/ --force-update
helm repo update cilium
helm upgrade --install cilium cilium/cilium \
  --namespace kube-system \
  --version "${CILIUM_CHART_VERSION}" \
  --values "${DEMO_DIR}/helm/cilium-values.yaml" \
  --wait \
  --timeout 10m

kubectl -n kube-system rollout status daemonset/cilium --timeout=5m
echo "KinD cluster ${KIND_CLUSTER_NAME} is ready with Cilium ${CILIUM_CHART_VERSION}."
