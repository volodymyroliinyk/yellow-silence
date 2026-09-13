# Changelog

Notable changes to this project are recorded in this file. Releases follow [Semantic Versioning](https://semver.org/).

## Unreleased

- Changed the default browser monitoring interval from 500 ms to 200 ms.
- Added regression coverage for a progress bar that changes length between screen captures.
- Limited screenshot and audio-command output, image dimensions, command duration, and configuration size.
- Removed child-process output from errors and logs.
- Added network, capability, IPC, and operating-system resource restrictions to the systemd service.
- The installer now corrects configuration permissions to owner-only read/write access.
- Added tests for mute ownership, restore deadlines, and failed unmute retries.
- Encoded screenshots and decoded pixel buffers are now cleared as soon as their detection cycle no longer needs them.
- Added Debian package build, install, and verified update scripts.
- GitHub releases now include a versioned Linux binary, a versioned Debian package, and SHA-256 checksums.
- Debian package ownership and modes are normalized independently of the builder's umask.
