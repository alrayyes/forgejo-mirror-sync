// Package ghsource lists public GitHub repos by shelling out to gh, which
// already holds whatever credential it needs — this package never sees a
// token.
package ghsource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alrayyes/forgejo-mirror-sync/internal/runner"
)

// Repo is a public, non-fork GitHub repo.
type Repo struct {
	Name     string
	CloneURL string
	Archived bool
}

// Lister lists an owner's public repos.
type Lister struct {
	Runner runner.Runner
}

type ghRepo struct {
	Name       string `json:"name"`
	IsArchived bool   `json:"isArchived"`
	IsFork     bool   `json:"isFork"`
}

// PublicRepos returns owner's public, non-fork repos. gh's own pagination
// cap (1000) is passed explicitly since gh defaults to 30.
func (l Lister) PublicRepos(ctx context.Context, owner string) ([]Repo, error) {
	out, err := l.Runner.Run(ctx, "gh", "repo", "list", owner,
		"--visibility", "public",
		"--limit", "1000",
		"--json", "name,isArchived,isFork",
	)
	if err != nil {
		return nil, fmt.Errorf("listing GitHub repos for %s: %w", owner, err)
	}

	var raw []ghRepo
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing gh repo list output: %w", err)
	}

	repos := make([]Repo, 0, len(raw))
	for _, r := range raw {
		if r.IsFork {
			continue
		}
		repos = append(repos, Repo{
			Name:     r.Name,
			CloneURL: fmt.Sprintf("https://github.com/%s/%s.git", owner, r.Name),
			Archived: r.IsArchived,
		})
	}

	return repos, nil
}
