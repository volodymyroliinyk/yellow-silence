# User guide

## How it works

The program polls continuously while at least one configured browser process is running. On every cycle it captures the current desktop and measures the bar again, so a growing or shrinking progress bar does not rely on a previous measurement. When a matching run reaches `minimum_length_px` for at least `minimum_thickness_px` rows, the default output device is muted.

Audio is restored only when all of these are true:

1. Yellow Silence performed the mute.
2. The bar is no longer visible, or no configured browser is running.
3. Less than `restore_within` has elapsed since the mute.

If the restore window expires, audio remains muted for safety. The user can unmute it normally.

## Installation

Install the dependencies listed in the README, clone the repository, and run:

```bash
./scripts/install.sh
```

This installs only into the current user's home directory and creates a user service. Root access is not required. To uninstall:

```bash
systemctl --user disable --now yellow-silence
rm ~/.config/systemd/user/yellow-silence.service ~/.local/bin/yellow-silence
systemctl --user daemon-reload
```

The command deliberately leaves your configuration in place.

## Configuration reference

The default path is `$XDG_CONFIG_HOME/yellow-silence/config.json` or `~/.config/yellow-silence/config.json`.

| Key | Meaning | Default |
| --- | --- | --- |
| `browsers` | Process names or executable paths. Matching uses the executable basename. | Common browsers |
| `color` | Target color in `#RRGGBB`. The default matches the supplied yellow video progress bar. | `#FFCC00` |
| `color_tolerance` | Maximum difference in each RGB channel, 0–64. | `24` |
| `minimum_length_px` | Minimum horizontal length. | `100` |
| `minimum_thickness_px` | Minimum consecutive rows. | `2` |
| `poll_interval` | Delay between consecutive screen checks while a browser is running. Go duration, at least `100ms`. | `200ms` |
| `restore_within` | Maximum time during which automatic unmute is allowed. | `5m` |
| `audio_backend` | `auto`, `wpctl`, or `pactl`. | `auto` |
| `log_level` | `debug`, `info`, `warn`, or `error`. | `info` |
| `capture_command` | Optional absolute executable path and arguments that write a PNG or JPEG to stdout. | auto-detect |

Validate after editing:

```bash
yellow-silence check
```

The service rejects a configuration that other users can modify. If validation reports unsafe permissions, run `chmod 0600 ~/.config/yellow-silence/config.json`.

For diagnosis, set `log_level` to `debug`, restart the service, and read:

```bash
journalctl --user -u yellow-silence --since today
systemctl --user status yellow-silence
```

## Wayland notes

Wayland compositors intentionally restrict screen capture. `grim` works directly on wlroots-based compositors. GNOME may display a capture permission prompt or reject unattended capture. If your desktop offers a trusted screenshot command, set `capture_command`; its first element must be an absolute path and it must emit PNG/JPEG bytes on standard output. Do not place secrets in this array because command arguments can be visible to operating-system process tools.

If a manually launched command works but the service does not, import the graphical environment and restart:

```bash
systemctl --user import-environment DISPLAY WAYLAND_DISPLAY XAUTHORITY XDG_CURRENT_DESKTOP
systemctl --user restart yellow-silence
```

## Performance and privacy

Screenshots remain in memory and are never saved or logged. Increase `poll_interval` to reduce CPU usage. The detector scans the full combined desktop, so its cost grows with total screen resolution.
