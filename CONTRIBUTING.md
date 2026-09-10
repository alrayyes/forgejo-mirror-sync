# Contributing

## Getting set up

- **Go 1.25 or newer**, for writing and building the code itself — the
  hooks below don't need it on your `PATH`.
- **Docker.** `gofmt`, `golangci-lint`, `go mod edit -fmt`,
  `go mod tidy -diff` and `go test` all run inside pinned images
  (`scripts/go-docker.sh`, `scripts/golangci-lint-docker.sh`) rather than
  whatever your host happens to have — a host `go`/`golangci-lint` pair
  can drift out of sync with each other, or with what `go.mod` declares,
  independently of anything in this repo. `rules/go-mod.md`,
  `rules/go-lint.md`. Also what `scripts/docker-build-check.sh` and
  hadolint both need, for the shipped image itself (see the README's
  Docker section).
- **[bun](https://bun.sh)** for the tooling that isn't Go — commitlint,
  Prettier, markdownlint, and the [lefthook](https://lefthook.dev) that runs
  the git hooks. There's a `package.json`, but nothing here is JavaScript;
  it exists only so those tools resolve and stay pinned.
- **[Vale](https://vale.sh)** on your `PATH`, for the style tier of the
  prose lint:

  ```sh
  go install github.com/errata-ai/vale/v3/cmd/vale@latest
  ```

  `ltex-cli-plus` needs nothing installed: the hook fetches and caches it
  on first use.

- **[`govulncheck`](https://go.dev/doc/tutorial/govulncheck)**:
  `go install golang.org/x/vuln/cmd/govulncheck@latest`.
- **[`goreleaser`](https://goreleaser.com)**, for `goreleaser check` — only
  validates `.goreleaser.yml`, never runs a real release locally.

One command installs the linters and the git hooks:

```sh
bun install
```

An uninstalled hook silently does nothing, which is worse than not having
one, so the `prepare` script runs `lefthook install` for you. You find out
at the pipeline otherwise, not at the commit.

## Everyday commands

Every one of these is what a hook or CI runs — see `lefthook.yml` and
`.github/workflows/*.yml` for exactly which.

```sh
go build ./...              # host go — nothing to pin, doesn't need a matched version
go vet ./...

./scripts/go-docker.sh go test ./...
./scripts/go-docker.sh go test -race -coverprofile=coverage.out -coverpkg=./... ./... && go tool cover -func=coverage.out
./scripts/go-docker.sh go tool gotestsum --junitfile junit.xml -- -race -coverprofile=coverage.out -coverpkg=./... ./...  # what CI runs, for the JUnit report Codecov Test Analytics reads
./scripts/golangci-lint-docker.sh golangci-lint run ./...
./scripts/golangci-lint-docker.sh golangci-lint fmt ./...   # the fixer; `run` stays the check
./scripts/go-docker.sh go mod tidy -diff   # broader than `go mod edit -fmt`: catches a stale require too
govulncheck ./...
goreleaser check
./scripts/docker-build-check.sh local-check   # proves the shipped image still builds
docker run --rm -i -v "$(pwd)/.hadolint.yaml:/.hadolint.yaml" \
  hadolint/hadolint@sha256:32dac94127fd60b7b7e3fbfc65e1383b9b5e25c9bfd7b8536de7a539fe68a12d < Dockerfile

bun run format:check       # prettier --check, add --write to fix
bun run lint:md
bun run lint:prose         # vale
bun run lint:mechanics     # ltex-cli-plus
```

## How it fits together

- `internal/plan` decides what to do — pure functions, no network, the
  package most worth reading first.
- `internal/ghsource` lists public GitHub repos, and `internal/forgejo`
  lists and mutates Forgejo repos. Both shell out to `gh`/`tea` through the
  `internal/runner.Runner` interface rather than talking HTTP directly, so
  neither package ever handles a token — the CLI already authenticated on
  the host does, and that's also what makes both packages testable with a
  fake `Runner` instead of a live network call.
- `internal/confirm` is the interactive yes/no prompt, with the reader and
  writer injectable for tests.
- `internal/config` holds the settings' shape, defaults and validation,
  and `init`'s starter-file writer — no cobra, no viper, nothing about
  where a value came from. That layering (flags > env > file > defaults)
  lives in `cmd/forgejo-mirror-sync/main.go` itself, wired through viper.
- `cmd/forgejo-mirror-sync/main.go` wires every package listed here
  together and owns the cobra command tree.

## No PII in output

Nothing here ever prints a token, a full request or response body, or any
identity beyond a repo name/URL and its archived state — verbose mode logs
the command that ran (`tea api -X POST /repos/migrate …`) and the decision
behind it, never the response body, since a Forgejo repo object or an error
message is not a place to assume a credential can't leak from. Keep new
logging to that same shape.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/):
`type(scope): description`, types `feat`/`fix`/`docs`/`style`/`refactor`/
`perf`/`test`/`build`/`ci`/`chore`/`revert`. Subject under 50 characters,
lowercase, no trailing full stop. commitlint enforces the shape at
commit-msg and again in CI; the length and case rules are tighter than what
it checks, so hold to them anyway.

## Branching, review, and release

Every change goes through a pull request — nothing is pushed straight to
`main`, including the bootstrapping that built this repo. The default
branch is protected (`Settings → Branches`): a pull request and at least
zero required reviews (solo repo) are required before a merge, so this
isn't just discipline.

The pull request **title** has to be a valid Conventional Commit too —
`pr-title.yml` checks it. commitlint only ever reads commit objects, and a
squash merge defaults its commit message to the pull request title, so this
is the only check standing between a badly titled pull request and a bad
message on `main`.

Once a pull request's checks are green, squash-merge it and delete the
branch. [release-please](https://github.com/googleapis/release-please)
reads the Conventional Commits on `main` and keeps a release pull request
open with the next version and changelog entry; merging that one tags the
release, and [goreleaser](https://goreleaser.com) builds the binaries onto
it. Nobody picks a version by hand.
