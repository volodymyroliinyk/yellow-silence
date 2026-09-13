#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"
if [[ ! -d .git ]]; then
  printf 'Update requires a Git checkout.\n' >&2
  exit 1
fi
if [[ -n "$(git status --porcelain)" ]]; then
  printf 'Refusing to update: the working tree has local changes.\n' >&2
  exit 1
fi
git pull --ff-only
"$project_dir/scripts/test.sh"
"$project_dir/scripts/install.sh"
