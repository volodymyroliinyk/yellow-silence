#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
version="${1:-}"
OUTPUT_DIR="$project_dir/dist" "$project_dir/scripts/build-deb.sh" "$version"
deb_arch="$(dpkg --print-architecture)"
package="$project_dir/dist/yellow-silence_${version}_${deb_arch}.deb"
sudo apt install "$package"
config_dir="${XDG_CONFIG_HOME:-${HOME:?}/.config}/yellow-silence"
if [[ ! -e "$config_dir/config.json" ]]; then
  /usr/bin/yellow-silence init-config
fi
chmod 0600 "$config_dir/config.json"
systemctl --user daemon-reload
systemctl --user enable --now yellow-silence.service
printf 'Installed %s and started the user service.\n' "$package"
