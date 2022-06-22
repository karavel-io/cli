# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2022-06-22

### Changed

- Init now uses the GitHub tags API to fetch the latest version and generates an empty `karavel.hcl`, instead of trying to download an example file from the release.
- The version/git commit fields are now filled even when building directly from source (eg. `go install github.com/karavel-io/cli@latest`)
- To avoid stale caches, each run of karavel will now make its own distinct temporary folders.
- Version tags are now prefixed with a `v` to comply with Go's expected versioning scheme for modules (so 0.3.0 is `v0.3.0`)
- 
### Fixed

- Karavel now uses the proper TEMP path on Windows instead of `/tmp`

## [0.2.1] - 2022-03-08

### Fixed

- Add new bootstrap apps to kustomization list

## [0.2.0] - 2022-03-08

### Added

 - Allow overriding component integration flags via the HCL file
 - Generate bootstrap and projects applications so that Argo can automatically deploy everything from the start

## [0.1.1] - 2022-02-01

### Changed

- `karavel render` now renders UNIX-style paths regardless of host operating system ([#4](https://github.com/karavel-io/cli/pull/4))

### Fixed

- Multiple data races and crashes have been fixed when rendering charts ([#5](https://github.com/karavel-io/cli/pull/5#pullrequestreview-868973482))

## [0.1.0] - 2021-10-25

- `karavel init` first implementation
- `karavel render` first implementation

[unreleased]: https://github.com/karavel-io/cli/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/karavel-io/cli/releases/compare/0.2.1...v0.3.0
[0.2.1]: https://github.com/karavel-io/cli/releases/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/karavel-io/cli/releases/compare/0.1.1...0.2.0
[0.1.1]: https://github.com/karavel-io/cli/releases/compare/0.1.0...0.1.1
[0.1.0]: https://github.com/karavel-io/cli/releases/tag/0.1.0
