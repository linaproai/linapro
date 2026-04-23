// This file boots the development-only framework upgrade tool invoked by
// `make upgrade`.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/errors/gerror"

	"upgrade-framework/internal/frameworkupgrade"
)

// upgradeCommandName stores the confirmation token and display name for the tool.
const upgradeCommandName = "upgrade"

// cliOptions stores the parsed upgrade command-line options.
type cliOptions struct {
	confirm string
	repoURL string
	target  string
	dryRun  bool
}

// main parses flags, validates confirmation, and executes the upgrade plan.
func main() {
	if err := run(context.Background(), parseOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// parseOptions parses the development-only upgrade flags.
func parseOptions() cliOptions {
	var options cliOptions

	flag.StringVar(&options.confirm, "confirm", "", "explicit confirmation value, must be 'upgrade'")
	flag.StringVar(&options.repoURL, "repo", "", "upstream framework git repository URL, default uses apps/lina-core/hack/config.yaml")
	flag.StringVar(&options.target, "target", "", "target framework tag or git reference, default uses the latest semantic version tag")
	flag.BoolVar(&options.dryRun, "dry-run", false, "only print the upgrade plan without modifying code or database")
	flag.Parse()
	return options
}

// run executes the full upgrade tool flow.
func run(ctx context.Context, options cliOptions) error {
	if err := requireCommandConfirmation(options.confirm); err != nil {
		return err
	}
	fmt.Println("升级前请先确认已经完成代码仓库和数据库备份。")

	svc := frameworkupgrade.New()
	plan, err := svc.BuildPlan(ctx, frameworkupgrade.BuildPlanInput{
		RepoURL:   options.repoURL,
		TargetRef: options.target,
	})
	if err != nil {
		return err
	}
	defer cleanupUpgradePlan(plan)

	printUpgradePlan(plan)
	if !plan.UpgradeNeeded {
		fmt.Println("当前项目已使用相同或更高版本的框架，无需升级。")
		return nil
	}
	if options.dryRun {
		fmt.Println("已启用 dry-run，仅输出升级计划，不执行代码覆盖和 SQL 升级。")
		return nil
	}

	result, err := svc.ExecutePlan(ctx, plan)
	if err != nil {
		return err
	}
	printUpgradeResult(result)
	return nil
}

// requireCommandConfirmation validates the explicit confirmation token.
func requireCommandConfirmation(confirmValue string) error {
	if strings.TrimSpace(confirmValue) == upgradeCommandName {
		return nil
	}
	return gerror.Newf(
		"命令 %s 涉及敏感升级或数据库操作，必须显式确认后才能执行。请使用 make %s confirm=%s 或 go run ./hack/upgrade-framework --confirm=%s",
		upgradeCommandName,
		upgradeCommandName,
		upgradeCommandName,
		upgradeCommandName,
	)
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
