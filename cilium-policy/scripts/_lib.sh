#!/usr/bin/env bash
set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../versions.env
source "${DEMO_DIR}/versions.env"

if command -v docker >/dev/null 2>&1; then
  CONTAINER_ENGINE=docker
elif command -v podman >/dev/null 2>&1; then
  CONTAINER_ENGINE=podman
  # KinD only talks to Podman when this provider is selected explicitly.
  export KIND_EXPERIMENTAL_PROVIDER=podman
else
  CONTAINER_ENGINE=
fi

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    return 1
  fi
}

require_container_engine() {
  if [[ -z "${CONTAINER_ENGINE}" ]]; then
    echo "Missing required command: docker or podman" >&2
    return 1
  fi
}

use_demo_context() {
  kubectl config use-context "kind-${KIND_CLUSTER_NAME}" >/dev/null
}

