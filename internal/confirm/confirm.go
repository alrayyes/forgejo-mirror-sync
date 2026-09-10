// Package confirm asks a yes/no question on the terminal. The reader and
// writer are both injectable so main.go's --yes/--dry-run flags can skip it
// entirely, and so it's testable without a real terminal.
package confirm

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Ask prints prompt followed by "[y/N]: " to w, reads one line from r, and
// reports whether the answer was an affirmative y/yes (case-insensitive).
// Anything else, including empty input, is no.
func Ask(w io.Writer, r io.Reader, prompt string) (bool, error) {
	if _, err := fmt.Fprintf(w, "%s [y/N]: ", prompt); err != nil {
		return false, err
	}

	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
