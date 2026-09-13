# Yellow Silence

`yellow-silence` is a headless Ubuntu/Linux service that watches the screen while a configured browser is running. It mutes the default audio output when a horizontal bar of the configured color is present and restores audio after the bar disappears, provided the configured restore window has not expired.

The default detector profile targets the `#FFCC00`, 3-pixel yellow video progress bar shown in the supplied reference screenshots; partial bars are detected as soon as they are at least 100 pixels long.

The service never unmutes audio that was already muted before detection.

## Requirements

- Go 1.22+ to build.
- Ubuntu or another Linux distribution with systemd user services.
- PipeWire/WirePlumber (`wpctl`) or PulseAudio (`pactl`).
- A screenshot tool: `grim` on Wayland, or `maim`/`scrot` on X11.

Ubuntu X11 example:

```bash
sudo apt install golang-go pulseaudio-utils maim
```

Ubuntu Wayland example:

```bash
sudo apt install golang-go wireplumber grim
```

Desktop security policies may require permission for screen capture. See [User guide](docs/USER.md).

## Quick start

```bash
./scripts/test.sh
./scripts/install.sh
journalctl --user -u yellow-silence -f
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
