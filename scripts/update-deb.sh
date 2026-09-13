#!/usr/bin/env bash
set -euo pipefail
repository="${YELLOW_SILENCE_REPOSITORY:-volodymyroliinyk/yellow-silence}"
command -v gh >/dev/null || { printf 'GitHub CLI (gh) is required.\n' >&2; exit 1; }
gh auth status >/dev/null
command -v dpkg >/dev/null || { printf 'dpkg is required.\n' >&2; exit 1; }
arch="$(dpkg --print-architecture)"
version="$(gh release view --repo "$repository" --json tagName --jq '.tagName | ltrimstr("v")')"
package="yellow-silence_${version}_${arch}.deb"
download_dir="$(mktemp -d)"
cleanup() { rm -rf "$download_dir"; }
trap cleanup EXIT
gh release download "v$version" --repo "$repository" --dir "$download_dir" --pattern "$package" --pattern SHA256SUMS
(
  cd "$download_dir"
  checksum_line="$(grep -F "  $package" SHA256SUMS)"
  [[ -n "$checksum_line" ]] || { printf 'Checksum for %s is missing.\n' "$package" >&2; exit 1; }
  printf '%s\n' "$checksum_line" | sha256sum -c -
)
sudo apt install "$download_dir/$package"
systemctl --user daemon-reload
systemctl --user restart yellow-silence.service
printf 'Updated Yellow Silence to %s.\n' "$version"
