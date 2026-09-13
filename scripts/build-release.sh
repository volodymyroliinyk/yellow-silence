#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
version="${1:-}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf 'Usage: %s VERSION (for example: 1.2.3)\n' "$0" >&2
  exit 2
fi
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"
[[ "$goos" == "linux" ]] || { printf 'Release binaries must be built on or for Linux.\n' >&2; exit 1; }
release_root="${RELEASE_ROOT:-$project_dir/dist/release}"
release_dir="$release_root/v$version"
if [[ -e "$release_dir" ]]; then
  printf 'Release output already exists: %s\n' "$release_dir" >&2
  exit 1
fi
mkdir -p "$release_dir"
VERSION="$version" OUTPUT_DIR="$release_dir" "$project_dir/scripts/build.sh"
binary_name="yellow-silence_${version}_linux_${goarch}"
mv "$release_dir/yellow-silence" "$release_dir/$binary_name"
OUTPUT_DIR="$release_dir" "$project_dir/scripts/build-deb.sh" "$version"
(
  cd "$release_dir"
  sha256sum yellow-silence_* >SHA256SUMS
)
printf 'Release artifacts are in %s\n' "$release_dir"
