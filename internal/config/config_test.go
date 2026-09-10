package config_test

import (
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateAcceptsBothOwnersSet(t *testing.T) {
	t.Parallel()

	err := config.Config{GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes"}.Validate()

	require.NoError(t, err)
}

func TestConfig_ValidateRejectsEmptyGitHubOwner(t *testing.T) {
	t.Parallel()

	err := config.Config{GitHubOwner: "", ForgejoOwner: "alrayyes"}.Validate()

	require.Error(t, err)
	require.ErrorIs(t, err, config.ErrMissingOwner)
	assert.Contains(t, err.Error(), "github_owner")
}

func TestConfig_ValidateRejectsEmptyForgejoOwner(t *testing.T) {
	t.Parallel()

	err := config.Config{GitHubOwner: "alrayyes", ForgejoOwner: ""}.Validate()

	require.Error(t, err)
	require.ErrorIs(t, err, config.ErrMissingOwner)
	assert.Contains(t, err.Error(), "forgejo_owner")
}
