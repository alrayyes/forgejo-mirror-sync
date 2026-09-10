FROM debian:bookworm-slim@sha256:88200866dfff7ea7f5cbcb6ec7c8a701889efe6fe859fe64d6990e4b07ea4171

# gh and tea, pinned by version and checksum. This tool shells out to both
# — see internal/runner's package doc — so the image needs them installed
# the same way a host would, not just a Go runtime to run the binary below.
# go-releases.md's "COPY, don't recompile" is about the Go binary this repo
# builds; it doesn't extend to dependencies that were never Go source here
# to begin with. amd64 only for now — arm64 is a follow-up, tracked in the
# repo's own issue tracker.
ARG GH_VERSION=2.100.0
ARG GH_SHA256=e4d4bb4498e8d007abe545b6568926793ace1b6447da598294a610018cb164be
ARG TEA_VERSION=0.16.0
ARG TEA_SHA256=92e0c966c98be0c6ca4c80b1912d08ff7887c7cb55e123542fa541842d875149

# bash, not the default dash, so `-o pipefail` actually exists — the
# sha256sum -c below is piped from echo, and dash would silently ignore a
# checksum-verification failure upstream of the pipe.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update \
    && apt-get install --no-install-recommends -y ca-certificates curl xz-utils \
    && curl -fsSL -o /tmp/gh.tar.gz "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz" \
    && echo "${GH_SHA256}  /tmp/gh.tar.gz" | sha256sum -c - \
    && tar -xzf /tmp/gh.tar.gz -C /tmp \
    && install -m 0755 "/tmp/gh_${GH_VERSION}_linux_amd64/bin/gh" /usr/local/bin/gh \
    && curl -fsSL -o /tmp/tea.xz "https://gitea.com/gitea/tea/releases/download/v${TEA_VERSION}/tea-${TEA_VERSION}-linux-amd64.xz" \
    && echo "${TEA_SHA256}  /tmp/tea.xz" | sha256sum -c - \
    && xz -d /tmp/tea.xz \
    && install -m 0755 /tmp/tea /usr/local/bin/tea \
    && apt-get purge -y --auto-remove curl xz-utils \
    && rm -rf /tmp/* /var/lib/apt/lists/*

# The binary goreleaser already cross-compiled — see go-releases.md's
# Dockerfile section for why this doesn't run `go build` itself.
COPY forgejo-mirror-sync /usr/local/bin/forgejo-mirror-sync

RUN useradd --system --create-home --home-dir /home/mirror-sync mirror-sync
# Go's os.UserHomeDir() (which gh and tea both use to find their config)
# only ever reads $HOME — it never falls back to the passwd entry the way
# a shell would, so this has to be set explicitly for a non-root, non-login
# USER.
ENV HOME=/home/mirror-sync
USER mirror-sync

ENTRYPOINT ["/usr/local/bin/forgejo-mirror-sync"]
