package forgejo_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/forgejo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type call struct {
	output []byte
	err    error
}

type fakeRunner struct {
	calls []call
	next  int
	seen  [][]string
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	got := append([]string{name}, args...)
	f.seen = append(f.seen, got)
	if f.next >= len(f.calls) {
		return nil, assert.AnError
	}
	c := f.calls[f.next]
	f.next++

	return c.output, c.err
}

func TestRepos_PaginatesUntilAShortPage(t *testing.T) {
	t.Parallel()

	page1 := make([]map[string]any, 50)
	for i := range page1 {
		page1[i] = map[string]any{"name": "repo", "mirror": true, "archived": false}
	}
	page1JSON, err := json.Marshal(page1)
	require.NoError(t, err)
	page2JSON := []byte(`[{"name": "last-one", "mirror": false, "archived": true}]`)

	fake := &fakeRunner{calls: []call{
		{output: page1JSON},
		{output: page2JSON},
	}}
	client := forgejo.Client{Runner: fake}

	repos, err := client.Repos(context.Background(), "alrayyes")

	require.NoError(t, err)
	assert.Len(t, repos, 51)
	assert.Equal(t, forgejo.Repo{Name: "last-one", IsMirror: false, Archived: true}, repos[50])
	require.Len(t, fake.seen, 2)
	assert.Equal(t, []string{"tea", "api", "/users/alrayyes/repos?limit=50&page=1"}, fake.seen[0])
	assert.Equal(t, []string{"tea", "api", "/users/alrayyes/repos?limit=50&page=2"}, fake.seen[1])
}

func TestRepos_StopsOnFirstEmptyPage(t *testing.T) {
	t.Parallel()

	fake := &fakeRunner{calls: []call{{output: []byte(`[]`)}}}
	client := forgejo.Client{Runner: fake}

	repos, err := client.Repos(context.Background(), "alrayyes")

	require.NoError(t, err)
	assert.Empty(t, repos)
	assert.Len(t, fake.seen, 1)
}

func TestCreateMirror_RunsExpectedTeaCommand(t *testing.T) {
	t.Parallel()

	fake := &fakeRunner{calls: []call{{output: []byte(`{}`)}}}
	client := forgejo.Client{Runner: fake}

	err := client.CreateMirror(context.Background(), "alrayyes", "widget", "https://github.com/alrayyes/widget.git")

	require.NoError(t, err)
	require.Len(t, fake.seen, 1)
	assert.Equal(t, []string{"tea", "api", "-X", "POST", "/repos/migrate", "-d"}, fake.seen[0][:6])

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(fake.seen[0][6]), &body))
	assert.Equal(t, "https://github.com/alrayyes/widget.git", body["clone_addr"])
	assert.Equal(t, "widget", body["repo_name"])
	assert.Equal(t, "alrayyes", body["repo_owner"])
	assert.Equal(t, true, body["mirror"])
	assert.Equal(t, false, body["private"])
	assert.Equal(t, "8h0m0s", body["mirror_interval"])
}

func TestSetArchived_RunsExpectedTeaCommand(t *testing.T) {
	t.Parallel()

	fake := &fakeRunner{calls: []call{{output: []byte(`{}`)}}}
	client := forgejo.Client{Runner: fake}

	err := client.SetArchived(context.Background(), "alrayyes", "widget", true)

	require.NoError(t, err)
	require.Len(t, fake.seen, 1)
	assert.Equal(t, []string{"tea", "api", "-X", "PATCH", "/repos/alrayyes/widget", "-d"}, fake.seen[0][:6])

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(fake.seen[0][6]), &body))
	assert.Equal(t, map[string]any{"archived": true}, body)
}

func TestCreateMirror_PropagatesRunnerError(t *testing.T) {
	t.Parallel()

	fake := &fakeRunner{calls: []call{{err: assert.AnError}}}
	client := forgejo.Client{Runner: fake}

	err := client.CreateMirror(context.Background(), "alrayyes", "widget", "https://github.com/alrayyes/widget.git")

	assert.Error(t, err)
}
