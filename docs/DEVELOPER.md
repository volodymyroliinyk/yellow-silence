# Developer guide

## Layout

- `cmd/yellow-silence`: CLI entry point and signal handling.
- `internal/config`: strict JSON parsing, defaults, and validation.
- `internal/process`: browser discovery through read-only `/proc` inspection.
- `internal/capture`: screenshot-command selection and in-memory decoding.
- `internal/detect`: dependency-free pixel detector.
- `internal/audio`: `wpctl`/`pactl` adapters.
- `internal/app`: monitoring state machine and mute ownership.
- `packaging`: systemd user unit.
- `packaging/debian`: Debian control template and package-specific systemd user unit.
- `scripts`: test, binary, Debian package, install, update, and release workflows.

## Development

```bash
./scripts/test.sh
./scripts/build.sh
./dist/yellow-silence check --config config/config.example.json
./scripts/build-deb.sh 0.1.0
./scripts/build-release.sh 0.1.0
```

Run interactively before enabling systemd:

```bash
./dist/yellow-silence run --config config/config.example.json
```

The project intentionally uses only the Go standard library. Keep OS integrations behind narrow package APIs. Commands must use `exec.CommandContext` with separate argument arrays; never invoke a shell with values from configuration. Preserve output-size, image-dimension, and execution-time limits. Never add child-process output to logs or returned errors. Every successful `Capture` call must pair pixel inspection with `defer capture.Release(img)` inside a narrow function scope so the decoded buffer is cleared before audio handling and all return paths remain covered.

## Git workflow

`main` contains released code. `develop` is the integration branch. Create feature and fix branches from `develop`, then merge them through pull requests. Do not commit or push directly to `main`.

Install the repository-managed local hooks after cloning:

```bash
./scripts/install-git-hooks.sh
```

After the repository has been created on GitHub and `origin` points to it, an administrator can apply branch protection:

```bash
gh auth login
./scripts/configure-github.sh
```

This requires pull requests and the `test` status check for changes to `main`. Local hooks provide an earlier warning but are not a substitute for GitHub branch protection.

## Detector contract

A pixel matches when every 8-bit RGB channel differs from the target by no more than the configured tolerance and alpha is at least 50%. Detection succeeds only for a horizontal run meeting both minimum length and consecutive-row thickness. The first match is returned for debug metadata.

## State and failure policy

The process keeps mute ownership only in memory. It never persists a request to unmute across restarts, because a later process cannot safely know whether the user changed audio state. Capture and audio failures are logged and retried on the next poll. Startup fails if no capture or audio backend exists.

Changes to ownership, timeout handling, process matching, or command execution require focused tests. Run the race detector and `go vet` before merging.

## Release/update

`scripts/update.sh` accepts only a clean Git worktree, performs a fast-forward-only pull, runs all checks, reinstalls, and restarts the user service.

Releases are initiated from a laptop with GitHub CLI. Start from a clean, up-to-date `main` branch:

```bash
git switch main
git pull --ff-only
./scripts/release.sh 1.2.3
```

The release script:

1. Verifies the branch, clean worktree, GitHub authentication, remote state, and version.
2. Builds release notes from commit subjects since the previous tag.
3. Creates `release/v1.2.3` and updates `CHANGELOG.md`.
4. Opens a pull request, waits for checks, and merges it into `main` through GitHub.
5. Updates local `main` and builds the release artifacts.
6. Creates and pushes an annotated tag, then creates the GitHub Release with the artifacts.

Release asset names are deterministic:

- `yellow-silence_<version>_linux_<goarch>` for the standalone Go binary;
- `yellow-silence_<version>_<debian-architecture>.deb` for the Debian package;
- `SHA256SUMS` for integrity verification.

Use concise commit subjects because they become changelog entries. Conventional Commit prefixes such as `feat:`, `fix:`, `docs:`, and `chore:` are recommended.
