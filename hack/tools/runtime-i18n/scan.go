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

// scannerRules stores every high-risk runtime i18n source-code pattern.
var scannerRules = []scanRule{
	{
		Identifier: "go-error-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`gerror\.(?:New|Newf|Wrap|Wrapf)\([^)\n]*"[^"]*[\x{3400}-\x{9fff}]`),
		Message: "Go errors visible to users must use structured runtime message errors.",
	},
	{
		Identifier: "go-message-field-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`\b(?:Reason|Message|Fallback)\s*:\s*"[^"]*[\x{3400}-\x{9fff}]`),
		Message: "Reason, Message, and Fallback fields must use message keys or rendered i18n text.",
	},
	{
		Identifier: "go-message-assignment-han",
		Category:   "UserMessage",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`\b[A-Za-z0-9_]+\.Message\s*=\s*(?:fmt\.Sprintf\()?"[^"]*[\x{3400}-\x{9fff}]`),
		Message: "Result messages must carry messageKey/messageParams instead of fixed-language text.",
	},
	{
		Identifier: "go-artifact-slice-han",
		Category:   "UserArtifact",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`\[\]string\s*\{[^}\n]*"[^"]*[\x{3400}-\x{9fff}]`),
		Message: "Export headers and template labels must be rendered through runtime i18n.",
	},
	{
		Identifier: "go-status-text-han",
		Category:   "UserArtifact",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`\b[A-Za-z0-9_]*Text\s*(?::=|=)\s*"[^"]*[\x{3400}-\x{9fff}]`),
		Message: "Status/display text variables must use translated enum labels.",
	},
	{
		Identifier: "go-hostcall-error-han",
		Category:   "DeveloperDiagnostic",
		Include: []string{
			"apps/lina-core/**/*.go",
			"apps/lina-plugins/**/*.go",
		},
		Pattern: regexp.MustCompile(`NewHostCallErrorResponse\([^)\n]*"[^"]*[\x{3400}-\x{9fff}]`),
		Message: "Plugin bridge errors must expose stable codes and English developer diagnostics.",
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

// excludedPatterns stores repository-relative glob patterns skipped by the scanner.
var excludedPatterns = []string{
	"**/*_test.go",
	"**/*.test.ts",
	"**/*.spec.ts",
	"**/internal/model/**",
	"**/internal/dao/**",
	"**/model/do/**",
	"**/model/entity/**",
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
	allowlist, err := loadAllowlist(options.allowlistPath)
	if err != nil {
		return nil, err
	}

	files, err := iterSourceFiles(repoRoot)
	if err != nil {
		return nil, err
	}

	findings := make([]scanFinding, 0)
	for _, path := range files {
		fileFindings, scanErr := scanSourceFile(repoRoot, path, allowlist)
		if scanErr != nil {
			return nil, scanErr
		}
		findings = append(findings, fileFindings...)
	}
	sort.Slice(findings, func(left int, right int) bool {
		if findings[left].Path != findings[right].Path {
			return findings[left].Path < findings[right].Path
		}
		if findings[left].Line != findings[right].Line {
			return findings[left].Line < findings[right].Line
		}
		return findings[left].Rule < findings[right].Rule
	})
	return findings, nil
}

// loadAllowlist reads the optional JSON allowlist.
func loadAllowlist(path string) (map[allowlistKey]struct{}, error) {
	result := make(map[allowlistKey]struct{})
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
		if sourcePath == "" || rule == "" || reason == "" || category == "" {
			return nil, fmt.Errorf("invalid allowlist entry #%d: path, rule, category, and reason are required", index+1)
		}
		line := 0
		if entry.Line != nil {
			line = *entry.Line
		}
		result[allowlistKey{Path: sourcePath, Rule: rule, Line: line}] = struct{}{}
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
			if pathMatchesAny(relPath, excludedPatterns) {
				return nil
			}
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
func scanSourceFile(repoRoot string, path string, allowlist map[allowlistKey]struct{}) ([]scanFinding, error) {
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return nil, err
	}
	relPath = filepath.ToSlash(relPath)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read source file %s: %w", relPath, err)
	}

	lines := strings.Split(string(content), "\n")
	findings := make([]scanFinding, 0)
	for index, line := range lines {
		lineNumber := index + 1
		if !hanPattern.MatchString(line) || isCommentOrI18NUsage(line) {
			continue
		}
		for _, rule := range scannerRules {
			if !pathMatchesAny(relPath, rule.Include) || !rule.Pattern.MatchString(line) {
				continue
			}
			if isAllowlisted(allowlist, relPath, rule.Identifier, lineNumber) {
				continue
			}
			findings = append(findings, scanFinding{
				Path:     relPath,
				Line:     lineNumber,
				Rule:     rule.Identifier,
				Category: rule.Category,
				Message:  rule.Message,
				Content:  strings.TrimSpace(line),
			})
		}
	}
	return findings, nil
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

// isAllowlisted reports whether one finding is allowlisted by exact or file-wide line.
func isAllowlisted(allowlist map[allowlistKey]struct{}, path string, rule string, line int) bool {
	if _, ok := allowlist[allowlistKey{Path: path, Rule: rule, Line: line}]; ok {
		return true
	}
	_, ok := allowlist[allowlistKey{Path: path, Rule: rule}]
	return ok
}

// emitScanFindings writes text or JSON scanner output.
func emitScanFindings(out io.Writer, findings []scanFinding, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(out)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(findings)
	}
	if len(findings) == 0 {
		return writeLine(out, "Runtime i18n scan passed: no high-risk hardcoded copy found.")
	}
	if err := writeLine(out, fmt.Sprintf("Runtime i18n scan found %d high-risk hardcoded copy item(s):", len(findings))); err != nil {
		return err
	}
	for _, finding := range findings {
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
