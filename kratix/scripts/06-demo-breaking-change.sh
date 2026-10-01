#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_DIR="$(cd "${SCRIPT_DIR}/../manifests" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
PROFILE="${PROFILE:-${REPO_ROOT}/artifacts/request-logger-http.profile.yaml}"
RELEASE="${REPO_ROOT}/kratix/generated/request-logger-breaking-application-release.yaml"

printf '[platform-demo] publishing breaking API catalog bundle\n'
kubectl apply -f "${MANIFEST_DIR}/catalog/todos-api-catalog-breaking.yaml"

"${SCRIPT_DIR}/materialize-application-release.sh" \
  "${MANIFEST_DIR}/apps/request-logger-breaking-application-release.base.yaml" \
  "${PROFILE}" \
  "${RELEASE}"

printf '[platform-demo] submitting an ApplicationRelease expected to fail contract validation\n'
sed "s|^  imagePullPolicy: Always|  imagePullPolicy: ${IMAGE_PULL_POLICY:-Always}|" "${RELEASE}" | kubectl apply -f -
set +e
kubectl -n demo wait applicationrelease/request-logger-breaking \
  --for=condition=ConfigureWorkflowCompleted \
  --timeout=120s
WAIT_STATUS=$?
set -e

if [[ "${WAIT_STATUS}" -eq 0 ]]; then
  printf '[platform-demo] breaking OpenAPI deployment unexpectedly succeeded\n' >&2
  exit 1
fi

printf '[platform-demo] ApplicationRelease failed as expected. Recent workflow logs:\n'
kubectl -n demo get pods -l kratix.io/promise-name=application-release || true
for pod in $(kubectl -n demo get pods -l kratix.io/promise-name=application-release -o name 2>/dev/null | tail -n 3); do
  printf '[platform-demo] logs for %s\n' "${pod}"
  kubectl -n demo logs "${pod}" --all-containers=true --tail=120 || true
done

printf '[platform-demo] restoring compatible API catalog bundle\n'
kubectl apply -f "${MANIFEST_DIR}/catalog/todos-api-catalog.yaml"

printf '[platform-demo] breaking change demo completed\n'
