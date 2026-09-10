# docker-distribution Specification

## Purpose

Distributes forgejo-mirror-sync as a Docker image that works the same way
running the native binary does, without requiring a Go toolchain on the
host.

## Requirements

### Requirement: The image is runnable without a host Go toolchain

The system SHALL be published as a Docker image that runs the tool's full
command surface (`--help`, `--version`, `init`, the default reconciliation
command) with no Go installation on the host.

#### Scenario: Running the image directly

- **WHEN** the published image is run with `--version`
- **THEN** it prints the tool's version, the same as the native binary
  would

### Requirement: The image bundles the external tools it shells out to

The system SHALL include pinned `gh` and `tea` binaries inside the image,
since the tool cannot do its own work without both.

#### Scenario: `gh` and `tea` are both present in the image

- **WHEN** the image is inspected for `gh` and `tea` on its `$PATH`
- **THEN** both are present and report their pinned versions

### Requirement: The image runs as a non-root user with a resolvable home directory

The system SHALL run as a non-root user whose `$HOME` is set explicitly,
so that a mounted `gh`/`tea` credential directory at the expected path is
actually found.

#### Scenario: Mounted credentials are found

- **WHEN** a host's `gh` config directory is mounted at the image's
  non-root user's home directory
- **THEN** `gh` inside the container reads that mounted configuration

### Requirement: The image's platform scope is explicit

The system SHALL document which platforms the image supports, and
SHALL NOT claim support for a platform it doesn't actually build for.

#### Scenario: Current scope

- **WHEN** the published image tag is inspected
- **THEN** it is `linux/amd64` only, and the README, Dockerfile and
  release configuration all say so
