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

### Debian package

Build and install a local package with:

```bash
./scripts/install-deb.sh 0.1.0
```

The package uses the Debian filename format `yellow-silence_<version>_<architecture>.deb` and installs the binary at `/usr/bin/yellow-silence` plus the user-service unit at `/usr/lib/systemd/user/yellow-silence.service`. The script creates the per-user configuration and starts the service after `apt` installs the package.

To update from the latest GitHub Release:

```bash
./scripts/update-deb.sh
```

The updater requires GitHub CLI authentication and `sudo` access for `apt`. It downloads the package matching the current Debian architecture, verifies it against the release `SHA256SUMS`, installs it through `apt`, and restarts the user service. Set `YELLOW_SILENCE_REPOSITORY=owner/repository` when using a fork.

## Configuration reference

The default path is `$XDG_CONFIG_HOME/yellow-silence/config.json` or `~/.config/yellow-silence/config.json`.

| Key | Meaning | Default |
| --- | --- | --- |
| `browsers` | Process names or executable paths. Matching uses the executable basename. | Common browsers |
| `color` | Target color in `#RRGGBB`. The default matches the supplied yellow video progress bar. | `#FFCC00` |
| `color_tolerance` | Maximum difference in each RGB channel, 0–64. | `24` |
| `minimum_length_px` | Minimum horizontal length. Lower values detect the bar earlier but increase false-positive risk. | `40` |
| `minimum_thickness_px` | Minimum consecutive rows. | `2` |
| `poll_interval` | Delay between consecutive screen checks while a browser is running. Go duration, at least `100ms`. | `100ms` |
| `capture_frame_rate` | Maximum Wayland Portal/PipeWire frames per second. Higher values react faster but use more CPU and memory bandwidth. | `10` |
| `disappearance_confirmation_frames` | Consecutive frames without the target bar required before audio is restored. Higher values prevent brief detection gaps from restoring audio. | `3` |
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

Wayland capture uses the XDG Desktop Portal and a restricted PipeWire stream.
When a configured browser first runs, the desktop asks which monitor to share.
The permission and stream last only until Yellow Silence stops; no permission
token is persisted. GNOME displays its normal screen-sharing indicator while
the stream is active.

Frames travel from PipeWire through GStreamer into the service as fixed-size
RGBA pixels over an anonymous pipe. They are never encoded or written to disk.
The service limits capture to ten frames per second and rejects unsafe image
dimensions before starting the frame receiver.

If your desktop offers a different trusted capture command, `capture_command`
still overrides the portal backend. Its first element must be an absolute path
and it must emit PNG/JPEG bytes on standard output. Do not place secrets in this
array because command arguments can be visible to operating-system process tools.

If a manually launched command works but the service does not, import the graphical environment and restart:

```bash
systemctl --user import-environment DISPLAY WAYLAND_DISPLAY XAUTHORITY XDG_CURRENT_DESKTOP XDG_SESSION_TYPE
systemctl --user restart yellow-silence
```

## Performance and privacy

Screenshots remain in memory and are never saved or logged. Increase `poll_interval` to reduce CPU usage. The detector scans the full combined desktop, so its cost grows with total screen resolution.
