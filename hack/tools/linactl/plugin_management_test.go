// This file verifies configured source-plugin workspace management commands.

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"linactl/internal/config"
	"linactl/internal/fileutil"
	"linactl/internal/plugins"
)

// TestValidatePluginConfigAcceptsStringItemsAndFilters verifies configured
// source and plugin filters keep items as plain string plugin IDs.
func TestValidatePluginConfigAcceptsStringItemsAndFilters(t *testing.T) {
	cfg := config.Plugins{
		Sources: map[string]config.PluginSource{
			"custom": {
				Repo:  "https://example.com/custom.git",
				Root:  "apps/lina-plugins",
				Ref:   "main",
				Items: []string{"linapro-content-notice"},
			},
			"official": {
				Repo:  "https://example.com/official.git",
				Root:  ".",
				Ref:   "main",
				Items: []string{"linapro-tenant-core", "linapro-org-core"},
			},
		},
	}

	plan, err := plugins.ValidateConfig(cfg, commandInput{Params: map[string]string{"p": "linapro-org-core", "source": "official"}})
	if err != nil {
		t.Fatalf("plugins.ValidateConfig returned error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Fatalf("expected one filtered item, got %#v", plan.Items)
	}
	item := plan.Items[0]
	if item.ID != "linapro-org-core" || item.Source != "official" || item.Root != "." {
		t.Fatalf("unexpected filtered item: %#v", item)
	}
}

// TestValidatePluginConfigRejectsDuplicatePluginIDs verifies duplicate plugin
// IDs across sources are rejected before any workspace write.
func TestValidatePluginConfigRejectsDuplicatePluginIDs(t *testing.T) {
	cfg := config.Plugins{
		Sources: map[string]config.PluginSource{
			"a": {Repo: "repo-a", Root: ".", Ref: "main", Items: []string{"linapro-tenant-core"}},
			"b": {Repo: "repo-b", Root: ".", Ref: "main", Items: []string{"linapro-tenant-core"}},
		},
	}

	_, err := plugins.ValidateConfig(cfg, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "multiple sources") {
		t.Fatalf("expected duplicate plugin validation error, got %v", err)
	}
}

// TestValidatePluginConfigRejectsWildcardMixedWithExplicitIDs verifies a
// source cannot mix "*" with individual plugin IDs.
func TestValidatePluginConfigRejectsWildcardMixedWithExplicitIDs(t *testing.T) {
	cfg := config.Plugins{
		Sources: map[string]config.PluginSource{
			"official": {Repo: "repo", Root: ".", Ref: "main", Items: []string{"*", "linapro-tenant-core"}},
		},
	}

	_, err := plugins.ValidateConfig(cfg, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "cannot mix wildcard") {
		t.Fatalf("expected wildcard mix validation error, got %v", err)
	}
}

// TestValidatePluginSourceRootRejectsUnsafePaths verifies source roots cannot
// escape the remote repository or use platform-specific drive paths.
func TestValidatePluginSourceRootRejectsUnsafePaths(t *testing.T) {
	invalid := []string{"", "..", "../plugins", "/tmp/plugins", `C:\plugins`, "C:/plugins", "apps/../secret", "apps\\plugins"}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			if _, err := plugins.ValidateSourceRoot(value); err == nil {
				t.Fatalf("expected invalid root %q to fail", value)
			}
		})
	}
}

// TestLoadPluginPlanRejectsNonStringItems verifies YAML objects in items fail
// because plugin items must remain a string array.
func TestLoadPluginPlanRejectsNonStringItems(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-core"), 0o755); err != nil {
		t.Fatalf("mkdir lina-core: %v", err)
	}
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), `plugins:
  sources:
    official:
      repo: "https://example.com/plugins.git"
      root: "."
      ref: "main"
      items:
        - id: linapro-tenant-core
`)

	_, err := plugins.LoadPlan(root, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Fatalf("expected non-string item YAML error, got %v", err)
	}
}

// TestRemoveGitSubmoduleSectionPreservesOtherSections verifies only the plugin
// submodule section is removed from a Git config-style file.
func TestRemoveGitSubmoduleSectionPreservesOtherSections(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".gitmodules")
	writeFile(t, configPath, `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
[submodule "docs"]
	path = docs
	url = https://example.com/docs.git
`)

	if err := plugins.RemoveGitSubmoduleSection(configPath, plugins.ManagedRootRelativePath); err != nil {
		t.Fatalf("plugins.RemoveGitSubmoduleSection returned error: %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "apps/lina-plugins") {
		t.Fatalf("target submodule section was not removed:\n%s", text)
	}
	if !strings.Contains(text, `[submodule "docs"]`) {
		t.Fatalf("unrelated submodule section was not preserved:\n%s", text)
	}
}

// TestRemoveGitSubmoduleSectionStopsAtAnyNextSection verifies submodule
// removal does not delete following non-submodule Git config sections.
func TestRemoveGitSubmoduleSectionStopsAtAnyNextSection(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config")
	writeFile(t, configPath, `[core]
	repositoryformatversion = 0
[submodule "apps/lina-plugins"]
	url = https://example.com/plugins.git
[remote "origin"]
	url = https://example.com/project.git
`)

	if err := plugins.RemoveGitSubmoduleSection(configPath, plugins.ManagedRootRelativePath); err != nil {
		t.Fatalf("plugins.RemoveGitSubmoduleSection returned error: %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "apps/lina-plugins") {
		t.Fatalf("target submodule section was not removed:\n%s", text)
	}
	if !strings.Contains(text, `[core]`) || !strings.Contains(text, `[remote "origin"]`) {
		t.Fatalf("non-submodule sections were not preserved:\n%s", text)
	}
}

// TestRunPluginsInitConvertsGitlinkAndPreservesFiles verifies plugins.init
// removes submodule metadata without deleting plugin files.
func TestRunPluginsInitConvertsGitlinkAndPreservesFiles(t *testing.T) {
	root := newGitRepo(t)
	writeFile(t, filepath.Join(root, ".gitmodules"), `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
[submodule "docs"]
	path = docs
	url = https://example.com/docs.git
`)
	writeFile(t, filepath.Join(root, ".git", "config"), `[core]
	repositoryformatversion = 0
[submodule "apps/lina-plugins"]
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "demo", "plugin.yaml"), "id: demo\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", ".git"), "gitdir: ../../.git/modules/apps/lina-plugins\n")
	writeFile(t, filepath.Join(root, ".git", "modules", "apps", "lina-plugins", "config"), "[core]\n")
	runGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,apps/lina-plugins")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInit(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInit returned error: %v", err)
	}

	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "demo", "plugin.yaml")) {
		t.Fatalf("plugin file was not preserved")
	}
	gitmodules, err := os.ReadFile(filepath.Join(root, ".gitmodules"))
	if err != nil {
		t.Fatalf("read .gitmodules: %v", err)
	}
	if strings.Contains(string(gitmodules), "apps/lina-plugins") || !strings.Contains(string(gitmodules), `"docs"`) {
		t.Fatalf("unexpected .gitmodules content:\n%s", string(gitmodules))
	}
	stage := runGitOutput(t, root, "ls-files", "--stage", "--", plugins.ManagedRootRelativePath)
	if strings.Contains(stage, "160000") {
		t.Fatalf("gitlink still exists after plugins.init: %s", stage)
	}
	if fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", ".git")) || fileutil.DirExists(filepath.Join(root, ".git", "modules", "apps", "lina-plugins")) {
		t.Fatalf("submodule metadata was not cleaned")
	}
}

// TestPluginsInstallAutoInitializesSubmoduleWorkspace verifies install runs
// the same workspace initialization as plugins.init before copying plugins.
func TestPluginsInstallAutoInitializesSubmoduleWorkspace(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")
	writeFile(t, filepath.Join(root, ".gitmodules"), `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, ".git", "config"), `[core]
	repositoryformatversion = 0
[submodule "apps/lina-plugins"]
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", ".git"), "gitdir: ../../.git/modules/apps/lina-plugins\n")
	writeFile(t, filepath.Join(root, ".git", "modules", "apps", "lina-plugins", "config"), "[core]\n")
	runGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,apps/lina-plugins")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Plugin workspace converted to ordinary directory") || !strings.Contains(output, "Installed plugin linapro-tenant-core") {
		t.Fatalf("expected install to auto-initialize workspace and continue, got:\n%s", output)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml")) {
		t.Fatalf("plugin was not installed after auto initialization")
	}
	stage := runGitOutput(t, root, "ls-files", "--stage", "--", plugins.ManagedRootRelativePath)
	if strings.Contains(stage, "160000") {
		t.Fatalf("gitlink still exists after plugins.install auto initialization: %s", stage)
	}
	if fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", ".git")) || fileutil.DirExists(filepath.Join(root, ".git", "modules", "apps", "lina-plugins")) {
		t.Fatalf("submodule metadata was not cleaned")
	}
}

// TestPluginsInstallUpdateAndStatusUseConfiguredSources verifies install,
// update, lock writing, and status output against a local source repository.
func TestPluginsInstallUpdateAndStatusUseConfiguredSources(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "README.md"), "v1\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	var installOut bytes.Buffer
	application := newApp(&installOut, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml")) {
		t.Fatalf("plugin was not installed")
	}
	for _, expected := range []string{
		"Preparing plugin installation for 1 configured item(s)...",
		"Installing 1 plugin(s)...",
		"Synchronizing plugin source official",
		"[1/1] installing plugin linapro-tenant-core from official...",
		"Installed plugin linapro-tenant-core",
	} {
		if !strings.Contains(installOut.String(), expected) {
			t.Fatalf("expected install output to contain %q, got:\n%s", expected, installOut.String())
		}
	}
	if fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", ".git")) || !fileutil.FileExists(plugins.LockPath(root)) {
		t.Fatalf("plugin metadata or lock state is incorrect")
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")

	if err := runPluginsInstall(context.Background(), application, commandInput{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected install to reject existing plugin, got %v", err)
	}

	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")
	if err := runPluginsUpdate(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsUpdate returned error: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml"))
	if err != nil {
		t.Fatalf("read updated plugin manifest: %v", err)
	}
	if !strings.Contains(string(content), "0.2.0") {
		t.Fatalf("plugin was not updated:\n%s", string(content))
	}

	var statusOut bytes.Buffer
	application.stdout = &statusOut
	if err = runPluginsStatus(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsStatus returned error: %v", err)
	}
	output := statusOut.String()
	for _, expected := range []string{
		"Plugin workspace:",
		"Querying configured plugin sources...",
		"Rendering status for 1 configured plugin(s)...",
		"| Plugin",
		"| Source",
		"| Version",
		"| Installed",
		"| Dirty",
		"| Remote",
		"| linapro-tenant-core",
		"| official",
		"| 0.2.0",
		"| true",
		"| current",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected status output to contain %q, got:\n%s", expected, output)
		}
	}

	var filteredStatusOut bytes.Buffer
	application.stdout = &filteredStatusOut
	if err = runPluginsStatus(context.Background(), application, commandInput{Params: map[string]string{"p": "linapro-tenant-core"}}); err != nil {
		t.Fatalf("filtered runPluginsStatus returned error: %v", err)
	}
	filteredOutput := filteredStatusOut.String()
	if !strings.Contains(filteredOutput, "| linapro-tenant-core") || !strings.Contains(filteredOutput, "| current") {
		t.Fatalf("expected filtered status table to include current plugin row, got:\n%s", filteredOutput)
	}
	if strings.Contains(filteredOutput, "remote=current") {
		t.Fatalf("filtered status output must use table columns, got legacy key-value output:\n%s", filteredOutput)
	}
}

// TestPluginsSourceCacheReusesCheckoutWithFetch verifies plugin source sync
// keeps one reusable checkout and refreshes it through later Git fetches.
func TestPluginsSourceCacheReusesCheckoutWithFetch(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	var firstOut bytes.Buffer
	application := newApp(&firstOut, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	cachePath := plugins.SourceCachePath(root, "official")
	if !fileutil.DirExists(filepath.Join(cachePath, ".git")) {
		t.Fatalf("expected reusable source cache at %s", cachePath)
	}
	assertNoLegacyPluginSourceTemps(t, root)

	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")

	var updateOut bytes.Buffer
	application.stdout = &updateOut
	if err := runPluginsUpdate(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsUpdate returned error: %v", err)
	}
	output := updateOut.String()
	if strings.Contains(output, "Cloning into") {
		t.Fatalf("expected update to reuse source cache instead of cloning again, got:\n%s", output)
	}
	content, err := os.ReadFile(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml"))
	if err != nil {
		t.Fatalf("read updated plugin manifest: %v", err)
	}
	if !strings.Contains(string(content), "0.2.0") {
		t.Fatalf("plugin update did not fetch latest source content:\n%s", string(content))
	}
	assertNoLegacyPluginSourceTemps(t, root)
}

// TestPluginsInstallExpandsWildcardItems verifies items ["*"] installs every
// plugin directory under the configured source root.
func TestPluginsInstallExpandsWildcardItems(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "plugins", "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "plugins", "linapro-org-core", "plugin.yaml"), "id: linapro-org-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "plugins", "not-plugin", "README.md"), "ignored\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "source plugins")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \"plugins\"\n      ref: \"master\"\n      items:\n        - \"*\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	for _, pluginID := range []string{"linapro-tenant-core", "linapro-org-core"} {
		if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", pluginID, "plugin.yaml")) {
			t.Fatalf("expected wildcard plugin %s to be installed", pluginID)
		}
	}
	if fileutil.DirExists(filepath.Join(root, "apps", "lina-plugins", "not-plugin")) {
		t.Fatalf("directory without plugin.yaml should not be installed")
	}
}

// TestPluginsInstallWildcardHonorsPluginFilter verifies p=<plugin-id> filters
// the plugins discovered from a wildcard source.
func TestPluginsInstallWildcardHonorsPluginFilter(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "linapro-org-core", "plugin.yaml"), "id: linapro-org-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "source plugins")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"*\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{Params: map[string]string{"p": "linapro-org-core"}}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-org-core", "plugin.yaml")) {
		t.Fatalf("expected filtered wildcard plugin to be installed")
	}
	if fileutil.DirExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core")) {
		t.Fatalf("unexpected unfiltered wildcard plugin installed")
	}
}

// TestRunPluginsUpdateRejectsLocalChangesUnlessForced verifies update protects
// local plugin edits unless the user explicitly passes force=1.
func TestRunPluginsUpdateRejectsLocalChangesUnlessForced(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "local.txt"), "local change\n")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")

	err := runPluginsUpdate(context.Background(), application, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "local changes") {
		t.Fatalf("expected dirty update rejection, got %v", err)
	}
	if err = runPluginsUpdate(context.Background(), application, commandInput{Params: map[string]string{"force": "1"}}); err != nil {
		t.Fatalf("forced update returned error: %v", err)
	}
}

// TestRunPluginsUpdateRejectsCommittedLockDrift verifies update protects
// committed local plugin edits when they differ from the tool lock hash.
func TestRunPluginsUpdateRejectsCommittedLockDrift(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "local.txt"), "committed local change\n")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "local plugin customization")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")

	err := runPluginsUpdate(context.Background(), application, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "local changes") {
		t.Fatalf("expected committed lock drift rejection, got %v", err)
	}
}

// TestPluginsStatusAutoInitializesSubmoduleWithoutPluginWrites verifies status
// initializes the workspace but still avoids plugin directory and lock writes.
func TestPluginsStatusAutoInitializesSubmoduleWithoutPluginWrites(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")
	writeFile(t, filepath.Join(root, ".gitmodules"), `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", ".git"), "gitdir: ../../.git/modules/apps/lina-plugins\n")
	runGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,apps/lina-plugins")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsStatus(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsStatus returned error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Plugin workspace converted to ordinary directory") || !strings.Contains(output, "Plugin workspace: apps/lina-plugins (ordinary)") {
		t.Fatalf("expected status to auto-initialize workspace and continue, got:\n%s", output)
	}
	stage := runGitOutput(t, root, "ls-files", "--stage", "--", plugins.ManagedRootRelativePath)
	if strings.Contains(stage, "160000") {
		t.Fatalf("gitlink still exists after plugins.status auto initialization: %s", stage)
	}
	if fileutil.DirExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core")) {
		t.Fatalf("status must not install plugin directories")
	}
	if fileutil.FileExists(plugins.LockPath(root)) {
		t.Fatalf("status must not write plugin lock state")
	}
}

// newGitRepo creates a minimal repository shaped like a LinaPro checkout.
func newGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	runGit(t, root, "config", "user.email", "linactl@example.com")
	runGit(t, root, "config", "user.name", "linactl")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-core"), 0o755); err != nil {
		t.Fatalf("mkdir lina-core: %v", err)
	}
	return root
}

// runGit executes a Git command in a test repository.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

// runGitOutput executes a Git command and returns its combined output.
func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

// assertNoLegacyPluginSourceTemps verifies source sync no longer creates
// one-shot plugin-source-* directories under temp.
func assertNoLegacyPluginSourceTemps(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "temp"))
	if err != nil {
		t.Fatalf("read temp directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "plugin-source-") {
			t.Fatalf("unexpected legacy plugin source temp directory: %s", entry.Name())
		}
	}
}
