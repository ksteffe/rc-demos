#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
demo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
workspace_root=$(CDPATH= cd -- "$demo_root/../.." && pwd)
output_dir=${DEMO_OUTPUT_DIR:-"$demo_root/artifacts"}
extensions_root=${EXTENSIONS_ROOT:-"$workspace_root/extensions"}

cd "$demo_root"

./scripts/generate-application-profile.sh

go run ./cmd/profile-complete \
  -application-profile "$output_dir/application.profiler.yaml" \
  -application-additions examples/application.conditions.yaml \
  -out "$output_dir/application.profile.yaml"

go run ./cmd/profile-validate \
  -profile "$output_dir/application.profile.yaml" \
  -extensions-root "$extensions_root"

go run ./cmd/profile-compose \
  -application-profile "$output_dir/application.profile.yaml" \
  -wrapper-additions examples/dev-container.conditions.yaml \
  -recipe examples/dev-container.compose.yaml \
  -out "$output_dir/dev-container.profile.yaml" \
  -provenance-out "$output_dir/dev-container.provenance.yaml"

go run ./cmd/profile-validate \
  -profile "$output_dir/dev-container.profile.yaml" \
  -extensions-root "$extensions_root"
