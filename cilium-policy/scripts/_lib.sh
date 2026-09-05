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

build_and_load_image() {
  local image_tag="$1"
  local build_context="$2"
  local image_archive
  local status=0

  "${CONTAINER_ENGINE}" build -t "${image_tag}" "${build_context}"

  # kind load docker-image shells out to the docker CLI, so load via an archive
  # to stay engine-agnostic.
  image_archive="$(mktemp -t rc-demos-image)"
  if ! "${CONTAINER_ENGINE}" save "${image_tag}" -o "${image_archive}" ||
    ! kind load image-archive "${image_archive}" --name "${KIND_CLUSTER_NAME}"; then
    status=1
  fi
  rm -f "${image_archive}"

  return "${status}"
}

