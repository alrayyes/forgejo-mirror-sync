// Package plan reconciles a GitHub repo list against a Forgejo one. GitHub
// is always the source of truth: a repo missing on Forgejo gets a new
// mirror, an archived-state mismatch on an existing mirror gets fixed to
// match GitHub, and everything else is left alone. Nothing here touches a
// network — that split is what makes the decision logic testable without
// either API.
package plan

import "sort"

// GitHubRepo is the subset of a GitHub repo this tool cares about. Forks are
// expected to already be filtered out by the caller.
type GitHubRepo struct {
	Name     string
	CloneURL string
	Archived bool
}

// ForgejoRepo is the subset of an existing Forgejo repo this tool cares
// about. IsMirror distinguishes a live pull mirror from a deliberately
// independent repo that happens to share a name — a scaffold's GitHub-native
// sibling, for instance.
type ForgejoRepo struct {
	Name     string
	IsMirror bool
	Archived bool
}

// Create is a Forgejo pull mirror that doesn't exist yet.
type Create struct {
	Name     string
	CloneURL string
}

// ArchiveFix is an existing mirror whose archived flag needs to change to
// match GitHub.
type ArchiveFix struct {
	Name         string
	WantArchived bool
}

// Skip is a GitHub repo left alone, and why.
type Skip struct {
	Name   string
	Reason string
}

// Result is what Compute decided. InSync lists repo names that already
// match, so a caller can report them in verbose mode without recomputing
// anything.
type Result struct {
	ToCreate     []Create
	ToArchiveFix []ArchiveFix
	InSync       []string
	Skipped      []Skip
}

// Compute reconciles github against forgejo. Both slices may be in any
// order; the result is sorted by repo name so output is stable across runs.
func Compute(github []GitHubRepo, forgejo []ForgejoRepo) Result {
	existing := make(map[string]ForgejoRepo, len(forgejo))
	for _, r := range forgejo {
		existing[r.Name] = r
	}

	var result Result
	for _, gh := range github {
		fr, ok := existing[gh.Name]
		switch {
		case !ok && gh.Archived:
			result.Skipped = append(result.Skipped, Skip{
				Name:   gh.Name,
				Reason: "archived on GitHub with no existing mirror; already settled, not worth mirroring",
			})
		case !ok:
			result.ToCreate = append(result.ToCreate, Create{Name: gh.Name, CloneURL: gh.CloneURL})
		case !fr.IsMirror:
			result.Skipped = append(result.Skipped, Skip{
				Name:   gh.Name,
				Reason: "existing Forgejo repo of the same name isn't a mirror; left alone",
			})
		case fr.Archived != gh.Archived:
			result.ToArchiveFix = append(result.ToArchiveFix, ArchiveFix{Name: gh.Name, WantArchived: gh.Archived})
		default:
			result.InSync = append(result.InSync, gh.Name)
		}
	}

	sort.Slice(result.ToCreate, func(i, j int) bool { return result.ToCreate[i].Name < result.ToCreate[j].Name })
	sort.Slice(result.ToArchiveFix, func(i, j int) bool { return result.ToArchiveFix[i].Name < result.ToArchiveFix[j].Name })
	sort.Slice(result.Skipped, func(i, j int) bool { return result.Skipped[i].Name < result.Skipped[j].Name })
	sort.Strings(result.InSync)

	return result
}
