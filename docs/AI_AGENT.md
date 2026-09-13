# AI-agent guide

This file is the operational context for automated coding agents working on Yellow Silence.

## Product invariants

1. Never unmute sound unless the current process successfully muted it and the restore deadline has not expired.
2. Never store screenshots, pixel samples, window titles, command output containing user data, or audio-device metadata in logs.
3. Never execute configuration through a shell. Preserve argument boundaries.
4. Browser entries represent process executable names/paths, not URLs.
5. Keep the app headless and keep the CLI usable without systemd.
6. Configuration parsing must reject unknown fields and unsafe/invalid values.
7. Do not commit or push directly to `main`; use a pull request from `develop`, a topic branch, or a release branch.

## Change workflow

Read `README.md`, `docs/DEVELOPER.md`, and the package being changed. Preserve the standard-library-only design unless a dependency has a documented security and maintenance justification. Add focused table-driven tests, then run `./scripts/test.sh`, `./scripts/build.sh`, and `git diff --check`.

Do not weaken the systemd sandbox without documenting the exact desktop/backend incompatibility that requires it. Do not add automatic privilege escalation. Installation remains per-user because screen capture and audio control belong to the graphical user session.

Documentation should use direct, neutral language. Avoid marketing claims, superlatives, and unsupported descriptions of quality or completeness.

## Extension points

- Add capture tools in `internal/capture`, keeping output in memory.
- Add sound servers by implementing `audio.Controller`.
- More efficient window-region capture may be added behind a capturer interface; preserve full-screen fallback.
- If persistence is introduced, it must never persist mute ownership across process or boot boundaries.
