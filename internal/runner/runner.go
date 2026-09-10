// Package runner is the one place this tool shells out to another command.
// gh and tea already hold whatever credential they need — nothing here ever
// sees, stores, or logs a token.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs a command and returns its standard output. Real code uses
// Exec; tests use a fake that returns canned output for a matched command,
// so neither internal/ghsource nor internal/forgejo ever needs a live
// network call to be tested.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Exec runs a real command on the host.
type Exec struct{}

// Run runs name with args and returns its stdout. On failure the error
// includes stderr, trimmed to one line — enough to say what went wrong
// without echoing a full API response body back to the terminal.
func (Exec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	//nolint:gosec // name/args always come from a call site inside this
	// repo (always "gh" or "tea" with fixed subcommands), never from
	// external input.
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		firstLine, _, _ := strings.Cut(strings.TrimSpace(stderr.String()), "\n")

		return nil, fmt.Errorf("%s: %w: %s", name, err, firstLine)
	}

	return stdout.Bytes(), nil
}
