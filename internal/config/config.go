// Package config holds forgejo-mirror-sync's settings and the defaults
// they fall back to. Loading (flags/env/file precedence) lives in
// cmd/forgejo-mirror-sync, wired through viper — this package only knows
// the shape of a valid config, not where its values came from.
package config

import (
	"errors"
	"fmt"
)

// Default values, used both as viper's own defaults and as what `init`
// writes into a fresh config file.
const (
	DefaultGitHubOwner  = "alrayyes"
	DefaultForgejoOwner = "alrayyes"
)

// Config is forgejo-mirror-sync's full set of settings.
type Config struct {
	GitHubOwner  string
	ForgejoOwner string
}

// ErrMissingOwner is returned, wrapped with which field, when a required
// owner is empty. Config from a file or the environment can lie, so this
// is checked once at startup rather than trusted as far as the first place
// the value's read.
var ErrMissingOwner = errors.New("owner must not be empty")

// Validate reports whether c is usable.
func (c Config) Validate() error {
	if c.GitHubOwner == "" {
		return fmt.Errorf("%w: github_owner", ErrMissingOwner)
	}

	if c.ForgejoOwner == "" {
		return fmt.Errorf("%w: forgejo_owner", ErrMissingOwner)
	}

	return nil
}
