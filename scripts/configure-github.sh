#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"
command -v gh >/dev/null || { printf 'GitHub CLI (gh) is required.\n' >&2; exit 1; }
gh auth status >/dev/null
repository="$(gh repo view --json nameWithOwner --jq .nameWithOwner)"
payload="$(mktemp)"
trap 'rm -f "$payload"' EXIT
printf '%s\n' '{
  "required_status_checks": {"strict": true, "contexts": ["test"]},
  "enforce_admins": true,
  "required_pull_request_reviews": {"required_approving_review_count": 0},
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}' >"$payload"
gh api --method PUT "repos/$repository/branches/main/protection" --input "$payload" >/dev/null
printf 'Branch protection configured for %s: pull requests and the test check are required.\n' "$repository"
