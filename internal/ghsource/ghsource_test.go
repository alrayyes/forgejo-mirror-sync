package ghsource_test

import (
	"context"
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/ghsource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	gotArgs []string
	output  []byte
	err     error
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.gotArgs = append([]string{name}, args...)
	return f.output, f.err
}

func TestPublicRepos_ParsesAndFiltersForks(t *testing.T) {
	fake := &fakeRunner{output: []byte(`[
		{"name": "widget", "isArchived": false, "isFork": false},
		{"name": "a-fork", "isArchived": false, "isFork": true},
		{"name": "old-thing", "isArchived": true, "isFork": false}
	]`)}
	lister := ghsource.Lister{Runner: fake}

	repos, err := lister.PublicRepos(context.Background(), "alrayyes")

	require.NoError(t, err)
	assert.Equal(t, []ghsource.Repo{
		{Name: "widget", CloneURL: "https://github.com/alrayyes/widget.git", Archived: false},
		{Name: "old-thing", CloneURL: "https://github.com/alrayyes/old-thing.git", Archived: true},
	}, repos)
}

func TestPublicRepos_RunsExpectedGhCommand(t *testing.T) {
	fake := &fakeRunner{output: []byte(`[]`)}
	lister := ghsource.Lister{Runner: fake}

	_, err := lister.PublicRepos(context.Background(), "alrayyes")

	require.NoError(t, err)
	assert.Equal(t, []string{
		"gh", "repo", "list", "alrayyes",
		"--visibility", "public",
		"--limit", "1000",
		"--json", "name,isArchived,isFork",
	}, fake.gotArgs)
}

func TestPublicRepos_PropagatesRunnerError(t *testing.T) {
	fake := &fakeRunner{err: assert.AnError}
	lister := ghsource.Lister{Runner: fake}

	_, err := lister.PublicRepos(context.Background(), "alrayyes")

	assert.Error(t, err)
}
