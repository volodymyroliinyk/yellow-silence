# Security policy

## Supported versions

Security fixes are provided for the latest release.

## Reporting

Please report vulnerabilities privately to the repository owner rather than opening a public issue. Include the affected version, reproduction steps, and expected impact. Do not include real screenshots, personal window titles, or credentials.

## Security model

Yellow Silence runs as the logged-in user, requires no root privileges, does not use the network, and keeps screenshots only in process memory. The systemd unit enables filesystem, privilege, kernel, and executable-memory restrictions. External capture/audio programs inherit the same service sandbox.

Configured commands are executed directly without a shell. Configuration should be writable only by its owner (`0600` is used at creation). A malicious screenshot utility or modified executable in `PATH` has the same user-level access as this service, so use distribution-provided packages and a trusted service-manager environment.
