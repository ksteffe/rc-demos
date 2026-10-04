#!/bin/sh
# Generate the canonical request-logger-http Profile from application source.
#
#   generate-profile.sh          write artifacts/request-logger-http.profile.yaml
#   generate-profile.sh --check  fail if the committed Profile differs from a fresh generation
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
workspace_root=$(CDPATH= cd -- "$repo_root/.." && pwd)
profiler_root=${PROFILER_ROOT:-"$workspace_root/go-rc-profiler"}
profile="$repo_root/artifacts/request-logger-http.profile.yaml"

mode=${1:-generate}

generated=$(mktemp)
trap 'rm -f "$generated" "$generated.raw"' EXIT

(
  cd "$profiler_root"
  go run . \
    -dir "$repo_root/apps/request-logger-http" \
    -name request-logger-http \
    -workload-uri github.com/runtimeconditions/rc-demos/apps/request-logger-http \
    -workload-version dev \
    -out "$generated.raw"
)

{
  echo "# Generated from apps/request-logger-http by portable-profile/scripts/generate-profile.sh."
  echo "# Do not edit; regenerate instead."
  cat "$generated.raw"
} >"$generated"

case "$mode" in
  --check)
    if ! diff -u "$profile" "$generated"; then
      echo "generate-profile: $profile is stale; run portable-profile/scripts/generate-profile.sh" >&2
      exit 1
    fi
    echo "generate-profile: $profile matches application source"
    ;;
  generate)
    mkdir -p "$(dirname -- "$profile")"
    cp "$generated" "$profile"
    echo "generate-profile: wrote $profile"
    ;;
  *)
    echo "usage: $0 [--check]" >&2
    exit 2
    ;;
esac
