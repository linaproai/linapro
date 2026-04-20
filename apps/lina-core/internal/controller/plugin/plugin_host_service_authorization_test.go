// This file tests plugin host-service authorization response projections.

package plugin

import (
	"testing"
	"time"

	"lina-core/internal/service/jobmeta"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginbridge"
)

// TestBuildHostServicePermissionItemsIncludesCronItems verifies the cron host
// service view includes discovered cron declaration summaries.
func TestBuildHostServicePermissionItemsIncludesCronItems(t *testing.T) {
	specs := []*pluginbridge.HostServiceSpec{
		{
			Service: pluginbridge.HostServiceCron,
			Methods: []string{pluginbridge.HostServiceMethodCronRegister},
		},
		{
			Service: pluginbridge.HostServiceData,
			Methods: []string{pluginbridge.HostServiceMethodDataList},
			Tables:  []string{"sys_plugin_node_state"},
		},
	}
	cronJobs := []pluginsvc.ManagedCronJob{
		{
			Name:           "heartbeat",
			DisplayName:    "Dynamic Heartbeat",
			Description:    "Runs one plugin heartbeat job.",
			Pattern:        "# */10 * * * *",
			Timezone:       "Asia/Shanghai",
			Scope:          jobmeta.JobScopeAllNode,
			Concurrency:    jobmeta.JobConcurrencySingleton,
			MaxConcurrency: 1,
			Timeout:        30 * time.Second,
		},
	}

	items := buildHostServicePermissionItems(
		specs,
		map[string]string{"sys_plugin_node_state": "Plugin node state"},
		cronJobs,
	)
	if len(items) != 2 {
		t.Fatalf("expected 2 host service items, got %d", len(items))
	}

	cronItem := items[0]
	if cronItem.Service != pluginbridge.HostServiceCron {
		t.Fatalf("expected first service to be cron, got %s", cronItem.Service)
	}
	if len(cronItem.CronItems) != 1 {
		t.Fatalf("expected 1 cron item, got %d", len(cronItem.CronItems))
	}
	if cronItem.CronItems[0].Name != "heartbeat" {
		t.Fatalf("expected cron name heartbeat, got %s", cronItem.CronItems[0].Name)
	}
	if cronItem.CronItems[0].Pattern != "# */10 * * * *" {
		t.Fatalf("expected cron pattern preserved, got %s", cronItem.CronItems[0].Pattern)
	}
	if cronItem.CronItems[0].Scope != string(jobmeta.JobScopeAllNode) {
		t.Fatalf("expected cron scope all_node, got %s", cronItem.CronItems[0].Scope)
	}
	if cronItem.CronItems[0].Concurrency != string(jobmeta.JobConcurrencySingleton) {
		t.Fatalf("expected cron concurrency singleton, got %s", cronItem.CronItems[0].Concurrency)
	}

	dataItem := items[1]
	if dataItem.Service != pluginbridge.HostServiceData {
		t.Fatalf("expected second service to be data, got %s", dataItem.Service)
	}
	if len(dataItem.CronItems) != 0 {
		t.Fatalf("expected non-cron service to have no cron items, got %d", len(dataItem.CronItems))
	}
	if len(dataItem.TableItems) != 1 {
		t.Fatalf("expected 1 table item, got %d", len(dataItem.TableItems))
	}
	if dataItem.TableItems[0].Comment != "Plugin node state" {
		t.Fatalf("expected table comment to be preserved, got %s", dataItem.TableItems[0].Comment)
	}
}

// TestBuildHostServicePermissionCronItemsSortsByDisplayName verifies cron
// items stay in a stable alphabetical order for the review UI.
func TestBuildHostServicePermissionCronItemsSortsByDisplayName(t *testing.T) {
	cronItems := buildHostServicePermissionCronItems(
		pluginbridge.HostServiceCron,
		[]pluginsvc.ManagedCronJob{
			{
				Name:        "zeta",
				DisplayName: "Zeta Job",
			},
			{
				Name:        "alpha",
				DisplayName: "Alpha Job",
			},
		},
	)
	if len(cronItems) != 2 {
		t.Fatalf("expected 2 cron items, got %d", len(cronItems))
	}
	if cronItems[0].Name != "alpha" {
		t.Fatalf("expected first cron item alpha, got %s", cronItems[0].Name)
	}
	if cronItems[1].Name != "zeta" {
		t.Fatalf("expected second cron item zeta, got %s", cronItems[1].Name)
	}
}
