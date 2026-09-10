package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedRunner fakes exactly the gh/tea calls Run makes, so these tests
// never touch a network or a real credential.
type scriptedRunner struct {
	ghOutput     []byte
	forgejoPages [][]byte
	pageIdx      int
	creates      []string        // repo_name pulled from each POST /repos/migrate body
	patches      []string        // "name=archived" pulled from each PATCH body
	failCreate   map[string]bool // repo names whose CreateMirror call should error
}

func (s *scriptedRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	switch {
	case name == "gh":
		return s.ghOutput, nil
	case name == "tea" && len(args) >= 2 && args[0] == "api" && strings.HasPrefix(args[1], "/users/"):
		out := s.forgejoPages[s.pageIdx]
		s.pageIdx++
		return out, nil
	case name == "tea" && len(args) >= 3 && args[1] == "-X" && args[2] == "POST":
		var body struct {
			RepoName string `json:"repo_name"`
		}
		_ = json.Unmarshal([]byte(args[len(args)-1]), &body)
		if s.failCreate[body.RepoName] {
			return nil, fmt.Errorf("simulated failure for %s", body.RepoName)
		}
		s.creates = append(s.creates, body.RepoName)
		return []byte(`{}`), nil
	case name == "tea" && len(args) >= 3 && args[1] == "-X" && args[2] == "PATCH":
		var body struct {
			Archived bool `json:"archived"`
		}
		_ = json.Unmarshal([]byte(args[len(args)-1]), &body)
		s.patches = append(s.patches, fmt.Sprintf("%s=%v", args[3], body.Archived))
		return []byte(`{}`), nil
	}
	return nil, fmt.Errorf("unexpected command: %s %v", name, args)
}

func newScriptedRunner() *scriptedRunner {
	return &scriptedRunner{
		ghOutput: []byte(`[
			{"name": "new-repo", "isArchived": false, "isFork": false},
			{"name": "widget", "isArchived": true, "isFork": false}
		]`),
		forgejoPages: [][]byte{
			[]byte(`[{"name": "widget", "mirror": true, "archived": false}]`),
		},
	}
}

func TestRun_DryRunMakesNoWrites(t *testing.T) {
	r := newScriptedRunner()
	var out bytes.Buffer

	err := Run(context.Background(), &out, strings.NewReader(""), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes", DryRun: true,
	})

	require.NoError(t, err)
	assert.Empty(t, r.creates)
	assert.Empty(t, r.patches)
	assert.Contains(t, out.String(), "Dry run")
}

func TestRun_YesSkipsPromptAndWrites(t *testing.T) {
	r := newScriptedRunner()
	var out bytes.Buffer

	err := Run(context.Background(), &out, strings.NewReader(""), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes", Yes: true,
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"new-repo"}, r.creates)
	assert.Equal(t, []string{"/repos/alrayyes/widget=true"}, r.patches)
}

func TestRun_DecliningPromptMakesNoWrites(t *testing.T) {
	r := newScriptedRunner()
	var out bytes.Buffer

	err := Run(context.Background(), &out, strings.NewReader("n\n"), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes",
	})

	require.NoError(t, err)
	assert.Empty(t, r.creates)
	assert.Empty(t, r.patches)
	assert.Contains(t, out.String(), "Aborted")
}

func TestRun_AcceptingPromptWrites(t *testing.T) {
	r := newScriptedRunner()
	var out bytes.Buffer

	err := Run(context.Background(), &out, strings.NewReader("y\n"), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes",
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"new-repo"}, r.creates)
	assert.Equal(t, []string{"/repos/alrayyes/widget=true"}, r.patches)
}

func TestRun_VerboseLogsSkipReasons(t *testing.T) {
	r := &scriptedRunner{
		ghOutput: []byte(`[{"name": "old-thing", "isArchived": true, "isFork": false}]`),
		forgejoPages: [][]byte{
			[]byte(`[]`),
		},
	}
	var out bytes.Buffer

	err := Run(context.Background(), &out, strings.NewReader(""), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes", DryRun: true, Verbose: true,
	})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "old-thing")
	assert.Contains(t, out.String(), "already settled, not worth mirroring")
}

func TestRun_ReportsFailedActionsWithoutStoppingTheRest(t *testing.T) {
	r := &scriptedRunner{
		ghOutput: []byte(`[
			{"name": "will-fail", "isArchived": false, "isFork": false},
			{"name": "will-succeed", "isArchived": false, "isFork": false}
		]`),
		forgejoPages: [][]byte{[]byte(`[]`)},
		failCreate:   map[string]bool{"will-fail": true},
	}
	var out bytes.Buffer

	err := Run(context.Background(), &out, strings.NewReader(""), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes", Yes: true, Verbose: true,
	})

	require.Error(t, err)
	assert.Equal(t, []string{"will-succeed"}, r.creates)
	assert.Contains(t, out.String(), "failed to create mirror for will-fail")
}

func TestRun_NothingToDoSkipsPrompt(t *testing.T) {
	r := &scriptedRunner{
		ghOutput:     []byte(`[{"name": "widget", "isArchived": false, "isFork": false}]`),
		forgejoPages: [][]byte{[]byte(`[{"name": "widget", "mirror": true, "archived": false}]`)},
	}
	var out bytes.Buffer

	// No stdin available to read from — if this tried to prompt, it would
	// block or error rather than return cleanly.
	err := Run(context.Background(), &out, strings.NewReader(""), r, Options{
		GitHubOwner: "alrayyes", ForgejoOwner: "alrayyes",
	})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Nothing to do")
}
