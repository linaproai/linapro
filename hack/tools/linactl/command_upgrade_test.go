package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeStableVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{in: "v1.2.3", want: "v1.2.3", wantOK: true},
		{in: "1.2.3", want: "v1.2.3", wantOK: true},
		{in: "v0.5.0", want: "v0.5.0", wantOK: true},
		{in: "v1.2.3-rc.1", wantOK: false},
		{in: "main", wantOK: false},
		{in: "", wantOK: false},
		{in: "v1.2", wantOK: false},
	}
	for _, tc := range cases {
		got, ok := normalizeStableVersion(tc.in)
		if ok != tc.wantOK {
			t.Fatalf("normalizeStableVersion(%q) ok=%v want %v", tc.in, ok, tc.wantOK)
		}
		if ok && got != tc.want {
			t.Fatalf("normalizeStableVersion(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestSelectLatestStableTag(t *testing.T) {
	t.Parallel()
	latest, err := selectLatestStableTag([]string{
		"v0.4.0",
		"v0.5.0",
		"v0.6.0-rc.1",
		"not-a-version",
		"v0.3.9",
		"0.5.1",
	})
	if err != nil {
		t.Fatalf("selectLatestStableTag returned error: %v", err)
	}
	if latest != "v0.5.1" {
		t.Fatalf("selectLatestStableTag=%q want v0.5.1", latest)
	}

	if _, err = selectLatestStableTag([]string{"v1.0.0-rc.1", "nightly"}); err == nil {
		t.Fatal("expected error when no stable tags exist")
	}
}

func TestUpgradeMergeRef(t *testing.T) {
	t.Parallel()
	ref, err := upgradeMergeRef(upgradeTarget{Kind: upgradeTargetTag, Name: "v0.5.0"})
	if err != nil || ref != "v0.5.0" {
		t.Fatalf("tag merge ref=%q err=%v", ref, err)
	}
	ref, err = upgradeMergeRef(upgradeTarget{Kind: upgradeTargetBranch, Name: "main"})
	if err != nil || ref != officialFrameworkRemoteName+"/main" {
		t.Fatalf("branch merge ref=%q err=%v", ref, err)
	}
}

func TestParseRemoteTagList(t *testing.T) {
	t.Parallel()
	tags := parseRemoteTagList(`
abc123	refs/tags/v0.4.0
def456	refs/tags/v0.5.0
ignored
`)
	if len(tags) != 2 || tags[0] != "v0.4.0" || tags[1] != "v0.5.0" {
		t.Fatalf("parseRemoteTagList=%v", tags)
	}
}

func TestCommandRegistryIncludesUpgrade(t *testing.T) {
	t.Parallel()
	registry := commandRegistry()
	spec, ok := registry["upgrade"]
	if !ok {
		t.Fatal("expected upgrade command to be registered")
	}
	if !strings.Contains(spec.Usage, "v=") {
		t.Fatalf("upgrade usage should document v parameter: %s", spec.Usage)
	}
	if strings.Contains(spec.Usage, "remote=") {
		t.Fatalf("upgrade must not accept remote=; got usage %s", spec.Usage)
	}
	if defaultOfficialFrameworkRepoURL != "https://github.com/linaproai/linapro.git" {
		t.Fatalf("official repo URL constant changed unexpectedly: %s", defaultOfficialFrameworkRepoURL)
	}
}

func TestRunFrameworkUpgradeRejectsRemoteParam(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{"remote": "origin"}})
	if err == nil || !strings.Contains(err.Error(), "remote= is not supported") {
		t.Fatalf("expected remote rejection, got %v", err)
	}
}

func TestRunFrameworkUpgradeDefaultLatestStable(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	if err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{}}); err != nil {
		t.Fatalf("runFrameworkUpgrade default: %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "v0.5.0") {
		t.Fatalf("expected merge of latest stable v0.5.0, output:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), officialFrameworkRepoURL) {
		t.Fatalf("expected official repository URL in output:\n%s", stdout.String())
	}
	// Pre-release must not win default selection.
	if strings.Contains(stdout.String(), "v0.6.0-rc.1") {
		t.Fatalf("pre-release should not be selected by default:\n%s", stdout.String())
	}
	content := readFileString(t, filepath.Join(local, "VERSION"))
	if strings.TrimSpace(content) != "0.5.0" {
		t.Fatalf("merged VERSION=%q want 0.5.0", content)
	}
	assertOfficialRemote(t, local)
	assertPluginsPreserved(t, local, "local-0.4.0")
}

func TestRunFrameworkUpgradeSpecificVersion(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	if err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{"v": "0.4.0"}}); err != nil {
		t.Fatalf("runFrameworkUpgrade v=0.4.0: %v\n%s", err, stdout.String())
	}
	content := readFileString(t, filepath.Join(local, "VERSION"))
	if strings.TrimSpace(content) != "0.4.0" {
		t.Fatalf("merged VERSION=%q want 0.4.0", content)
	}
}

func TestRunFrameworkUpgradeMainBranch(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	// Advance official main beyond the stable tags so v=main is distinct.
	official := filepath.Join(filepath.Dir(local), "official.git")
	work := t.TempDir()
	runGit(t, work, "clone", official, "src")
	src := filepath.Join(work, "src")
	runGit(t, src, "config", "user.email", "linactl@example.com")
	runGit(t, src, "config", "user.name", "linactl")
	writeFile(t, filepath.Join(src, "VERSION"), "main-tip\n")
	runGit(t, src, "add", "VERSION")
	runGit(t, src, "commit", "-m", "main tip")
	runGit(t, src, "push", "origin", "HEAD:main")

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	if err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{"v": "main"}}); err != nil {
		t.Fatalf("runFrameworkUpgrade v=main: %v\n%s", err, stdout.String())
	}
	content := readFileString(t, filepath.Join(local, "VERSION"))
	if strings.TrimSpace(content) != "main-tip" {
		t.Fatalf("merged VERSION=%q want main-tip", content)
	}
	if !strings.Contains(stdout.String(), officialFrameworkRemoteName+"/main") {
		t.Fatalf("expected merge ref %s/main in output:\n%s", officialFrameworkRemoteName, stdout.String())
	}
}

func TestRunFrameworkUpgradeIgnoresOriginFork(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	// Point origin at an empty unrelated bare repo so a mistaken origin fetch
	// cannot supply the upgrade tags.
	fork := filepath.Join(filepath.Dir(local), "fork.git")
	runGit(t, filepath.Dir(local), "init", "--bare", "-b", "main", fork)
	runGit(t, local, "remote", "set-url", "origin", fork)

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	if err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{}}); err != nil {
		t.Fatalf("upgrade should use official remote, not origin: %v\n%s", err, stdout.String())
	}
	assertOfficialRemote(t, local)
	content := readFileString(t, filepath.Join(local, "VERSION"))
	if strings.TrimSpace(content) != "0.5.0" {
		t.Fatalf("merged VERSION=%q want 0.5.0 from official remote", content)
	}
}

func TestRunFrameworkUpgradeDoesNotUpdatePlugins(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	if err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{"v": "v0.5.0"}}); err != nil {
		t.Fatalf("runFrameworkUpgrade: %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "plugins are not auto-updated") {
		t.Fatalf("expected plugins-not-auto-updated message, output:\n%s", stdout.String())
	}
	// Host framework upgraded.
	if strings.TrimSpace(readFileString(t, filepath.Join(local, "VERSION"))) != "0.5.0" {
		t.Fatalf("VERSION not upgraded")
	}
	// Official v0.5.0 plugins marker must not replace local plugins content.
	assertPluginsPreserved(t, local, "local-0.4.0")
	// Ensure the merge commit still tracks the pre-upgrade plugins blob, not official.
	pluginBlob := strings.TrimSpace(runGitOutput(t, local, "show", "HEAD:apps/lina-plugins/MARKER"))
	if pluginBlob != "local-0.4.0\n" && pluginBlob != "local-0.4.0" {
		t.Fatalf("HEAD plugins MARKER=%q want local-0.4.0", pluginBlob)
	}
}

func TestRunFrameworkUpgradeRejectsDirtyWorktree(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	writeFile(t, filepath.Join(local, "dirty.txt"), "dirty\n")

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{}})
	if err == nil || !strings.Contains(err.Error(), "worktree is not clean") {
		t.Fatalf("expected dirty worktree error, got %v", err)
	}
}

func TestRunFrameworkUpgradeForceAllowsDirtyWorktree(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	writeFile(t, filepath.Join(local, "dirty.txt"), "dirty\n")

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	if err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{"force": "1", "v": "v0.5.0"}}); err != nil {
		t.Fatalf("force upgrade should continue: %v\n%s", err, stdout.String())
	}
}

func TestRunFrameworkUpgradeRejectsDetachedHEAD(t *testing.T) {
	local, cleanup := setupUpgradeRepos(t)
	defer cleanup()

	runGit(t, local, "checkout", "--detach", "HEAD")

	var stdout bytes.Buffer
	application := newApp(&stdout, &stdout, strings.NewReader(""))
	application.root = local

	err := runFrameworkUpgrade(context.Background(), application, commandInput{Params: map[string]string{"force": "1"}})
	if err == nil || !strings.Contains(err.Error(), "detached HEAD") {
		t.Fatalf("expected detached HEAD error, got %v", err)
	}
}

// setupUpgradeRepos creates a bare "official" repo with tags and a local clone
// whose origin can differ. Tests override officialFrameworkRepoURL to the bare
// official path so no network access is required.
func setupUpgradeRepos(t *testing.T) (local string, cleanup func()) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for upgrade integration tests")
	}

	base := t.TempDir()
	seed := filepath.Join(base, "seed")
	if err := os.MkdirAll(seed, 0o755); err != nil {
		t.Fatalf("mkdir seed: %v", err)
	}
	runGit(t, seed, "init", "-b", "main")
	runGit(t, seed, "config", "user.email", "linactl@example.com")
	runGit(t, seed, "config", "user.name", "linactl")
	writeFile(t, filepath.Join(seed, "VERSION"), "0.1.0\n")
	// Keep go.work so discoverRepoRoot-like checks stay irrelevant; upgrade uses a.root.
	writeFile(t, filepath.Join(seed, "go.work"), "go 1.25.0\n")
	writePluginMarker(t, seed, "official-0.1.0")
	runGit(t, seed, "add", ".")
	runGit(t, seed, "commit", "-m", "init")

	writeFile(t, filepath.Join(seed, "VERSION"), "0.4.0\n")
	writePluginMarker(t, seed, "official-0.4.0")
	runGit(t, seed, "add", ".")
	runGit(t, seed, "commit", "-m", "v0.4.0")
	runGit(t, seed, "tag", "v0.4.0")

	writeFile(t, filepath.Join(seed, "VERSION"), "0.5.0\n")
	writePluginMarker(t, seed, "official-0.5.0")
	runGit(t, seed, "add", ".")
	runGit(t, seed, "commit", "-m", "v0.5.0")
	runGit(t, seed, "tag", "v0.5.0")

	writeFile(t, filepath.Join(seed, "VERSION"), "0.6.0-rc.1\n")
	writePluginMarker(t, seed, "official-0.6.0-rc.1")
	runGit(t, seed, "add", ".")
	runGit(t, seed, "commit", "-m", "v0.6.0-rc.1")
	runGit(t, seed, "tag", "v0.6.0-rc.1")

	// Reset main back to v0.5.0 so main is not ahead of the latest stable tag for default tests.
	runGit(t, seed, "reset", "--hard", "v0.5.0")

	official := filepath.Join(base, "official.git")
	runGit(t, base, "clone", "--bare", seed, official)

	// Local clone still has origin for realism; upgrade must not depend on it.
	local = filepath.Join(base, "local")
	runGit(t, base, "clone", official, local)
	runGit(t, local, "config", "user.email", "linactl@example.com")
	runGit(t, local, "config", "user.name", "linactl")
	// Start app branch from an older point so merges have work to do.
	runGit(t, local, "checkout", "-b", "app", "v0.4.0")
	// Local customization of plugins workspace must survive framework upgrade.
	writePluginMarker(t, local, "local-0.4.0")
	runGit(t, local, "add", "apps/lina-plugins/MARKER")
	runGit(t, local, "commit", "-m", "local plugin customization")

	previousURL := officialFrameworkRepoURL
	officialFrameworkRepoURL = official
	cleanup = func() {
		officialFrameworkRepoURL = previousURL
	}
	t.Cleanup(cleanup)
	return local, cleanup
}

func writePluginMarker(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, "apps", "lina-plugins", "MARKER")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir plugins: %v", err)
	}
	writeFile(t, path, content+"\n")
}

func assertPluginsPreserved(t *testing.T, root string, want string) {
	t.Helper()
	got := strings.TrimSpace(readFileString(t, filepath.Join(root, "apps", "lina-plugins", "MARKER")))
	if got != want {
		t.Fatalf("plugins MARKER=%q want %q (plugins must not be auto-updated)", got, want)
	}
}

func assertOfficialRemote(t *testing.T, root string) {
	t.Helper()
	url := strings.TrimSpace(runGitOutput(t, root, "remote", "get-url", officialFrameworkRemoteName))
	if url != officialFrameworkRepoURL {
		t.Fatalf("official remote URL=%q want %q", url, officialFrameworkRepoURL)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
