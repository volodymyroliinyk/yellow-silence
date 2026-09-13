#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bin_dir="${XDG_BIN_HOME:-${HOME:?}/.local/bin}"
config_dir="${XDG_CONFIG_HOME:-${HOME:?}/.config}/yellow-silence"
unit_dir="${XDG_CONFIG_HOME:-${HOME:?}/.config}/systemd/user"
OUTPUT_DIR="$project_dir/dist" "$project_dir/scripts/build.sh"
install -d -m 0755 "$bin_dir" "$unit_dir"
install -d -m 0700 "$config_dir"
install -m 0755 "$project_dir/dist/yellow-silence" "$bin_dir/yellow-silence"
install -m 0644 "$project_dir/packaging/yellow-silence.service" "$unit_dir/yellow-silence.service"
if [[ ! -e "$config_dir/config.json" ]]; then
  install -m 0600 "$project_dir/config/config.example.json" "$config_dir/config.json"
fi
systemctl --user daemon-reload
systemctl --user enable --now yellow-silence.service
printf 'Installed and started yellow-silence. Edit %s/config.json if needed.\n' "$config_dir"
