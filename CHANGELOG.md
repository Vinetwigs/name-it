# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2026-10-10

### Added

- Version number is now shown in the application header.
- Automatic update check at startup. When a newer release is published, a
  themed modal tells the user which version is available and links to it.
- `internal/version` package as the single source of truth for the version
  string.
- `internal/updater` package for querying the GitHub releases feed and
  comparing semantic versions.
- GitHub Actions workflow that builds, strips and UPX-compresses releases for
  Windows, Linux and macOS on every version tag, plus a continuous integration
  workflow.
- `build.ps1` for producing a compressed, stripped build locally.

### Fixed

- The filename scorer now splits words from trailing digits, so names such as
  `finale2` are correctly flagged as a suspicious token with a numeric suffix.
- Escalation detection now works: replays of the same file name
  (`poster.psd`, `poster2.psd`, `poster3.psd`) are recognized and punished.
- `trailingNumber` no longer always returns zero.

### Changed

- Internal configuration access is now guarded by a mutex to avoid data races
  between the UI thread and the filesystem watcher.

### Removed

- Unused `internal/analyzer` package.
- Prebuilt archives were removed from the repository. Binary releases are
  distributed through GitHub Releases.

## [2.0.0] - 2026-09-21

### Added

- Naming pattern alerts.
- `ALERT EVERY FILE` mode.
- `filename_vX.extension` naming pattern.
- `YYYY-MM-DD_filename_vX.extension` naming pattern.
- `CUSTOM PATTERN` mode supporting the `filename`, `extension`, `X`, `YYYY`,
  `MM` and `DD` placeholders.
- Pattern examples and a custom pattern mini guide in the UI.
- Filtering so files that already match the selected pattern do not trigger an
  alert.

### Changed

- Naming pattern filtering is handled separately from filename scoring.
- Name It can now react only when a filename does not follow the selected
  naming convention.
- The default custom pattern is now `filename_vX.extension`.

## [1.0.0] - 2026-09-21

### Added

- Initial release.
- Watches a chosen folder and reacts to newly created files.
- Deterministic filename scoring with severity levels (`ignore`, `gentle`,
  `normal`, `passive`, `aggressive`, `maximum`).
- Full-window interruption overlay and native OS notification.
- Pause, aggression and session controls.

[Unreleased]: https://github.com/Vinetwigs/name-it/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/Vinetwigs/name-it/compare/v2.0...v2.1.0
[2.0.0]: https://github.com/Vinetwigs/name-it/compare/v1.0...v2.0
[1.0.0]: https://github.com/Vinetwigs/name-it/releases/tag/v1.0
