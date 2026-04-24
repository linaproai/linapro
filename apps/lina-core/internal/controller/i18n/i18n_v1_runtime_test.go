// This file verifies the runtime i18n controller endpoints.

package i18n

import (
	"context"
	"strings"
	"testing"

	v1 "lina-core/api/i18n/v1"
	"lina-core/internal/model"
	i18nsvc "lina-core/internal/service/i18n"

	"github.com/gogf/gf/v2/os/gctx"
)

// TestRuntimeMessagesUsesExplicitLangOverride verifies that the runtime
// messages endpoint honors the explicit lang query parameter.
func TestRuntimeMessagesUsesExplicitLangOverride(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{i18nSvc: i18nsvc.New()}
	ctx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.DefaultLocale},
	)

	res, err := controller.RuntimeMessages(ctx, &v1.RuntimeMessagesReq{Lang: i18nsvc.EnglishLocale})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Locale != i18nsvc.EnglishLocale {
		t.Fatalf("expected runtime locale %q, got %q", i18nsvc.EnglishLocale, res.Locale)
	}

	actual, ok := lookupRuntimeMessage(res.Messages, "menu.dashboard.title")
	if !ok {
		t.Fatal("expected menu.dashboard.title to exist in runtime messages")
	}
	if actual != "Workbench" {
		t.Fatalf("expected English runtime message %q, got %q", "Workbench", actual)
	}
}

// TestRuntimeLocalesReturnsLocalizedDescriptors verifies that the runtime
// locale endpoint returns localized display names with stable native names.
func TestRuntimeLocalesReturnsLocalizedDescriptors(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{i18nSvc: i18nsvc.New()}
	ctx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.DefaultLocale},
	)

	res, err := controller.RuntimeLocales(ctx, &v1.RuntimeLocalesReq{Lang: i18nsvc.EnglishLocale})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Locale != i18nsvc.EnglishLocale {
		t.Fatalf("expected runtime locale %q, got %q", i18nsvc.EnglishLocale, res.Locale)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 locale descriptors, got %d", len(res.Items))
	}

	zhLocale, ok := findRuntimeLocale(res.Items, i18nsvc.DefaultLocale)
	if !ok {
		t.Fatalf("expected locale %q in runtime locale list", i18nsvc.DefaultLocale)
	}
	if zhLocale.Name != "Chinese (Simplified)" || zhLocale.NativeName != "简体中文" || !zhLocale.IsDefault {
		t.Fatalf("unexpected zh-CN locale descriptor: %+v", zhLocale)
	}
}

// lookupRuntimeMessage reads one dotted runtime message path from the nested response payload.
func lookupRuntimeMessage(messages map[string]interface{}, key string) (string, bool) {
	current := interface{}(messages)
	for _, segment := range strings.Split(strings.TrimSpace(key), ".") {
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}
		current, ok = currentMap[segment]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}

// findRuntimeLocale locates one locale descriptor by locale code.
func findRuntimeLocale(items []v1.RuntimeLocaleItem, locale string) (v1.RuntimeLocaleItem, bool) {
	for _, item := range items {
		if item.Locale == locale {
			return item, true
		}
	}
	return v1.RuntimeLocaleItem{}, false
}
