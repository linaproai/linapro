// This file provides git-command helpers used by framework-upgrade planning.

package frameworkupgrade

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// detectRepoRoot resolves the current git repository root.
func detectRepoRoot(ctx context.Context) (string, error) {
	output, err := runGitCommand(ctx, "", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", gerror.Wrap(err, "解析当前仓库根目录失败")
	}
	repoRoot := strings.TrimSpace(output)
	if repoRoot == "" {
		return "", gerror.New("当前目录不在 Git 仓库内")
	}
	return repoRoot, nil
}

// listDirtyGitFiles returns all working-tree paths reported by git status --porcelain.
func listDirtyGitFiles(ctx context.Context, repoRoot string) ([]string, error) {
	output, err := runGitCommand(ctx, repoRoot, "status", "--porcelain")
	if err != nil {
		return nil, gerror.Wrap(err, "读取 Git 工作区状态失败")
	}
	return parseGitStatusPorcelain(output), nil
}

// resolveLatestTargetRef resolves the newest semver-like tag from the upstream repository.
func resolveLatestTargetRef(ctx context.Context, repoURL string) (string, error) {
	output, err := runGitCommand(ctx, "", "ls-remote", "--tags", "--refs", repoURL)
	if err != nil {
		return "", gerror.Wrap(err, "读取远端标签失败")
	}
	tags := parseRemoteTagsOutput(output)
	if len(tags) == 0 {
		return "", gerror.New("远端仓库未找到可用的语义化标签")
	}
	sorted, err := sortSemanticVersionsDesc(tags)
	if err != nil {
		return "", err
	}
	return sorted[0], nil
}

// parseRemoteTagsOutput extracts tag names from git ls-remote --tags output.
func parseRemoteTagsOutput(output string) []string {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	seen := make(map[string]struct{})
	items := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) != 2 {
			continue
		}
		ref := strings.TrimSpace(parts[1])
		if !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}
		tag := strings.TrimPrefix(ref, "refs/tags/")
		if _, err := parseSemanticVersion(tag); err != nil {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		items = append(items, tag)
	}
	return items
}

// cloneRepositoryAtRef clones the target repository ref into one temporary directory.
func cloneRepositoryAtRef(ctx context.Context, repoURL string, ref string) (string, error) {
	tempDir, err := os.MkdirTemp("", "lina-framework-upgrade-*")
	if err != nil {
		return "", gerror.Wrap(err, "创建临时升级目录失败")
	}
	if _, err = runGitCommand(ctx, "", "clone", "--depth", "1", "--branch", ref, repoURL, tempDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", gerror.Wrapf(err, "克隆目标版本代码失败: %s@%s", repoURL, ref)
	}
	return filepath.Clean(tempDir), nil
}

// runGitCommand executes one git command and returns stdout as a string.
func runGitCommand(ctx context.Context, dir string, args ...string) (string, error) {
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "git", args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return "", gerror.Wrapf(err, "git %s 失败: %s", strings.Join(args, " "), message)
		}
		return "", gerror.Wrapf(err, "git %s 失败", strings.Join(args, " "))
	}
	return string(output), nil
}

// parseGitStatusPorcelain extracts repo-relative paths from git status --porcelain output.
func parseGitStatusPorcelain(output string) []string {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	lines := strings.Split(output, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) >= 3 {
			items = append(items, strings.TrimSpace(line[3:]))
			continue
		}
		items = append(items, strings.TrimSpace(line))
	}
	return items
}
