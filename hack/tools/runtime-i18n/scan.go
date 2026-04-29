// This file scans runtime-visible source code copy that should be routed
// through i18n resources.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// hanPattern identifies Han characters in runtime-visible source lines.
var hanPattern = regexp.MustCompile(`[\x{3400}-\x{9fff}]`)

// goStringLiteralHanPattern identifies interpreted or raw Go string literals
// that contain Han characters.
var goStringLiteralHanPattern = `(?:"[^"\n]*[\x{3400}-\x{9fff}][^"\n]*"|` + "`" + `[^` + "`" + `]*[\x{3400}-\x{9fff}][^` + "`" + `]*` + "`" + `)`

// scannerRules stores every high-risk runtime i18n source-code pattern.
var scannerRules = []scanRule{
	{
		Identifier: "go-caller-error-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`(?:gerror\.(?:New|Newf|Wrap|Wrapf|NewCode|NewCodef|WrapCode|WrapCodef)|errors\.New|fmt\.Errorf)\([^)\n]*` + goStringLiteralHanPattern),
		Message: "Go errors visible to users must use structured runtime message errors.",
	},
	{
		Identifier: "go-message-field-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`\b(?:Reason|Message|Fallback|Label|Title|DisabledReason)\s*:\s*` + goStringLiteralHanPattern),
		Message: "Reason, Message, Fallback, Label, Title, and DisabledReason fields must use message keys or rendered i18n text.",
	},
	{
		Identifier: "go-message-assignment-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`\b(?:[A-Za-z0-9_]+\.)?(?:Reason|Message|Fallback|Label|Title|DisabledReason)\s*=\s*(?:fmt\.Sprintf\()?` + goStringLiteralHanPattern),
		Message: "Result display fields must carry message keys, message params, or rendered runtime i18n text.",
	},
	{
		Identifier: "go-artifact-slice-han",
		Category:   "UserArtifact",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`(?:\[\]string\s*\{[^}\n]*|(?:headers?|columns?|titles?)\s*(?::=|=)\s*\[\]string\s*\{?[^}\n]*)` + goStringLiteralHanPattern),
		Message: "Export headers and template labels must be rendered through runtime i18n.",
	},
	{
		Identifier: "go-status-label-map-han",
		Category:   "UserArtifact",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`(?i)\b(?:status|type|label|title)[A-Za-z0-9_]*(?:Map|Labels?|Titles?)?\b.*(?:map\[|\{).*` + goStringLiteralHanPattern),
		Message: "Status, type, and label maps must use dictionaries or runtime i18n keys.",
	},
	{
		Identifier: "go-hostcall-error-han",
		Category:   "DeveloperDiagnostic",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`(?:NewHostCallErrorResponse|NewHostServiceErrorResponse)\([^)\n]*` + goStringLiteralHanPattern),
		Message: "Plugin bridge errors must expose stable codes and English developer diagnostics.",
	},
	{
		Identifier: "go-string-literal-han",
		Category:   "Unclassified",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(goStringLiteralHanPattern),
		Message: "Hand-written backend Go string literals must not contain fixed-language Chinese copy.",
	},
	{
		Identifier: "frontend-property-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-vben/apps/web-antd/src/**/*.ts",
			"apps/lina-vben/apps/web-antd/src/**/*.vue",
			"apps/lina-plugins/*/frontend/**/*.vue",
		},
		Pattern: regexp.MustCompile(`\b(?:title|label|placeholder|content|emptyText)\s*:\s*['"][^'"]*[\x{3400}-\x{9fff}]`),
		Message: "Frontend visible properties must use $t or runtime i18n keys.",
	},
	{
		Identifier: "frontend-message-call-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-vben/apps/web-antd/src/**/*.ts",
			"apps/lina-vben/apps/web-antd/src/**/*.vue",
			"apps/lina-plugins/*/frontend/**/*.vue",
		},
		Pattern: regexp.MustCompile(`\b(?:message|notification)\.(?:success|error|warning|info|loading)\([^)\n]*['"][^'"]*[\x{3400}-\x{9fff}]`),
		Message: "Frontend toast/notification text must use $t or runtime i18n keys.",
	},
	{
		Identifier: "frontend-template-text-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-vben/apps/web-antd/src/**/*.vue",
			"apps/lina-plugins/*/frontend/**/*.vue",
		},
		Pattern: regexp.MustCompile(`>[^<>{}$]*[\x{3400}-\x{9fff}][^<>{}$]*<`),
		Message: "Vue template text nodes must use $t or runtime i18n keys.",
	},
}

// excludedPathParts stores path components skipped by the scanner.
var excludedPathParts = map[string]struct{}{
	".git":              {},
	"node_modules":      {},
	"dist":              {},
	"playwright-report": {},
	"test-results":      {},
	"temp":              {},
}

// generatedSourcePatterns stores generated Go source patterns reported as
// statistics instead of runtime-source violations.
var generatedSourcePatterns = []string{
	"**/internal/dao/**",
	"**/internal/model/do/**",
	"**/internal/model/entity/**",
}

// testFixturePatterns stores test source patterns reported as statistics
// instead of runtime-source violations.
var testFixturePatterns = []string{
	"**/*_test.go",
	"**/*.test.ts",
	"**/*.spec.ts",
}

// runtimeExcludedPatterns stores repository-relative glob patterns skipped by
// the scanner because they are resource bundles or delivery assets.
var runtimeExcludedPatterns = []string{
	"**/locales/**",
	"**/manifest/i18n/**",
	"**/manifest/sql/**",
}

// commentPrefixes stores line prefixes that mark comments or empty comment tails.
var commentPrefixes = []string{"//", "*", "/*", "<!--", "{/*", "*/"}

// scanOptions stores command-line options for the hard-coded copy scanner.
type scanOptions struct {
	allowlistPath string
	format        string
}

// scanRule identifies one source-code pattern that flags runtime-visible copy.
type scanRule struct {
	Identifier string
	Category   string
	Include    []string
	Pattern    *regexp.Regexp
	Message    string
}

// scanFinding describes one runtime i18n scanner finding.
type scanFinding struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Content  string `json:"content"`
}

// scanAllowlistHit describes one matched allowlist entry.
type scanAllowlistHit struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Category string `json:"category"`
	Reason   string `json:"reason"`
	Scope    string `json:"scope"`
	Content  string `json:"content"`
}

// scanSummary stores aggregate scan counts for review reports and CI output.
type scanSummary struct {
	Violations       int            `json:"violations"`
	ViolationFiles   int            `json:"violationFiles"`
	AllowlistHits    int            `json:"allowlistHits"`
	GeneratedFiles   int            `json:"generatedFiles"`
	GeneratedItems   int            `json:"generatedItems"`
	TestFixtureFiles int            `json:"testFixtureFiles"`
	TestFixtureItems int            `json:"testFixtureItems"`
	ByCategory       map[string]int `json:"byCategory"`
}

// scanReport stores the complete hard-coded copy scan report.
type scanReport struct {
	Summary       scanSummary        `json:"summary"`
	Findings      []scanFinding      `json:"findings"`
	AllowlistHits []scanAllowlistHit `json:"allowlistHits"`
}

// scanSourceResult stores one file scan result before report aggregation.
type scanSourceResult struct {
	Findings      []scanFinding
	AllowlistHits []scanAllowlistHit
}

// allowlistPayload stores the JSON allowlist document.
type allowlistPayload struct {
	Entries []allowlistEntry `json:"entries"`
}

// allowlistEntry stores one allowed scanner finding.
type allowlistEntry struct {
	Path     string `json:"path"`
	Rule     string `json:"rule"`
	Category string `json:"category"`
	Reason   string `json:"reason"`
	Scope    string `json:"scope"`
	Line     *int   `json:"line,omitempty"`
}

// allowlistKey stores the normalized lookup key for one allowed finding.
type allowlistKey struct {
	Path string
	Rule string
	Line int
}

// scanRuntimeI18N scans source files and returns all non-allowlisted findings.
func scanRuntimeI18N(repoRoot string, options scanOptions) ([]scanFinding, error) {
	report, err := scanRuntimeI18NReport(repoRoot, options)
	if err != nil {
		return nil, err
	}
	return report.Findings, nil
}

// scanRuntimeI18NReport scans source files and returns a categorized report.
func scanRuntimeI18NReport(repoRoot string, options scanOptions) (*scanReport, error) {
	allowlist, err := loadAllowlist(options.allowlistPath)
	if err != nil {
		return nil, err
	}

	files, err := iterSourceFiles(repoRoot)
	if err != nil {
		return nil, err
	}

	report := &scanReport{
		Summary: scanSummary{
			ByCategory: make(map[string]int),
		},
		Findings:      make([]scanFinding, 0),
		AllowlistHits: make([]scanAllowlistHit, 0),
	}
	for _, path := range files {
		relPath, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return nil, relErr
		}
		relPath = filepath.ToSlash(relPath)
		if pathMatchesAny(relPath, generatedSourcePatterns) {
			if statErr := collectHanLiteralStats(path, &report.Summary.GeneratedFiles, &report.Summary.GeneratedItems); statErr != nil {
				return nil, statErr
			}
			continue
		}
		if pathMatchesAny(relPath, testFixturePatterns) {
			if statErr := collectHanLiteralStats(path, &report.Summary.TestFixtureFiles, &report.Summary.TestFixtureItems); statErr != nil {
				return nil, statErr
			}
			continue
		}
		if pathMatchesAny(relPath, runtimeExcludedPatterns) {
			continue
		}

		fileResult, scanErr := scanSourceFile(repoRoot, path, allowlist)
		if scanErr != nil {
			return nil, scanErr
		}
		report.Findings = append(report.Findings, fileResult.Findings...)
		report.AllowlistHits = append(report.AllowlistHits, fileResult.AllowlistHits...)
	}
	sort.Slice(report.Findings, func(left int, right int) bool {
		if report.Findings[left].Path != report.Findings[right].Path {
			return report.Findings[left].Path < report.Findings[right].Path
		}
		if report.Findings[left].Line != report.Findings[right].Line {
			return report.Findings[left].Line < report.Findings[right].Line
		}
		return report.Findings[left].Rule < report.Findings[right].Rule
	})
	sort.Slice(report.AllowlistHits, func(left int, right int) bool {
		if report.AllowlistHits[left].Path != report.AllowlistHits[right].Path {
			return report.AllowlistHits[left].Path < report.AllowlistHits[right].Path
		}
		if report.AllowlistHits[left].Line != report.AllowlistHits[right].Line {
			return report.AllowlistHits[left].Line < report.AllowlistHits[right].Line
		}
		return report.AllowlistHits[left].Rule < report.AllowlistHits[right].Rule
	})
	populateScanSummary(report)
	return report, nil
}

// loadAllowlist reads the optional JSON allowlist.
func loadAllowlist(path string) (map[allowlistKey]allowlistEntry, error) {
	result := make(map[allowlistKey]allowlistEntry)
	normalizedPath := strings.TrimSpace(path)
	if normalizedPath == "" {
		return result, nil
	}
	content, err := os.ReadFile(normalizedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("read allowlist %s: %w", normalizedPath, err)
	}

	payload := &allowlistPayload{}
	if err = json.Unmarshal(content, payload); err != nil {
		return nil, fmt.Errorf("invalid allowlist JSON %s: %w", normalizedPath, err)
	}
	for index, entry := range payload.Entries {
		sourcePath := strings.TrimSpace(entry.Path)
		rule := strings.TrimSpace(entry.Rule)
		reason := strings.TrimSpace(entry.Reason)
		category := strings.TrimSpace(entry.Category)
		scope := strings.TrimSpace(entry.Scope)
		if sourcePath == "" || rule == "" || reason == "" || category == "" || scope == "" {
			return nil, fmt.Errorf("invalid allowlist entry #%d: path, rule, category, reason, and scope are required", index+1)
		}
		entry.Path = sourcePath
		entry.Rule = rule
		entry.Category = category
		entry.Reason = reason
		entry.Scope = scope
		line := 0
		if entry.Line != nil {
			line = *entry.Line
		}
		result[allowlistKey{Path: sourcePath, Rule: rule, Line: line}] = entry
	}
	return result, nil
}

// iterSourceFiles returns source files that can contain runtime-visible copy.
func iterSourceFiles(repoRoot string) ([]string, error) {
	roots := []string{
		filepath.Join(repoRoot, "apps", "lina-core"),
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src"),
		filepath.Join(repoRoot, "apps", "lina-plugins"),
	}
	files := make([]string, 0)
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("stat scan root %s: %w", root, err)
		}
		walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if _, ok := excludedPathParts[entry.Name()]; ok {
					return filepath.SkipDir
				}
				return nil
			}
			if !isScannableSourceSuffix(path) {
				return nil
			}
			relPath, relErr := filepath.Rel(repoRoot, path)
			if relErr != nil {
				return relErr
			}
			relPath = filepath.ToSlash(relPath)
			files = append(files, path)
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("scan source root %s: %w", root, walkErr)
		}
	}
	sort.Strings(files)
	return files, nil
}

// isScannableSourceSuffix reports whether one file suffix is included in scanning.
func isScannableSourceSuffix(path string) bool {
	switch filepath.Ext(path) {
	case ".go", ".ts", ".vue":
		return true
	default:
		return false
	}
}

// scanSourceFile scans one file and returns high-risk runtime i18n findings.
func scanSourceFile(repoRoot string, path string, allowlist map[allowlistKey]allowlistEntry) (scanSourceResult, error) {
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return scanSourceResult{}, err
	}
	relPath = filepath.ToSlash(relPath)

	content, err := os.ReadFile(path)
	if err != nil {
		return scanSourceResult{}, fmt.Errorf("read source file %s: %w", relPath, err)
	}

	lines := strings.Split(string(content), "\n")
	result := scanSourceResult{
		Findings:      make([]scanFinding, 0),
		AllowlistHits: make([]scanAllowlistHit, 0),
	}
	for index, line := range lines {
		lineNumber := index + 1
		if !hanPattern.MatchString(line) || isCommentOrI18NUsage(line) {
			continue
		}
		for _, rule := range scannerRules {
			if !pathMatchesAny(relPath, rule.Include) || !rule.Pattern.MatchString(line) {
				continue
			}
			trimmedLine := strings.TrimSpace(line)
			if entry, ok := findAllowlistEntry(allowlist, relPath, rule.Identifier, lineNumber); ok {
				result.AllowlistHits = append(result.AllowlistHits, scanAllowlistHit{
					Path:     relPath,
					Line:     lineNumber,
					Rule:     rule.Identifier,
					Category: rule.Category,
					Reason:   entry.Reason,
					Scope:    entry.Scope,
					Content:  trimmedLine,
				})
				break
			}
			result.Findings = append(result.Findings, scanFinding{
				Path:     relPath,
				Line:     lineNumber,
				Rule:     rule.Identifier,
				Category: rule.Category,
				Message:  rule.Message,
				Content:  trimmedLine,
			})
			break
		}
	}
	return result, nil
}

// isCommentOrI18NUsage skips comments and frontend lines already routed through i18n.
func isCommentOrI18NUsage(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	if strings.Contains(trimmed, "$t(") {
		return true
	}
	return strings.Contains(trimmed, "t(") && !strings.Contains(trimmed, "message.")
}

// findAllowlistEntry reports whether one finding is allowlisted by exact line
// or by a file-wide rule entry.
func findAllowlistEntry(allowlist map[allowlistKey]allowlistEntry, path string, rule string, line int) (allowlistEntry, bool) {
	if entry, ok := allowlist[allowlistKey{Path: path, Rule: rule, Line: line}]; ok {
		return entry, true
	}
	entry, ok := allowlist[allowlistKey{Path: path, Rule: rule}]
	return entry, ok
}

// collectHanLiteralStats counts non-comment Han-containing lines for an
// excluded source category.
func collectHanLiteralStats(path string, fileCount *int, itemCount *int) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read source file %s: %w", path, err)
	}
	items := 0
	for _, line := range strings.Split(string(content), "\n") {
		if hanPattern.MatchString(line) && !isCommentOrI18NUsage(line) {
			items++
		}
	}
	if items == 0 {
		return nil
	}
	(*fileCount)++
	*itemCount += items
	return nil
}

// populateScanSummary derives aggregate counts from detailed report entries.
func populateScanSummary(report *scanReport) {
	report.Summary.Violations = len(report.Findings)
	report.Summary.AllowlistHits = len(report.AllowlistHits)
	violationFiles := make(map[string]struct{})
	for _, finding := range report.Findings {
		violationFiles[finding.Path] = struct{}{}
		report.Summary.ByCategory[finding.Category]++
	}
	report.Summary.ViolationFiles = len(violationFiles)
}

// emitScanFindings writes text or JSON scanner output.
func emitScanFindings(out io.Writer, findings []scanFinding, format string) error {
	report := &scanReport{
		Summary: scanSummary{
			ByCategory: make(map[string]int),
		},
		Findings: findings,
	}
	populateScanSummary(report)
	return emitScanReport(out, report, format)
}

// emitScanReport writes text or JSON scanner output.
func emitScanReport(out io.Writer, report *scanReport, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(out)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	if err := writeLine(out, "Runtime i18n scan summary:"); err != nil {
		return err
	}
	summaryLines := []string{
		fmt.Sprintf("  violations: %d across %d file(s)", report.Summary.Violations, report.Summary.ViolationFiles),
		fmt.Sprintf("  allowlist hits: %d", report.Summary.AllowlistHits),
		fmt.Sprintf("  generated source Han literal lines: %d across %d file(s)", report.Summary.GeneratedItems, report.Summary.GeneratedFiles),
		fmt.Sprintf("  test fixture Han literal lines: %d across %d file(s)", report.Summary.TestFixtureItems, report.Summary.TestFixtureFiles),
	}
	for _, line := range summaryLines {
		if err := writeLine(out, line); err != nil {
			return err
		}
	}
	for _, category := range sortedMapKeys(report.Summary.ByCategory) {
		if err := writeLine(out, fmt.Sprintf("  %s violations: %d", category, report.Summary.ByCategory[category])); err != nil {
			return err
		}
	}
	if len(report.Findings) == 0 {
		return writeLine(out, "Runtime i18n scan passed: no high-risk hardcoded copy found.")
	}
	if err := writeLine(out, fmt.Sprintf("Runtime i18n scan found %d high-risk hardcoded copy item(s):", len(report.Findings))); err != nil {
		return err
	}
	for _, finding := range report.Findings {
		message := fmt.Sprintf(
			"%s:%d: %s [%s] %s\n  %s",
			finding.Path,
			finding.Line,
			finding.Rule,
			finding.Category,
			finding.Message,
			finding.Content,
		)
		if err := writeLine(out, message); err != nil {
			return err
		}
	}
	return nil
}
