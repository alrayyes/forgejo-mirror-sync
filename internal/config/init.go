package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrConfigExists is returned by WriteDefaultFile when path already exists
// — init never overwrites a config someone has already customized.
var ErrConfigExists = errors.New("config file already exists")

const defaultConfigYAML = `# forgejo-mirror-sync config, written by "forgejo-mirror-sync init".
# Flags and FORGEJO_MIRROR_SYNC_* environment variables override these.

github_owner: ` + DefaultGitHubOwner + `
forgejo_owner: ` + DefaultForgejoOwner + `
`

// WriteDefaultFile writes a starter config, populated with the defaults
// forgejo-mirror-sync would otherwise fall back to, to path — creating any
// missing parent directory. It refuses to touch a file that already
// exists.
func WriteDefaultFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrConfigExists, path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking for an existing config at %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(defaultConfigYAML), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
