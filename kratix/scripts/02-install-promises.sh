#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_DIR="$(cd "${SCRIPT_DIR}/../manifests" && pwd)"
IMAGE_PULL_POLICY="${IMAGE_PULL_POLICY:-Always}"

printf '[platform-demo] installing Redis Promise\n'
sed "s|imagePullPolicy: Always|imagePullPolicy: ${IMAGE_PULL_POLICY}|" "${MANIFEST_DIR}/promises/redis.yaml" | kubectl apply -f -
kubectl wait --for=create crd/redis.platform.demoteam.io --timeout=120s
kubectl wait --for=condition=Established crd/redis.platform.demoteam.io --timeout=120s

printf '[platform-demo] installing ApplicationRelease Promise\n'
sed "s|imagePullPolicy: Always|imagePullPolicy: ${IMAGE_PULL_POLICY}|" "${MANIFEST_DIR}/promises/application-release.yaml" | kubectl apply -f -
kubectl wait --for=create crd/applicationreleases.platform.demoteam.io --timeout=120s
kubectl wait --for=condition=Established crd/applicationreleases.platform.demoteam.io --timeout=120s

printf '[platform-demo] platform Promises are installed\n'
kubectl get crds -l kratix.io/promise-name
