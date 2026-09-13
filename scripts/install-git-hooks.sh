#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"
git rev-parse --is-inside-work-tree >/dev/null
git config core.hooksPath .githooks
printf 'Git hooks installed from .githooks.\n'
