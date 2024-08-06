# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->

[keep a changelog]: https://keepachangelog.com/en/1.0.0/
[semantic versioning]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

### Added

- Added `/health` endpoint.

### Changed

- Do not log HTTP requests for `/assets` and `/health` endpoints.

### Fixed

- Fixed panic when response from GitHub API is `nil`.

## [0.1.9] - 2024-08-06

### Changed

- Bumped Go runtime to v1.22

### Fixed

- Marked GitHub app environment variables as "sensitive", so they will not be
  printed to `STDOUT` when they are invalid.

## [0.1.8] - 2024-01-19

- Bumped `configkit` to v0.12.2

## [0.1.7] - 2023-09-27

- Build on Go v1.21

## [0.1.6] - 2023-05-05

- Add support for Dogma v0.12.0 `Routes()` configuration

## [0.1.5] - 2023-03-03

- Generate read-only installation tokens that can be used to access private Git
  repositories for the duration of analysis. This makes the `GITHUB_USER_TOKEN`
  environment variable unnecessary.

## [0.1.4] - 2023-03-03

- Fix another misnamed environment variable, all environment variables are now
  named as they were in v0.1.1

## [0.1.3] - 2023-03-03

- Return environment variable naming to match v0.1.1, as the change in v0.1.2
  was not intended.

## [0.1.2] - 2023-03-03

- Use Go v1.20 at runtime (via `golang:latest` docker image)

## [0.1.1] - 2023-03-02

- Build with Go v1.20

## [0.1.0] - 2021-08-30

- Initial release

<!-- references -->

[unreleased]: https://github.com/dogmatiq/browser
[0.1.0]: https://github.com/dogmatiq/browser/releases/v0.1.0
[0.1.1]: https://github.com/dogmatiq/browser/releases/v0.1.1
[0.1.2]: https://github.com/dogmatiq/browser/releases/v0.1.2
[0.1.3]: https://github.com/dogmatiq/browser/releases/v0.1.3
[0.1.4]: https://github.com/dogmatiq/browser/releases/v0.1.4
[0.1.5]: https://github.com/dogmatiq/browser/releases/v0.1.5
[0.1.6]: https://github.com/dogmatiq/browser/releases/v0.1.6
[0.1.7]: https://github.com/dogmatiq/browser/releases/v0.1.7
[0.1.8]: https://github.com/dogmatiq/browser/releases/v0.1.8
[0.1.9]: https://github.com/dogmatiq/browser/releases/v0.1.9

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
