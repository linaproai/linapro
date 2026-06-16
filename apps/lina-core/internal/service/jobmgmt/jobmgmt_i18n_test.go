// This file tests scheduled-job display metadata localization helpers.

package jobmgmt

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
)

// fakeJobmgmtI18nTranslator provides deterministic source-text translations
// for scheduled-job metadata localization tests.
type fakeJobmgmtI18nTranslator struct {
	values map[string]string
}

// TranslateSourceText returns a keyed fake translation or sourceText.
func (f fakeJobmgmtI18nTranslator) TranslateSourceText(_ context.Context, key string, sourceText string) string {
	if value := f.values[key]; value != "" {
		return value
	}
	return sourceText
}

// TestTranslateHandlerSourceTextUsesPluginHandlerKey verifies plugin-owned
// built-in jobs are localized by their stable Jobs handler i18n key.
func TestTranslateHandlerSourceTextUsesPluginHandlerKey(t *testing.T) {
	handlerRef := "plugin:linapro-demo-source/jobs:heartbeat"
	nameKey := jobmeta.HandlerI18nKey(handlerRef, jobNameI18nField)
	descriptionKey := jobmeta.HandlerI18nKey(handlerRef, jobDescriptionI18nField)

	svc := &serviceImpl{
		i18nSvc: fakeJobmgmtI18nTranslator{
			values: map[string]string{
				nameKey:        "源码插件心跳",
				descriptionKey: "执行源码插件注册的内置定时任务。",
			},
		},
	}

	if actual := svc.localizeBuiltinJobName(context.Background(), handlerRef, "Source Plugin Heartbeat", 1); actual != "源码插件心跳" {
		t.Fatalf("expected plugin job name translation, got %q", actual)
	}
	if actual := svc.localizeBuiltinJobDescription(context.Background(), handlerRef, "Runs the plugin built-in job.", 1); actual != "执行源码插件注册的内置定时任务。" {
		t.Fatalf("expected plugin job description translation, got %q", actual)
	}
}

// TestListJobsKeywordMatchesLocalizedBuiltinJobName verifies the management
// list can find built-in plugin jobs by the display name rendered to the user.
func TestListJobsKeywordMatchesLocalizedBuiltinJobName(t *testing.T) {
	ctx := context.Background()
	handlerRef := "plugin:linapro-demo-source/jobs:" + uniqueTestName("source-plugin-echo-inspection")
	sourceName := uniqueTestName("Source Plugin Echo Inspection")
	nameKey := jobmeta.HandlerI18nKey(handlerRef, jobNameI18nField)
	registry := jobhandler.New()
	if err := registry.Register(jobhandler.HandlerDef{
		Ref:          handlerRef,
		DisplayName:  sourceName,
		Description:  "Runs a lightweight source-plugin inspection task for scheduler integration validation.",
		ParamsSchema: `{"type":"object","properties":{}}`,
		Source:       jobmeta.HandlerSourcePlugin,
		PluginID:     "linapro-demo-source",
		Invoke:       func(context.Context, json.RawMessage) (any, error) { return nil, nil },
	}); err != nil {
		t.Fatalf("register plugin job handler: %v", err)
	}
	svc := newTestServiceWithRegistry(t, registry, nil)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{UserId: 1}})
	svc.i18nSvc = fakeJobmgmtI18nTranslator{
		values: map[string]string{
			nameKey: "源码插件回显巡检",
		},
	}

	jobID := syncBuiltinHandlerJob(t, ctx, svc, BuiltinJobDef{
		GroupCode:      "default",
		Name:           sourceName,
		Description:    "Runs a lightweight source-plugin inspection task for scheduler integration validation.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     handlerRef,
		Params:         map[string]any{},
		Timeout:        30 * time.Second,
		Pattern:        "@every 1m",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobmeta.JobStatusEnabled,
	})
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	out, err := svc.ListJobs(ctx, ListJobsInput{Keyword: "源码插件回显巡检", PageNum: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list jobs by localized name: %v", err)
	}
	var matched *JobListItem
	for _, item := range out.List {
		if item != nil && item.SysJob != nil && item.SysJob.Id == jobID {
			matched = item
			break
		}
	}
	if matched == nil {
		t.Fatalf("expected localized keyword to find builtin job %d, got total=%d list=%#v", jobID, out.Total, out.List)
	}
	if matched.Name != "源码插件回显巡检" {
		t.Fatalf("expected localized list name, got %q", matched.Name)
	}

	var stored *struct {
		Name string `orm:"name"`
	}
	if err = dao.SysJob.Ctx(ctx).Fields(dao.SysJob.Columns().Name).Where(do.SysJob{Id: jobID}).Scan(&stored); err != nil {
		t.Fatalf("read stored builtin job name: %v", err)
	}
	if stored == nil || stored.Name != sourceName {
		t.Fatalf("expected source text to stay stored in DB, got %#v", stored)
	}
}
