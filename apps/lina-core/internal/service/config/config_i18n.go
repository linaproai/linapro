// This file defines runtime internationalization configuration loading.

package config

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// I18nConfig holds runtime internationalization settings loaded from config.yaml.
type I18nConfig struct {
	Default string             `json:"default"` // Default is used when a request locale is missing or unsupported.
	Enabled bool               `json:"enabled"` // Enabled controls whether multiple runtime languages may be selected.
	Locales []I18nLocaleConfig `json:"locales"` // Locales defines enabled runtime locale order and native names.
}

// I18nLocaleConfig stores metadata for one enabled runtime locale.
type I18nLocaleConfig struct {
	Locale     string `json:"locale"`     // Locale is the canonical locale code.
	NativeName string `json:"nativeName"` // NativeName is the locale's own display name.
}

// GetI18n reads runtime internationalization config from configuration file.
func (s *serviceImpl) GetI18n(ctx context.Context) *I18nConfig {
	return cloneI18nConfig(processStaticConfigCaches.i18n.load(func() *I18nConfig {
		cfg := &I18nConfig{}
		mustScanConfig(ctx, "i18n", cfg)
		normalizeAndValidateI18nConfig("runtime config", cfg)
		return cfg
	}))
}

// normalizeAndValidateI18nConfig trims runtime i18n config and rejects missing
// required fields that would otherwise reintroduce implicit language defaults.
func normalizeAndValidateI18nConfig(source string, cfg *I18nConfig) {
	if cfg == nil {
		panic(gerror.Newf("%s i18n config cannot be nil", source))
	}
	cfg.Default = strings.TrimSpace(cfg.Default)
	if cfg.Default == "" {
		panic(gerror.Newf("%s i18n.default cannot be empty", source))
	}
	cfg.Locales = normalizeI18nLocaleConfigs(cfg.Locales)
}

// normalizeI18nLocaleConfigs trims duplicate locale metadata while preserving declaration order.
func normalizeI18nLocaleConfigs(locales []I18nLocaleConfig) []I18nLocaleConfig {
	if len(locales) == 0 {
		return nil
	}

	normalized := make([]I18nLocaleConfig, 0, len(locales))
	seenLocales := make(map[string]struct{}, len(locales))
	for _, item := range locales {
		locale := strings.TrimSpace(item.Locale)
		if locale == "" {
			continue
		}
		if _, ok := seenLocales[locale]; ok {
			continue
		}
		seenLocales[locale] = struct{}{}
		normalized = append(normalized, I18nLocaleConfig{
			Locale:     locale,
			NativeName: strings.TrimSpace(item.NativeName),
		})
	}
	return normalized
}
