package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteDefaultFile_WritesUsableDefaults(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "config.yaml")

	err := config.WriteDefaultFile(path)

	require.NoError(t, err)
	contents, err := os.ReadFile(path) //nolint:gosec // path is t.TempDir()-derived, not external input
	require.NoError(t, err)
	assert.Contains(t, string(contents), "github_owner: "+config.DefaultGitHubOwner)
	assert.Contains(t, string(contents), "forgejo_owner: "+config.DefaultForgejoOwner)
}

func TestWriteDefaultFile_RefusesToOverwriteAnExistingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("github_owner: someone-else\n"), 0o600))

	err := config.WriteDefaultFile(path)

	require.Error(t, err)
	require.ErrorIs(t, err, config.ErrConfigExists)
	contents, readErr := os.ReadFile(path) //nolint:gosec // path is t.TempDir()-derived, not external input
	require.NoError(t, readErr)
	assert.Equal(t, "github_owner: someone-else\n", string(contents))
}
