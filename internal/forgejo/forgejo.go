// Package forgejo lists and mutates repos on git.higherlearning.eu by
// shelling out to tea, which already holds whatever credential it needs —
// this package never sees a token, and every write body it sends carries
// nothing beyond a repo name, clone URL, and archived flag.
package forgejo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alrayyes/forgejo-mirror-sync/internal/runner"
)

const pageSize = 50

// Repo is an existing Forgejo repo. IsMirror distinguishes a live pull
// mirror from a deliberately independent repo that happens to share a name.
type Repo struct {
	Name     string
	IsMirror bool
	Archived bool
}

// Client lists and mutates repos under a Forgejo user namespace.
type Client struct {
	Runner runner.Runner
}

type teaRepo struct {
	Name     string `json:"name"`
	Mirror   bool   `json:"mirror"`
	Archived bool   `json:"archived"`
}

// Repos returns every repo under owner's user namespace, paginating until a
// page comes back short of pageSize.
func (c Client) Repos(ctx context.Context, owner string) ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		path := fmt.Sprintf("/users/%s/repos?limit=%d&page=%d", owner, pageSize, page)
		out, err := c.Runner.Run(ctx, "tea", "api", path)
		if err != nil {
			return nil, fmt.Errorf("listing Forgejo repos for %s: %w", owner, err)
		}

		var raw []teaRepo
		if err := json.Unmarshal(out, &raw); err != nil {
			return nil, fmt.Errorf("parsing tea api repos response: %w", err)
		}
		for _, r := range raw {
			all = append(all, Repo{Name: r.Name, IsMirror: r.Mirror, Archived: r.Archived})
		}

		if len(raw) < pageSize {
			break
		}
	}
	return all, nil
}

// CreateMirror creates a new Forgejo pull mirror of cloneURL under
// owner/name, syncing every 8 hours — the same shape documented in
// skills/repo-creation for mirroring a GitHub repo onto Forgejo.
func (c Client) CreateMirror(ctx context.Context, owner, name, cloneURL string) error {
	body, err := json.Marshal(map[string]any{
		"clone_addr":      cloneURL,
		"repo_name":       name,
		"repo_owner":      owner,
		"mirror":          true,
		"private":         false,
		"mirror_interval": "8h0m0s",
	})
	if err != nil {
		return fmt.Errorf("encoding migrate request for %s: %w", name, err)
	}

	if _, err := c.Runner.Run(ctx, "tea", "api", "-X", "POST", "/repos/migrate", "-d", string(body)); err != nil {
		return fmt.Errorf("creating mirror for %s: %w", name, err)
	}
	return nil
}

// SetArchived sets owner/name's archived flag.
func (c Client) SetArchived(ctx context.Context, owner, name string, archived bool) error {
	body, err := json.Marshal(map[string]any{"archived": archived})
	if err != nil {
		return fmt.Errorf("encoding archive request for %s: %w", name, err)
	}

	path := fmt.Sprintf("/repos/%s/%s", owner, name)
	if _, err := c.Runner.Run(ctx, "tea", "api", "-X", "PATCH", path, "-d", string(body)); err != nil {
		return fmt.Errorf("setting archived=%v for %s: %w", archived, name, err)
	}
	return nil
}
