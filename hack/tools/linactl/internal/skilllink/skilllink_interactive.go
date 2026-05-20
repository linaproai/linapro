// This file implements the interactive selection flow used by skills.link
// and skills.unlink commands when callers omit the agent parameter on a TTY.
// The interactive flow uses only the Go standard library; it relies on
// os.File.Stat()'s ModeCharDevice bit to detect a real terminal so callers
// in CI or piped contexts continue to receive non-interactive output.

package skilllink

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

// ReadLine reads a single trimmed line from the provided reader. EOF is
// treated as an empty line so callers can interpret it as cancellation.
// Other read errors are wrapped with context for upstream display.
func ReadLine(in io.Reader) (string, error) {
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read line: %w", err)
	}
	return strings.TrimSpace(strings.ToLower(line)), nil
}

// SelectableEntry describes one candidate row for interactive selection.
type SelectableEntry struct {
	// Spec is the agent backing this row.
	Spec AgentSpec
	// CurrentStatus is the current Inspect status, used to render context.
	CurrentStatus Status
	// Detail mirrors Inspect's detail field for the same row.
	Detail string
}

// LinkCandidates returns selectable entries for skills.link interactive mode.
// native agents are excluded because they require no action; rootCollision
// agents are excluded because they require explicit FORCE=1.
func LinkCandidates(repoRoot string) []SelectableEntry {
	out := make([]SelectableEntry, 0)
	for _, spec := range agents {
		if spec.Category != CategoryLink {
			continue
		}
		result := Inspect(repoRoot, spec)
		out = append(out, SelectableEntry{
			Spec:          spec,
			CurrentStatus: result.Status,
			Detail:        result.Detail,
		})
	}
	return out
}

// UnlinkCandidates returns selectable entries for skills.unlink interactive
// mode. Only agents whose project path is currently a managed symlink (i.e.
// pointing at .agents/skills) are returned, since unlink is a no-op for any
// other state.
func UnlinkCandidates(repoRoot string) []SelectableEntry {
	out := make([]SelectableEntry, 0)
	for _, spec := range agents {
		if spec.Category == CategoryNative {
			continue
		}
		result := Inspect(repoRoot, spec)
		if result.Status != StatusOK {
			continue
		}
		out = append(out, SelectableEntry{
			Spec:          spec,
			CurrentStatus: result.Status,
			Detail:        result.Detail,
		})
	}
	return out
}

// PromptSelection runs an interactive numbered prompt and returns the agent
// names selected by the user. The function renders the candidate list to
// out, reads one line of input from in, and parses comma-separated indexes,
// "all" or "q" (cancel). An empty selection is treated as cancellation.
//
// Returned agent names are deduplicated and ordered by the candidate list.
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
			names = append(names, entry.Spec.Name)
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
		names = append(names, candidates[index-1].Spec.Name)
	}
	return names, nil
}

// PromptYesNo asks a yes/no question with a default answer used when the
// user submits an empty line. Used by skills.link to confirm FORCE rebuilds.
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
// list of link-class agents fits within the typical 24-row terminal viewport.
// Each cell shows the numbered index, a single-character status glyph and the
// agent name. A legend is printed before the grid so users can map glyphs
// back to their full status meanings without expanding the grid width.
//
// Glyphs:
//
//	[+] linked         ok
//	[~] mismatch       linked but pointing at a foreign target
//	[.] absent         not linked yet
//	[!] conflict       a real directory or file blocks linking
//	[*] root collision agent uses the repo-root skills/ path (openclaw)
//	[?] error          inspection failed; see status table for details
//
// Callers that need the full status text and project path can run the
// non-interactive listing via `make skills.link` without an AGENT argument.
func renderCandidateGrid(out io.Writer, candidates []SelectableEntry) error {
	const columns = 3
	if _, err := fmt.Fprintln(out, "  Legend: [+] linked  [~] mismatch  [.] absent  [!] conflict  [*] root-collision  [?] error"); err != nil {
		return fmt.Errorf("write candidate legend: %w", err)
	}
	maxName := 0
	for _, entry := range candidates {
		if width := len(entry.Spec.Name); width > maxName {
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
				statusGlyph(entry.CurrentStatus),
				maxName, entry.Spec.Name,
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

// statusGlyph maps a Status to a single-character indicator used in the
// interactive grid. Unknown statuses fall back to "?" so the grid never
// silently drops a state.
func statusGlyph(status Status) string {
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
