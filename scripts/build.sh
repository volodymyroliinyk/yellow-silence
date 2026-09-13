#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
output_dir="${OUTPUT_DIR:-$project_dir/dist}"
mkdir -p "$output_dir"
cd "$project_dir"
CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags='-s -w' -o "$output_dir/yellow-silence" ./cmd/yellow-silence
printf 'Built %s/yellow-silence\n' "$output_dir"
