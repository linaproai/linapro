// This file implements the upgrade command that merges upstream framework
// refs into the current local branch.

package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"linactl/internal/plugins"
)

// Official LinaPro framework repository used by make upgrade / linactl upgrade.
// Forks and local origin remotes are intentionally ignored so secondary
// development always upgrades from the canonical upstream.
const (
	defaultOfficialFrameworkRepoURL = "https://github.com/linaproai/linapro.git"
	// officialFrameworkRemoteName is a tool-managed remote that always points at
	// the official framework repository. It is created or corrected on each run.
	officialFrameworkRemoteName = "linapro"
)

// officialFrameworkRepoURL is the repository URL fetched by upgrade.
// Tests may override it to point at a temporary bare repository.
var officialFrameworkRepoURL = defaultOfficialFrameworkRepoURL

// stableReleaseVersionPattern matches optional-v MAJOR.MINOR.PATCH only.
// Pre-release suffixes such as -rc.1 are intentionally excluded from the
// default "latest stable" selection.
var stableReleaseVersionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// upgradeTargetKind distinguishes tag merges from branch merges.
type upgradeTargetKind string

const (
	upgradeTargetTag    upgradeTargetKind = "tag"
	upgradeTargetBranch upgradeTargetKind = "branch"
)

// upgradeTarget describes the resolved Git ref that should be merged.
type upgradeTarget struct {
	Kind upgradeTargetKind
	// Name is the tag name (vX.Y.Z) or branch name (main).
	Name string
}

// runFrameworkUpgrade merges the selected official framework ref into the
// current branch. The handler name is intentionally not runUpgrade so it does
// not collide with db.upgrade's runUpgrade.
func runFrameworkUpgrade(ctx context.Context, a *app, input commandInput) error {
	if strings.TrimSpace(input.Get("remote")) != "" {
		return fmt.Errorf("remote= is not supported; upgrade always fetches %s via remote %q", officialFrameworkRepoURL, officialFrameworkRemoteName)
	}
	force, err := input.Bool("force", false)
	if err != nil {
		return fmt.Errorf("parse force: %w", err)
	}
	versionParam := strings.TrimSpace(input.Get("v"))

	if err = ensureGitAvailable(ctx, a); err != nil {
		return err
	}
	branch, err := currentGitBranch(ctx, a)
	if err != nil {
		return err
	}
	if err = ensureCleanWorktree(ctx, a, force); err != nil {
		return err
	}
	if err = ensureOfficialFrameworkRemote(ctx, a); err != nil {
		return err
	}

	if err = a.runCommand(ctx, commandOptions{Dir: a.root}, "git", "fetch", officialFrameworkRemoteName, "--tags", "--prune"); err != nil {
		return fmt.Errorf("fetch official framework repository %s: %w", officialFrameworkRepoURL, err)
	}

	target, err := resolveUpgradeTarget(ctx, a, versionParam)
	if err != nil {
		return err
	}

	mergeRef, err := upgradeMergeRef(target)
	if err != nil {
		return err
	}
	if err = verifyGitCommitRef(ctx, a, mergeRef); err != nil {
		return err
	}

	preMergeHEAD, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("resolve pre-upgrade HEAD: %w", err)
	}
	preMergeHEAD = strings.TrimSpace(preMergeHEAD)

	fmt.Fprintf(a.stdout, "Upgrading current branch %s from official repository %s by merging %s (%s)\n", branch, officialFrameworkRepoURL, mergeRef, target.Kind)
	fmt.Fprintf(a.stdout, "Preserving local %s (plugins are not auto-updated; use make plugins.update separately)\n", plugins.ManagedRootRelativePath)

	mergeErr := a.runCommand(ctx, commandOptions{Dir: a.root}, "git", "merge", "--no-edit", "--no-commit", mergeRef)
	if mergeErr != nil && !mergeInProgress(ctx, a) {
		return fmt.Errorf("merge %s into %s failed; resolve conflicts manually or run git merge --abort: %w", mergeRef, branch, mergeErr)
	}
	if !mergeInProgress(ctx, a) {
		// Already up to date: nothing to merge.
		fmt.Fprintf(a.stdout, "Framework upgrade complete: branch %s already contains %s (plugins unchanged)\n", branch, mergeRef)
		return nil
	}

	if err = preserveLocalPluginsPath(ctx, a, preMergeHEAD); err != nil {
		_ = a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "merge", "--abort")
		return err
	}

	if hasUnmergedPaths(ctx, a) {
		return fmt.Errorf("merge %s into %s has conflicts outside %s; resolve them manually or run git merge --abort", mergeRef, branch, plugins.ManagedRootRelativePath)
	}

	hasChanges, err := mergeIndexHasChanges(ctx, a)
	if err != nil {
		return err
	}
	if !hasChanges {
		if abortErr := a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "merge", "--abort"); abortErr != nil {
			return fmt.Errorf("abort empty upgrade merge after preserving plugins: %w", abortErr)
		}
		fmt.Fprintf(a.stdout, "Framework upgrade complete: no host framework changes after preserving %s\n", plugins.ManagedRootRelativePath)
		return nil
	}

	if err = a.runCommand(ctx, commandOptions{Dir: a.root}, "git", "commit", "--no-edit"); err != nil {
		return fmt.Errorf("commit upgrade merge of %s into %s failed: %w", mergeRef, branch, err)
	}
	fmt.Fprintf(a.stdout, "Framework upgrade complete: merged %s into %s from %s\n", mergeRef, branch, officialFrameworkRepoURL)
	fmt.Fprintf(a.stdout, "Left %s unchanged; run make plugins.update only when you intentionally want plugin updates\n", plugins.ManagedRootRelativePath)
	return nil
}

// preserveLocalPluginsPath restores apps/lina-plugins to the pre-upgrade state so
// official submodule pointer or plugin tree changes are never applied by upgrade.
func preserveLocalPluginsPath(ctx context.Context, a *app, preMergeHEAD string) error {
	path := plugins.ManagedRootRelativePath
	if gitPathExists(ctx, a, preMergeHEAD, path) {
		if err := a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "checkout", preMergeHEAD, "--", path); err != nil {
			return fmt.Errorf("preserve local %s from pre-upgrade commit: %w", path, err)
		}
		return nil
	}
	// Pre-upgrade tree had no plugins workspace: drop any path introduced by official merge.
	if gitPathExists(ctx, a, "", path) || indexHasPath(ctx, a, path) {
		if err := a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "rm", "-r", "--cached", "-f", "--ignore-unmatch", "--", path); err != nil {
			return fmt.Errorf("exclude official %s introduced by upgrade: %w", path, err)
		}
	}
	return nil
}

// mergeInProgress reports whether a merge is waiting for commit or conflict resolution.
func mergeInProgress(ctx context.Context, a *app) bool {
	_, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "rev-parse", "-q", "--verify", "MERGE_HEAD")
	return err == nil
}

// gitPathExists reports whether path exists in a commit tree (or the index when
// commit is empty).
func gitPathExists(ctx context.Context, a *app, commit string, path string) bool {
	if commit == "" {
		_, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "ls-files", "--error-unmatch", "--", path)
		return err == nil
	}
	_, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "rev-parse", "--verify", commit+":"+path)
	return err == nil
}

// indexHasPath reports whether the index tracks path (including gitlinks).
func indexHasPath(ctx context.Context, a *app, path string) bool {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "ls-files", "--stage", "--", path)
	return err == nil && strings.TrimSpace(output) != ""
}

// hasUnmergedPaths reports whether the index still has merge conflicts.
func hasUnmergedPaths(ctx context.Context, a *app) bool {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "diff", "--name-only", "--diff-filter=U")
	return err == nil && strings.TrimSpace(output) != ""
}

// mergeIndexHasChanges reports whether the merge index differs from HEAD.
func mergeIndexHasChanges(ctx context.Context, a *app) (bool, error) {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "diff", "--cached", "--name-only")
	if err != nil {
		return false, fmt.Errorf("inspect upgrade merge index: %w", err)
	}
	return strings.TrimSpace(output) != "", nil
}

// ensureOfficialFrameworkRemote creates or rewrites the tool-managed remote so
// it always points at the official LinaPro repository URL.
func ensureOfficialFrameworkRemote(ctx context.Context, a *app) error {
	url := strings.TrimSpace(officialFrameworkRepoURL)
	if url == "" {
		return fmt.Errorf("official framework repository URL is empty")
	}
	current, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "remote", "get-url", officialFrameworkRemoteName)
	if err != nil {
		if addErr := a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "remote", "add", officialFrameworkRemoteName, url); addErr != nil {
			return fmt.Errorf("add official remote %q (%s): %w", officialFrameworkRemoteName, url, addErr)
		}
		return nil
	}
	if strings.TrimSpace(current) == url {
		return nil
	}
	if err = a.runCommand(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "remote", "set-url", officialFrameworkRemoteName, url); err != nil {
		return fmt.Errorf("set official remote %q to %s: %w", officialFrameworkRemoteName, url, err)
	}
	return nil
}

// ensureGitAvailable checks that git is on PATH.
func ensureGitAvailable(ctx context.Context, a *app) error {
	if _, err := a.lookPath("git"); err != nil {
		return fmt.Errorf("required tool %q is not available in PATH: %w", "git", err)
	}
	// Confirm the repository root is a Git work tree.
	if _, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("repository root is not a git work tree: %w", err)
	}
	return nil
}

// currentGitBranch returns the current branch name or rejects detached HEAD.
func currentGitBranch(ctx context.Context, a *app) (string, error) {
	branch, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("resolve current branch: %w", err)
	}
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "", fmt.Errorf("detached HEAD is not supported; checkout a branch before running upgrade")
	}
	return branch, nil
}

// ensureCleanWorktree rejects uncommitted changes unless force is set.
func ensureCleanWorktree(ctx context.Context, a *app, force bool) error {
	if force {
		return nil
	}
	status, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("check worktree status: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("worktree is not clean; commit or stash changes, or pass force=1 to continue")
	}
	return nil
}

// resolveUpgradeTarget decides whether to merge a tag or a remote branch from
// the official framework remote.
func resolveUpgradeTarget(ctx context.Context, a *app, versionParam string) (upgradeTarget, error) {
	versionParam = strings.TrimSpace(versionParam)
	if versionParam == "" {
		tags, err := listRemoteGitTags(ctx, a, officialFrameworkRemoteName)
		if err != nil {
			return upgradeTarget{}, err
		}
		latest, err := selectLatestStableTag(tags)
		if err != nil {
			return upgradeTarget{}, err
		}
		return upgradeTarget{Kind: upgradeTargetTag, Name: latest}, nil
	}
	if tag, ok := normalizeStableVersion(versionParam); ok {
		return upgradeTarget{Kind: upgradeTargetTag, Name: tag}, nil
	}
	if strings.Contains(versionParam, "/") || strings.Contains(versionParam, "..") {
		return upgradeTarget{}, fmt.Errorf("invalid branch name %q; pass a simple branch name such as main", versionParam)
	}
	if versionParam == officialFrameworkRemoteName {
		return upgradeTarget{}, fmt.Errorf("invalid v=%q; use a version tag or branch name such as main", versionParam)
	}
	return upgradeTarget{Kind: upgradeTargetBranch, Name: versionParam}, nil
}

// upgradeMergeRef builds the Git ref passed to git merge.
func upgradeMergeRef(target upgradeTarget) (string, error) {
	switch target.Kind {
	case upgradeTargetTag:
		if target.Name == "" {
			return "", fmt.Errorf("upgrade tag is empty")
		}
		return target.Name, nil
	case upgradeTargetBranch:
		if target.Name == "" {
			return "", fmt.Errorf("upgrade branch is empty")
		}
		return officialFrameworkRemoteName + "/" + target.Name, nil
	default:
		return "", fmt.Errorf("unknown upgrade target kind %q", target.Kind)
	}
}

// verifyGitCommitRef ensures the resolved ref points at a commit.
func verifyGitCommitRef(ctx context.Context, a *app, ref string) error {
	if _, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "rev-parse", "--verify", ref+"^{commit}"); err != nil {
		return fmt.Errorf("upgrade target %q was not found after fetch from %s: %w", ref, officialFrameworkRepoURL, err)
	}
	return nil
}

// listRemoteGitTags returns tag names advertised by the remote.
// Local-only tags are ignored so "latest stable" always reflects the remote.
func listRemoteGitTags(ctx context.Context, a *app, remote string) ([]string, error) {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Quiet: true}, "git", "ls-remote", "--tags", "--refs", remote)
	if err != nil {
		return nil, fmt.Errorf("list tags on remote %q: %w", remote, err)
	}
	return parseRemoteTagList(output), nil
}

// parseRemoteTagList extracts tag names from git ls-remote --tags output.
func parseRemoteTagList(output string) []string {
	lines := strings.Split(output, "\n")
	tags := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: <hash>\trefs/tags/<name>
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := fields[len(fields)-1]
		const prefix = "refs/tags/"
		if !strings.HasPrefix(ref, prefix) {
			continue
		}
		tag := strings.TrimPrefix(ref, prefix)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// normalizeStableVersion accepts vX.Y.Z or X.Y.Z and returns the canonical tag.
func normalizeStableVersion(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !stableReleaseVersionPattern.MatchString(raw) {
		return "", false
	}
	if strings.HasPrefix(raw, "v") || strings.HasPrefix(raw, "V") {
		return "v" + raw[1:], true
	}
	return "v" + raw, true
}

// selectLatestStableTag picks the highest stable semver tag from a list.
func selectLatestStableTag(tags []string) (string, error) {
	var (
		bestTag string
		best    stableVersion
		found   bool
	)
	for _, tag := range tags {
		normalized, ok := normalizeStableVersion(tag)
		if !ok {
			continue
		}
		parsed, ok := parseStableVersion(normalized)
		if !ok {
			continue
		}
		if !found || parsed.greaterThan(best) {
			best = parsed
			bestTag = normalized
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("no stable release tags found on %s; pass v=<version> or v=main", officialFrameworkRepoURL)
	}
	return bestTag, nil
}

// stableVersion is a parsed MAJOR.MINOR.PATCH triple.
type stableVersion struct {
	Major int
	Minor int
	Patch int
}

func parseStableVersion(tag string) (stableVersion, bool) {
	normalized, ok := normalizeStableVersion(tag)
	if !ok {
		return stableVersion{}, false
	}
	parts := strings.Split(strings.TrimPrefix(normalized, "v"), ".")
	if len(parts) != 3 {
		return stableVersion{}, false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return stableVersion{}, false
	}
	return stableVersion{Major: major, Minor: minor, Patch: patch}, true
}

func (v stableVersion) greaterThan(other stableVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}
