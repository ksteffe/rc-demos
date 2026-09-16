#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
demo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
output_dir=${DEMO_OUTPUT_DIR:-"$demo_root/artifacts"}

cd "$demo_root"

./scripts/generate-application-profile.sh

go run ./cmd/profile-compose \
  -application-profile "$output_dir/application.profiler.yaml" \
  -application-additions examples/application.conditions.yaml \
  -wrapper-additions examples/dev-container.conditions.yaml \
  -recipe examples/dev-container.compose.yaml \
  -out "$output_dir/dev-container.profile.yaml" \
  -provenance-out "$output_dir/dev-container.provenance.yaml"
