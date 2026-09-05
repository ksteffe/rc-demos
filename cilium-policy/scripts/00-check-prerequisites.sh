#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"

missing=0
for command_name in docker kind helm kubectl curl; do
  if require_command "${command_name}"; then
    printf "ok  %s\n" "${command_name}"
  else
    missing=1
  fi
done

if ! docker info >/dev/null 2>&1; then
  echo "Docker is installed but its daemon is not available." >&2
  missing=1
fi

if [[ "${missing}" -ne 0 ]]; then
  exit 1
fi

echo "All prerequisites are ready."

