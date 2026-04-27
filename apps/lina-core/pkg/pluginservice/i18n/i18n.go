// Package i18n exposes a narrowed host runtime-translation contract to source
// plugins without requiring them to import host-internal service packages.
package i18n

import (
	"context"
	"sort"
	"strings"

	internali18n "lina-core/internal/service/i18n"
)

// Service defines the runtime translation operations published to source plugins.
type Service interface {
	// GetLocale returns the effective request locale stored in host business context.
	GetLocale(ctx context.Context) string
	// Translate returns the localized value for one runtime i18n key and fallback text.
	Translate(ctx context.Context, key string, fallback string) string
	// FindMessageKeys returns runtime i18n keys under prefix whose localized value
	// contains keyword in the current request language.
	FindMessageKeys(ctx context.Context, prefix string, keyword string) []string
}

// serviceAdapter bridges the internal i18n service into the published plugin contract.
type serviceAdapter struct {
	service hostRuntimeTranslator
}

// hostRuntimeTranslator defines the host i18n capabilities used by the plugin adapter.
type hostRuntimeTranslator interface {
	// GetLocale returns the locale stored in request business context.
	GetLocale(ctx context.Context) string
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
	// ExportMessages exports flat runtime messages for one locale.
	ExportMessages(ctx context.Context, locale string, raw bool) internali18n.MessageExportOutput
}

// New creates and returns the published i18n service adapter.
func New() Service {
	return &serviceAdapter{service: internali18n.New()}
}

// GetLocale returns the effective request locale stored in host business context.
func (s *serviceAdapter) GetLocale(ctx context.Context) string {
	if s == nil || s.service == nil {
		return internali18n.DefaultLocale
	}
	return s.service.GetLocale(ctx)
}

// Translate returns the localized value for one runtime i18n key and fallback text.
func (s *serviceAdapter) Translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.service == nil {
		return fallback
	}
	return s.service.Translate(ctx, key, fallback)
}

// FindMessageKeys returns runtime i18n keys under prefix whose localized value
// contains keyword in the current request language.
func (s *serviceAdapter) FindMessageKeys(ctx context.Context, prefix string, keyword string) []string {
	if s == nil || s.service == nil {
		return []string{}
	}

	trimmedKeyword := strings.TrimSpace(keyword)
	if trimmedKeyword == "" {
		return []string{}
	}
	normalizedKeyword := strings.ToLower(trimmedKeyword)
	trimmedPrefix := strings.TrimSpace(prefix)

	messages := s.service.ExportMessages(ctx, s.service.GetLocale(ctx), false).Messages
	keys := make([]string, 0)
	for key, value := range messages {
		if trimmedPrefix != "" && !strings.HasPrefix(key, trimmedPrefix) {
			continue
		}
		if strings.Contains(strings.ToLower(value), normalizedKeyword) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
