// Package plugingovernance implements plugin governance scans for linactl. It
// checks plugin production paths under apps/lina-plugins for host core table
// generation, direct host storage access, legacy host-service declarations,
// dynamic data-service grants, and owner capability imports that would bypass
// plugin dependency governance.
package plugingovernance

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	// formatText prints a concise human-readable report.
	formatText = "text"
	// formatJSON prints the complete report as JSON for CI consumers.
	formatJSON = "json"
)

// Options configures one plugin governance scan run.
type Options struct {
	// Format controls report output. Supported values are "text" and "json".
	Format string
}

// Finding reports one plugin governance violation.
type Finding struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Content  string `json:"content"`
}

// Summary aggregates scan coverage and finding counts.
type Summary struct {
	Findings      int            `json:"findings"`
	FilesScanned  int            `json:"filesScanned"`
	ConfigFiles   int            `json:"configFiles"`
	ManifestFiles int            `json:"manifestFiles"`
	GoFiles       int            `json:"goFiles"`
	ByRule        map[string]int `json:"byRule"`
	ByCategory    map[string]int `json:"byCategory"`
}

// Report stores one complete plugin governance scan result.
type Report struct {
	Summary  Summary   `json:"summary"`
	Findings []Finding `json:"findings"`
}

// RunCheck scans repoRoot, writes the selected report format, and returns an
// error when any governance finding is present.
func RunCheck(repoRoot string, out io.Writer, options Options) error {
	format := strings.TrimSpace(options.Format)
	if format == "" {
		format = formatText
	}
	if format != formatText && format != formatJSON {
		return fmt.Errorf("unsupported plugin check report format %q", format)
	}

	report, err := Scan(repoRoot)
	if err != nil {
		return err
	}
	if err = emitReport(out, report, format); err != nil {
		return err
	}
	if len(report.Findings) > 0 {
		return fmt.Errorf("plugin check failed with %d finding(s)", len(report.Findings))
	}
	return nil
}

// newReport creates a report with initialized maps.
func newReport() *Report {
	return &Report{
		Summary: Summary{
			ByRule:     make(map[string]int),
			ByCategory: make(map[string]int),
		},
		Findings: make([]Finding, 0),
	}
}

// addFinding appends one sorted-report finding.
func addFinding(report *Report, path string, line int, rule string, category string, message string, content string) {
	report.Findings = append(report.Findings, Finding{
		Path:     path,
		Line:     line,
		Rule:     rule,
		Category: category,
		Message:  message,
		Content:  strings.TrimSpace(content),
	})
}

// finalizeReport sorts findings and updates summary counters.
func finalizeReport(report *Report) {
	sort.Slice(report.Findings, func(left int, right int) bool {
		if report.Findings[left].Path != report.Findings[right].Path {
			return report.Findings[left].Path < report.Findings[right].Path
		}
		if report.Findings[left].Line != report.Findings[right].Line {
			return report.Findings[left].Line < report.Findings[right].Line
		}
		return report.Findings[left].Rule < report.Findings[right].Rule
	})

	report.Summary.Findings = len(report.Findings)
	report.Summary.FilesScanned = report.Summary.ConfigFiles + report.Summary.ManifestFiles + report.Summary.GoFiles
	for _, finding := range report.Findings {
		report.Summary.ByRule[finding.Rule]++
		report.Summary.ByCategory[finding.Category]++
	}
}

// emitReport writes text or JSON output for one report.
func emitReport(out io.Writer, report *Report, format string) error {
	switch format {
	case formatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	default:
		return emitTextReport(out, report)
	}
}

// emitTextReport writes a stable human-readable scan report.
func emitTextReport(out io.Writer, report *Report) error {
	if _, err := fmt.Fprintln(out, "Plugin check summary:"); err != nil {
		return err
	}
	lines := []string{
		fmt.Sprintf("  files scanned: %d", report.Summary.FilesScanned),
		fmt.Sprintf("  config files: %d", report.Summary.ConfigFiles),
		fmt.Sprintf("  manifest files: %d", report.Summary.ManifestFiles),
		fmt.Sprintf("  Go files: %d", report.Summary.GoFiles),
		fmt.Sprintf("  findings: %d", len(report.Findings)),
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	if len(report.Findings) == 0 {
		_, err := fmt.Fprintln(out, "Plugin check passed: no production governance finding found.")
		return err
	}

	if _, err := fmt.Fprintln(out, "Plugin check findings:"); err != nil {
		return err
	}
	for _, finding := range report.Findings {
		if _, err := fmt.Fprintf(
			out,
			"  %s:%d [%s/%s] %s",
			finding.Path,
			finding.Line,
			finding.Rule,
			finding.Category,
			finding.Message,
		); err != nil {
			return err
		}
		if finding.Content != "" {
			if _, err := fmt.Fprintf(out, " -> %s", finding.Content); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}
	return nil
}
