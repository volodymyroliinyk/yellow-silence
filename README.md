# Yellow Silence

`yellow-silence` is a headless Ubuntu/Linux service that watches the screen while a configured browser is running. It mutes the default audio output when a horizontal bar of the configured color is present and restores audio after the bar disappears, provided the configured restore window has not expired.

The default detector profile targets the `#FFCC00`, 3-pixel yellow video progress bar shown in the supplied reference screenshots. While a configured browser is running, the current screen is checked every 150 ms by default. A changing partial bar is detected whenever its current length is at least 100 pixels.

The service never unmutes audio that was already muted before detection.

## Requirements

- [Go 1.27.1+](https://go.dev/dl/) to build.
- Ubuntu or another Linux distribution with systemd user services.
- PipeWire/WirePlumber (`wpctl`) or PulseAudio (`pactl`).
- XDG Desktop Portal, PipeWire, and GStreamer with its PipeWire plugin on Wayland.
- `maim` or `scrot` when using the legacy X11 fallback.
- `dpkg-deb` and `gzip` when building a Debian package.

Ubuntu X11 example:

```bash
sudo apt install pulseaudio-utils maim
```

Ubuntu Wayland example:

```bash
sudo apt install wireplumber xdg-desktop-portal-gnome gstreamer1.0-tools gstreamer1.0-pipewire
```

On Wayland, the desktop presents a monitor-sharing dialog the first time the
service needs a frame. The permission lasts until the service stops.

Desktop security policies may require permission for screen capture. See [User guide](docs/USER.md).

## Quick start

```bash
./scripts/test.sh
./scripts/install.sh
journalctl --user -u yellow-silence -f
```

To build and install a Debian package instead:

```bash
./scripts/install-deb.sh 0.1.0
```

Edit `~/.config/yellow-silence/config.json`, then reload with:

```bash
systemctl --user restart yellow-silence
```

CLI commands:

```text
yellow-silence init-config [--config PATH]
yellow-silence check [--config PATH]
yellow-silence run [--config PATH]
yellow-silence version
```

## Documentation

- [User guide](docs/USER.md)
- [Developer guide](docs/DEVELOPER.md)
- [Architecture and AI-agent guide](docs/AI_AGENT.md)
- [Security](docs/SECURITY.md)

Development takes place on `develop` and topic branches. Direct commits and pushes to `main` are blocked by repository hooks and should also be blocked by GitHub branch protection. See the [developer guide](docs/DEVELOPER.md#git-workflow) for setup and release commands.

## License

MIT. Note that the MIT license permits commercial as well as non-commercial use. See [LICENSE](LICENSE).
