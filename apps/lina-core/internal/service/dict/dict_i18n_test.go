// This file verifies dictionary-owned localization projection rules.

package dict

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
)

// dictTestTranslator stubs the narrow dictionary translation dependency.
type dictTestTranslator struct {
	locale       string
	translations map[string]string
}

// ResolveLocale returns the configured locale.
func (t dictTestTranslator) ResolveLocale(_ context.Context, _ string) string {
	if t.locale == "" {
		return i18nsvc.DefaultLocale
	}
	return t.locale
}

// Translate returns a configured translation or the caller fallback.
func (t dictTestTranslator) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := t.translations[key]; ok {
		return value
	}
	return fallback
}

// TestLocalizeDictTypeSkipsDefaultLocale verifies editable default-locale values stay untouched.
func TestLocalizeDictTypeSkipsDefaultLocale(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: dictTestTranslator{
			locale: i18nsvc.DefaultLocale,
			translations: map[string]string{
				"dict.sys_status.name": "Status",
			},
		},
	}
	item := &entity.SysDictType{
		Type: "sys_status",
		Name: "状态",
	}

	svc.localizeDictTypeEntity(context.Background(), item)

	if item.Name != "状态" {
		t.Fatalf("expected default-locale dictionary name to remain unchanged, got %q", item.Name)
	}
}

// TestLocalizeDictDataTranslatesNonDefaultLocale verifies dict data uses its owning key convention.
func TestLocalizeDictDataTranslatesNonDefaultLocale(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: dictTestTranslator{
			locale: i18nsvc.EnglishLocale,
			translations: map[string]string{
				"dict.sys_status.enabled.label": "Enabled",
			},
		},
	}
	item := &entity.SysDictData{
		DictType: "sys_status",
		Value:    "enabled",
		Label:    "启用",
	}

	svc.localizeDictDataEntity(context.Background(), item)

	if item.Label != "Enabled" {
		t.Fatalf("expected dictionary label to be localized, got %q", item.Label)
	}
}
