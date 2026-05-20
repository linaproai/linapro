// This file implements tabular result rendering and aggregate hint emission
// for skills.link and skills.unlink command output.

package skilllink

import (
	"fmt"
	"io"
)

// Render writes the result table for a list of results.
func Render(out io.Writer, results []Result) error {
	const (
		columnAgent    = "AGENT"
		columnPath     = "PROJECT PATH"
		columnCategory = "CATEGORY"
		columnStatus   = "STATUS"
		columnDetail   = "DETAIL"
	)
	maxAgent := len(columnAgent)
	maxPath := len(columnPath)
	maxCategory := len(columnCategory)
	maxStatus := len(columnStatus)
	for _, result := range results {
		if width := len(result.Spec.Name); width > maxAgent {
			maxAgent = width
		}
		if width := len(result.Spec.ProjectPath); width > maxPath {
			maxPath = width
		}
		if width := len(string(result.Spec.Category)); width > maxCategory {
			maxCategory = width
		}
		if width := len(string(result.Status)); width > maxStatus {
			maxStatus = width
		}
	}
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s\n",
		maxAgent, columnAgent,
		maxPath, columnPath,
		maxCategory, columnCategory,
		maxStatus, columnStatus,
		columnDetail,
	)
	if _, err := io.WriteString(out, header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, result := range results {
		row := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s\n",
			maxAgent, result.Spec.Name,
			maxPath, result.Spec.ProjectPath,
			maxCategory, string(result.Spec.Category),
			maxStatus, string(result.Status),
			result.Detail,
		)
		if _, err := io.WriteString(out, row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	return nil
}

// EmitHints writes follow-up hints derived from a result list.
func EmitHints(out io.Writer, results []Result) error {
	hasMismatch := false
	hasConflict := false
	hasRootCollision := false
	hasError := false
	for _, result := range results {
		switch result.Status {
		case StatusMismatch:
			hasMismatch = true
		case StatusConflict:
			hasConflict = true
		case StatusSkippedRootCollision:
			hasRootCollision = true
		case StatusError:
			hasError = true
		}
	}
	if hasMismatch {
		if _, err := fmt.Fprintln(out, "Hint: rerun with FORCE=1 to rebuild mismatched links."); err != nil {
			return fmt.Errorf("write hint: %w", err)
		}
	}
	if hasConflict {
		if _, err := fmt.Fprintln(out, "Hint: real directories or files block linking. Resolve listed paths manually."); err != nil {
			return fmt.Errorf("write hint: %w", err)
		}
	}
	if hasRootCollision {
		if _, err := fmt.Fprintln(out, "Hint: rootCollision agents (e.g. openclaw) require explicit AGENT=<name> FORCE=1."); err != nil {
			return fmt.Errorf("write hint: %w", err)
		}
	}
	if hasError {
		if _, err := fmt.Fprintln(out, "Hint: see DETAIL column for the underlying error message."); err != nil {
			return fmt.Errorf("write hint: %w", err)
		}
	}
	return nil
}

// HasError reports whether the result list contains any StatusError entries.
func HasError(results []Result) bool {
	for _, result := range results {
		if result.IsError() {
			return true
		}
	}
	return false
}
