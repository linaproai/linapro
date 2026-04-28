// This file tests plugin host-service authorization response projections.

package plugin

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/service/jobmeta"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginbridge"
)

// fakePluginI18nTranslator provides deterministic translation values for
// authorization projection tests.
type fakePluginI18nTranslator struct {
	values        map[string]string
	dynamicValues map[string]string
}

// Translate returns a keyed fake translation or the supplied fallback.
func (f fakePluginI18nTranslator) Translate(_ context.Context, key string, fallback string) string {
	if value := f.values[key]; value != "" {
		return value
	}
	return fallback
}

// TranslateSourceText returns a keyed fake source-text translation or sourceText.
func (f fakePluginI18nTranslator) TranslateSourceText(_ context.Context, key string, sourceText string) string {
	if value := f.values[key]; value != "" {
		return value
	}
	return sourceText
}

// TranslateOrKey returns a keyed fake translation or the key itself.
func (f fakePluginI18nTranslator) TranslateOrKey(_ context.Context, key string) string {
	if value := f.values[key]; value != "" {
		return value
	}
	return key
}

// TranslateWithDefaultLocale returns a keyed fake translation using default locale semantics.
func (f fakePluginI18nTranslator) TranslateWithDefaultLocale(_ context.Context, key string, fallback string) string {
	return f.Translate(context.Background(), key, fallback)
}

// LocalizeError returns the error string for fake localizer tests.
func (f fakePluginI18nTranslator) LocalizeError(_ context.Context, err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TranslateDynamicPluginSourceText returns artifact-local fake translations.
func (f fakePluginI18nTranslator) TranslateDynamicPluginSourceText(
	_ context.Context,
	_ string,
	key string,
	sourceText string,
) string {
	if value := f.dynamicValues[key]; value != "" {
		return value
	}
	return sourceText
}

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

// TestLocalizeManagedCronJobsUsesDynamicPluginArtifactKeys verifies install
// review can use artifact-local translations before a dynamic plugin is enabled.
func TestLocalizeManagedCronJobsUsesDynamicPluginArtifactKeys(t *testing.T) {
	handlerRef, err := pluginbridge.BuildPluginCronHandlerRef("plugin-demo-dynamic", "heartbeat")
	if err != nil {
		t.Fatalf("build handler ref failed: %v", err)
	}
	nameKey := jobmeta.HandlerI18nKey(handlerRef, cronDisplayNameI18nField)
	descriptionKey := jobmeta.HandlerI18nKey(handlerRef, cronDescriptionI18nField)

	localizedJobs := localizeManagedCronJobs(
		context.Background(),
		[]pluginsvc.ManagedCronJob{
			{
				PluginID:    "plugin-demo-dynamic",
				Name:        "heartbeat",
				DisplayName: "Dynamic Plugin Heartbeat",
				Description: "Runs the dynamic plugin built-in job.",
			},
		},
		fakePluginI18nTranslator{
			dynamicValues: map[string]string{
				nameKey:        "动态插件心跳",
				descriptionKey: "通过 Wasm bridge 执行动态插件内置定时任务。",
			},
		},
	)

	if localizedJobs[0].DisplayName != "动态插件心跳" {
		t.Fatalf("expected dynamic artifact display name, got %q", localizedJobs[0].DisplayName)
	}
	if localizedJobs[0].Description != "通过 Wasm bridge 执行动态插件内置定时任务。" {
		t.Fatalf("expected dynamic artifact description, got %q", localizedJobs[0].Description)
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

// TestLocalizeManagedCronJobsUsesSourceTextKeys verifies authorization review
// cron items use the same runtime i18n keys as scheduled-job management.
func TestLocalizeManagedCronJobsUsesSourceTextKeys(t *testing.T) {
	handlerRef, err := pluginbridge.BuildPluginCronHandlerRef("plugin-demo-dynamic", "heartbeat")
	if err != nil {
		t.Fatalf("build handler ref failed: %v", err)
	}
	nameKey := jobmeta.HandlerI18nKey(handlerRef, cronDisplayNameI18nField)
	descriptionKey := jobmeta.HandlerI18nKey(handlerRef, cronDescriptionI18nField)

	sourceJobs := []pluginsvc.ManagedCronJob{
		{
			PluginID:    "plugin-demo-dynamic",
			Name:        "heartbeat",
			DisplayName: "Dynamic Plugin Heartbeat",
			Description: "Runs the dynamic plugin built-in job.",
		},
	}
	localizedJobs := localizeManagedCronJobs(
		context.Background(),
		sourceJobs,
		fakePluginI18nTranslator{
			values: map[string]string{
				nameKey:        "动态插件心跳",
				descriptionKey: "通过 Wasm bridge 执行动态插件内置定时任务。",
			},
		},
	)

	if localizedJobs[0].DisplayName != "动态插件心跳" {
		t.Fatalf("expected localized display name, got %q", localizedJobs[0].DisplayName)
	}
	if localizedJobs[0].Description != "通过 Wasm bridge 执行动态插件内置定时任务。" {
		t.Fatalf("expected localized description, got %q", localizedJobs[0].Description)
	}
	if sourceJobs[0].DisplayName != "Dynamic Plugin Heartbeat" {
		t.Fatalf("source job should not be mutated, got %q", sourceJobs[0].DisplayName)
	}
}
