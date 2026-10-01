#!/usr/bin/env bash
# Materialize an ApplicationRelease request by embedding a generated Runtime
# Conditions Profile, unchanged, as spec.profile.
#
#   materialize-application-release.sh BASE_MANIFEST PROFILE OUTPUT
set -euo pipefail

if [[ $# -ne 3 ]]; then
  printf 'usage: %s BASE_MANIFEST PROFILE OUTPUT\n' "$0" >&2
  exit 2
fi

base="$1"
profile="$2"
output="$3"

mkdir -p "$(dirname "${output}")"
{
  cat "${base}"
  printf '  profile: |\n'
  sed 's/^/    /' "${profile}"
} >"${output}"

printf '[platform-demo] materialized %s from %s\n' "${output}" "${profile}"
