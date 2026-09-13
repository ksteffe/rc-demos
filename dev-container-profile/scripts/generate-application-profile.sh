#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
demo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
workspace_root=$(CDPATH= cd -- "$demo_root/../.." && pwd)
output_dir=${DEMO_OUTPUT_DIR:-"$demo_root/artifacts"}

mkdir -p "$output_dir"
cd "$workspace_root/go-rc-profiler"
go run . \
  -dir "$demo_root/app" \
  -extensions-root "$workspace_root/extensions" \
  -name web-profile-demo \
  -workload-uri https://github.com/runtimeconditions/rc-demos/tree/main/dev-container-profile/app \
  -workload-version demo \
  -out "$output_dir/application.profiler.yaml"

echo "profiler output: $output_dir/application.profiler.yaml"
