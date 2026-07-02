// This file keeps runtime i18n parsing and cache reset helpers scoped to tests.

package i18n

import (
	"strings"

	"lina-core/pkg/i18nresource"
)

// normalizeAcceptLanguage converts an Accept-Language header into the first valid locale tag.
func normalizeAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		languageTag := strings.TrimSpace(strings.Split(part, ";")[0])
		if locale := normalizeLocale(languageTag); locale != "" {
			return locale
		}
	}
	return ""
}

// invalidateRuntimeBundleCache clears all runtime i18n bundle state for tests.
func invalidateRuntimeBundleCache() {
	runtimeBundleCache.invalidate(InvalidateScope{})
	resetRuntimeLocaleCache()
}

// invalidateRuntimeLocaleCache clears the cached locale descriptors for tests.
func invalidateRuntimeLocaleCache() {
	resetRuntimeLocaleCache()
}

// parseLocaleJSON unmarshals one locale JSON file into a flat message catalog for tests.
func parseLocaleJSON(content []byte) map[string]string {
	bundle, err := i18nresource.ParseCatalog(content, i18nresource.ValueModeStringifyScalars)
	if err != nil {
		return map[string]string{}
	}
	return bundle
}

// lookupMessageString retrieves one string message by dotted key path for tests.
func lookupMessageString(messages map[string]interface{}, key string) (string, bool) {
	if len(messages) == 0 {
		return "", false
	}

	current := interface{}(messages)
	for _, segment := range strings.Split(strings.TrimSpace(key), ".") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return "", false
		}
		nested, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}
		current, ok = nested[segment]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}
