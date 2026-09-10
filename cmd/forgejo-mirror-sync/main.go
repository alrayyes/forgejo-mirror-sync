// Command forgejo-mirror-sync finds a GitHub account's public, non-fork
// repos that have no Forgejo pull mirror yet, creates one for each, and
// fixes any archived-state mismatch on mirrors that already exist. GitHub
// is always the source of truth; nothing here ever writes back to it.
//
// It shells out to gh and tea for everything — see internal/runner's
// package doc for why: neither of those credentials is ever something this
// program itself handles.
//
// Configuration layers flags over environment variables
// (FORGEJO_MIRROR_SYNC_*) over an XDG config file over built-in defaults —
// see internal/config's package doc for the file's shape, and `init`
// below for how one gets written.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/alrayyes/forgejo-mirror-sync/internal/config"
	"github.com/alrayyes/forgejo-mirror-sync/internal/confirm"
	"github.com/alrayyes/forgejo-mirror-sync/internal/forgejo"
	"github.com/alrayyes/forgejo-mirror-sync/internal/ghsource"
	"github.com/alrayyes/forgejo-mirror-sync/internal/plan"
	"github.com/alrayyes/forgejo-mirror-sync/internal/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// version is stamped in at build time by goreleaser, from the tag.
var version = "dev"

// envPrefix is the prefix every config setting's environment variable
// carries — FORGEJO_MIRROR_SYNC_GITHUB_OWNER, and so on.
const envPrefix = "FORGEJO_MIRROR_SYNC"

// configRelPath is this tool's config file, relative to an XDG config
// directory.
const configRelPath = "forgejo-mirror-sync/config.yaml"

var (
	errNoTerminal    = errors.New("no terminal to confirm on: pass --yes to proceed or --dry-run to only preview")
	errActionsFailed = errors.New("some actions failed")
)

// Options are the fully-resolved settings Run acts on — flags, env vars
// and the config file already layered into one value. Interactive says
// whether stdin is a real TTY — set from main, never from a flag — and
// gates every prompt: a piped or scripted invocation with no --yes must
// fail closed rather than block forever on a read nothing will ever send.
type Options struct {
	GitHubOwner  string
	ForgejoOwner string
	DryRun       bool
	Yes          bool
	Verbose      bool
	Interactive  bool
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	v := viper.New()

	cmd := &cobra.Command{
		Use:           "forgejo-mirror-sync",
		Short:         "Mirror public GitHub repos to Forgejo and keep archived state in sync",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRoot(cmd, v)
		},
	}

	cmd.Flags().String("github-owner", config.DefaultGitHubOwner, "GitHub account to read public repos from")
	cmd.Flags().String("forgejo-owner", config.DefaultForgejoOwner, "Forgejo namespace mirrors live under")
	cmd.Flags().Bool("dry-run", false, "Print the plan and exit; never prompts, never writes")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts, and write a missing config non-interactively")
	cmd.Flags().Bool("verbose", false, "Log why each repo was skipped and the exact API calls made")

	bindConfig(v, cmd)
	cmd.AddCommand(newInitCmd())

	return cmd
}

// bindConfig wires viper's three layers — flags, environment, defaults —
// together. The config file layer is read separately in runRoot, since it
// depends on whether one actually exists.
func bindConfig(v *viper.Viper, cmd *cobra.Command) {
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	for key, flag := range map[string]string{
		"github_owner":  "github-owner",
		"forgejo_owner": "forgejo-owner",
		"dry_run":       "dry-run",
		"yes":           "yes",
		"verbose":       "verbose",
	} {
		_ = v.BindPFlag(key, cmd.Flags().Lookup(flag))
	}

	v.SetDefault("github_owner", config.DefaultGitHubOwner)
	v.SetDefault("forgejo_owner", config.DefaultForgejoOwner)
}

func runRoot(cmd *cobra.Command, v *viper.Viper) error {
	configPath, configExists, err := readConfigFile(v)
	if err != nil {
		return err
	}

	opts, err := LoadOptions(v)
	if err != nil {
		return err
	}
	opts.Interactive = term.IsTerminal(int(os.Stdin.Fd()))

	out := cmd.OutOrStdout()
	if err := MaybeOfferInit(out, cmd.InOrStdin(), configPath, configExists, relevantEnvSet(), opts.Interactive, opts.Yes); err != nil {
		return err
	}

	return Run(cmd.Context(), out, cmd.InOrStdin(), runner.Exec{}, opts)
}

// readConfigFile locates forgejo-mirror-sync's config file under the XDG
// config search path and, if one exists, reads it into v. It always
// returns the path a config file would live at in the primary XDG config
// directory — for display, and for maybeOfferInit to write to — regardless
// of whether one was found.
func readConfigFile(v *viper.Viper) (path string, exists bool, err error) {
	displayPath := filepath.Join(xdg.ConfigHome, configRelPath)

	found, searchErr := xdg.SearchConfigFile(configRelPath)
	if searchErr != nil {
		return displayPath, false, nil
	}

	v.SetConfigFile(found)
	if err := v.ReadInConfig(); err != nil {
		return found, true, fmt.Errorf("reading config file %s: %w", found, err)
	}

	return found, true, nil
}

func relevantEnvSet() bool {
	for _, key := range []string{"GITHUB_OWNER", "FORGEJO_OWNER", "DRY_RUN", "YES", "VERBOSE"} {
		if _, ok := os.LookupEnv(envPrefix + "_" + key); ok {
			return true
		}
	}

	return false
}

// LoadOptions reads v's already-layered flag/env/default values into an
// Options and validates it. v is expected to already have the config file
// layer read in, if one was found.
func LoadOptions(v *viper.Viper) (Options, error) {
	opts := Options{
		GitHubOwner:  v.GetString("github_owner"),
		ForgejoOwner: v.GetString("forgejo_owner"),
		DryRun:       v.GetBool("dry_run"),
		Yes:          v.GetBool("yes"),
		Verbose:      v.GetBool("verbose"),
	}

	cfg := config.Config{GitHubOwner: opts.GitHubOwner, ForgejoOwner: opts.ForgejoOwner}
	if err := cfg.Validate(); err != nil {
		return Options{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return opts, nil
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Write a starter config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := xdg.ConfigFile(configRelPath)
			if err != nil {
				return fmt.Errorf("resolving config path: %w", err)
			}

			if err := config.WriteDefaultFile(path); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", path)

			return nil
		},
	}
}

// MaybeOfferInit is the first-run prompt: an unconfigured run (no file, no
// relevant env var) is exactly the moment someone needs pointing at
// `init`, not left to find it in --help — but only until a config file
// exists; after that this is a no-op every time.
func MaybeOfferInit(out io.Writer, in io.Reader, path string, configExists, envSet, interactive, yes bool) error {
	if configExists || envSet {
		return nil
	}

	if yes {
		if err := writeDefaultConfig(out, path, "No config file found — wrote defaults to %s.\n"); err != nil {
			return err
		}

		return nil
	}

	if !interactive {
		_, _ = fmt.Fprintf(out, "No config file found; run %q to create one. Continuing on built-in defaults.\n", "forgejo-mirror-sync init")

		return nil
	}

	ok, err := confirm.Ask(out, in, fmt.Sprintf("No config file found. Write one with today's defaults to %s?", path))
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}

	if !ok {
		return nil
	}

	return writeDefaultConfig(out, path, "Wrote %s.\n")
}

func writeDefaultConfig(out io.Writer, path, format string) error {
	if err := config.WriteDefaultFile(path); err != nil && !errors.Is(err, config.ErrConfigExists) {
		return fmt.Errorf("writing default config: %w", err)
	}

	_, _ = fmt.Fprintf(out, format, path)

	return nil
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
		if !opts.Interactive {
			return errNoTerminal
		}

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
		return fmt.Errorf("%w: %d", errActionsFailed, failures)
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
