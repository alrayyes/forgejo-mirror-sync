#!/usr/bin/env bash
# Proves the image still builds — hadolint reads the Dockerfile as text
# and a syntactically fine one can still fail to build. GOOS/GOARCH is
# pinned to what the Dockerfile's COPY expects; goreleaser picks the
# matching per-platform build at release time, but a local check needs
# one binary to copy in itself.
set -euo pipefail

cd "$(dirname "$0")/.."

tag="${1:?usage: docker-build-check.sh <tag>}"

GOOS=linux GOARCH=amd64 go build -o forgejo-mirror-sync ./cmd/forgejo-mirror-sync
trap 'rm -f forgejo-mirror-sync; docker rmi -f "forgejo-mirror-sync:${tag}" >/dev/null 2>&1 || true' EXIT

docker build -t "forgejo-mirror-sync:${tag}" .
