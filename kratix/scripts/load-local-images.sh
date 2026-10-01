#!/usr/bin/env bash
# Build the demo images under the names the manifests use and load them into a
# KinD cluster. Run the other scripts with IMAGE_PULL_POLICY=IfNotPresent so
# the cluster uses these images instead of pulling from ghcr.io.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
EXTENSIONS_ROOT="${EXTENSIONS_ROOT:-${REPO_ROOT}/../extensions}"
KIND_CLUSTER="${KIND_CLUSTER:-rc-demos}"
REGISTRY="ghcr.io/runtimeconditions"

build() {
  printf '[platform-demo] building %s\n' "$1"
  docker build -q -t "${REGISTRY}/$1:latest" "${@:2}"
}

build application-release-pipeline "${REPO_ROOT}/kratix/promises/application-release/pipeline"
build redis-pipeline "${REPO_ROOT}/kratix/promises/redis/pipeline"
build todos-api "${REPO_ROOT}/apps/todos-api"

# request-logger's Dockerfile expects extensions/ and rc-demos/ side by side.
context="$(mktemp -d)"
trap 'rm -rf "${context}"' EXIT
mkdir -p "${context}/rc-demos/apps"
cp -R "${EXTENSIONS_ROOT}" "${context}/extensions"
rm -rf "${context}/extensions/.git"
cp -R "${REPO_ROOT}/apps/request-logger-http" "${context}/rc-demos/apps/"
build request-logger -f "${context}/rc-demos/apps/request-logger-http/Dockerfile" "${context}"

for image in application-release-pipeline redis-pipeline todos-api request-logger; do
  kind load docker-image "${REGISTRY}/${image}:latest" --name "${KIND_CLUSTER}"
done

printf '[platform-demo] local images loaded into kind cluster %s\n' "${KIND_CLUSTER}"
