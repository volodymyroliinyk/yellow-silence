#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"

version="${1:-}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf 'Usage: %s VERSION (for example: 1.2.3)\n' "$0" >&2
  exit 2
fi
tag="v$version"
release_branch="release/$tag"

command -v gh >/dev/null || { printf 'GitHub CLI (gh) is required.\n' >&2; exit 1; }
gh auth status >/dev/null
[[ "$(git branch --show-current)" == "main" ]] || { printf 'Releases must start from the main branch.\n' >&2; exit 1; }
[[ -z "$(git status --porcelain)" ]] || { printf 'The working tree must be clean.\n' >&2; exit 1; }
git remote get-url origin >/dev/null
git fetch --prune origin
[[ "$(git rev-parse HEAD)" == "$(git rev-parse origin/main)" ]] || { printf 'Local main must match origin/main.\n' >&2; exit 1; }
if git rev-parse "$tag" >/dev/null 2>&1; then
  printf 'Tag %s already exists.\n' "$tag" >&2
  exit 1
fi

previous_tag="$(git describe --tags --abbrev=0 2>/dev/null || true)"
range="HEAD"
[[ -z "$previous_tag" ]] || range="$previous_tag..HEAD"
mapfile -t commits < <(git log "$range" --no-merges --pretty=format:'%s' | sed '/^$/d')
(( ${#commits[@]} > 0 )) || { printf 'There are no commits to release.\n' >&2; exit 1; }

notes="$(mktemp)"
updated_changelog="$(mktemp)"
cleanup() { rm -f "$notes" "$updated_changelog"; }
trap cleanup EXIT
{
  printf '## %s - %s\n\n' "$tag" "$(date -u +%F)"
  for subject in "${commits[@]}"; do printf -- '- %s\n' "$subject"; done
  printf '\n'
} >"$notes"
awk -v notes_file="$notes" '
  { print }
  /^## Unreleased$/ {
    print ""
    while ((getline line < notes_file) > 0) print line
    close(notes_file)
  }
' CHANGELOG.md >"$updated_changelog"

git switch -c "$release_branch"
cp "$updated_changelog" CHANGELOG.md
git add CHANGELOG.md
git commit -m "chore(release): $tag"
git push -u origin "$release_branch"
pr_url="$(gh pr create --base main --head "$release_branch" --title "Release $tag" --body "Updates CHANGELOG.md and prepares $tag.")"
gh pr checks "$pr_url" --watch
gh pr merge "$pr_url" --merge --delete-branch
git switch main
git pull --ff-only origin main
release_dir="$project_dir/dist/release/$tag"
"$project_dir/scripts/build-release.sh" "$version"
git tag -a "$tag" -m "Release $tag"
git push origin "$tag"
gh release create "$tag" "$release_dir"/* --title "$tag" --notes-file "$notes" --verify-tag
printf 'Created GitHub release %s.\n' "$tag"
