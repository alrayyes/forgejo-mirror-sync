## Context

`skills/repo-creation`'s "Mirror every `github.com` repo onto
`git.higherlearning.eu`" section already documents the intended shape for a
single repo, run by hand at bootstrap time: `tea api -X POST
/repos/migrate`, skip forks, skip an already-archived repo. This tool
automates that same shape across every existing public repo at once, plus
the ongoing archived-state check nothing was doing before.

The `personal-github` Forgejo org looked like a plausible mirror
destination at first glance (it already held a couple of GitHub-sourced
repos), but `tea api /orgs/personal-github/repos` showed every entry there
as `mirror: false, archived: true` — a static historical import, not a
live pull mirror. The `alrayyes` user namespace was the real, actively
maintained destination (`tea api /users/alrayyes/repos` showed the
expected `mirror: true` on the repos already mirrored by hand).

## Goals / Non-Goals

**Goals:**

- Find every public, non-fork GitHub repo with no Forgejo mirror yet and
  create one, without ever touching a repo that's deliberately not a
  mirror.
- Keep an existing mirror's `archived` flag honest against GitHub without
  anyone having to remember to check by hand.
- Never handle a GitHub or Forgejo credential directly.

**Non-Goals:**

- Scheduling or unattended operation — this is a manual, on-demand tool;
  automating it is explicitly out of scope for this change (tracked
  separately if it's ever wanted).
- A config file, environment-variable layer, or `init` command — flags
  only for now (tracked as its own follow-up, `#7`, once it was noticed
  this doesn't yet match `rules/cli.md`'s full configuration-layering
  requirement).
- Mirroring anything other than `alrayyes`'s own public GitHub repos —
  the owner is a flag with that default, not hardcoded, but no other
  target was exercised.

## Decisions

- **Shell out to `gh`/`tea` rather than call the GitHub/Forgejo APIs
  directly.** Both CLIs are already authenticated on any machine this
  runs on; a direct HTTP client would need its own token handling, which
  is exactly the kind of credential surface `internal/runner`'s package
  doc says this tool never wants. The cost is one small seam,
  `internal/runner.Runner` (real `exec.Command` in production, a
  scripted fake in tests) — cheap, and it's also what makes
  `internal/ghsource` and
  `internal/forgejo` testable without a network call.
- **`internal/plan` is pure — no network, no `Runner` — and owns the
  entire reconcile-or-skip decision.** `internal/ghsource` and
  `internal/forgejo` only fetch and mutate; `cmd/forgejo-mirror-sync`
  converts their types into `plan`'s and prints the result. This is what
  let the decision logic (five distinct outcomes: create, fix archived,
  skip-archived-no-mirror, skip-non-mirror, already-in-sync) get a full
  table of unit tests with no fake network setup at all.
- **Confirm once, by default, rather than per-action.** A single batched
  prompt (`Create N mirror(s) and fix M archived flag(s)?`) reads better
  for a CLI than confirming each of what could be 30+ individual writes,
  and it's still skippable with `--yes` for the day this runs
  unattended.
- **Fail closed with no TTY, rather than blocking on a read that will
  never arrive.** `golang.org/x/term.IsTerminal` gates the prompt; a
  piped or scripted invocation with no `--yes` gets a clear error instead
  of hanging (`rules/cli.md`'s TTY-check requirement, fixed shortly after
  the initial implementation shipped, in the same repo's follow-up
  rules-compliance pass).
- **No PII in any output, including `--verbose`.** Verbose mode logs the
  `gh`/`tea` command and the decision behind it, never a response body —
  a Forgejo repo object or an error message is not somewhere to assume a
  credential can't leak from.

## Risks / Trade-offs

- Shelling out to `gh`/`tea` means this tool's behaviour depends on those
  binaries' own CLI surface staying stable across versions — a flag
  rename in either would break parsing rather than failing a compile.
  Accepted: both are mature, already-pinned-elsewhere tools, and the
  alternative (a direct API client) trades that risk for a
  credential-handling one, which is worse for this tool's purpose.
- No automated schedule means drift can build back up between manual
  runs. Accepted as the deliberate Non-Goal above; revisit if it turns
  out to need running more often than someone remembers to.
