# configuration Specification

## Purpose

Layers this tool's settings across flags, environment variables, a config
file, and built-in defaults, and offers to create that config file the
first time none of the above is set.

## Requirements

### Requirement: Flags override environment variables

The system SHALL prefer a value given on the command line over the same
setting's environment variable, when both are set.

#### Scenario: Flag and environment variable both set

- **WHEN** `--github-owner` and `FORGEJO_MIRROR_SYNC_GITHUB_OWNER` are
  both set to different values
- **THEN** the flag's value is used

### Requirement: Environment variables override the config file

The system SHALL prefer a setting's environment variable over the same
setting's value in the config file, when both are set and no flag
overrides either.

#### Scenario: Environment variable and config file both set

- **WHEN** `FORGEJO_MIRROR_SYNC_GITHUB_OWNER` is set and the config file
  sets a different `github_owner`, with no `--github-owner` flag given
- **THEN** the environment variable's value is used

### Requirement: The config file overrides built-in defaults

The system SHALL prefer a setting's value in the config file over its
built-in default, when the file sets it and no flag or environment
variable overrides it.

#### Scenario: config file sets a value, nothing else does

- **WHEN** the config file sets `github_owner` and no flag or
  environment variable for it is given
- **THEN** the config file's value is used

### Requirement: Configuration is validated once at startup

The system SHALL reject an invalid fully layered configuration (an empty
required owner) with a clear error before doing any work, rather than
failing partway through or on whichever line first reads the bad value.

#### Scenario: Empty owner after layering

- **WHEN** the fully layered `github_owner` or `forgejo_owner` is empty
- **THEN** the system reports an error naming which setting is invalid
  and does not attempt to list or mutate any repo

### Requirement: `init` writes a starter config file

The system SHALL, given the `init` command, write a config file
populated with the tool's built-in defaults to the XDG config
directory, and SHALL NOT overwrite a config file that already exists
there.

#### Scenario: No config file yet

- **WHEN** `init` runs and no config file exists at the XDG config path
- **THEN** a new config file is written there, populated with the
  built-in defaults

#### Scenario: A config file already exists

- **WHEN** `init` runs and a config file already exists at the XDG
  config path
- **THEN** the system makes no change to that file and reports that one
  already exists

### Requirement: A first run with nothing configured offers to write a config file

The system SHALL, when no config file exists and no relevant environment
variable is set, offer to create one — interactively when possible,
automatically under `--yes`, and neither when there is no terminal to
ask on — rather than silently running on defaults with no path to
persisting a preference. Once a config file exists, the system SHALL NOT
make this offer again.

#### Scenario: Interactive first run, offer accepted

- **WHEN** no config file exists, no relevant environment variable is
  set, the run is interactive, and `--yes` is not given
- **THEN** the system asks whether to write a config file, and writes
  one if the answer is yes

#### Scenario: Interactive first run, offer declined

- **WHEN** the same conditions hold and the answer is no
- **THEN** the system writes no config file and continues the run on
  built-in defaults

#### Scenario: Non-interactive first run

- **WHEN** no config file exists, no relevant environment variable is
  set, and the run is not interactive (no terminal, `--yes` not given)
- **THEN** the system does not prompt, notes on its output that `init`
  is available, and continues the run on built-in defaults

#### Scenario: --yes on a first run

- **WHEN** no config file exists, no relevant environment variable is
  set, and `--yes` is given
- **THEN** the system writes a config file populated with the built-in
  defaults without prompting

#### Scenario: A config file already exists at first run

- **WHEN** a config file already exists
- **THEN** the system makes no first-run offer, regardless of the other
  conditions above
