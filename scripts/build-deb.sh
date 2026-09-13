#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
version="${1:-}"
output_dir="${OUTPUT_DIR:-$project_dir/dist}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.+~][0-9A-Za-z.+~-]+)?$ ]]; then
  printf 'Usage: %s VERSION (for example: 1.2.3)\n' "$0" >&2
  exit 2
fi
command -v dpkg-deb >/dev/null || { printf 'dpkg-deb is required.\n' >&2; exit 1; }
deb_arch="$(dpkg --print-architecture)"
case "$deb_arch" in
  amd64) go_arch=amd64 ;;
  arm64) go_arch=arm64 ;;
  armhf) go_arch=arm ;;
  *) printf 'Unsupported Debian architecture: %s\n' "$deb_arch" >&2; exit 1 ;;
esac

staging="$(mktemp -d)"
cleanup() { rm -rf "$staging"; }
trap cleanup EXIT
cd "$project_dir"
chmod 0755 "$staging"
install -d -m 0755 "$staging/DEBIAN" "$staging/usr/bin" "$staging/usr/lib/systemd/user" "$staging/usr/share/doc/yellow-silence" "$output_dir"
sed -e "s/@VERSION@/$version/g" -e "s/@ARCH@/$deb_arch/g" "$project_dir/packaging/debian/control.in" >"$staging/DEBIAN/control"
GOOS=linux GOARCH="$go_arch" CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -X main.version=$version" -o "$staging/usr/bin/yellow-silence" ./cmd/yellow-silence
chmod 0755 "$staging/usr/bin/yellow-silence"
install -m 0644 "$project_dir/packaging/debian/yellow-silence.service" "$staging/usr/lib/systemd/user/yellow-silence.service"
install -m 0644 "$project_dir/config/config.example.json" "$staging/usr/share/doc/yellow-silence/config.example.json"
install -m 0644 "$project_dir/README.md" "$staging/usr/share/doc/yellow-silence/README.md"
install -m 0644 "$project_dir/LICENSE" "$staging/usr/share/doc/yellow-silence/copyright"
gzip -n -9 -c "$project_dir/CHANGELOG.md" >"$staging/usr/share/doc/yellow-silence/changelog.gz"
chmod 0644 "$staging/usr/share/doc/yellow-silence/changelog.gz"
package="$output_dir/yellow-silence_${version}_${deb_arch}.deb"
dpkg-deb --root-owner-group --build "$staging" "$package" >/dev/null
printf 'Built %s\n' "$package"
