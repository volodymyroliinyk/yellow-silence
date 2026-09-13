# Security policy

## Supported versions

Security fixes are provided for the latest release.

## Reporting

Please report vulnerabilities privately to the repository owner rather than opening a public issue. Include the affected version, reproduction steps, and expected impact. Do not include real screenshots, personal window titles, or credentials.

## Security model

Yellow Silence runs as the logged-in user and requires no root privileges. Screenshots remain in process memory and are subject to encoded-size and pixel-count limits. Encoded screenshot bytes are cleared immediately after decoding. The decoded pixel buffer is cleared when the current detection cycle finishes, before the next polling interval. Capture and audio commands have execution deadlines. Their stderr and device details are discarded rather than included in logs.

The systemd unit denies IP networking and capabilities and enables filesystem, privilege, kernel, hostname, clock, IPC, and executable-memory restrictions. Unix sockets remain available because Wayland/X11, PipeWire, PulseAudio, and D-Bus use them. External capture/audio programs inherit the same service sandbox.

Configured commands are executed directly without a shell. A custom capture executable must use an absolute path. Configuration is limited to 1 MiB, must be a regular file, and is rejected when group- or world-writable (`0600` is used at creation). A malicious screenshot utility has the same user-level screen access as this service, so use distribution-provided executables and review custom commands before enabling them.

Screen capture is inherently privacy-sensitive. The service scans pixels only for the configured color pattern; it does not save screenshots, send network requests, perform OCR, or log image contents. Memory clearing applies to buffers owned by Yellow Silence. The operating system, compositor, Go runtime, and selected screenshot utility can have their own temporary copies and remain part of the trusted computing base.

Release artifacts include `SHA256SUMS`. The Debian updater verifies the downloaded package before invoking `apt`. Checksums detect corruption and mismatched files but do not replace GitHub account security or package signing; users should verify the repository owner and release source.
