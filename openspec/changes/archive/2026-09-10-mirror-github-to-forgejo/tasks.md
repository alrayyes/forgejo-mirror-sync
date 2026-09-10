## 1. Bootstrap the repo

- [x] 1.1 Create `github.com/alrayyes/forgejo-mirror-sync`, public, GPL-3.0,
      tooling stripped from the `scaffold-go-api` template (API/Docker
      pieces removed, this is a CLI) — verified by the repo existing with
      CI, lefthook, commitlint, OpenSpec, issue/PR templates in place.
- [x] 1.2 File the forge ticket (`#1`) before implementation code, per
      the personal-project ticket-first rule — verified by `#1` predating
      the first `feat:` commit.

## 2. Reconciliation logic

- [x] 2.1 `internal/plan.Compute` decides create/archive-fix/skip/in-sync
      for a GitHub repo list against a Forgejo repo list, with no
      network dependency — verified by `internal/plan/plan_test.go`'s
      table of scenarios (missing mirror, archived-no-mirror,
      non-mirror-same-name, archived mismatch both directions, matching
      mirror, orphaned Forgejo repo, deterministic ordering).
- [x] 2.2 `internal/ghsource.Lister` lists public, non-fork GitHub repos
      via `gh repo list`, and `internal/forgejo.Client` lists/creates/
      patches Forgejo repos via `tea api`, both through the injectable
      `internal/runner.Runner` interface — verified by each package's
      tests using a fake `Runner`, no live network call.

## 3. CLI behaviour

- [x] 3.1 Wire `--dry-run`, `--yes`/`-y`, `--verbose`,
      `--github-owner`, `--forgejo-owner` flags in
      `cmd/forgejo-mirror-sync/main.go` — verified by
      `cmd/forgejo-mirror-sync/main_test.go`'s dry-run/yes/verbose/
      failure-reporting cases.
- [x] 3.2 Batch confirmation prompt by default, skippable with `--yes`,
      bypassed entirely by `--dry-run` — verified by
      `TestRun_DecliningPromptMakesNoWrites`,
      `TestRun_AcceptingPromptWrites`,
      `TestRun_YesSkipsPromptAndWrites`.
- [x] 3.3 Fail closed instead of hanging when stdin isn't a terminal and
      no `--yes` is given (`golang.org/x/term.IsTerminal`, gating
      `Options.Interactive`) — verified by
      `TestRun_NonInteractiveWithNoYesFailsClosed`. Landed slightly after
      the initial implementation, in the repo's own follow-up
      rules-compliance pass, once `rules/cli.md`'s TTY-check requirement
      was checked against.
- [x] 3.4 No credential or full response body in any output, including
      `--verbose` — verified by inspection: `internal/runner` never
      handles a token (it shells out), and verbose logging only ever
      prints repo name/owner/archived-value, never a response body.

## 4. Land it

- [x] 4.1 Open the pull request linked to `#1` (`Closes #1`), merged once
      CI was green — verified by `#1` being closed by the merge and the
      PR history on `main`.
