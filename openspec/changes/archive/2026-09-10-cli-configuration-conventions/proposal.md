## Why

The tool shipped (`#1`/`#2`) with flags only — no environment-variable
layer, no config file, no way to persist a preference between runs. A
rules-compliance audit against `rules/cli.md` (this account's cross-language
CLI conventions) found the gap and filed it as `#7`. This tool also had no
Docker image at all, despite `rules/cli.md` asking every CLI to ship one
alongside its native binary.

Written up retroactively — the OpenSpec change for this work was missed at
the time (`#7`'s PR merged without one, the same gap `mirror-github-to-forgejo`'s
backfill had already found once), and is being backfilled now, matching
what actually shipped rather than re-deriving the design fresh.

## What Changes

- Layer configuration: flags override `FORGEJO_MIRROR_SYNC_*` environment
  variables, which override an XDG config file, which overrides built-in
  defaults.
- Add a `forgejo-mirror-sync init` command that writes a starter config
  file, populated with today's defaults, to
  `$XDG_CONFIG_HOME/forgejo-mirror-sync/config.yaml` — refusing to
  overwrite one that already exists.
- On a run with no config file and no relevant environment variable set,
  offer to run `init` right there (interactively) or write one
  automatically under `--yes`, rather than silently falling back to
  defaults with no path forward. Never ask again once a config file
  exists.
- Ship a Docker image, `ghcr.io/alrayyes/forgejo-mirror-sync`, bundling
  pinned `gh` and `tea` binaries alongside the tool itself —
  `linux/amd64` only for now, a documented gap rather than a silent one.

## Capabilities

### New Capabilities

- `configuration`: the flag/environment/file/default precedence, the
  `init` command, and the first-run offer to create a config file.
- `docker-distribution`: the shipped Docker image, what it bundles, and
  its current platform scope.

## Impact

- `cmd/forgejo-mirror-sync` gains `internal/config` as a dependency and
  moves from the standard library `flag` package to `spf13/cobra` +
  `spf13/viper`; `github.com/adrg/xdg` resolves the config file's path.
- New files: `Dockerfile`, `.dockerignore`, `.hadolint.yaml`,
  `scripts/docker-build-check.sh`; `.goreleaser.yml` gains a `dockers:`
  block; the release workflow gains a GHCR login step.
- No change to the reconciliation behaviour `mirror-sync`'s own spec
  already covers — this is additive, about how the tool is configured
  and distributed, not what it decides to mirror.
