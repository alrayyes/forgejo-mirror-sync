// Package plan decides what to do; nothing in here talks to a network.
package plan_test

import (
	"testing"

	"github.com/alrayyes/forgejo-mirror-sync/internal/plan"
	"github.com/stretchr/testify/assert"
)

func TestCompute_MissingMirrorIsCreated(t *testing.T) {
	github := []plan.GitHubRepo{
		{Name: "widget", CloneURL: "https://github.com/alrayyes/widget.git", Archived: false},
	}

	result := plan.Compute(github, nil)

	assert.Equal(t, []plan.Create{
		{Name: "widget", CloneURL: "https://github.com/alrayyes/widget.git"},
	}, result.ToCreate)
	assert.Empty(t, result.ToArchiveFix)
	assert.Empty(t, result.InSync)
	assert.Empty(t, result.Skipped)
}

func TestCompute_ArchivedGitHubRepoWithNoMirrorIsSkipped(t *testing.T) {
	// Already settled — not worth mirroring for the first time.
	github := []plan.GitHubRepo{
		{Name: "old-thing", CloneURL: "https://github.com/alrayyes/old-thing.git", Archived: true},
	}

	result := plan.Compute(github, nil)

	assert.Empty(t, result.ToCreate)
	assert.Equal(t, []plan.Skip{
		{Name: "old-thing", Reason: "archived on GitHub with no existing mirror; already settled, not worth mirroring"},
	}, result.Skipped)
}

func TestCompute_NonMirrorSameNameRepoIsSkipped(t *testing.T) {
	// A scaffold's GitHub-native sibling, or any other deliberately
	// independent repo sharing a name — never overwritten.
	github := []plan.GitHubRepo{
		{Name: "scaffold-go-api", Archived: false},
	}
	forgejo := []plan.ForgejoRepo{
		{Name: "scaffold-go-api", IsMirror: false, Archived: false},
	}

	result := plan.Compute(github, forgejo)

	assert.Empty(t, result.ToCreate)
	assert.Empty(t, result.ToArchiveFix)
	assert.Equal(t, []plan.Skip{
		{Name: "scaffold-go-api", Reason: "existing Forgejo repo of the same name isn't a mirror; left alone"},
	}, result.Skipped)
}

func TestCompute_ArchivedMismatchOnExistingMirrorIsFixed(t *testing.T) {
	github := []plan.GitHubRepo{
		{Name: "widget", Archived: true},
	}
	forgejo := []plan.ForgejoRepo{
		{Name: "widget", IsMirror: true, Archived: false},
	}

	result := plan.Compute(github, forgejo)

	assert.Empty(t, result.ToCreate)
	assert.Equal(t, []plan.ArchiveFix{
		{Name: "widget", WantArchived: true},
	}, result.ToArchiveFix)
	assert.Empty(t, result.InSync)
}

func TestCompute_UnarchivedMismatchOnExistingMirrorIsFixed(t *testing.T) {
	github := []plan.GitHubRepo{
		{Name: "widget", Archived: false},
	}
	forgejo := []plan.ForgejoRepo{
		{Name: "widget", IsMirror: true, Archived: true},
	}

	result := plan.Compute(github, forgejo)

	assert.Equal(t, []plan.ArchiveFix{
		{Name: "widget", WantArchived: false},
	}, result.ToArchiveFix)
}

func TestCompute_MatchingMirrorIsLeftAlone(t *testing.T) {
	github := []plan.GitHubRepo{
		{Name: "widget", Archived: true},
	}
	forgejo := []plan.ForgejoRepo{
		{Name: "widget", IsMirror: true, Archived: true},
	}

	result := plan.Compute(github, forgejo)

	assert.Empty(t, result.ToCreate)
	assert.Empty(t, result.ToArchiveFix)
	assert.Equal(t, []string{"widget"}, result.InSync)
}

func TestCompute_ExtraForgejoRepoIsIgnored(t *testing.T) {
	// A repo that only exists on Forgejo isn't this tool's business — it
	// might be hand-created, or mirror a GitHub repo that's since gone
	// private or been deleted. Never touched either way.
	result := plan.Compute(nil, []plan.ForgejoRepo{{Name: "orphan", IsMirror: true, Archived: false}})

	assert.Empty(t, result.ToCreate)
	assert.Empty(t, result.ToArchiveFix)
	assert.Empty(t, result.InSync)
	assert.Empty(t, result.Skipped)
}

func TestCompute_IsDeterministicallyOrdered(t *testing.T) {
	github := []plan.GitHubRepo{
		{Name: "zeta", Archived: false},
		{Name: "alpha", Archived: false},
	}

	result := plan.Compute(github, nil)

	assert.Equal(t, []plan.Create{
		{Name: "alpha"},
		{Name: "zeta"},
	}, result.ToCreate)
}
