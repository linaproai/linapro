// This file implements development-only source-plugin upgrade planning and execution.

package sourceupgrade

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	hostsourceupgrade "lina-core/pkg/sourceupgrade"
	"upgrade-source/internal/frameworkupgrade"
)

// newSourcePluginUpgradeService builds the source-plugin upgrade service used by
// the development-only command. Tests replace it with fakes to isolate CLI flow.
var newSourcePluginUpgradeService = func() hostsourceupgrade.Service {
	return hostsourceupgrade.New()
}

// sourcePluginPlan stores the selected source-plugin upgrade items for one command run.
type sourcePluginPlan struct {
	// Target stores the operator-selected plugin target, either one plugin ID or all.
	Target string
	// Items stores all selected source-plugin upgrade items in deterministic order.
	Items []*sourcePluginPlanItem
}

// sourcePluginPlanItem stores one source-plugin status together with the
// derived action that the development-only upgrade command should take.
type sourcePluginPlanItem struct {
	// Status stores the underlying source-plugin upgrade status.
	Status *hostsourceupgrade.SourcePluginStatus
	// SkipReason explains why no upgrade work should run for this plugin.
	SkipReason string
	// RejectReason explains why the command must reject this plugin selection.
	RejectReason string
}

// runSourcePluginUpgrade executes the source-plugin upgrade flow for the given CLI options.
func runSourcePluginUpgrade(ctx context.Context, repoRoot string, options cliOptions) error {
	if strings.TrimSpace(options.repoURL) != "" {
		return gerror.New("scope=source-plugin 不支持 --repo 参数，请仅在框架升级时使用")
	}
	if strings.TrimSpace(options.target) != "" {
		return gerror.New("scope=source-plugin 不支持 --target 参数，请使用 --plugin 指定源码插件")
	}
	if err := frameworkupgrade.ConfigureGoFrameConfig(repoRoot); err != nil {
		return err
	}

	targetPlugin := strings.TrimSpace(options.plugin)
	if targetPlugin == "" {
		return gerror.New("scope=source-plugin 时必须通过 --plugin 指定插件 ID 或 all")
	}

	svc := newSourcePluginUpgradeService()
	statuses, err := svc.ListSourcePluginStatuses(ctx)
	if err != nil {
		return err
	}
	plan, err := buildSourcePluginPlan(statuses, targetPlugin)
	if err != nil {
		return err
	}

	printSourcePluginPlan(plan)

	rejects := collectRejectedSourcePluginPlanItems(plan)
	if len(rejects) > 0 {
		return gerror.New(strings.Join(rejects, "\n"))
	}
	upgradeItems := collectExecutableSourcePluginPlanItems(plan)
	if len(upgradeItems) == 0 {
		fmt.Println("当前选定范围内没有待升级的源码插件。")
		return nil
	}
	if options.dryRun {
		fmt.Println("已启用 dry-run，仅输出源码插件升级计划，不执行 upgrade SQL 或治理切换。")
		return nil
	}

	executedCount := 0
	for _, item := range plan.Items {
		if item == nil || item.Status == nil {
			continue
		}
		if strings.TrimSpace(item.SkipReason) != "" {
			fmt.Printf("- skipped: %s (%s)\n", item.Status.PluginID, item.SkipReason)
			continue
		}
		result, err := svc.UpgradeSourcePlugin(ctx, item.Status.PluginID)
		if err != nil {
			return err
		}
		if result == nil {
			continue
		}
		if result.Executed {
			executedCount++
			fmt.Printf(
				"- upgraded: %s %s -> %s\n",
				result.PluginID,
				displaySourcePluginVersion(result.FromVersion),
				displaySourcePluginVersion(result.ToVersion),
			)
			continue
		}
		fmt.Printf("- skipped: %s (%s)\n", result.PluginID, result.Message)
	}

	fmt.Printf("Source plugin upgrade completed. executed=%d\n", executedCount)
	return nil
}

// buildSourcePluginPlan filters one source-plugin status list according to the
// operator-selected plugin target and derives command actions for every item.
func buildSourcePluginPlan(
	statuses []*hostsourceupgrade.SourcePluginStatus,
	targetPlugin string,
) (*sourcePluginPlan, error) {
	normalizedTarget := strings.TrimSpace(targetPlugin)
	if normalizedTarget == "" {
		return nil, gerror.New("源码插件目标不能为空")
	}

	plan := &sourcePluginPlan{
		Target: normalizedTarget,
		Items:  make([]*sourcePluginPlanItem, 0),
	}
	if normalizedTarget == "all" {
		for _, status := range statuses {
			if item := newSourcePluginPlanItem(status); item != nil {
				plan.Items = append(plan.Items, item)
			}
		}
		return plan, nil
	}

	for _, status := range statuses {
		if status == nil || status.PluginID != normalizedTarget {
			continue
		}
		plan.Items = append(plan.Items, newSourcePluginPlanItem(status))
		return plan, nil
	}
	return nil, gerror.Newf("未找到源码插件: %s", normalizedTarget)
}

// newSourcePluginPlanItem derives one command action for one source-plugin status item.
func newSourcePluginPlanItem(status *hostsourceupgrade.SourcePluginStatus) *sourcePluginPlanItem {
	if status == nil {
		return nil
	}

	item := &sourcePluginPlanItem{Status: status}
	switch {
	case status.Installed != hostsourceupgrade.SourcePluginInstalledYes:
		item.SkipReason = "未安装，跳过。"
	case status.DowngradeDetected:
		item.RejectReason = fmt.Sprintf(
			"源码插件 %s 发现版本 %s 低于当前生效版本 %s，当前不支持降级或回滚。",
			status.PluginID,
			displaySourcePluginVersion(status.DiscoveredVersion),
			displaySourcePluginVersion(status.EffectiveVersion),
		)
	case !status.NeedsUpgrade:
		item.SkipReason = "当前已是生效版本，无需升级。"
	}
	return item
}

// printSourcePluginPlan renders the selected source-plugin upgrade plan for CLI output.
func printSourcePluginPlan(plan *sourcePluginPlan) {
	if plan == nil {
		return
	}

	fmt.Println("Source plugin upgrade plan")
	fmt.Printf("- target: %s\n", plan.Target)
	fmt.Printf("- selected source plugins: %d\n", len(plan.Items))
	if len(plan.Items) == 0 {
		return
	}
	fmt.Println("- items:")
	for _, item := range plan.Items {
		if item == nil || item.Status == nil {
			continue
		}
		action := "upgrade"
		detail := ""
		switch {
		case strings.TrimSpace(item.RejectReason) != "":
			action = "reject"
			detail = item.RejectReason
		case strings.TrimSpace(item.SkipReason) != "":
			action = "skip"
			detail = item.SkipReason
		default:
			detail = fmt.Sprintf(
				"upgrade %s -> %s",
				displaySourcePluginVersion(item.Status.EffectiveVersion),
				displaySourcePluginVersion(item.Status.DiscoveredVersion),
			)
		}
		fmt.Printf(
			"  - %s [%s] current=%s discovered=%s %s\n",
			item.Status.PluginID,
			action,
			displaySourcePluginVersion(item.Status.EffectiveVersion),
			displaySourcePluginVersion(item.Status.DiscoveredVersion),
			detail,
		)
	}
}

// collectRejectedSourcePluginPlanItems returns all reject reasons found in one selected plan.
func collectRejectedSourcePluginPlanItems(plan *sourcePluginPlan) []string {
	if plan == nil {
		return nil
	}

	items := make([]string, 0)
	for _, item := range plan.Items {
		if item == nil || strings.TrimSpace(item.RejectReason) == "" {
			continue
		}
		items = append(items, item.RejectReason)
	}
	return items
}

// collectExecutableSourcePluginPlanItems returns all plan items that should execute upgrade work.
func collectExecutableSourcePluginPlanItems(plan *sourcePluginPlan) []*sourcePluginPlanItem {
	if plan == nil {
		return nil
	}

	items := make([]*sourcePluginPlanItem, 0)
	for _, item := range plan.Items {
		if item == nil || item.Status == nil {
			continue
		}
		if strings.TrimSpace(item.SkipReason) != "" || strings.TrimSpace(item.RejectReason) != "" {
			continue
		}
		items = append(items, item)
	}
	return items
}

// displaySourcePluginVersion renders one possibly-empty source-plugin version for CLI output.
func displaySourcePluginVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return "<empty>"
	}
	return strings.TrimSpace(version)
}
