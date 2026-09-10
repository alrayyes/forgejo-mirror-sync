# mirror-sync Specification

## Purpose

Reconciles a GitHub account's public, non-fork repos against their pull
mirrors on a Forgejo instance — creating a mirror where one is missing and
keeping an existing mirror's archived state in sync with GitHub, which is
always the source of truth.

## Requirements

### Requirement: Create a mirror for an unmirrored, non-archived GitHub repo

The system SHALL create a Forgejo pull mirror for any public, non-fork
GitHub repo that has no same-named repo under the target Forgejo
namespace yet, unless that GitHub repo is archived.

#### Scenario: New public repo with no existing mirror

- **WHEN** a public, non-fork GitHub repo has no same-named repo under
  the target Forgejo namespace and is not archived
- **THEN** the system creates a new Forgejo pull mirror for it

#### Scenario: Fork is never mirrored

- **WHEN** a GitHub repo is a fork
- **THEN** the system does not create a mirror for it, regardless of its
  archived or mirrored state

### Requirement: Skip an archived GitHub repo with no existing mirror

The system SHALL NOT create a mirror for a GitHub repo that is archived
and has no existing Forgejo mirror.

#### Scenario: Archived repo, never mirrored

- **WHEN** a GitHub repo is archived and has no same-named repo under the
  target Forgejo namespace
- **THEN** the system creates no mirror for it and reports it as skipped

### Requirement: Never overwrite a Forgejo repo that isn't a mirror

The system SHALL NOT create or modify a Forgejo repo that shares a name
with a GitHub repo but is not itself a pull mirror.

#### Scenario: Same-named deliberate native repo

- **WHEN** a Forgejo repo exists under the target namespace with the same
  name as a GitHub repo, but that Forgejo repo is not a pull mirror
- **THEN** the system makes no change to that Forgejo repo and reports it
  as skipped, whatever the GitHub repo's archived state

### Requirement: Keep an existing mirror's archived state in sync with GitHub

The system SHALL update an existing Forgejo mirror's archived flag to
match its GitHub source whenever the two disagree, and SHALL NOT modify a
mirror whose archived flag already matches.

#### Scenario: GitHub repo archived after its mirror was created

- **WHEN** an existing Forgejo mirror's archived flag is false and its
  GitHub source is archived
- **THEN** the system sets the Forgejo mirror's archived flag to true

#### Scenario: GitHub repo unarchived after its mirror was created

- **WHEN** an existing Forgejo mirror's archived flag is true and its
  GitHub source is not archived
- **THEN** the system sets the Forgejo mirror's archived flag to false

#### Scenario: Mirror already matches

- **WHEN** an existing Forgejo mirror's archived flag already matches its
  GitHub source
- **THEN** the system makes no write for that repo and reports it as
  already in sync

### Requirement: Never write back to GitHub

The system SHALL treat GitHub as read-only and SHALL NOT create, modify,
or delete anything on GitHub under any flag combination.

#### Scenario: Any run of the tool

- **WHEN** the system runs in any mode, including a run that creates
  mirrors or fixes archived flags
- **THEN** no GitHub repo, setting, or resource is created, modified, or
  deleted

### Requirement: Ignore a Forgejo repo with no matching GitHub repo

The system SHALL NOT modify or report on a Forgejo repo under the target
namespace that has no same-named GitHub repo.

#### Scenario: Forgejo-only repo

- **WHEN** a Forgejo repo exists under the target namespace with no
  same-named repo on GitHub
- **THEN** the system leaves it untouched and does not include it in its
  summary

### Requirement: Preview mode makes no writes

The system SHALL, when run with `--dry-run`, print the same plan it would
otherwise act on and make no Forgejo API writes, without prompting for
confirmation.

#### Scenario: Dry run with pending changes

- **WHEN** the system runs with `--dry-run` and the reconciliation finds
  at least one mirror to create or one archived flag to fix
- **THEN** it prints the planned creates and fixes, makes no write, and
  does not prompt

### Requirement: Confirm once before writing, unless told not to

The system SHALL, by default, print a single summary of all pending
mirror creations and archived-flag fixes and ask for one confirmation
before making any of them, unless `--yes`/`-y` or `--dry-run` is given.

#### Scenario: Interactive run with pending changes, no flag given

- **WHEN** the system runs interactively with at least one pending change
  and neither `--yes` nor `--dry-run` is given
- **THEN** it prints the pending changes and asks for one confirmation
  covering all of them before writing anything

#### Scenario: Confirmation declined

- **WHEN** the user declines the confirmation prompt
- **THEN** the system makes no writes and exits without error

#### Scenario: --yes skips the prompt

- **WHEN** the system runs with `--yes` (or `-y`) and at least one
  pending change
- **THEN** it makes the pending writes without prompting

### Requirement: Fail closed with no terminal to confirm on

The system SHALL refuse to block waiting for confirmation input when its
standard input is not an interactive terminal, unless `--yes` or
`--dry-run` is given.

#### Scenario: Piped or scripted invocation with no --yes

- **WHEN** the system runs with standard input that is not a terminal,
  at least one pending change, and neither `--yes` nor `--dry-run` given
- **THEN** it exits with an error and makes no writes, rather than
  waiting to read a confirmation that will never arrive

### Requirement: Verbose mode explains every decision without leaking credentials

The system SHALL, when run with `--verbose`, report why each GitHub repo
was skipped and the exact command it ran for each write, and SHALL NOT
include a credential, token, or full API response body in that output
under any flag combination.

#### Scenario: Verbose skip reporting

- **WHEN** the system runs with `--verbose` and a GitHub repo is skipped
- **THEN** the output includes that repo's name and the reason it was
  skipped

#### Scenario: Verbose write reporting

- **WHEN** the system runs with `--verbose` and creates a mirror or fixes
  an archived flag
- **THEN** the output includes the command it ran, identified by repo
  name/owner and the archived value being set, and never a token or a
  full response body

### Requirement: Nothing here handles a GitHub or Forgejo credential directly

The system SHALL read from GitHub and write to Forgejo only through
already-authenticated external command-line tools, and SHALL NOT read,
store, or transmit a credential itself.

#### Scenario: Any GitHub or Forgejo operation

- **WHEN** the system lists GitHub repos, lists Forgejo repos, creates a
  mirror, or sets an archived flag
- **THEN** it does so by invoking an external, already-authenticated
  command-line tool rather than making a direct, credentialed API call
  itself
