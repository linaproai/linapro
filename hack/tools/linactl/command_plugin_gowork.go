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

// officialPluginBackendImport stores one source-plugin backend import path for
// the generated aggregate module.
type officialPluginBackendImport struct {
	PluginID string
	Module   string
	Import   string
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
	aggregateUse, err := writeOfficialPluginAggregateModule(root, workspace)
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
	if aggregateUse != "" {
		normalized := normalizeGoWorkUse(root, aggregateUse)
		if normalized != "" {
			if _, ok := seen[normalized]; !ok {
				seen[normalized] = struct{}{}
				normalizedUses = append(normalizedUses, normalized)
			}
		}
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

// writeOfficialPluginAggregateModule resolves the module that satisfies the
// host's `import _ "lina-plugins"` bridge. Official plugin workspaces can
// provide this aggregate module at their root; older local fixtures without a
// root module still receive an ignored generated fallback module.
func writeOfficialPluginAggregateModule(root string, workspace officialPluginWorkspace) (string, error) {
	existingUse, err := existingOfficialPluginAggregateModule(root, workspace)
	if err != nil {
		return "", err
	}
	if existingUse != "" {
		if err = os.RemoveAll(officialPluginAggregateModuleDir(root)); err != nil {
			return "", fmt.Errorf("clean stale official plugin aggregate module: %w", err)
		}
		return existingUse, nil
	}

	imports, err := officialPluginBackendImports(workspace)
	if err != nil {
		return "", err
	}
	moduleDir := officialPluginAggregateModuleDir(root)
	if err = os.RemoveAll(moduleDir); err != nil {
		return "", fmt.Errorf("clean official plugin aggregate module: %w", err)
	}
	if err = os.MkdirAll(moduleDir, 0o755); err != nil {
		return "", fmt.Errorf("create official plugin aggregate module: %w", err)
	}
	if err = os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(renderOfficialPluginAggregateGoMod()), 0o644); err != nil {
		return "", fmt.Errorf("write official plugin aggregate go.mod: %w", err)
	}
	if err = os.WriteFile(filepath.Join(moduleDir, "plugins.go"), []byte(renderOfficialPluginAggregateGo(imports)), 0o644); err != nil {
		return "", fmt.Errorf("write official plugin aggregate imports: %w", err)
	}
	relativePath, err := filepath.Rel(root, moduleDir)
	if err != nil {
		return "", fmt.Errorf("resolve official plugin aggregate module path: %w", err)
	}
	return "./" + filepath.ToSlash(relativePath), nil
}

// existingOfficialPluginAggregateModule returns the official plugin root module
// when it already owns the host bridge import path.
func existingOfficialPluginAggregateModule(root string, workspace officialPluginWorkspace) (string, error) {
	goModPath := filepath.Join(workspace.Root, "go.mod")
	if !fileExists(goModPath) {
		return "", nil
	}
	moduleName, err := readGoModuleName(goModPath)
	if err != nil {
		return "", err
	}
	if moduleName != officialPluginAggregateModuleName {
		return "", nil
	}
	relativePath, err := filepath.Rel(root, workspace.Root)
	if err != nil {
		return "", fmt.Errorf("resolve existing official plugin aggregate module path: %w", err)
	}
	return "./" + filepath.ToSlash(relativePath), nil
}

// officialPluginBackendImports discovers source-plugin backend packages that
// must be imported by the generated aggregate module.
func officialPluginBackendImports(workspace officialPluginWorkspace) ([]officialPluginBackendImport, error) {
	var imports []officialPluginBackendImport
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
		moduleDir := filepath.Dir(path)
		manifestPath := filepath.Join(moduleDir, "plugin.yaml")
		backendPath := filepath.Join(moduleDir, "backend", "plugin.go")
		if !fileExists(manifestPath) || !fileExists(backendPath) {
			return nil
		}
		manifest, err := readPluginManifest(manifestPath)
		if err != nil {
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(manifest.Type), "source") {
			return nil
		}
		moduleName, err := readGoModuleName(path)
		if err != nil {
			return err
		}
		pluginID := filepath.Base(moduleDir)
		imports = append(imports, officialPluginBackendImport{
			PluginID: pluginID,
			Module:   moduleName,
			Import:   moduleName + "/backend",
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover official source-plugin backend imports: %w", err)
	}
	sort.Slice(imports, func(left int, right int) bool {
		return imports[left].PluginID < imports[right].PluginID
	})
	return imports, nil
}

// readGoModuleName reads the module directive from a go.mod file.
func readGoModuleName(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = stripGoWorkLineComment(line)
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("%s is missing a module directive", path)
}

// renderOfficialPluginAggregateGoMod renders the generated aggregate go.mod.
func renderOfficialPluginAggregateGoMod() string {
	return fmt.Sprintf("module %s\n\ngo 1.25.0\n", officialPluginAggregateModuleName)
}

// renderOfficialPluginAggregateGo renders the generated aggregate import file.
func renderOfficialPluginAggregateGo(imports []officialPluginBackendImport) string {
	var builder strings.Builder
	builder.WriteString("// This file is generated by linactl for plugin-full builds.\n\n")
	builder.WriteString("package linaplugins\n")
	if len(imports) == 0 {
		return builder.String()
	}
	builder.WriteString("\nimport (\n")
	for _, item := range imports {
		fmt.Fprintf(&builder, "\t_ %q\n", item.Import)
	}
	builder.WriteString(")\n")
	return builder.String()
}

// officialPluginAggregateModuleDir returns the generated module directory used
// by plugin-full builds.
func officialPluginAggregateModuleDir(root string) string {
	return filepath.Join(root, "temp", officialPluginAggregateDir)
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
