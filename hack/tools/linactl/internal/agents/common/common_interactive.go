// This file implements the interactive selection helpers shared by every
// resource subpackage. The flow uses only the Go standard library: it
// detects a real terminal via os.File.Stat()'s ModeCharDevice bit, reads
// numbered selections from a generic io.Reader, and renders candidate
// agents in a 3-column glyph grid with a legend so users can understand
// each candidate's current status without scrolling.

package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// IsInteractiveTerminal reports whether the provided file looks like an
// interactive terminal. Returning false also covers nil files, pipes and
// regular files used by tests.
func IsInteractiveTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// ReadLine reads a single trimmed and lower-cased line from the provided
// reader. EOF is treated as an empty line so callers can interpret it as
// cancellation. Other read errors are wrapped with context for upstream
// display.
func ReadLine(in io.Reader) (string, error) {
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read line: %w", err)
	}
	return strings.TrimSpace(strings.ToLower(line)), nil
}

// SelectableEntry describes one candidate row for interactive selection.
// The Spec field is held as SpecLike so the helper works for any resource
// subpackage's AgentSpec type.
type SelectableEntry struct {
	// Spec is the agent backing this row.
	Spec SpecLike
	// CurrentStatus is the current Inspect status, used to render context.
	CurrentStatus Status
	// Detail mirrors Inspect's detail field for the same row.
	Detail string
}

// PromptSelection runs an interactive numbered prompt and returns the
// agent names selected by the user. The function renders the candidate
// list to out, reads one line of input from in, and parses comma-separated
// indexes, "all" or "q" (cancel). An empty selection is treated as
// cancellation. Returned agent names are deduplicated and ordered by the
// candidate list.
func PromptSelection(in io.Reader, out io.Writer, title string, candidates []SelectableEntry) ([]string, error) {
	if len(candidates) == 0 {
		fmt.Fprintln(out, title+": no candidates available.")
		return nil, nil
	}
	fmt.Fprintln(out, title)
	if err := renderCandidateGrid(out, candidates); err != nil {
		return nil, err
	}
	fmt.Fprintln(out, "Enter numbers separated by commas (e.g. 1,3,5), 'all' for everything, or 'q' to cancel:")
	fmt.Fprint(out, "> ")

	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read selection: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" || strings.EqualFold(line, "q") || strings.EqualFold(line, "quit") {
		fmt.Fprintln(out, "Cancelled.")
		return nil, nil
	}
	if strings.EqualFold(line, "all") {
		names := make([]string, 0, len(candidates))
		for _, entry := range candidates {
			names = append(names, entry.Spec.SpecName())
		}
		return names, nil
	}

	tokens := strings.Split(line, ",")
	picked := make(map[int]struct{}, len(tokens))
	names := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		index, parseErr := strconv.Atoi(token)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid selection %q: expected a number", token)
		}
		if index < 1 || index > len(candidates) {
			return nil, fmt.Errorf("selection %d out of range [1..%d]", index, len(candidates))
		}
		if _, dup := picked[index]; dup {
			continue
		}
		picked[index] = struct{}{}
		names = append(names, candidates[index-1].Spec.SpecName())
	}
	return names, nil
}

// PromptYesNo asks a yes/no question with a default answer used when the
// user submits an empty line. Used by link flows to confirm FORCE rebuilds.
func PromptYesNo(in io.Reader, out io.Writer, question string, defaultYes bool) (bool, error) {
	suffix := " [y/N] "
	if defaultYes {
		suffix = " [Y/n] "
	}
	fmt.Fprint(out, question+suffix)
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultYes, nil
	}
	switch line {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid answer %q: expected y or n", line)
	}
}

// renderCandidateGrid lays out candidates in a 3-column grid so the entire
// list of link-class agents fits within the typical 24-row terminal
// viewport. Each cell shows the numbered index, a single-character status
// glyph and the agent name. A legend is printed before the grid so users
// can map glyphs back to their full status meanings without expanding the
// grid width.
//
// Glyphs:
//
//	[+] linked         ok / created / rebuilt / removed
//	[~] mismatch       linked but pointing at a foreign target
//	[.] absent         not linked yet (or native, no action)
//	[!] conflict       a real directory or file blocks linking
//	[*] root collision agent uses a colliding repo-root path (openclaw)
//	[?] error          inspection failed; see status table for details
func renderCandidateGrid(out io.Writer, candidates []SelectableEntry) error {
	const columns = 3
	if _, err := fmt.Fprintln(out, "  Legend: [+] linked  [~] mismatch  [.] absent  [!] conflict  [*] root-collision  [?] error"); err != nil {
		return fmt.Errorf("write candidate legend: %w", err)
	}
	maxName := 0
	for _, entry := range candidates {
		if width := len(entry.Spec.SpecName()); width > maxName {
			maxName = width
		}
	}
	rows := (len(candidates) + columns - 1) / columns
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			index := column*rows + row
			if index >= len(candidates) {
				continue
			}
			entry := candidates[index]
			separator := "  "
			if column == columns-1 || index == len(candidates)-1 {
				separator = ""
			}
			if _, err := fmt.Fprintf(out, "  [%2d] %s %-*s%s",
				index+1,
				StatusGlyph(entry.CurrentStatus),
				maxName, entry.Spec.SpecName(),
				separator,
			); err != nil {
				return fmt.Errorf("write candidate grid: %w", err)
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return fmt.Errorf("write candidate grid: %w", err)
		}
	}
	return nil
}

// StatusGlyph maps a Status to a single-character indicator used in the
// interactive grid. Unknown statuses fall back to "?" so the grid never
// silently drops a state.
func StatusGlyph(status Status) string {
	switch status {
	case StatusOK, StatusCreated, StatusRebuilt, StatusRemoved:
		return "[+]"
	case StatusMismatch:
		return "[~]"
	case StatusAbsent, StatusNative:
		return "[.]"
	case StatusConflict:
		return "[!]"
	case StatusSkippedRootCollision:
		return "[*]"
	case StatusSkippedForeignTarget, StatusSkippedNotManaged:
		return "[~]"
	case StatusError:
		return "[?]"
	default:
		return "[?]"
	}
}
