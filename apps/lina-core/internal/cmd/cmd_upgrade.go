// This file implements the formal framework source-upgrade command.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/frameworkupgrade"
)

// UpgradeInput defines the command-line options for the framework-upgrade command.
type UpgradeInput struct {
	g.Meta  `name:"upgrade" brief:"upgrade framework source from upstream tag and replay host SQL, requires --confirm=upgrade"`
	Confirm string `name:"confirm" brief:"explicit confirmation value, must be 'upgrade'"`
	Repo    string `name:"repo" brief:"upstream framework git repository URL, default uses framework repositoryUrl from metadata.yaml"`
	Target  string `name:"target" brief:"target framework tag or git reference, default uses the latest semantic version tag"`
	DryRun  bool   `name:"dry-run" brief:"only print the upgrade plan without modifying code or database"`
}

// UpgradeOutput carries the command result placeholder.
type UpgradeOutput struct{}

// Upgrade executes one framework source upgrade.
func (m *Main) Upgrade(ctx context.Context, in UpgradeInput) (out *UpgradeOutput, err error) {
	if err = requireCommandConfirmation(upgradeCommandName, in.Confirm); err != nil {
		return nil, err
	}
	fmt.Println("升级前请先确认已经完成代码仓库和数据库备份。")

	svc := frameworkupgrade.New()
	plan, err := svc.BuildPlan(ctx, frameworkupgrade.BuildPlanInput{
		RepoURL:   in.Repo,
		TargetRef: in.Target,
	})
	if err != nil {
		return nil, err
	}
	defer cleanupUpgradePlan(plan)

	printUpgradePlan(plan)
	if !plan.UpgradeNeeded {
		fmt.Println("当前项目已使用相同或更高版本的框架，无需升级。")
		return &UpgradeOutput{}, nil
	}
	if in.DryRun {
		fmt.Println("已启用 dry-run，仅输出升级计划，不执行代码覆盖和 SQL 升级。")
		return &UpgradeOutput{}, nil
	}

	result, err := svc.ExecutePlan(ctx, plan)
	if err != nil {
		return nil, err
	}
	printUpgradeResult(result)
	return &UpgradeOutput{}, nil
}

// printUpgradePlan writes the resolved upgrade plan summary to stdout.
func printUpgradePlan(plan *frameworkupgrade.Plan) {
	if plan == nil {
		return
	}
	fmt.Println("Framework upgrade plan")
	fmt.Printf("- repo root: %s\n", plan.RepoRoot)
	fmt.Printf("- upstream repo: %s\n", plan.RepoURL)
	fmt.Printf("- target ref: %s\n", plan.TargetRef)
	fmt.Printf("- current version: %s\n", plan.CurrentVersion)
	fmt.Printf("- target version: %s\n", plan.TargetVersion)
	fmt.Printf("- latest target sql: %s\n", displayUpgradeValue(plan.LatestSQLFile))
	fmt.Printf("- host sql replay count: %d\n", len(plan.SQLFiles))
	if len(plan.SQLFiles) == 0 {
		return
	}
	fmt.Println("- host sql files:")
	for _, item := range plan.SQLFiles {
		fmt.Printf("  - %s\n", upgradePathBaseName(item))
	}
}

// printUpgradeResult writes the final upgrade result summary to stdout.
func printUpgradeResult(result *frameworkupgrade.ExecuteResult) {
	if result == nil {
		return
	}
	fmt.Println("Framework upgrade completed.")
	fmt.Printf("- target version: %s\n", result.TargetVersion)
	fmt.Printf("- executed sql count: %d\n", len(result.ExecutedSQLFiles))
	if len(result.ExecutedSQLFiles) == 0 {
		return
	}
	fmt.Println("- executed sql files:")
	for _, item := range result.ExecutedSQLFiles {
		fmt.Printf("  - %s\n", item)
	}
}

// cleanupUpgradePlan removes the temporary target checkout after planning or execution completes.
func cleanupUpgradePlan(plan *frameworkupgrade.Plan) {
	if plan == nil || strings.TrimSpace(plan.TargetCloneDir) == "" {
		return
	}
	if err := os.RemoveAll(plan.TargetCloneDir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", gerror.Wrapf(err, "清理升级临时目录失败: %s", plan.TargetCloneDir))
	}
}

// displayUpgradeValue renders one possibly-empty string value for CLI output.
func displayUpgradeValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<empty>"
	}
	return value
}

// upgradePathBaseName returns the basename rendered in CLI output.
func upgradePathBaseName(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	parts := strings.Split(strings.ReplaceAll(path, `\\`, "/"), "/")
	return parts[len(parts)-1]
}
