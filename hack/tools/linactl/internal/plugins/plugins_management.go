// This file contains source-plugin workspace management and configured plugin
// installation helpers used by plugins.* commands. It owns the Git checkout,
// lock-file, status-table, and submodule conversion details behind a narrow
// Runtime interface supplied by root command files.

package plugins

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"linactl/internal/config"
	"linactl/internal/fileutil"
	"linactl/internal/toolrun"
	"linactl/internal/toolutil"

	"gopkg.in/yaml.v3"
)

const (
	// ManagedRootRelativePath is the fixed source-plugin workspace path.
	ManagedRootRelativePath = "apps/lina-plugins"
	// pluginLockFileName stores tool-generated source-plugin state.
	pluginLockFileName = ".linapro-plugins.lock.yaml"
	// pluginSourceCacheDirName stores reusable Git checkouts for configured sources.
	pluginSourceCacheDirName = "plugin-sources"
	// pluginWildcardItem expands to every plugin directory under the source root.
	pluginWildcardItem = "*"
)

var (
	// pluginIDPattern matches the repository plugin ID convention.
	pluginIDPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	// pluginSourceNamePattern matches safe source names for diagnostics.
	pluginSourceNamePattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9_.-]*[A-Za-z0-9])?$`)
)

// ManagedWorkspaceState classifies the user-project plugin workspace.
type ManagedWorkspaceState string

const (
	// ManagedWorkspaceMissing means apps/lina-plugins does not exist.
	ManagedWorkspaceMissing ManagedWorkspaceState = "missing"
	// ManagedWorkspaceOrdinary means apps/lina-plugins is a normal directory.
	ManagedWorkspaceOrdinary ManagedWorkspaceState = "ordinary"
	// ManagedWorkspaceSubmodule means apps/lina-plugins is tracked as a gitlink.
	ManagedWorkspaceSubmodule ManagedWorkspaceState = "submodule"
	// ManagedWorkspaceNestedGit means apps/lina-plugins contains Git metadata.
	ManagedWorkspaceNestedGit ManagedWorkspaceState = "nested-git"
	// ManagedWorkspaceInvalid means apps/lina-plugins exists but is not a directory.
	ManagedWorkspaceInvalid ManagedWorkspaceState = "invalid"
)

// ManagedWorkspace describes apps/lina-plugins for plugin management.
type ManagedWorkspace struct {
	Root  string
	State ManagedWorkspaceState
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

// pluginSourceCheckout stores one source checkout and resolved commit.
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

// pluginStatusRow stores one printable configured plugin status row.
type pluginStatusRow struct {
	Plugin    string
	Source    string
	Version   string
	Installed string
	Dirty     string
	Remote    string
	Note      string
}

// InstallOrUpdate executes install or update for selected plugins.
func InstallOrUpdate(ctx context.Context, runtime Runtime, input Input, update bool) error {
	workspace, err := InspectManagedWorkspace(ctx, runtime)
	if err != nil {
		return err
	}
	if workspace.State == ManagedWorkspaceSubmodule {
		return errors.New("apps/lina-plugins is still a submodule; run `make plugins.init` first")
	}
	if workspace.State == ManagedWorkspaceNestedGit {
		return errors.New("apps/lina-plugins contains nested Git metadata; run `make plugins.init` or remove nested metadata first")
	}
	if workspace.State == ManagedWorkspaceInvalid {
		return fmt.Errorf("apps/lina-plugins is invalid: %s", workspace.Root)
	}
	if err = os.MkdirAll(workspace.Root, 0o755); err != nil {
		return fmt.Errorf("create plugin workspace: %w", err)
	}

	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}
	plan, err := LoadPlan(runtime.Root, input)
	if err != nil {
		return err
	}
	lock, err := ReadLock(runtime.Root)
	if err != nil {
		return err
	}

	action := "Installing"
	prepAction := "installation"
	if update {
		action = "Updating"
		prepAction = "update"
	}
	if _, err = fmt.Fprintf(runtime.Stdout, "Preparing plugin %s for %d configured item(s)...\n", prepAction, len(plan.Items)); err != nil {
		return fmt.Errorf("write plugin progress: %w", err)
	}
	checkouts, sourceErrors, err := checkoutPluginSources(ctx, runtime, plan.Items)
	if err != nil {
		return err
	}
	if len(sourceErrors) > 0 {
		return firstPluginSourceError(sourceErrors)
	}
	plan, err = expandPluginPlanFromCheckouts(plan, checkouts)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(runtime.Stdout, "%s %d plugin(s)...\n", action, len(plan.Items)); err != nil {
		return fmt.Errorf("write plugin progress: %w", err)
	}
	for index, item := range plan.Items {
		checkout := checkouts[item.Source]
		if _, err = fmt.Fprintf(runtime.Stdout, "[%d/%d] %s plugin %s from %s...\n", index+1, len(plan.Items), strings.ToLower(action), item.ID, item.Source); err != nil {
			return fmt.Errorf("write plugin progress: %w", err)
		}
		if err = applyPluginFromCheckout(ctx, runtime, item, checkout, &lock, update, force); err != nil {
			return err
		}
	}
	if err = WriteLock(runtime.Root, lock); err != nil {
		return err
	}
	return nil
}

// Status prints read-only plugin workspace and source status.
func Status(ctx context.Context, runtime Runtime, input Input) error {
	workspace, err := InspectManagedWorkspace(ctx, runtime)
	if err != nil {
		return err
	}
	fmt.Fprintf(runtime.Stdout, "Plugin workspace: %s (%s)\n", toolutil.RelativePath(runtime.Root, workspace.Root), workspace.State)
	if workspace.State == ManagedWorkspaceSubmodule {
		fmt.Fprintln(runtime.Stdout, "Action required: run `make plugins.init` before installing or updating user-project plugins.")
		return nil
	}

	plan, err := LoadPlan(runtime.Root, input)
	if err != nil {
		return err
	}
	lock, err := ReadLock(runtime.Root)
	if err != nil {
		return err
	}
	lockByID := lock.entriesByID()

	if _, err = fmt.Fprintln(runtime.Stdout, "Querying configured plugin sources..."); err != nil {
		return fmt.Errorf("write plugin status progress: %w", err)
	}
	checkouts, sourceErrors, err := checkoutPluginSources(ctx, runtime, plan.Items)
	if err != nil {
		return err
	}
	expandedPlan, expandErr := expandPluginPlanFromCheckouts(plan, checkouts)
	if expandErr != nil {
		return expandErr
	}

	if _, err = fmt.Fprintf(runtime.Stdout, "Rendering status for %d configured plugin(s)...\n", len(expandedPlan.Items)); err != nil {
		return fmt.Errorf("write plugin status progress: %w", err)
	}
	rows := make([]pluginStatusRow, 0, len(expandedPlan.Items))
	for _, item := range expandedPlan.Items {
		target := ManagedPath(runtime.Root, item.ID)
		exists := fileutil.DirExists(target)
		version := "-"
		contentHash := ""
		if exists {
			if manifest, readErr := ReadManifest(filepath.Join(target, "plugin.yaml")); readErr == nil {
				version = toolutil.FirstNonEmpty(manifest.Version, "-")
			}
			if hash, hashErr := fileutil.HashDirectory(target); hashErr == nil {
				contentHash = hash
			}
		}

		dirty := "unknown"
		if exists {
			changed, dirtyErr := pluginHasLocalChanges(ctx, runtime, item.ID, lockByID[item.ID])
			if dirtyErr == nil {
				dirty = fmt.Sprintf("%t", changed)
			}
		} else {
			dirty = "false"
		}

		remote := "unknown"
		note := ""
		if sourceErr, ok := sourceErrors[item.Source]; ok {
			note = sourceErr.Error()
			rows = append(rows, pluginStatusRow{
				Plugin:    item.ID,
				Source:    item.Source,
				Version:   version,
				Installed: fmt.Sprintf("%t", exists),
				Dirty:     dirty,
				Remote:    remote,
				Note:      note,
			})
			continue
		}
		if checkout, ok := checkouts[item.Source]; ok {
			sourceDir := filepath.Join(checkout.Dir, filepath.FromSlash(sourcePluginRelativePath(item)))
			if !fileutil.FileExists(filepath.Join(sourceDir, "plugin.yaml")) {
				remote = "missing"
			} else if remoteHash, hashErr := fileutil.HashDirectory(sourceDir); hashErr == nil {
				remote = remoteStatus(lockByID[item.ID], contentHash, remoteHash)
			}
		}
		rows = append(rows, pluginStatusRow{
			Plugin:    item.ID,
			Source:    item.Source,
			Version:   version,
			Installed: fmt.Sprintf("%t", exists),
			Dirty:     dirty,
			Remote:    remote,
			Note:      note,
		})
	}
	if _, err = fmt.Fprintln(runtime.Stdout, "Configured plugins:"); err != nil {
		return fmt.Errorf("write plugin status progress: %w", err)
	}
	if err = printPluginStatusTable(runtime.Stdout, rows); err != nil {
		return err
	}

	if err = printUnconfiguredPluginStatus(runtime, expandedPlan, lockByID); err != nil {
		return err
	}
	return nil
}

// LoadPlan loads and validates plugin source configuration.
func LoadPlan(root string, input Input) (pluginPlan, error) {
	cfg, err := LoadConfig(root, input)
	if err != nil {
		return pluginPlan{}, err
	}
	return ValidateConfig(cfg.Plugins, input)
}

// ValidateConfig normalizes configured plugin sources and applies
// command-level source or plugin filters.
func ValidateConfig(cfg config.Plugins, input Input) (pluginPlan, error) {
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

// checkoutPluginSources synchronizes each configured source once.
func checkoutPluginSources(ctx context.Context, runtime Runtime, items []pluginPlanItem) (map[string]pluginSourceCheckout, map[string]error, error) {
	checkouts := map[string]pluginSourceCheckout{}
	sourceErrors := map[string]error{}
	for _, item := range items {
		if _, ok := checkouts[item.Source]; ok {
			continue
		}
		if _, ok := sourceErrors[item.Source]; ok {
			continue
		}
		if _, err := fmt.Fprintf(runtime.Stdout, "Synchronizing plugin source %s from %s (%s)...\n", item.Source, item.Repo, item.Ref); err != nil {
			return checkouts, sourceErrors, fmt.Errorf("write plugin source progress: %w", err)
		}
		checkout, err := checkoutPluginSource(ctx, runtime, item)
		if err != nil {
			sourceErrors[item.Source] = err
			continue
		}
		checkouts[item.Source] = checkout
		if _, err = fmt.Fprintf(runtime.Stdout, "Resolved plugin source %s at %s\n", item.Source, checkout.Commit); err != nil {
			return checkouts, sourceErrors, fmt.Errorf("write plugin source progress: %w", err)
		}
	}
	return checkouts, sourceErrors, nil
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
		if !fileutil.FileExists(filepath.Join(root, id, "plugin.yaml")) {
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
func validatePluginSourceConfig(sourceName string, source config.PluginSource) (string, error) {
	if strings.TrimSpace(source.Repo) == "" {
		return "", fmt.Errorf("plugins.sources.%s.repo is required", sourceName)
	}
	root, err := ValidateSourceRoot(source.Root)
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

// ValidateSourceRoot validates a repository-internal source root.
func ValidateSourceRoot(root string) (string, error) {
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

// InspectManagedWorkspace classifies apps/lina-plugins for management.
func InspectManagedWorkspace(ctx context.Context, runtime Runtime) (ManagedWorkspace, error) {
	root := ManagedRoot(runtime.Root)
	if isGitlink(ctx, runtime, ManagedRootRelativePath) {
		return ManagedWorkspace{Root: root, State: ManagedWorkspaceSubmodule}, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ManagedWorkspace{Root: root, State: ManagedWorkspaceMissing}, nil
		}
		return ManagedWorkspace{}, fmt.Errorf("stat plugin workspace: %w", err)
	}
	if !info.IsDir() {
		return ManagedWorkspace{Root: root, State: ManagedWorkspaceInvalid}, nil
	}
	if fileutil.FileExists(filepath.Join(root, ".git")) || fileutil.DirExists(filepath.Join(root, ".git")) {
		return ManagedWorkspace{Root: root, State: ManagedWorkspaceNestedGit}, nil
	}
	return ManagedWorkspace{Root: root, State: ManagedWorkspaceOrdinary}, nil
}

// ConvertSubmoduleToDirectory removes submodule metadata while keeping files.
func ConvertSubmoduleToDirectory(ctx context.Context, runtime Runtime, workspace ManagedWorkspace) error {
	if err := os.MkdirAll(workspace.Root, 0o755); err != nil {
		return fmt.Errorf("create plugin workspace: %w", err)
	}
	if err := RemoveGitSubmoduleSection(filepath.Join(runtime.Root, ".gitmodules"), ManagedRootRelativePath); err != nil {
		return err
	}
	if err := RemoveGitSubmoduleSection(filepath.Join(runtime.Root, ".git", "config"), ManagedRootRelativePath); err != nil {
		return err
	}
	if err := fileutil.RemoveDirectoryIfExists(filepath.Join(runtime.Root, ".git", "modules", "apps", "lina-plugins")); err != nil {
		return err
	}
	pluginGitPath := filepath.Join(workspace.Root, ".git")
	if fileutil.FileExists(pluginGitPath) || fileutil.DirExists(pluginGitPath) {
		if err := os.RemoveAll(pluginGitPath); err != nil {
			return fmt.Errorf("remove plugin git metadata: %w", err)
		}
	}
	if err := runtime.run(ctx, toolrun.Options{Dir: runtime.Root, Quiet: true}, "git", "update-index", "--force-remove", "--", ManagedRootRelativePath); err != nil {
		return err
	}
	fmt.Fprintf(runtime.Stdout, "Plugin workspace converted to ordinary directory: %s\n", toolutil.RelativePath(runtime.Root, workspace.Root))
	return nil
}

// RemoveGitSubmoduleSection removes the target submodule section from a git config file.
func RemoveGitSubmoduleSection(configPath string, submodulePath string) error {
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
func isGitlink(ctx context.Context, runtime Runtime, relative string) bool {
	output, err := runtime.output(ctx, toolrun.Options{Dir: runtime.Root, Quiet: true}, "git", "ls-files", "--stage", "--", relative)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(output), "160000 ")
}

// checkoutPluginSource synchronizes the configured source into the reusable cache.
func checkoutPluginSource(ctx context.Context, runtime Runtime, item pluginPlanItem) (pluginSourceCheckout, error) {
	cacheDir := SourceCachePath(runtime.Root, item.Source)
	reusable, err := pluginSourceCacheReusable(ctx, runtime, cacheDir, item.Repo)
	if err != nil {
		return pluginSourceCheckout{}, err
	}
	if !reusable {
		if err = rebuildPluginSourceCache(ctx, runtime, item, cacheDir); err != nil {
			return pluginSourceCheckout{}, err
		}
	} else if err = fetchPluginSourceCache(ctx, runtime, cacheDir); err != nil {
		return pluginSourceCheckout{}, err
	}

	commit, err := resolvePluginSourceRef(ctx, runtime, cacheDir, item.Ref)
	if err != nil {
		return pluginSourceCheckout{}, err
	}
	if err = runtime.run(ctx, toolrun.Options{Dir: cacheDir, Quiet: true}, "git", "checkout", "--quiet", "--force", "--detach", commit); err != nil {
		return pluginSourceCheckout{}, err
	}
	if err = runtime.run(ctx, toolrun.Options{Dir: cacheDir, Quiet: true}, "git", "reset", "--hard", "--quiet", commit); err != nil {
		return pluginSourceCheckout{}, err
	}
	if err = runtime.run(ctx, toolrun.Options{Dir: cacheDir, Quiet: true}, "git", "clean", "-fdx", "--quiet"); err != nil {
		return pluginSourceCheckout{}, err
	}
	return pluginSourceCheckout{Dir: cacheDir, Commit: commit}, nil
}

// pluginSourceCacheReusable reports whether an existing source cache can be reused.
func pluginSourceCacheReusable(ctx context.Context, runtime Runtime, cacheDir string, repo string) (bool, error) {
	info, err := os.Stat(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat plugin source cache: %w", err)
	}
	if !info.IsDir() {
		return false, nil
	}
	if _, err = runtime.output(ctx, toolrun.Options{Dir: cacheDir, Quiet: true}, "git", "rev-parse", "--git-dir"); err != nil {
		return false, nil
	}
	remote, err := runtime.output(ctx, toolrun.Options{Dir: cacheDir, Quiet: true}, "git", "remote", "get-url", "origin")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(remote) == strings.TrimSpace(repo), nil
}

// rebuildPluginSourceCache recreates one source cache from the configured repo.
func rebuildPluginSourceCache(ctx context.Context, runtime Runtime, item pluginPlanItem, cacheDir string) error {
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("remove stale plugin source cache: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
		return fmt.Errorf("create plugin source cache parent: %w", err)
	}
	if err := runtime.run(ctx, toolrun.Options{Dir: runtime.Root, Stderr: runtime.Stdout}, "git", "clone", "--progress", "--no-checkout", item.Repo, cacheDir); err != nil {
		return err
	}
	return nil
}

// fetchPluginSourceCache updates one reusable source cache from origin.
func fetchPluginSourceCache(ctx context.Context, runtime Runtime, cacheDir string) error {
	if err := runtime.run(ctx, toolrun.Options{Dir: cacheDir, Stderr: runtime.Stdout}, "git", "fetch", "--prune", "--progress", "origin"); err != nil {
		return err
	}
	return nil
}

// resolvePluginSourceRef resolves a configured source ref to a commit.
func resolvePluginSourceRef(ctx context.Context, runtime Runtime, cacheDir string, ref string) (string, error) {
	candidates := []string{"refs/remotes/origin/" + ref, ref}
	for _, candidate := range candidates {
		commit, err := runtime.output(ctx, toolrun.Options{Dir: cacheDir, Quiet: true}, "git", "rev-parse", "--verify", candidate+"^{commit}")
		if err == nil && strings.TrimSpace(commit) != "" {
			return strings.TrimSpace(commit), nil
		}
	}
	return "", fmt.Errorf("resolve plugin source ref %s", ref)
}

// applyPluginFromCheckout copies one plugin from a source checkout.
func applyPluginFromCheckout(ctx context.Context, runtime Runtime, item pluginPlanItem, checkout pluginSourceCheckout, lock *pluginLockFile, update bool, force bool) error {
	sourceDir := filepath.Join(checkout.Dir, filepath.FromSlash(sourcePluginRelativePath(item)))
	if !fileutil.FileExists(filepath.Join(sourceDir, "plugin.yaml")) {
		return fmt.Errorf("source plugin %s is missing plugin.yaml at %s", item.ID, sourcePluginRelativePath(item))
	}
	targetDir := ManagedPath(runtime.Root, item.ID)
	targetExists := fileutil.DirExists(targetDir)
	if !update && targetExists && !force {
		return fmt.Errorf("plugin %s already exists; use `make plugins.update` or force=1", item.ID)
	}
	if update && targetExists && !force {
		changed, err := pluginHasLocalChanges(ctx, runtime, item.ID, lock.entryByID(item.ID))
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

	tempTarget := filepath.Join(ManagedRoot(runtime.Root), "."+item.ID+".tmp")
	if err := os.RemoveAll(tempTarget); err != nil {
		return fmt.Errorf("remove stale temp plugin dir: %w", err)
	}
	if err := fileutil.CopyPluginDir(sourceDir, tempTarget); err != nil {
		return err
	}
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("remove existing plugin %s: %w", item.ID, err)
	}
	if err := os.Rename(tempTarget, targetDir); err != nil {
		return fmt.Errorf("replace plugin %s: %w", item.ID, err)
	}
	manifest, err := ReadManifest(filepath.Join(targetDir, "plugin.yaml"))
	if err != nil {
		return err
	}
	hash, err := fileutil.HashDirectory(targetDir)
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
	fmt.Fprintf(runtime.Stdout, "%s plugin %s from %s@%s\n", action, item.ID, item.Source, checkout.Commit)
	return nil
}

// pluginHasLocalChanges reports whether a plugin differs from Git or lock state.
func pluginHasLocalChanges(ctx context.Context, runtime Runtime, pluginID string, lock *pluginLockEntry) (bool, error) {
	relative := filepath.ToSlash(filepath.Join(ManagedRootRelativePath, pluginID))
	output, err := runtime.output(ctx, toolrun.Options{Dir: runtime.Root, Quiet: true}, "git", "status", "--porcelain", "--", relative)
	if err == nil {
		if strings.TrimSpace(output) != "" {
			return true, nil
		}
	}
	if lock == nil || lock.ContentHash == "" {
		return false, nil
	}
	currentHash, hashErr := fileutil.HashDirectory(ManagedPath(runtime.Root, pluginID))
	if hashErr != nil {
		return false, hashErr
	}
	return currentHash != lock.ContentHash, nil
}

// ReadManifest reads the manifest fields needed by plugin management.
func ReadManifest(path string) (Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read plugin manifest %s: %w", path, err)
	}
	var manifest Manifest
	if err = yaml.Unmarshal(content, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse plugin manifest %s: %w", path, err)
	}
	return manifest, nil
}

// ReadLock reads the tool-generated plugin lock file.
func ReadLock(root string) (pluginLockFile, error) {
	path := LockPath(root)
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

// WriteLock writes the tool-generated plugin lock file atomically.
func WriteLock(root string, lock pluginLockFile) error {
	sort.Slice(lock.Plugins, func(left int, right int) bool {
		return lock.Plugins[left].ID < lock.Plugins[right].ID
	})
	path := LockPath(root)
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
func printUnconfiguredPluginStatus(runtime Runtime, plan pluginPlan, lockByID map[string]*pluginLockEntry) error {
	configured := map[string]struct{}{}
	for _, item := range plan.Items {
		configured[item.ID] = struct{}{}
	}
	var local []string
	entries, err := os.ReadDir(ManagedRoot(runtime.Root))
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
		if fileutil.FileExists(filepath.Join(ManagedRoot(runtime.Root), entry.Name(), "plugin.yaml")) {
			if _, ok := configured[entry.Name()]; !ok {
				local = append(local, entry.Name())
			}
		}
	}
	sort.Strings(local)
	if len(local) > 0 {
		fmt.Fprintf(runtime.Stdout, "Unconfigured local plugins: %s\n", strings.Join(local, ", "))
	}
	var orphaned []string
	for id := range lockByID {
		if _, ok := configured[id]; !ok {
			orphaned = append(orphaned, id)
		}
	}
	sort.Strings(orphaned)
	if len(orphaned) > 0 {
		fmt.Fprintf(runtime.Stdout, "Orphaned lock entries: %s\n", strings.Join(orphaned, ", "))
	}
	return nil
}

// printPluginStatusTable renders configured plugin status as an aligned ASCII table.
func printPluginStatusTable(out io.Writer, rows []pluginStatusRow) error {
	headers := []string{"Plugin", "Source", "Version", "Installed", "Dirty", "Remote", "Note"}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		values := row.values()
		for i, value := range values {
			if len(value) > widths[i] {
				widths[i] = len(value)
			}
		}
	}

	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	if err := printTableRow(out, widths, headers); err != nil {
		return err
	}
	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := printTableRow(out, widths, row.values()); err != nil {
			return err
		}
	}
	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	return nil
}

// printTableBorder prints one ASCII border line for a table.
func printTableBorder(out io.Writer, widths []int) error {
	if _, err := fmt.Fprint(out, "+"); err != nil {
		return fmt.Errorf("write table border: %w", err)
	}
	for _, width := range widths {
		if _, err := fmt.Fprint(out, strings.Repeat("-", width+2)); err != nil {
			return fmt.Errorf("write table border: %w", err)
		}
		if _, err := fmt.Fprint(out, "+"); err != nil {
			return fmt.Errorf("write table border: %w", err)
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return fmt.Errorf("write table border: %w", err)
	}
	return nil
}

// printTableRow prints one padded ASCII table row.
func printTableRow(out io.Writer, widths []int, values []string) error {
	if _, err := fmt.Fprint(out, "|"); err != nil {
		return fmt.Errorf("write table row: %w", err)
	}
	for i, value := range values {
		if _, err := fmt.Fprintf(out, " %-*s |", widths[i], value); err != nil {
			return fmt.Errorf("write table row: %w", err)
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return fmt.Errorf("write table row: %w", err)
	}
	return nil
}

// values returns the printable table cells for one configured plugin row.
func (r pluginStatusRow) values() []string {
	return []string{r.Plugin, r.Source, r.Version, r.Installed, r.Dirty, r.Remote, r.Note}
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
func ManagedRoot(root string) string {
	return filepath.Join(root, filepath.FromSlash(ManagedRootRelativePath))
}

// managedPluginPath returns one local plugin directory path.
func ManagedPath(root string, pluginID string) string {
	return filepath.Join(ManagedRoot(root), pluginID)
}

// pluginLockPath returns the tool-generated lock file path.
func LockPath(root string) string {
	return filepath.Join(ManagedRoot(root), pluginLockFileName)
}

// SourceCachePath returns the reusable checkout path for one configured source.
func SourceCachePath(root string, source string) string {
	return filepath.Join(root, "temp", pluginSourceCacheDirName, source)
}
