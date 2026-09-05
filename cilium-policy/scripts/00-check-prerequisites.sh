#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"

missing=0
if require_container_engine; then
  printf "ok  %s\n" "${CONTAINER_ENGINE}"
  if ! "${CONTAINER_ENGINE}" info >/dev/null 2>&1; then
    echo "${CONTAINER_ENGINE} is installed but its daemon is not available." >&2
    missing=1
  fi
else
  missing=1
fi

for command_name in kind helm kubectl curl; do
  if require_command "${command_name}"; then
    printf "ok  %s\n" "${command_name}"
  else
    missing=1
  fi
done

if [[ "${missing}" -ne 0 ]]; then
  exit 1
fi

echo "All prerequisites are ready."

