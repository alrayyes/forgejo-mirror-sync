package confirm_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/confirm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsk_AcceptsY(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ok, err := confirm.Ask(&out, strings.NewReader("y\n"), "Proceed?")

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, out.String(), "Proceed?")
}

func TestAsk_AcceptsYes(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ok, err := confirm.Ask(&out, strings.NewReader("yes\n"), "Proceed?")

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAsk_IsCaseInsensitive(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ok, err := confirm.Ask(&out, strings.NewReader("YES\n"), "Proceed?")

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAsk_DefaultsToNoOnEmptyInput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ok, err := confirm.Ask(&out, strings.NewReader("\n"), "Proceed?")

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestAsk_RejectsAnythingElse(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ok, err := confirm.Ask(&out, strings.NewReader("n\n"), "Proceed?")

	require.NoError(t, err)
	assert.False(t, ok)
}
