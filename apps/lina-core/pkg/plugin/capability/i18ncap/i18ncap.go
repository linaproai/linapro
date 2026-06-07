// This file defines the source-plugin visible runtime-translation contract.

package i18ncap

import "context"

// Service defines the runtime translation operations published to source plugins.
type Service interface {
	// GetLocale returns the effective request locale stored in host business context.
	GetLocale(ctx context.Context) string
	// Translate returns the localized value for one runtime i18n key and fallback text.
	Translate(ctx context.Context, key string, fallback string) string
	// FindMessageKeys returns runtime i18n keys under prefix whose localized value matches keyword.
	FindMessageKeys(ctx context.Context, prefix string, keyword string) []string
}
