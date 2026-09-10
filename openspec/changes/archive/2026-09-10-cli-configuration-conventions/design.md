## Context

`rules/cli.md` sets the cross-language baseline every CLI here follows:
flags over environment variables over a config file over defaults, a
`<tool> init` that writes the file, a first-run offer when none exists,
and a Docker image alongside the native binary. `rules/go.md`'s Command
line/Configuration sections name the Go-specific pairing —
`spf13/cobra` for the command surface, `spf13/viper` for the layered
config — as the reason those two libraries were picked over the standard
library.

## Goals / Non-Goals

**Goals:**

- Make configuration behave the same way every other CLI on this account
  does — no bespoke precedence rules to relearn per tool.
- Give the tool a starter-config path (`init`) instead of leaving someone
  to hand-author the file format from documentation alone.
- Let the tool run somewhere without a Go toolchain installed, the same
  way installing it natively would.

**Non-Goals:**

- `arm64` Docker support — the image bundles `linux/amd64`-only `gh`/
  `tea` binaries; a real gap, tracked rather than solved here.
- Secrets in the config file — none of this tool's settings (GitHub/
  Forgejo owner names) are credentials, so `rules/cli.md`'s
  `<field>_command` pattern for secret fields doesn't apply; nothing here
  invents a field just to exercise it.
- Multi-platform Docker manifests (`docker_manifests:` / buildx) — moot
  while there's only one platform to publish.

## Decisions

- **`internal/config` stays free of cobra/viper.** It only knows a
  config's shape, its defaults, and whether one is valid
  (`Config.Validate`) — the actual flag/env/file layering lives in
  `cmd/forgejo-mirror-sync/main.go`, wired through a `*viper.Viper`. That
  split is what let the precedence rules (flag beats env beats file beats
  default) get tested directly against a hand-built `viper.Viper`, with
  no cobra command, no XDG filesystem, and no real environment variable
  beyond what `t.Setenv` sets for one test.
- **`LoadOptions` and `MaybeOfferInit` are exported,** matching this
  file's existing `Run`/`Options` precedent, rather than adding an
  internal-only test file — `rules/go-test.md`'s "when the only way in is
  an unexported function, the test is aimed at the wrong thing" pushed
  toward exporting a real, reusable unit instead.
- **The config file is YAML**, matching every other config-shaped file in
  this repo (`lefthook.yml`, the GitHub Actions workflows) rather than
  introducing TOML or JSON as a one-off.
- **The Docker image bundles `gh` and `tea` directly, pinned by version
  and checksum, rather than assuming the host mounts them in.** This
  tool's whole purpose is shelling out to both — a container with just
  the Go binary and no `gh`/`tea` on `$PATH` would fail on its first real
  command. `go-releases.md`'s "the Dockerfile's job is to `COPY` the
  binary, not recompile" still holds for the one binary this repo
  actually builds; it says nothing about `gh`/`tea`, which were never Go
  source in this repository to begin with.
- **`debian:bookworm-slim`, not a distroless or `scratch` base.**
  Installing `gh`/`tea` needs a shell and `curl`/`tar`/`xz` at build
  time, and glibc at run time for GitHub's own `gh` release tarball —
  `containers.md`'s "smallest base that does the job" still points at
  Debian once "the job" includes installing two more binaries, not just
  running one static Go binary.
- **`$HOME` is set explicitly** (`ENV HOME=/home/mirror-sync`) rather
  than left to the `mirror-sync` user's passwd entry. Go's
  `os.UserHomeDir()` — what `gh` and `tea` both use to find their own
  config — only ever reads `$HOME`; it never falls back to a passwd
  lookup the way a login shell would. Confirmed live: without this, a
  volume-mounted `~/.config/gh` never resolves inside the container.

## Risks / Trade-offs

- Bundling `gh`/`tea` binaries in the image means this repo now tracks
  two more pinned versions and checksums, updated by hand — Dependabot
  has no ecosystem for "binary fetched by a Dockerfile `RUN curl`" the
  way it does for `go.mod`. Accepted: the alternative (assume the host
  mounts them in) makes the image nearly useless on its own, which
  defeats the point of shipping one.
- `linux/amd64`-only is a real capability gap, not a cosmetic one —
  `arm64` hardware can't run the image at all yet. Documented in the
  README, the Dockerfile, and `.goreleaser.yml`'s own comment rather than
  silently shipping a broken multi-arch claim.
