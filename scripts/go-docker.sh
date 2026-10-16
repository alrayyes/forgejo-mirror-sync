#!/usr/bin/env bash
# Runs a Go toolchain command inside the pinned golang image rather than
# the host binary, so the version doing the work is what go.mod's own `go`
# directive says, not whatever the host package manager currently has —
# those two can drift independently and did, once, on this account.
# rules/go-mod.md has the full reasoning and the exact invocation this
# mirrors.
set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p "$HOME/.cache/go-build-docker" "$HOME/.cache/go-mod-docker"

docker run --rm --user "$(id -u):$(id -g)" \
  -v "$(pwd):/src" -w /src \
  -e GOCACHE=/gocache -v "$HOME/.cache/go-build-docker:/gocache" \
  -e GOMODCACHE=/gomod -v "$HOME/.cache/go-mod-docker:/gomod" \
  golang:1.26.9-bookworm@sha256:d9c68c2c51161e12fd77e4c6320687c9cd86e1af1e3ad6e6cd63ff970641453c \
  "$@"
