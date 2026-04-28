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
	dynamicValues map[string]string
	values        map[string]string
}

// TranslateSourceText returns a keyed fake translation or sourceText.
func (f fakeJobmgmtI18nTranslator) TranslateSourceText(_ context.Context, key string, sourceText string) string {
	if value := f.values[key]; value != "" {
		return value
	}
	return sourceText
}

// TranslateDynamicPluginSourceText returns artifact-local fake translations.
func (f fakeJobmgmtI18nTranslator) TranslateDynamicPluginSourceText(
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

// TestTranslateHandlerSourceTextUsesDynamicPluginArtifactFallback verifies
// plugin-owned built-in jobs can be localized from dynamic plugin artifacts
// before the plugin has contributed to the enabled runtime bundle.
func TestTranslateHandlerSourceTextUsesDynamicPluginArtifactFallback(t *testing.T) {
	handlerRef := "plugin:plugin-demo-dynamic/cron:heartbeat"
	nameKey := jobmeta.HandlerI18nKey(handlerRef, jobNameI18nField)
	descriptionKey := jobmeta.HandlerI18nKey(handlerRef, jobDescriptionI18nField)

	svc := &serviceImpl{
		i18nSvc: fakeJobmgmtI18nTranslator{
			dynamicValues: map[string]string{
				nameKey:        "动态插件心跳",
				descriptionKey: "通过 Wasm bridge 执行动态插件内置定时任务。",
			},
		},
	}

	if actual := svc.localizeBuiltinJobName(context.Background(), handlerRef, "Dynamic Plugin Heartbeat", 1); actual != "动态插件心跳" {
		t.Fatalf("expected dynamic plugin job name translation, got %q", actual)
	}
	if actual := svc.localizeBuiltinJobDescription(context.Background(), handlerRef, "Runs the dynamic plugin built-in job.", 1); actual != "通过 Wasm bridge 执行动态插件内置定时任务。" {
		t.Fatalf("expected dynamic plugin job description translation, got %q", actual)
	}
}
