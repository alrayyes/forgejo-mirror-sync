package main_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	main "github.com/alrayyes/forgejo-mirror-sync/cmd/forgejo-mirror-sync"
	"github.com/alrayyes/forgejo-mirror-sync/internal/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestViper wires the same flag/env/default layers newRootCmd does, so
// these tests exercise the real precedence rather than a stand-in for it.
func newTestViper(t *testing.T) (*viper.Viper, *pflag.FlagSet) {
	t.Helper()

	v := viper.New()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("github-owner", config.DefaultGitHubOwner, "")
	fs.String("forgejo-owner", config.DefaultForgejoOwner, "")
	fs.Bool("dry-run", false, "")
	fs.BoolP("yes", "y", false, "")
	fs.Bool("verbose", false, "")

	v.SetEnvPrefix("FORGEJO_MIRROR_SYNC")
	v.AutomaticEnv()
	for key, flag := range map[string]string{
		"github_owner":  "github-owner",
		"forgejo_owner": "forgejo-owner",
		"dry_run":       "dry-run",
		"yes":           "yes",
		"verbose":       "verbose",
	} {
		require.NoError(t, v.BindPFlag(key, fs.Lookup(flag)))
	}
	v.SetDefault("github_owner", config.DefaultGitHubOwner)
	v.SetDefault("forgejo_owner", config.DefaultForgejoOwner)

	return v, fs
}

func TestLoadOptions_DefaultsWhenNothingElseSet(t *testing.T) {
	t.Parallel()

	v, _ := newTestViper(t)

	opts, err := main.LoadOptions(v)

	require.NoError(t, err)
	assert.Equal(t, config.DefaultGitHubOwner, opts.GitHubOwner)
	assert.Equal(t, config.DefaultForgejoOwner, opts.ForgejoOwner)
}

func TestLoadOptions_ConfigFileOverridesDefault(t *testing.T) {
	t.Parallel()

	v, _ := newTestViper(t)
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("github_owner: from-file\n")))

	opts, err := main.LoadOptions(v)

	require.NoError(t, err)
	assert.Equal(t, "from-file", opts.GitHubOwner)
}

// Not t.Parallel(): t.Setenv forbids it, since env is shared process state
// a concurrent test would race on.
func TestLoadOptions_EnvOverridesConfigFile(t *testing.T) {
	v, _ := newTestViper(t)
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("github_owner: from-file\n")))
	t.Setenv("FORGEJO_MIRROR_SYNC_GITHUB_OWNER", "from-env")

	opts, err := main.LoadOptions(v)

	require.NoError(t, err)
	assert.Equal(t, "from-env", opts.GitHubOwner)
}

// Not t.Parallel(): t.Setenv forbids it.
func TestLoadOptions_FlagOverridesEnv(t *testing.T) {
	v, fs := newTestViper(t)
	t.Setenv("FORGEJO_MIRROR_SYNC_GITHUB_OWNER", "from-env")
	require.NoError(t, fs.Set("github-owner", "from-flag"))

	opts, err := main.LoadOptions(v)

	require.NoError(t, err)
	assert.Equal(t, "from-flag", opts.GitHubOwner)
}

func TestLoadOptions_PropagatesValidationError(t *testing.T) {
	t.Parallel()

	v, fs := newTestViper(t)
	require.NoError(t, fs.Set("github-owner", ""))

	_, err := main.LoadOptions(v)

	require.Error(t, err)
	assert.ErrorIs(t, err, config.ErrMissingOwner)
}

func TestMaybeOfferInit_NoOpWhenConfigExists(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	err := main.MaybeOfferInit(&out, strings.NewReader(""), "/unused", true, false, true, false)

	require.NoError(t, err)
	assert.Empty(t, out.String())
}

func TestMaybeOfferInit_NoOpWhenRelevantEnvSet(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	err := main.MaybeOfferInit(&out, strings.NewReader(""), "/unused", false, true, true, false)

	require.NoError(t, err)
	assert.Empty(t, out.String())
}

func TestMaybeOfferInit_YesWritesConfigWithoutPrompting(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	var out bytes.Buffer

	err := main.MaybeOfferInit(&out, strings.NewReader(""), path, false, false, true, true)

	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.Contains(t, out.String(), "wrote defaults")
}

func TestMaybeOfferInit_NonInteractiveWithoutYesJustNotes(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	var out bytes.Buffer

	err := main.MaybeOfferInit(&out, strings.NewReader(""), path, false, false, false, false)

	require.NoError(t, err)
	assert.NoFileExists(t, path)
	assert.Contains(t, out.String(), "forgejo-mirror-sync init")
}

func TestMaybeOfferInit_InteractiveAcceptedWritesConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	var out bytes.Buffer

	err := main.MaybeOfferInit(&out, strings.NewReader("y\n"), path, false, false, true, false)

	require.NoError(t, err)
	assert.FileExists(t, path)
}

func TestMaybeOfferInit_InteractiveDeclinedLeavesNoFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	var out bytes.Buffer

	err := main.MaybeOfferInit(&out, strings.NewReader("n\n"), path, false, false, true, false)

	require.NoError(t, err)
	assert.NoFileExists(t, path)
}
