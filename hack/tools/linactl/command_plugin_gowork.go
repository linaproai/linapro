// This file generates and parses temporary plugin Go workspaces.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// goWorkUse stores one normalized go.work use path and its raw file value.
type goWorkUse struct {
	Raw        string
	Normalized string
}

// officialPluginGoWorkUses discovers plugin module directories that must be
// present in the temporary plugin workspace for plugin-full builds.
func officialPluginGoWorkUses(root string, workspace officialPluginWorkspace) ([]string, error) {
	if err := requireOfficialPluginWorkspaceState(root, workspace); err != nil {
		return nil, err
	}
	var uses []string
	err := filepath.WalkDir(workspace.Root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() != "go.mod" {
			return nil
		}
		dir := filepath.Dir(path)
		relativePath, relErr := filepath.Rel(root, dir)
		if relErr != nil {
			return relErr
		}
		uses = append(uses, "./"+filepath.ToSlash(relativePath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan official plugin Go modules: %w", err)
	}
	sort.Slice(uses, func(left int, right int) bool {
		leftDepth := strings.Count(uses[left], "/")
		rightDepth := strings.Count(uses[right], "/")
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return uses[left] < uses[right]
	})
	return uses, nil
}

// writeOfficialPluginWorkspace generates the temporary plugin-full workspace
// from the host workspace and the discovered official plugin Go modules.
func writeOfficialPluginWorkspace(root string, workspace officialPluginWorkspace) (string, error) {
	pluginUses, err := officialPluginGoWorkUses(root, workspace)
	if err != nil {
		return "", err
	}
	hostUses, err := readGoWorkUses(root)
	if err != nil {
		return "", err
	}
	version, err := readGoWorkVersion(root)
	if err != nil {
		return "", err
	}

	workspacePath := officialPluginGoWorkPath(root)
	normalizedUses := make([]string, 0, len(hostUses)+len(pluginUses))
	seen := make(map[string]struct{}, len(hostUses)+len(pluginUses))
	addUse := func(use string) {
		normalized := normalizeGoWorkUse(root, use)
		if normalized == "" || isOfficialPluginGoWorkUse(normalized) {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		normalizedUses = append(normalizedUses, normalized)
	}
	for _, use := range hostUses {
		addUse(use.Raw)
	}
	for _, use := range pluginUses {
		normalized := normalizeGoWorkUse(root, use)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		normalizedUses = append(normalizedUses, normalized)
	}

	content, err := renderPluginGoWork(root, workspacePath, version, normalizedUses)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(filepath.Dir(workspacePath), 0o755); err != nil {
		return "", fmt.Errorf("create plugin workspace directory: %w", err)
	}
	if err = os.WriteFile(workspacePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write plugin workspace: %w", err)
	}
	return workspacePath, nil
}

// officialPluginGoWorkPath returns the ignored temporary workspace path used by
// plugin-full builds.
func officialPluginGoWorkPath(root string) string {
	return filepath.Join(root, "temp", officialPluginWorkspaceFile)
}

// readGoWorkVersion reads the Go version directive from the root workspace.
func readGoWorkVersion(root string) (string, error) {
	path := filepath.Join(root, "go.work")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read go.work: %w", err)
	}
	version := parseGoWorkVersion(string(content))
	if version == "" {
		return "", fmt.Errorf("go.work is missing a go version directive")
	}
	return version, nil
}

// parseGoWorkVersion extracts the Go version directive from go.work content.
func parseGoWorkVersion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = stripGoWorkLineComment(line)
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

// renderPluginGoWork converts repository-relative module paths to paths
// relative to the temporary workspace file.
func renderPluginGoWork(root string, workspacePath string, version string, normalizedUses []string) (string, error) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "go %s\n", version)
	if len(normalizedUses) == 0 {
		builder.WriteString("\n")
		return builder.String(), nil
	}
	builder.WriteString("\nuse (\n")
	for _, normalized := range normalizedUses {
		usePath, err := renderGoWorkUsePath(root, workspacePath, normalized)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&builder, "\t%s\n", usePath)
	}
	builder.WriteString(")\n")
	return builder.String(), nil
}

// renderGoWorkUsePath renders one module path relative to workspacePath.
func renderGoWorkUsePath(root string, workspacePath string, normalized string) (string, error) {
	modulePath := filepath.FromSlash(normalized)
	if !filepath.IsAbs(modulePath) {
		modulePath = filepath.Join(root, modulePath)
	}
	relativePath, err := filepath.Rel(filepath.Dir(workspacePath), modulePath)
	if err != nil {
		return "", fmt.Errorf("render go.work use path for %s: %w", normalized, err)
	}
	relativePath = filepath.ToSlash(relativePath)
	if !strings.HasPrefix(relativePath, ".") {
		relativePath = "./" + relativePath
	}
	if strings.ContainsAny(relativePath, " \t\"") {
		return strconv.Quote(relativePath), nil
	}
	return relativePath, nil
}

// readGoWorkUses reads go.work use entries without requiring every listed
// module path to exist.
func readGoWorkUses(root string) ([]goWorkUse, error) {
	path := filepath.Join(root, "go.work")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read go.work: %w", err)
	}
	return parseGoWorkUses(root, string(content)), nil
}

// parseGoWorkUses extracts use entries from a go.work file.
func parseGoWorkUses(root string, content string) []goWorkUse {
	var (
		uses       []goWorkUse
		inUseBlock bool
	)
	for _, line := range strings.Split(content, "\n") {
		line = stripGoWorkLineComment(line)
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "use (") {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if trimmed == ")" {
				inUseBlock = false
				continue
			}
			if use := firstGoWorkField(trimmed); use != "" {
				uses = append(uses, goWorkUse{Raw: use, Normalized: normalizeGoWorkUse(root, use)})
			}
			continue
		}
		if strings.HasPrefix(trimmed, "use ") {
			if use := firstGoWorkField(strings.TrimSpace(strings.TrimPrefix(trimmed, "use"))); use != "" && use != "(" {
				uses = append(uses, goWorkUse{Raw: use, Normalized: normalizeGoWorkUse(root, use)})
			}
		}
	}
	return uses
}

// stripGoWorkLineComment removes a line comment from simple go.work syntax.
func stripGoWorkLineComment(line string) string {
	if index := strings.Index(line, "//"); index >= 0 {
		return line[:index]
	}
	return line
}

// firstGoWorkField returns the first path-like token from one go.work line.
func firstGoWorkField(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	field := fields[0]
	if unquoted, err := strconv.Unquote(field); err == nil {
		return unquoted
	}
	return field
}

// normalizeGoWorkUse maps go.work use paths to repository-relative slash paths.
func normalizeGoWorkUse(root string, use string) string {
	path := use
	if filepath.IsAbs(path) {
		if relativePath, err := filepath.Rel(root, path); err == nil {
			path = relativePath
		}
	}
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "./")
	return path
}

// isOfficialPluginGoWorkUse reports whether a normalized use path points at the
// official plugin workspace.
func isOfficialPluginGoWorkUse(normalized string) bool {
	return normalized == "apps/lina-plugins" || strings.HasPrefix(normalized, "apps/lina-plugins/")
}
