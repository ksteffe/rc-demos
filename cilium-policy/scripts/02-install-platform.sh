#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"
use_demo_context

kubectl apply -f "${DEMO_DIR}/cluster/namespaces.yaml"
kubectl apply -f "${DEMO_DIR}/manifests/dependencies.yaml"
kubectl apply -k "${DEMO_DIR}/backstage"
kubectl apply -f "${DEMO_DIR}/manifests/backstage-rbac.yaml"

helm repo add nats https://nats-io.github.io/k8s/helm/charts/ --force-update
helm repo add backstage https://backstage.github.io/charts --force-update
helm repo update nats backstage

helm upgrade --install event-bus nats/nats \
  --namespace dependencies \
  --version "${NATS_CHART_VERSION}" \
  --values "${DEMO_DIR}/helm/nats-values.yaml" \
  --wait \
  --timeout 10m

helm upgrade --install backstage backstage/backstage \
  --namespace backstage \
  --version "${BACKSTAGE_CHART_VERSION}" \
  --values "${DEMO_DIR}/helm/backstage-values.yaml" \
  --wait \
  --timeout 10m

kubectl -n dependencies rollout status statefulset/resource-database --timeout=5m
kubectl -n dependencies rollout status statefulset/availability-cache --timeout=5m
kubectl -n backstage rollout status deployment/backstage --timeout=5m

echo "Backstage, Postgres, Redis, and NATS are ready."
echo "Browse the catalog with: kubectl -n backstage port-forward svc/backstage 7007:7007"

