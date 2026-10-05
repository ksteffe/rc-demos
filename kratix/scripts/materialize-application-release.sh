#!/usr/bin/env bash
# Materialize an ApplicationRelease request by embedding a generated Runtime
# Conditions Profile, unchanged, as spec.profile.
#
#   materialize-application-release.sh NAME PROFILE OUTPUT
set -euo pipefail

if [[ $# -ne 3 ]]; then
  printf 'usage: %s NAME PROFILE OUTPUT\n' "$0" >&2
  exit 2
fi

name="$1"
profile="$2"
output="$3"

mkdir -p "$(dirname "${output}")"
{
  cat <<EOF_MANIFEST
apiVersion: platform.demoteam.io/v1alpha1
kind: ApplicationRelease
metadata:
  name: ${name}
  namespace: demo
spec:
  image: ghcr.io/runtimeconditions/request-logger:latest
  imagePullPolicy: Always
  port: 8080
  readinessPath: /ready
  catalog:
    configMapRef:
      name: platform-api-catalog
      namespace: platform-demo-system
  profile: |
EOF_MANIFEST
  sed 's/^/    /' "${profile}"
} >"${output}"

printf '[platform-demo] materialized %s from %s\n' "${output}" "${profile}"
