// This file verifies runtime internationalization configuration loading.

package config

import (
	"context"
	"testing"
)

// TestGetI18nUsesConfigFileValues verifies the i18n section is loaded from the
// default runtime config instead of a standalone locale metadata file.
func TestGetI18nUsesConfigFileValues(t *testing.T) {
	setTestConfigContent(t, `
i18n:
  default: en-US
  enabled: false
  locales:
    - locale: en-US
      nativeName: English
`)

	cfg := New().GetI18n(context.Background())

	if cfg.Default != "en-US" {
		t.Fatalf("expected default locale %q, got %q", "en-US", cfg.Default)
	}
	if cfg.Enabled {
		t.Fatal("expected i18n.enabled=false to be loaded")
	}
	if len(cfg.Locales) != 1 || cfg.Locales[0].Locale != "en-US" {
		t.Fatalf("expected only en-US locale metadata, got %+v", cfg.Locales)
	}
}

// TestGetI18nRejectsMissingConfig verifies i18n defaults must come from the
// already loaded system config rather than a config-template file path.
func TestGetI18nRejectsMissingConfig(t *testing.T) {
	setTestConfigContent(t, `{}`)

	defer func() {
		if recover() == nil {
			t.Fatal("expected missing i18n.default to panic")
		}
	}()
	_ = New().GetI18n(context.Background())
}

// TestGetI18nRejectsMissingDefault verifies the runtime default language is a
// required system config value.
func TestGetI18nRejectsMissingDefault(t *testing.T) {
	setTestConfigContent(t, `
i18n:
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
`)

	defer func() {
		if recover() == nil {
			t.Fatal("expected missing i18n.default to panic")
		}
	}()
	_ = New().GetI18n(context.Background())
}

// TestGetI18nAllowsDefaultOutsideSelectableLocales verifies the default locale
// remains valid even when it is not listed as a switchable locale.
func TestGetI18nAllowsDefaultOutsideSelectableLocales(t *testing.T) {
	setTestConfigContent(t, `
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
`)

	cfg := New().GetI18n(context.Background())

	if cfg.Default != "zh-CN" {
		t.Fatalf("expected configured default locale %q, got %q", "zh-CN", cfg.Default)
	}
	if len(cfg.Locales) != 1 || cfg.Locales[0].Locale != "en-US" {
		t.Fatalf("expected only selectable en-US metadata, got %+v", cfg.Locales)
	}
}
