// Command forgejo-mirror-sync finds a GitHub account's public, non-fork
// repos that have no Forgejo pull mirror yet, creates one for each, and
// fixes any archived-state mismatch on mirrors that already exist. GitHub
// is always the source of truth; nothing here ever writes back to it.
//
// It shells out to gh and tea for everything — see internal/runner's
// package doc for why: neither of those credentials is ever something this
// program itself handles.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/alrayyes/forgejo-mirror-sync/internal/confirm"
	"github.com/alrayyes/forgejo-mirror-sync/internal/forgejo"
	"github.com/alrayyes/forgejo-mirror-sync/internal/ghsource"
	"github.com/alrayyes/forgejo-mirror-sync/internal/plan"
	"github.com/alrayyes/forgejo-mirror-sync/internal/runner"
)

// version is stamped in at build time by goreleaser, from the tag.
var version = "dev"

// Options are the flags Run acts on.
type Options struct {
	GitHubOwner  string
	ForgejoOwner string
	DryRun       bool
	Yes          bool
	Verbose      bool
}

func main() {
	githubOwner := flag.String("github-owner", "alrayyes", "GitHub account to read public repos from")
	forgejoOwner := flag.String("forgejo-owner", "alrayyes", "Forgejo namespace mirrors live under")
	dryRun := flag.Bool("dry-run", false, "Print the plan and exit; never prompts, never writes")
	yes := flag.Bool("yes", false, "Skip the confirmation prompt")
	flag.BoolVar(yes, "y", false, "Shorthand for --yes")
	verbose := flag.Bool("verbose", false, "Log why each repo was skipped and the exact API calls made")
	showVersion := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	opts := Options{
		GitHubOwner:  *githubOwner,
		ForgejoOwner: *forgejoOwner,
		DryRun:       *dryRun,
		Yes:          *yes,
		Verbose:      *verbose,
	}

	if err := Run(context.Background(), os.Stdout, os.Stdin, runner.Exec{}, opts); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// Run does the actual work: list both sides, compute the plan, print it,
// confirm, then act. r is injectable so tests never shell out for real.
func Run(ctx context.Context, out io.Writer, in io.Reader, r runner.Runner, opts Options) error {
	githubRepos, err := (ghsource.Lister{Runner: r}).PublicRepos(ctx, opts.GitHubOwner)
	if err != nil {
		return fmt.Errorf("listing GitHub repos: %w", err)
	}

	forgejoClient := forgejo.Client{Runner: r}
	forgejoRepos, err := forgejoClient.Repos(ctx, opts.ForgejoOwner)
	if err != nil {
		return fmt.Errorf("listing Forgejo repos: %w", err)
	}

	result := plan.Compute(toPlanGitHub(githubRepos), toPlanForgejo(forgejoRepos))
	printSummary(out, result, opts.Verbose)

	if len(result.ToCreate) == 0 && len(result.ToArchiveFix) == 0 {
		_, _ = fmt.Fprintln(out, "Nothing to do.")
		return nil
	}

	if opts.DryRun {
		_, _ = fmt.Fprintln(out, "Dry run: no changes made.")
		return nil
	}

	if !opts.Yes {
		prompt := fmt.Sprintf("Create %d mirror(s) and fix %d archived flag(s)?", len(result.ToCreate), len(result.ToArchiveFix))
		ok, err := confirm.Ask(out, in, prompt)
		if err != nil {
			return fmt.Errorf("reading confirmation: %w", err)
		}
		if !ok {
			_, _ = fmt.Fprintln(out, "Aborted: no changes made.")
			return nil
		}
	}

	return apply(ctx, out, forgejoClient, opts, result)
}

func apply(ctx context.Context, out io.Writer, client forgejo.Client, opts Options, result plan.Result) error {
	var failures int

	for _, c := range result.ToCreate {
		if opts.Verbose {
			_, _ = fmt.Fprintf(out, "  running: tea api -X POST /repos/migrate (repo_owner=%s repo_name=%s)\n", opts.ForgejoOwner, c.Name)
		}
		if err := client.CreateMirror(ctx, opts.ForgejoOwner, c.Name, c.CloneURL); err != nil {
			_, _ = fmt.Fprintf(out, "  failed to create mirror for %s: %v\n", c.Name, err)
			failures++
			continue
		}
		_, _ = fmt.Fprintf(out, "  created mirror: %s\n", c.Name)
	}

	for _, f := range result.ToArchiveFix {
		if opts.Verbose {
			_, _ = fmt.Fprintf(out, "  running: tea api -X PATCH /repos/%s/%s (archived=%v)\n", opts.ForgejoOwner, f.Name, f.WantArchived)
		}
		if err := client.SetArchived(ctx, opts.ForgejoOwner, f.Name, f.WantArchived); err != nil {
			_, _ = fmt.Fprintf(out, "  failed to set archived=%v for %s: %v\n", f.WantArchived, f.Name, err)
			failures++
			continue
		}
		_, _ = fmt.Fprintf(out, "  archived=%v: %s\n", f.WantArchived, f.Name)
	}

	if failures > 0 {
		return fmt.Errorf("%d action(s) failed", failures)
	}
	return nil
}

func printSummary(out io.Writer, result plan.Result, verbose bool) {
	_, _ = fmt.Fprintf(out, "%d to create, %d archived-flag fix(es), %d already in sync, %d skipped\n",
		len(result.ToCreate), len(result.ToArchiveFix), len(result.InSync), len(result.Skipped))

	for _, c := range result.ToCreate {
		_, _ = fmt.Fprintf(out, "  create: %s\n", c.Name)
	}
	for _, f := range result.ToArchiveFix {
		_, _ = fmt.Fprintf(out, "  fix archived=%v: %s\n", f.WantArchived, f.Name)
	}

	if !verbose {
		return
	}
	for _, s := range result.Skipped {
		_, _ = fmt.Fprintf(out, "  skip %s: %s\n", s.Name, s.Reason)
	}
	for _, n := range result.InSync {
		_, _ = fmt.Fprintf(out, "  in sync: %s\n", n)
	}
}

func toPlanGitHub(repos []ghsource.Repo) []plan.GitHubRepo {
	out := make([]plan.GitHubRepo, len(repos))
	for i, r := range repos {
		out[i] = plan.GitHubRepo{Name: r.Name, CloneURL: r.CloneURL, Archived: r.Archived}
	}
	return out
}

func toPlanForgejo(repos []forgejo.Repo) []plan.ForgejoRepo {
	out := make([]plan.ForgejoRepo, len(repos))
	for i, r := range repos {
		out[i] = plan.ForgejoRepo{Name: r.Name, IsMirror: r.IsMirror, Archived: r.Archived}
	}
	return out
}
