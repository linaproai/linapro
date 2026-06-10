// This file tests scheduled-job display metadata localization helpers.

package jobmgmt

import (
	"context"
	"testing"

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
