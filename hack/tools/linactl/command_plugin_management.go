// This file manages user-project source-plugin workspaces and configured
// plugin installation sources.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// managedPluginRootRelativePath is the fixed source-plugin workspace path.
	managedPluginRootRelativePath = "apps/lina-plugins"
	// pluginLockFileName stores tool-generated source-plugin state.
	pluginLockFileName = ".linapro-plugins.lock.yaml"
	// pluginWildcardItem expands to every plugin directory under the source root.
	pluginWildcardItem = "*"
)

var (
	// pluginIDPattern matches the repository plugin ID convention.
	pluginIDPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	// pluginSourceNamePattern matches safe source names for diagnostics.
	pluginSourceNamePattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9_.-]*[A-Za-z0-9])?$`)
)

// managedPluginWorkspaceState classifies the user-project plugin workspace.
type managedPluginWorkspaceState string

const (
	// managedPluginWorkspaceMissing means apps/lina-plugins does not exist.
	managedPluginWorkspaceMissing managedPluginWorkspaceState = "missing"
	// managedPluginWorkspaceOrdinary means apps/lina-plugins is a normal directory.
	managedPluginWorkspaceOrdinary managedPluginWorkspaceState = "ordinary"
	// managedPluginWorkspaceSubmodule means apps/lina-plugins is tracked as a gitlink.
	managedPluginWorkspaceSubmodule managedPluginWorkspaceState = "submodule"
	// managedPluginWorkspaceNestedGit means apps/lina-plugins contains Git metadata.
	managedPluginWorkspaceNestedGit managedPluginWorkspaceState = "nested-git"
	// managedPluginWorkspaceInvalid means apps/lina-plugins exists but is not a directory.
	managedPluginWorkspaceInvalid managedPluginWorkspaceState = "invalid"
)

// managedPluginWorkspace describes apps/lina-plugins for plugin management.
type managedPluginWorkspace struct {
	Root  string
	State managedPluginWorkspaceState
}

// pluginPlan stores validated plugin source configuration for one command.
type pluginPlan struct {
	Items  []pluginPlanItem
	Filter map[string]struct{}
}

// pluginPlanItem stores one configured plugin operation target.
type pluginPlanItem struct {
	ID     string
	Source string
	Repo   string
	Root   string
	Ref    string
	All    bool
}

// pluginSourceCheckout stores one temporary checkout and resolved commit.
type pluginSourceCheckout struct {
	Dir    string
	Commit string
}

// pluginLockFile stores tool-generated source-plugin installation state.
type pluginLockFile struct {
	Plugins []pluginLockEntry `yaml:"plugins"`
}

// pluginLockEntry stores one installed plugin's source and content state.
type pluginLockEntry struct {
	ID             string `yaml:"id"`
	Source         string `yaml:"source"`
	Repo           string `yaml:"repo"`
	Root           string `yaml:"root"`
	Ref            string `yaml:"ref"`
	ResolvedCommit string `yaml:"resolvedCommit"`
	Version        string `yaml:"version"`
	ContentHash    string `yaml:"contentHash"`
}

// runPluginsInit converts apps/lina-plugins from a submodule into an ordinary
// source-plugin directory while preserving files.
func runPluginsInit(ctx context.Context, a *app, _ commandInput) error {
	workspace, err := inspectManagedPluginWorkspace(ctx, a)
	if err != nil {
		return err
	}
	switch workspace.State {
	case managedPluginWorkspaceOrdinary:
		fmt.Fprintf(a.stdout, "Plugin workspace already ordinary: %s\n", relativePath(a.root, workspace.Root))
		return nil
	case managedPluginWorkspaceMissing:
		if err = os.MkdirAll(workspace.Root, 0o755); err != nil {
			return fmt.Errorf("create plugin workspace: %w", err)
		}
		fmt.Fprintf(a.stdout, "Plugin workspace created: %s\n", relativePath(a.root, workspace.Root))
		return nil
	case managedPluginWorkspaceInvalid:
		return fmt.Errorf("plugin workspace is invalid: %s", relativePath(a.root, workspace.Root))
	case managedPluginWorkspaceNestedGit:
		return fmt.Errorf("plugin workspace contains nested Git metadata; remove %s manually before plugins.init", relativePath(a.root, filepath.Join(workspace.Root, ".git")))
	case managedPluginWorkspaceSubmodule:
		return convertPluginSubmoduleToDirectory(ctx, a, workspace)
	default:
		return fmt.Errorf("unknown plugin workspace state: %s", workspace.State)
	}
}

// runPluginsInstall installs configured plugins into apps/lina-plugins.
func runPluginsInstall(ctx context.Context, a *app, input commandInput) error {
	return runPluginInstallOrUpdate(ctx, a, input, false)
}

// runPluginsUpdate updates configured plugins in apps/lina-plugins.
func runPluginsUpdate(ctx context.Context, a *app, input commandInput) error {
	return runPluginInstallOrUpdate(ctx, a, input, true)
}

// runPluginsStatus prints read-only plugin workspace and source status.
func runPluginsStatus(ctx context.Context, a *app, input commandInput) error {
	workspace, err := inspectManagedPluginWorkspace(ctx, a)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.stdout, "Plugin workspace: %s (%s)\n", relativePath(a.root, workspace.Root), workspace.State)
	if workspace.State == managedPluginWorkspaceSubmodule {
		fmt.Fprintln(a.stdout, "Action required: run `make plugins.init` before installing or updating user-project plugins.")
		return nil
	}

	plan, err := loadPluginPlan(a.root, input)
	if err != nil {
		return err
	}
	lock, err := readPluginLock(a.root)
	if err != nil {
		return err
	}
	lockByID := lock.entriesByID()

	checkouts, sourceErrors := checkoutPluginSources(ctx, a, plan.Items)
	defer cleanupPluginSourceCheckouts(a, checkouts)
	expandedPlan, expandErr := expandPluginPlanFromCheckouts(plan, checkouts)
	if expandErr != nil {
		return expandErr
	}

	fmt.Fprintln(a.stdout, "Configured plugins:")
	for _, item := range expandedPlan.Items {
		target := managedPluginPath(a.root, item.ID)
		exists := dirExists(target)
		version := "-"
		contentHash := ""
		if exists {
			if manifest, readErr := readPluginManifest(filepath.Join(target, "plugin.yaml")); readErr == nil {
				version = firstNonEmpty(manifest.Version, "-")
			}
			if hash, hashErr := hashDirectory(target); hashErr == nil {
				contentHash = hash
			}
		}

		dirty := "unknown"
		if exists {
			changed, dirtyErr := pluginHasLocalChanges(ctx, a, item.ID, lockByID[item.ID])
			if dirtyErr == nil {
				dirty = fmt.Sprintf("%t", changed)
			}
		} else {
			dirty = "false"
		}

		remote := "unknown"
		if sourceErr, ok := sourceErrors[item.Source]; ok {
			remote = "unknown"
			fmt.Fprintf(a.stdout, "  - %s source=%s version=%s installed=%t dirty=%s remote=%s error=%v\n", item.ID, item.Source, version, exists, dirty, remote, sourceErr)
			continue
		}
		if checkout, ok := checkouts[item.Source]; ok {
			sourceDir := filepath.Join(checkout.Dir, filepath.FromSlash(sourcePluginRelativePath(item)))
			if !fileExists(filepath.Join(sourceDir, "plugin.yaml")) {
				remote = "missing"
			} else if remoteHash, hashErr := hashDirectory(sourceDir); hashErr == nil {
				remote = remoteStatus(lockByID[item.ID], contentHash, remoteHash)
			}
		}
		fmt.Fprintf(a.stdout, "  - %s source=%s version=%s installed=%t dirty=%s remote=%s\n", item.ID, item.Source, version, exists, dirty, remote)
	}

	if err = printUnconfiguredPluginStatus(a, expandedPlan, lockByID); err != nil {
		return err
	}
	return nil
}

// runPluginInstallOrUpdate executes install or update for selected plugins.
func runPluginInstallOrUpdate(ctx context.Context, a *app, input commandInput, update bool) error {
	workspace, err := inspectManagedPluginWorkspace(ctx, a)
	if err != nil {
		return err
	}
	if workspace.State == managedPluginWorkspaceSubmodule {
		return errors.New("apps/lina-plugins is still a submodule; run `make plugins.init` first")
	}
	if workspace.State == managedPluginWorkspaceNestedGit {
		return errors.New("apps/lina-plugins contains nested Git metadata; run `make plugins.init` or remove nested metadata first")
	}
	if workspace.State == managedPluginWorkspaceInvalid {
		return fmt.Errorf("apps/lina-plugins is invalid: %s", workspace.Root)
	}
	if err = os.MkdirAll(workspace.Root, 0o755); err != nil {
		return fmt.Errorf("create plugin workspace: %w", err)
	}

	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}
	plan, err := loadPluginPlan(a.root, input)
	if err != nil {
		return err
	}
	lock, err := readPluginLock(a.root)
	if err != nil {
		return err
	}

	checkouts, sourceErrors := checkoutPluginSources(ctx, a, plan.Items)
	defer cleanupPluginSourceCheckouts(a, checkouts)
	if len(sourceErrors) > 0 {
		return firstPluginSourceError(sourceErrors)
	}
	plan, err = expandPluginPlanFromCheckouts(plan, checkouts)
	if err != nil {
		return err
	}
	for _, item := range plan.Items {
		checkout := checkouts[item.Source]
		if err = applyPluginFromCheckout(ctx, a, item, checkout, &lock, update, force); err != nil {
			return err
		}
	}
	if err = writePluginLock(a.root, lock); err != nil {
		return err
	}
	return nil
}

// loadPluginPlan loads and validates plugin source configuration.
func loadPluginPlan(root string, input commandInput) (pluginPlan, error) {
	cfg, err := loadRootConfig(root, input)
	if err != nil {
		return pluginPlan{}, err
	}
	return validatePluginConfig(cfg.Plugins, input)
}

// validatePluginConfig normalizes configured plugin sources and applies
// command-level source or plugin filters.
func validatePluginConfig(cfg pluginsConfig, input commandInput) (pluginPlan, error) {
	if len(cfg.Sources) == 0 {
		return pluginPlan{}, errors.New("plugins.sources is empty in hack/config.yaml")
	}
	sourceFilter := strings.TrimSpace(input.Get("source"))
	pluginFilter := splitPluginFilter(input.Get("p"))
	sourceNames := make([]string, 0, len(cfg.Sources))
	for name := range cfg.Sources {
		sourceNames = append(sourceNames, name)
	}
	sort.Strings(sourceNames)

	seen := map[string]string{}
	hasWildcard := false
	var items []pluginPlanItem
	for _, sourceName := range sourceNames {
		source := cfg.Sources[sourceName]
		if err := validatePluginSourceName(sourceName); err != nil {
			return pluginPlan{}, err
		}
		if sourceFilter != "" && sourceName != sourceFilter {
			continue
		}
		root, err := validatePluginSourceConfig(sourceName, source)
		if err != nil {
			return pluginPlan{}, err
		}
		for _, id := range source.Items {
			id = strings.TrimSpace(id)
			all := id == pluginWildcardItem
			if all && len(source.Items) > 1 {
				return pluginPlan{}, fmt.Errorf("source %s cannot mix wildcard %q with explicit plugin ids", sourceName, pluginWildcardItem)
			}
			if all {
				hasWildcard = true
			}
			if !all {
				if err = validatePluginID(id); err != nil {
					return pluginPlan{}, fmt.Errorf("source %s has invalid plugin id %q: %w", sourceName, id, err)
				}
			}
			if previous, ok := seen[id]; ok {
				return pluginPlan{}, fmt.Errorf("plugin %q is declared by multiple sources: %s, %s", id, previous, sourceName)
			}
			if !all {
				seen[id] = sourceName
			}
			if len(pluginFilter) > 0 {
				if all {
					// Wildcard expansion happens after checkout, where the plugin
					// filter can be applied to discovered plugin IDs.
				} else if _, ok := pluginFilter[id]; !ok {
					continue
				}
			}
			items = append(items, pluginPlanItem{
				ID:     id,
				Source: sourceName,
				Repo:   strings.TrimSpace(source.Repo),
				Root:   root,
				Ref:    strings.TrimSpace(source.Ref),
				All:    all,
			})
		}
	}
	if sourceFilter != "" {
		if _, ok := cfg.Sources[sourceFilter]; !ok {
			return pluginPlan{}, fmt.Errorf("plugin source %q is not configured", sourceFilter)
		}
	}
	if len(pluginFilter) > 0 {
		if !hasWildcard {
			for requested := range pluginFilter {
				if _, ok := seen[requested]; !ok {
					return pluginPlan{}, fmt.Errorf("plugin %q is not configured", requested)
				}
			}
		}
	}
	if len(items) == 0 {
		return pluginPlan{}, errors.New("no configured plugins match the requested filters")
	}
	return pluginPlan{Items: items, Filter: pluginFilter}, nil
}

// checkoutPluginSources checks out each configured source once.
func checkoutPluginSources(ctx context.Context, a *app, items []pluginPlanItem) (map[string]pluginSourceCheckout, map[string]error) {
	checkouts := map[string]pluginSourceCheckout{}
	sourceErrors := map[string]error{}
	for _, item := range items {
		if _, ok := checkouts[item.Source]; ok {
			continue
		}
		if _, ok := sourceErrors[item.Source]; ok {
			continue
		}
		checkout, err := checkoutPluginSource(ctx, a, item)
		if err != nil {
			sourceErrors[item.Source] = err
			continue
		}
		checkouts[item.Source] = checkout
	}
	return checkouts, sourceErrors
}

// cleanupPluginSourceCheckouts removes temporary source checkout directories.
func cleanupPluginSourceCheckouts(a *app, checkouts map[string]pluginSourceCheckout) {
	for _, checkout := range checkouts {
		if cleanupErr := os.RemoveAll(checkout.Dir); cleanupErr != nil {
			fmt.Fprintf(a.stderr, "warning: remove temporary plugin checkout %s: %v\n", checkout.Dir, cleanupErr)
		}
	}
}

// firstPluginSourceError returns the first deterministic source checkout error.
func firstPluginSourceError(sourceErrors map[string]error) error {
	names := make([]string, 0, len(sourceErrors))
	for name := range sourceErrors {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil
	}
	return fmt.Errorf("checkout plugin source %s: %w", names[0], sourceErrors[names[0]])
}

// expandPluginPlanFromCheckouts expands wildcard source items into concrete plugins.
func expandPluginPlanFromCheckouts(plan pluginPlan, checkouts map[string]pluginSourceCheckout) (pluginPlan, error) {
	var expanded []pluginPlanItem
	seen := map[string]string{}
	for _, item := range plan.Items {
		if !item.All {
			if previous, ok := seen[item.ID]; ok {
				return pluginPlan{}, fmt.Errorf("plugin %q is declared by multiple sources: %s, %s", item.ID, previous, item.Source)
			}
			seen[item.ID] = item.Source
			expanded = append(expanded, item)
			continue
		}
		checkout, ok := checkouts[item.Source]
		if !ok {
			continue
		}
		discovered, err := discoverSourcePlugins(checkout.Dir, item.Root)
		if err != nil {
			return pluginPlan{}, fmt.Errorf("discover plugins for source %s: %w", item.Source, err)
		}
		for _, pluginID := range discovered {
			if len(plan.Filter) > 0 {
				if _, ok := plan.Filter[pluginID]; !ok {
					continue
				}
			}
			if previous, ok := seen[pluginID]; ok {
				return pluginPlan{}, fmt.Errorf("plugin %q is declared by multiple sources: %s, %s", pluginID, previous, item.Source)
			}
			seen[pluginID] = item.Source
			next := item
			next.ID = pluginID
			next.All = false
			expanded = append(expanded, next)
		}
	}
	if len(expanded) == 0 {
		return pluginPlan{}, errors.New("no configured plugins match the requested filters")
	}
	sort.SliceStable(expanded, func(left int, right int) bool {
		if expanded[left].Source != expanded[right].Source {
			return expanded[left].Source < expanded[right].Source
		}
		return expanded[left].ID < expanded[right].ID
	})
	return pluginPlan{Items: expanded, Filter: plan.Filter}, nil
}

// discoverSourcePlugins lists all plugin directories directly under a source root.
func discoverSourcePlugins(checkoutDir string, sourceRoot string) ([]string, error) {
	root := filepath.Join(checkoutDir, filepath.FromSlash(sourceRoot))
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read source root %s: %w", sourceRoot, err)
	}
	var plugins []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if !fileExists(filepath.Join(root, id, "plugin.yaml")) {
			continue
		}
		if err = validatePluginID(id); err != nil {
			return nil, fmt.Errorf("invalid plugin directory %q: %w", id, err)
		}
		plugins = append(plugins, id)
	}
	sort.Strings(plugins)
	return plugins, nil
}

// validatePluginSourceConfig checks required source fields and returns a safe root.
func validatePluginSourceConfig(sourceName string, source pluginSourceConfig) (string, error) {
	if strings.TrimSpace(source.Repo) == "" {
		return "", fmt.Errorf("plugins.sources.%s.repo is required", sourceName)
	}
	root, err := validatePluginSourceRoot(source.Root)
	if err != nil {
		return "", fmt.Errorf("plugins.sources.%s.root is invalid: %w", sourceName, err)
	}
	if strings.TrimSpace(source.Ref) == "" {
		return "", fmt.Errorf("plugins.sources.%s.ref is required", sourceName)
	}
	if len(source.Items) == 0 {
		return "", fmt.Errorf("plugins.sources.%s.items must contain at least one plugin id", sourceName)
	}
	return root, nil
}

// validatePluginSourceName checks that a source name is safe for diagnostics.
func validatePluginSourceName(name string) error {
	if !pluginSourceNamePattern.MatchString(strings.TrimSpace(name)) {
		return fmt.Errorf("invalid plugin source name %q", name)
	}
	return nil
}

// validatePluginID checks that a plugin ID cannot escape the plugin root.
func validatePluginID(id string) error {
	if !pluginIDPattern.MatchString(strings.TrimSpace(id)) {
		return errors.New("expected kebab-case lowercase letters, numbers, and hyphens")
	}
	return nil
}

// validatePluginSourceRoot validates a repository-internal source root.
func validatePluginSourceRoot(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("root is required")
	}
	if strings.Contains(root, "\\") {
		return "", errors.New("backslash path separators are not allowed")
	}
	if strings.Contains(root, ":") {
		return "", errors.New("drive paths and colon characters are not allowed")
	}
	if filepath.IsAbs(root) || filepath.VolumeName(root) != "" || path.IsAbs(root) {
		return "", errors.New("absolute paths are not allowed")
	}
	segments := strings.Split(root, "/")
	for _, segment := range segments {
		if segment == ".." {
			return "", errors.New("parent directory segments are not allowed")
		}
	}
	cleaned := path.Clean(root)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("root escapes the repository")
	}
	if cleaned == "" {
		return "", errors.New("root is required")
	}
	return cleaned, nil
}

// splitPluginFilter parses a comma-separated plugin filter.
func splitPluginFilter(value string) map[string]struct{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	result := map[string]struct{}{}
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result[item] = struct{}{}
		}
	}
	return result
}

// inspectManagedPluginWorkspace classifies apps/lina-plugins for management.
func inspectManagedPluginWorkspace(ctx context.Context, a *app) (managedPluginWorkspace, error) {
	root := managedPluginRoot(a.root)
	if isGitlink(ctx, a, managedPluginRootRelativePath) {
		return managedPluginWorkspace{Root: root, State: managedPluginWorkspaceSubmodule}, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return managedPluginWorkspace{Root: root, State: managedPluginWorkspaceMissing}, nil
		}
		return managedPluginWorkspace{}, fmt.Errorf("stat plugin workspace: %w", err)
	}
	if !info.IsDir() {
		return managedPluginWorkspace{Root: root, State: managedPluginWorkspaceInvalid}, nil
	}
	if fileExists(filepath.Join(root, ".git")) || dirExists(filepath.Join(root, ".git")) {
		return managedPluginWorkspace{Root: root, State: managedPluginWorkspaceNestedGit}, nil
	}
	return managedPluginWorkspace{Root: root, State: managedPluginWorkspaceOrdinary}, nil
}

// convertPluginSubmoduleToDirectory removes submodule metadata while keeping files.
func convertPluginSubmoduleToDirectory(ctx context.Context, a *app, workspace managedPluginWorkspace) error {
	if err := os.MkdirAll(workspace.Root, 0o755); err != nil {
		return fmt.Errorf("create plugin workspace: %w", err)
	}
	if err := removeGitSubmoduleSection(filepath.Join(a.root, ".gitmodules"), managedPluginRootRelativePath); err != nil {
		return err
	}
	if err := removeGitSubmoduleSection(filepath.Join(a.root, ".git", "config"), managedPluginRootRelativePath); err != nil {
		return err
	}
	if err := removeDirectoryIfExists(filepath.Join(a.root, ".git", "modules", "apps", "lina-plugins")); err != nil {
		return err
	}
	pluginGitPath := filepath.Join(workspace.Root, ".git")
	if fileExists(pluginGitPath) || dirExists(pluginGitPath) {
		if err := os.RemoveAll(pluginGitPath); err != nil {
			return fmt.Errorf("remove plugin git metadata: %w", err)
		}
	}
	if err := a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "update-index", "--force-remove", "--", managedPluginRootRelativePath); err != nil {
		return err
	}
	fmt.Fprintf(a.stdout, "Plugin workspace converted to ordinary directory: %s\n", relativePath(a.root, workspace.Root))
	return nil
}

// removeGitSubmoduleSection removes the target submodule section from a git config file.
func removeGitSubmoduleSection(configPath string, submodulePath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", configPath, err)
	}
	lines := strings.SplitAfter(string(content), "\n")
	var (
		output   []string
		inTarget bool
		removed  bool
	)
	for _, line := range lines {
		if name, ok := parseGitSubmoduleSection(line); ok {
			inTarget = name == submodulePath
			if inTarget {
				removed = true
				continue
			}
		} else if isGitConfigSection(line) {
			inTarget = false
		}
		if inTarget {
			continue
		}
		output = append(output, line)
	}
	if !removed {
		return nil
	}
	outputContent := strings.Join(output, "")
	if strings.TrimSpace(outputContent) == "" {
		if err = os.Remove(configPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove empty %s: %w", configPath, err)
		}
		return nil
	}
	if err = os.WriteFile(configPath, []byte(outputContent), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", configPath, err)
	}
	return nil
}

// parseGitSubmoduleSection extracts submodule path names from section headers.
func parseGitSubmoduleSection(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, `[submodule "`) || !strings.HasSuffix(trimmed, `"]`) {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(trimmed, `[submodule "`), `"]`), true
}

// isGitConfigSection reports whether a line starts any Git config section.
func isGitConfigSection(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

// isGitlink reports whether a repository path is tracked as a gitlink.
func isGitlink(ctx context.Context, a *app, relative string) bool {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "ls-files", "--stage", "--", relative)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(output), "160000 ")
}

// checkoutPluginSource checks out the configured source into a temporary directory.
func checkoutPluginSource(ctx context.Context, a *app, item pluginPlanItem) (pluginSourceCheckout, error) {
	tempRoot := filepath.Join(a.root, "temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return pluginSourceCheckout{}, fmt.Errorf("create temp directory: %w", err)
	}
	dir, err := os.MkdirTemp(tempRoot, "plugin-source-"+item.Source+"-")
	if err != nil {
		return pluginSourceCheckout{}, fmt.Errorf("create plugin source checkout: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			if cleanupErr := os.RemoveAll(dir); cleanupErr != nil {
				fmt.Fprintf(a.stderr, "warning: remove temporary plugin checkout %s: %v\n", dir, cleanupErr)
			}
		}
	}()
	if err = a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "clone", "--quiet", "--no-checkout", item.Repo, dir); err != nil {
		return pluginSourceCheckout{}, err
	}
	if err = a.runCommand(ctx, commandOptions{Dir: dir, Quiet: true}, "git", "checkout", "--quiet", item.Ref); err != nil {
		return pluginSourceCheckout{}, err
	}
	commit, err := a.runCommandOutput(ctx, commandOptions{Dir: dir, Quiet: true}, "git", "rev-parse", "HEAD")
	if err != nil {
		return pluginSourceCheckout{}, err
	}
	cleanup = false
	return pluginSourceCheckout{Dir: dir, Commit: strings.TrimSpace(commit)}, nil
}

// applyPluginFromCheckout copies one plugin from a temporary checkout.
func applyPluginFromCheckout(ctx context.Context, a *app, item pluginPlanItem, checkout pluginSourceCheckout, lock *pluginLockFile, update bool, force bool) error {
	sourceDir := filepath.Join(checkout.Dir, filepath.FromSlash(sourcePluginRelativePath(item)))
	if !fileExists(filepath.Join(sourceDir, "plugin.yaml")) {
		return fmt.Errorf("source plugin %s is missing plugin.yaml at %s", item.ID, sourcePluginRelativePath(item))
	}
	targetDir := managedPluginPath(a.root, item.ID)
	targetExists := dirExists(targetDir)
	if !update && targetExists && !force {
		return fmt.Errorf("plugin %s already exists; use `make plugins.update` or force=1", item.ID)
	}
	if update && targetExists && !force {
		changed, err := pluginHasLocalChanges(ctx, a, item.ID, lock.entryByID(item.ID))
		if err != nil {
			return err
		}
		if changed {
			return fmt.Errorf("plugin %s has local changes; commit them or use force=1", item.ID)
		}
	}
	if update && !targetExists && !force {
		return fmt.Errorf("plugin %s is not installed; use `make plugins.install`", item.ID)
	}

	tempTarget := filepath.Join(managedPluginRoot(a.root), "."+item.ID+".tmp")
	if err := os.RemoveAll(tempTarget); err != nil {
		return fmt.Errorf("remove stale temp plugin dir: %w", err)
	}
	if err := copyPluginDir(sourceDir, tempTarget); err != nil {
		return err
	}
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("remove existing plugin %s: %w", item.ID, err)
	}
	if err := os.Rename(tempTarget, targetDir); err != nil {
		return fmt.Errorf("replace plugin %s: %w", item.ID, err)
	}
	manifest, err := readPluginManifest(filepath.Join(targetDir, "plugin.yaml"))
	if err != nil {
		return err
	}
	hash, err := hashDirectory(targetDir)
	if err != nil {
		return err
	}
	lock.upsert(pluginLockEntry{
		ID:             item.ID,
		Source:         item.Source,
		Repo:           item.Repo,
		Root:           item.Root,
		Ref:            item.Ref,
		ResolvedCommit: checkout.Commit,
		Version:        manifest.Version,
		ContentHash:    hash,
	})
	action := "Installed"
	if update {
		action = "Updated"
	}
	fmt.Fprintf(a.stdout, "%s plugin %s from %s@%s\n", action, item.ID, item.Source, checkout.Commit)
	return nil
}

// pluginHasLocalChanges reports whether a plugin differs from Git or lock state.
func pluginHasLocalChanges(ctx context.Context, a *app, pluginID string, lock *pluginLockEntry) (bool, error) {
	relative := filepath.ToSlash(filepath.Join(managedPluginRootRelativePath, pluginID))
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "status", "--porcelain", "--", relative)
	if err == nil {
		if strings.TrimSpace(output) != "" {
			return true, nil
		}
	}
	if lock == nil || lock.ContentHash == "" {
		return false, nil
	}
	currentHash, hashErr := hashDirectory(managedPluginPath(a.root, pluginID))
	if hashErr != nil {
		return false, hashErr
	}
	return currentHash != lock.ContentHash, nil
}

// readPluginManifest reads the manifest fields needed by plugin management.
func readPluginManifest(path string) (pluginManifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return pluginManifest{}, fmt.Errorf("read plugin manifest %s: %w", path, err)
	}
	var manifest pluginManifest
	if err = yaml.Unmarshal(content, &manifest); err != nil {
		return pluginManifest{}, fmt.Errorf("parse plugin manifest %s: %w", path, err)
	}
	return manifest, nil
}

// readPluginLock reads the tool-generated plugin lock file.
func readPluginLock(root string) (pluginLockFile, error) {
	path := pluginLockPath(root)
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return pluginLockFile{}, nil
		}
		return pluginLockFile{}, fmt.Errorf("read plugin lock: %w", err)
	}
	var lock pluginLockFile
	if err = yaml.Unmarshal(content, &lock); err != nil {
		return pluginLockFile{}, fmt.Errorf("parse plugin lock: %w", err)
	}
	return lock, nil
}

// writePluginLock writes the tool-generated plugin lock file atomically.
func writePluginLock(root string, lock pluginLockFile) error {
	sort.Slice(lock.Plugins, func(left int, right int) bool {
		return lock.Plugins[left].ID < lock.Plugins[right].ID
	})
	path := pluginLockPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create plugin lock parent: %w", err)
	}
	content, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("marshal plugin lock: %w", err)
	}
	tempPath := path + ".tmp"
	if err = os.WriteFile(tempPath, content, 0o644); err != nil {
		return fmt.Errorf("write temporary plugin lock: %w", err)
	}
	if err = os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace plugin lock: %w", err)
	}
	return nil
}

// entriesByID maps lock entries by plugin ID.
func (l pluginLockFile) entriesByID() map[string]*pluginLockEntry {
	result := map[string]*pluginLockEntry{}
	for index := range l.Plugins {
		entry := &l.Plugins[index]
		result[entry.ID] = entry
	}
	return result
}

// entryByID returns one lock entry by plugin ID.
func (l pluginLockFile) entryByID(id string) *pluginLockEntry {
	for index := range l.Plugins {
		if l.Plugins[index].ID == id {
			return &l.Plugins[index]
		}
	}
	return nil
}

// upsert inserts or replaces one lock entry.
func (l *pluginLockFile) upsert(entry pluginLockEntry) {
	for index := range l.Plugins {
		if l.Plugins[index].ID == entry.ID {
			l.Plugins[index] = entry
			return
		}
	}
	l.Plugins = append(l.Plugins, entry)
}

// printUnconfiguredPluginStatus prints local plugins and lock entries absent from config.
func printUnconfiguredPluginStatus(a *app, plan pluginPlan, lockByID map[string]*pluginLockEntry) error {
	configured := map[string]struct{}{}
	for _, item := range plan.Items {
		configured[item.ID] = struct{}{}
	}
	var local []string
	entries, err := os.ReadDir(managedPluginRoot(a.root))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read plugin workspace: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if fileExists(filepath.Join(managedPluginRoot(a.root), entry.Name(), "plugin.yaml")) {
			if _, ok := configured[entry.Name()]; !ok {
				local = append(local, entry.Name())
			}
		}
	}
	sort.Strings(local)
	if len(local) > 0 {
		fmt.Fprintf(a.stdout, "Unconfigured local plugins: %s\n", strings.Join(local, ", "))
	}
	var orphaned []string
	for id := range lockByID {
		if _, ok := configured[id]; !ok {
			orphaned = append(orphaned, id)
		}
	}
	sort.Strings(orphaned)
	if len(orphaned) > 0 {
		fmt.Fprintf(a.stdout, "Orphaned lock entries: %s\n", strings.Join(orphaned, ", "))
	}
	return nil
}

// remoteStatus compares local, lock, and remote content states.
func remoteStatus(lock *pluginLockEntry, currentHash string, remoteHash string) string {
	if lock == nil {
		return "not-locked"
	}
	if currentHash != "" && lock.ContentHash != "" && currentHash != lock.ContentHash {
		return "local-drift"
	}
	if remoteHash != "" && lock.ContentHash != "" && remoteHash != lock.ContentHash {
		return "update-available"
	}
	return "current"
}

// sourcePluginRelativePath renders the plugin path inside a source checkout.
func sourcePluginRelativePath(item pluginPlanItem) string {
	if item.Root == "." {
		return item.ID
	}
	return path.Join(item.Root, item.ID)
}

// managedPluginRoot returns the fixed local source-plugin workspace path.
func managedPluginRoot(root string) string {
	return filepath.Join(root, filepath.FromSlash(managedPluginRootRelativePath))
}

// managedPluginPath returns one local plugin directory path.
func managedPluginPath(root string, pluginID string) string {
	return filepath.Join(managedPluginRoot(root), pluginID)
}

// pluginLockPath returns the tool-generated lock file path.
func pluginLockPath(root string) string {
	return filepath.Join(managedPluginRoot(root), pluginLockFileName)
}

// runCommandOutput executes a child command and returns stdout.
func (a *app) runCommandOutput(ctx context.Context, options commandOptions, name string, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	options.Quiet = false
	options.Stdout = &stdout
	options.Stderr = &stderr
	err := a.runCommand(ctx, options, name, args...)
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return "", err
	}
	return stdout.String(), nil
}
