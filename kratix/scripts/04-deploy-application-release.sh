#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
PROFILE="${PROFILE:-${REPO_ROOT}/artifacts/request-logger-http.profile.yaml}"
RELEASE="${REPO_ROOT}/kratix/generated/request-logger-application-release.yaml"

"${SCRIPT_DIR}/materialize-application-release.sh" \
  "request-logger" \
  "${PROFILE}" \
  "${RELEASE}"

printf '[platform-demo] submitting ApplicationRelease request through Kratix\n'
sed "s|^  imagePullPolicy: Always|  imagePullPolicy: ${IMAGE_PULL_POLICY:-Always}|" "${RELEASE}" | kubectl apply -f -

printf '[platform-demo] waiting for ApplicationRelease configure workflow\n'
kubectl -n demo wait applicationrelease/request-logger \
  --for=condition=ConfigureWorkflowCompleted \
  --timeout=180s

printf '[platform-demo] waiting for generated Redis request\n'
kubectl -n demo wait redis/request-logger-cache \
  --for=create \
  --timeout=180s

kubectl -n demo wait redis/request-logger-cache \
  --for=condition=ConfigureWorkflowCompleted \
  --timeout=180s

printf '[platform-demo] waiting for generated application Deployment\n'
kubectl -n demo rollout status deployment/request-logger --timeout=240s

kubectl -n demo get applicationrelease request-logger
kubectl -n demo get redis request-logger-cache
kubectl -n demo get deployment request-logger
