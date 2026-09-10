<!-- vale Vale.Terms = NO -->

# forgejo-mirror-sync

<!-- vale Vale.Terms = YES -->

[![CI](https://github.com/alrayyes/forgejo-mirror-sync/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/alrayyes/forgejo-mirror-sync/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/alrayyes/forgejo-mirror-sync/graph/badge.svg)](https://codecov.io/gh/alrayyes/forgejo-mirror-sync)
[![Go Reference](https://pkg.go.dev/badge/github.com/alrayyes/forgejo-mirror-sync.svg)](https://pkg.go.dev/github.com/alrayyes/forgejo-mirror-sync)
[![release](https://img.shields.io/github/v/release/alrayyes/forgejo-mirror-sync?sort=semver)](https://github.com/alrayyes/forgejo-mirror-sync/releases/latest)
[![Licence: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)

Finds `alrayyes`'s public, non-fork GitHub repos that have no pull mirror
under the `alrayyes` user namespace on git.higherlearning.eu yet, creates
one for each, and fixes any archived-state mismatch on mirrors that already
exist. GitHub is always the source of truth; nothing here ever writes back
to GitHub.

A manual, on-demand tool — run it when you think of it, not on a schedule.
It reuses whatever `gh` and `tea` are already authenticated as on the
machine it runs on; it never handles a credential itself.

## What it does, and doesn't do

- Skips forks — a fork isn't something you created, so mirroring it here
  would just duplicate someone else's repo under your own name.
- Skips a GitHub repo that's already archived and has no mirror yet — an
  archived repo is settled; not worth mirroring for the first time.
- Skips a name that already exists on Forgejo as a non-mirror repo (a
  `scaffold-*` template's GitHub-native sibling, say) — those are
  deliberately independent repos, not something to overwrite.
- Never touches a Forgejo repo that has no matching GitHub repo — it might
  be hand-created, or mirror something that's since gone private or been
  deleted.
- Only ever mutates two things on Forgejo: creating a new pull mirror, and
  flipping an existing mirror's `archived` flag to match GitHub.

## Requirements

- **Go 1.25 or newer.**
- **[bun](https://bun.sh)**, for the tooling that isn't Go — commitlint,
  Prettier, markdownlint, and the [lefthook](https://lefthook.dev) that
  runs the git hooks. There's a `package.json`, but nothing here is
  JavaScript; it exists only so those tools resolve and stay pinned.
- **[golangci-lint](https://golangci-lint.run)**, pinned in
  [CONTRIBUTING.md](CONTRIBUTING.md#getting-set-up).
- **[`gh`](https://cli.github.com)**, already authenticated as the account
  that owns the GitHub repos.
- **[`tea`](https://forgejo.org/docs/next/user/tea/)**, already
  authenticated against git.higherlearning.eu with write access to your
  own namespace there.

## Installation

```sh
go install github.com/alrayyes/forgejo-mirror-sync/cmd/forgejo-mirror-sync@latest
```

Pin a specific release instead of `@latest` for a reproducible install —
`@v0.1.0`, say.

## Usage

```sh
./forgejo-mirror-sync
```

Without a flag it prints everything it's about to do — mirrors to create,
archived-state fixes — and asks for one confirmation before making any of
it happen.

```text
--github-owner string    GitHub account to read public repos from (default "alrayyes")
--forgejo-owner string   Forgejo namespace mirrors live under (default "alrayyes")
--dry-run                Print the plan and exit; never prompts, never writes
--yes, -y                Skip the confirmation prompt (for a later scripted run)
--verbose                Log why each repo was skipped and the exact API calls made
```

`--dry-run` and `--yes` both bypass the prompt; `--dry-run` additionally
guarantees nothing is written, `--yes` still writes.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the toolchain, the hooks, and
how a change gets reviewed and released.

## Licence

[GPL-3.0](LICENSE).
