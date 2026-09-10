## 1. Configuration layering

- [x] 1.1 `internal/config.Config` holds the settings' shape, defaults,
      and `Validate()` — verified by `internal/config/config_test.go`'s
      precedence-agnostic validation cases.
- [x] 1.2 `internal/config.WriteDefaultFile` writes a starter config,
      refusing to overwrite an existing one — verified by
      `internal/config/init_test.go`.
- [x] 1.3 `cmd/forgejo-mirror-sync/main.go` wires spf13/cobra (command
      surface) and spf13/viper (flag/env/file/default layering) —
      verified by `cmd/forgejo-mirror-sync/config_test.go`'s
      `TestLoadOptions_*` precedence table (defaults, config file, env,
      flag, each overriding the one before it) built against a real
      `*viper.Viper`, not a stand-in for one.
- [x] 1.4 `init` subcommand and the first-run offer
      (`MaybeOfferInit`) — verified by
      `cmd/forgejo-mirror-sync/config_test.go`'s
      `TestMaybeOfferInit_*` cases (config exists, env set, `--yes`,
      non-interactive, interactive accept/decline).
- [x] 1.5 Manual smoke test: built the binary, ran `--help`,
      `--version`, `init` against a temporary `XDG_CONFIG_HOME`, and
      confirmed the written file's defaults matched what the tool
      actually falls back to.

## 2. Docker image

- [x] 2.1 `Dockerfile` installs pinned, checksum-verified `gh` and `tea`
      binaries on `debian:bookworm-slim`, then `COPY`s the
      goreleaser-built binary in — verified by `docker build .` and
      running `gh --version`/`tea --version` inside the built image.
- [x] 2.2 Non-root user with `$HOME` set explicitly — verified live:
      mounting a fake `gh` config at
      `/home/mirror-sync/.config/gh` and reading it back from inside
      the container.
- [x] 2.3 `.goreleaser.yml`'s `dockers:` block, scoped to `linux/amd64`
      — verified by `goreleaser release --snapshot --clean` building
      and running the resulting image end to end.
- [x] 2.4 hadolint (`.hadolint.yaml`) and a real `docker build`
      (`scripts/docker-build-check.sh`) wired into both the git hooks
      and CI — verified by `lefthook run pre-commit`/`pre-push` and the
      CI `dockerfile`/`docker` jobs, all green.
- [x] 2.5 Release workflow logs in to `ghcr.io` before goreleaser runs,
      with `packages: write` permission — verified by the merged
      release run actually pushing
      `ghcr.io/alrayyes/forgejo-mirror-sync`.

## 3. Land it

- [x] 3.1 Open the pull request linked to `#7` (`Closes #7`), merged
      once CI was green — verified by `#7` being closed by the merge
      and the PR history on `main`.
