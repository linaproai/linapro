// This file builds and executes framework-upgrade plans.

package frameworkupgrade

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

// sqlExecutor executes one SQL statement during framework upgrade.
type sqlExecutor func(ctx context.Context, sql string) error

// BuildPlan resolves the target release and calculates the pending upgrade work.
func (s *serviceImpl) BuildPlan(ctx context.Context, input BuildPlanInput) (*Plan, error) {
	repoRoot, err := detectRepoRoot(ctx)
	if err != nil {
		return nil, err
	}
	if err = ConfigureGoFrameConfig(repoRoot); err != nil {
		return nil, err
	}
	dirtyGitFiles, err := listDirtyGitFiles(ctx, repoRoot)
	if err != nil {
		return nil, err
	}
	if len(dirtyGitFiles) > 0 {
		return &Plan{RepoRoot: repoRoot, DirtyGitFiles: dirtyGitFiles}, gerror.Newf(
			"检测到当前 Git 工作区存在未提交修改，请先提交或 stash 后再升级: %s",
			strings.Join(dirtyGitFiles, ", "),
		)
	}

	currentFramework, err := readCurrentUpgradeMetadata(repoRoot)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(currentFramework.Version) == "" {
		return nil, gerror.New("当前项目 apps/lina-core/hack/config.yaml 缺少 frameworkUpgrade.version，无法执行升级")
	}
	if _, err = parseSemanticVersion(currentFramework.Version); err != nil {
		return nil, err
	}

	repoURL := resolveRepoURL(strings.TrimSpace(input.RepoURL), currentFramework.RepositoryURL)
	if repoURL == "" {
		return nil, gerror.New("未配置上游框架仓库地址，请通过 --repo 显式传入")
	}
	targetRef := strings.TrimSpace(input.TargetRef)
	if targetRef == "" {
		targetRef, err = resolveLatestTargetRef(ctx, repoURL)
		if err != nil {
			return nil, err
		}
	}

	targetCloneDir, err := cloneRepositoryAtRef(ctx, repoURL, targetRef)
	if err != nil {
		return nil, err
	}
	targetFramework, err := readTargetUpgradeMetadata(targetCloneDir)
	if err != nil {
		_ = os.RemoveAll(targetCloneDir)
		return nil, err
	}

	compareResult, err := compareSemanticVersions(currentFramework.Version, targetFramework.Version)
	if err != nil {
		_ = os.RemoveAll(targetCloneDir)
		return nil, err
	}

	targetSQLFiles, err := scanTargetSQLFiles(targetCloneDir)
	if err != nil {
		_ = os.RemoveAll(targetCloneDir)
		return nil, err
	}

	plan := &Plan{
		RepoRoot:         repoRoot,
		RepoURL:          repoURL,
		TargetRef:        targetRef,
		TargetCloneDir:   targetCloneDir,
		CurrentFramework: currentFramework,
		TargetFramework:  *targetFramework,
		CurrentVersion:   currentFramework.Version,
		TargetVersion:    targetFramework.Version,
		LatestSQLFile:    fileBaseName(latestSQLFile(targetSQLFiles)),
		SQLFiles:         targetSQLFiles,
		UpgradeNeeded:    compareResult < 0,
	}
	return plan, nil
}

// ExecutePlan runs the resolved framework-upgrade plan.
func (s *serviceImpl) ExecutePlan(ctx context.Context, plan *Plan) (*ExecuteResult, error) {
	if plan == nil {
		return nil, gerror.New("升级计划不能为空")
	}
	if err := ConfigureGoFrameConfig(plan.RepoRoot); err != nil {
		return nil, err
	}

	result := &ExecuteResult{
		TargetVersion: plan.TargetVersion,
		UpgradeNeeded: plan.UpgradeNeeded,
	}
	if !plan.UpgradeNeeded {
		return result, nil
	}
	if err := syncRepositoryFromTarget(plan.TargetCloneDir, plan.RepoRoot); err != nil {
		return nil, err
	}

	executedSQLFiles, err := executeUpgradeSQLFiles(ctx, plan.SQLFiles)
	if err != nil {
		return nil, err
	}
	result.ExecutedSQLFiles = executedSQLFiles
	return result, nil
}

// executeUpgradeSQLFiles runs the provided host SQL files in order.
func executeUpgradeSQLFiles(ctx context.Context, files []string) ([]string, error) {
	return executeUpgradeSQLFilesWithExecutor(ctx, files, func(ctx context.Context, sql string) error {
		_, err := g.DB().Exec(ctx, sql)
		return err
	})
}

// executeUpgradeSQLFilesWithExecutor reads SQL files and delegates execution to
// the provided executor, which allows unit tests to verify replay order without
// touching a real database.
func executeUpgradeSQLFilesWithExecutor(ctx context.Context, files []string, executor sqlExecutor) ([]string, error) {
	executed := make([]string, 0, len(files))
	for _, file := range files {
		sqlContent := gfile.GetContents(file)
		if strings.TrimSpace(sqlContent) == "" {
			continue
		}
		fmt.Printf("Executing framework upgrade SQL file: %s\n", fileBaseName(file))
		if err := executor(ctx, sqlContent); err != nil {
			return executed, gerror.Wrapf(err, "执行升级 SQL 文件失败: %s", fileBaseName(file))
		}
		executed = append(executed, fileBaseName(file))
	}
	return executed, nil
}

// scanTargetSQLFiles scans target-release host SQL files in stable lexical order.
func scanTargetSQLFiles(targetCloneDir string) ([]string, error) {
	sqlDir := filepath.Join(targetCloneDir, "apps", "lina-core", "manifest", "sql")
	if !gfile.Exists(sqlDir) {
		return nil, nil
	}
	files, err := gfile.ScanDirFile(sqlDir, "*.sql", false)
	if err != nil {
		return nil, gerror.Wrap(err, "扫描目标版本宿主 SQL 文件失败")
	}
	sort.Strings(files)
	return files, nil
}

// syncRepositoryFromTarget copies target-release files into the current project repository.
func syncRepositoryFromTarget(targetCloneDir string, repoRoot string) error {
	return filepath.WalkDir(targetCloneDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, err := filepath.Rel(targetCloneDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		if relPath == "." {
			return nil
		}
		if shouldSkipSyncPath(relPath) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(repoRoot, relPath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		return copyFileWithMode(path, targetPath)
	})
}

// shouldSkipSyncPath reports whether one path inside the target clone must be skipped.
func shouldSkipSyncPath(relPath string) bool {
	normalized := strings.TrimSpace(filepath.ToSlash(relPath))
	if normalized == ".git" || strings.HasPrefix(normalized, ".git/") {
		return true
	}
	return false
}

// copyFileWithMode copies one file and preserves the source file mode bits.
func copyFileWithMode(sourcePath string, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return gerror.Wrapf(err, "创建升级目标目录失败: %s", filepath.Dir(targetPath))
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return gerror.Wrapf(err, "打开升级源文件失败: %s", sourcePath)
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	info, err := sourceFile.Stat()
	if err != nil {
		return gerror.Wrapf(err, "读取升级源文件属性失败: %s", sourcePath)
	}
	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return gerror.Wrapf(err, "打开升级目标文件失败: %s", targetPath)
	}
	defer func() {
		_ = targetFile.Close()
	}()
	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		return gerror.Wrapf(err, "复制升级文件失败: %s", sourcePath)
	}
	if err = os.Chmod(targetPath, info.Mode()); err != nil {
		return gerror.Wrapf(err, "更新升级目标文件权限失败: %s", targetPath)
	}
	return nil
}

// latestSQLFile returns the last SQL file path from one sorted file list.
func latestSQLFile(sqlFiles []string) string {
	if len(sqlFiles) == 0 {
		return ""
	}
	return sqlFiles[len(sqlFiles)-1]
}

// fileBaseName returns one filesystem basename in slash-agnostic form.
func fileBaseName(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.Base(path)
}

// resolveRepoURL resolves the effective upstream repository URL.
func resolveRepoURL(input string, fallback string) string {
	if strings.TrimSpace(input) != "" {
		return strings.TrimSpace(input)
	}
	return strings.TrimSpace(fallback)
}
