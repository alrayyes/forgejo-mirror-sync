package runner_test

import (
	"context"
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExec_ReturnsStdout(t *testing.T) {
	t.Parallel()

	out, err := runner.Exec{}.Run(context.Background(), "echo", "-n", "hello")

	require.NoError(t, err)
	assert.Equal(t, "hello", string(out))
}

func TestExec_WrapsFailureWithFirstStderrLine(t *testing.T) {
	t.Parallel()

	_, err := runner.Exec{}.Run(context.Background(), "sh", "-c", "printf 'first line\\nsecond line\\n' >&2; exit 1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "first line")
	assert.NotContains(t, err.Error(), "second line")
}
