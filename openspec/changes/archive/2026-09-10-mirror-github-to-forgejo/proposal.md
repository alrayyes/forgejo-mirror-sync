## Why

`alrayyes`'s public, non-fork GitHub repos were being mirrored onto
git.higherlearning.eu by hand — `tea api -X POST /repos/migrate`, one repo
at a time, whenever someone remembered. Nothing checked which repos already
had a mirror, nothing caught a repo whose GitHub archived state had since
diverged from its Forgejo copy, and the account had accumulated over 100
public repos by the time this was written, most already mirrored, a few
not. A script was needed to find the gap and close it, and to keep an
existing mirror's archived flag honest without anyone having to remember
to check.

Written up retroactively (dotfiles CLAUDE.md's "the ticket comes before
the work" rule applies to the forge ticket — filed as `#1` before this
tool's implementation code was written — but this OpenSpec change itself
was missed at the time and is being backfilled now that the gap was
noticed).

## What Changes

- List `alrayyes`'s public, non-fork GitHub repos and the existing repos
  under the `alrayyes` namespace on git.higherlearning.eu, by shelling
  out to `gh` and `tea` — never a direct HTTP call, so neither credential
  is ever something this tool itself handles.
- Reconcile the two lists: create a Forgejo pull mirror for any GitHub
  repo missing one, unless the GitHub repo is already archived and never
  had a mirror (settled, not worth mirroring for the first time) or a
  same-named Forgejo repo exists that isn't itself a mirror (a
  scaffold's GitHub-native sibling, say — never overwritten).
- Fix an existing mirror's `archived` flag to match GitHub whenever the
  two disagree, since a repo can be archived well after its mirror was
  created.
- `--dry-run` prints the plan and makes no writes at all, without
  prompting. `--verbose` additionally logs why each repo was skipped and
  the exact `gh`/`tea` commands the tool is about to run. `--yes`/`-y`
  skips the interactive confirmation for a later scripted run. With none
  of those flags, the tool prints the plan and asks for one confirmation
  before writing anything — and refuses to prompt at all (fails closed,
  doesn't hang) when stdin isn't a real terminal.

## Capabilities

### New Capabilities

- `mirror-sync`: reconciling a GitHub account's public repos against
  their Forgejo pull mirrors — what gets mirrored, what gets skipped, and
  how an existing mirror's archived state stays in sync.

## Impact

- New repo: `github.com/alrayyes/forgejo-mirror-sync`, a standalone Go
  CLI, no impact on any other repo.
- Read-only against GitHub; the only Forgejo writes are creating a new
  mirror and flipping an existing mirror's `archived` flag — nothing
  else on either forge is touched.
