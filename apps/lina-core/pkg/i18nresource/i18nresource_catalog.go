// This file implements JSON catalog parsing, flattening, filtering, and key normalization.

package i18nresource

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/logger"
)

var keyInvalidCharsPattern = regexp.MustCompile(`[^A-Za-z0-9_]+`)

// ParseCatalog parses one JSON locale resource into a flat dotted-key catalog.
func ParseCatalog(content []byte, valueMode ValueMode) (map[string]string, error) {
	result := make(map[string]interface{})
	if len(content) == 0 {
		return map[string]string{}, nil
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	flatMessages := make(map[string]string)
	if err := flattenCatalogValue("", result, flatMessages, normalizeValueMode(valueMode)); err != nil {
		return nil, err
	}
	return flatMessages, nil
}

// SanitizeKeyPart normalizes one dotted-key segment for structured i18n keys.
func SanitizeKeyPart(part string) string {
	trimmedPart := strings.TrimSpace(part)
	trimmedPart = strings.Trim(trimmedPart, "{}")
	trimmedPart = strings.ReplaceAll(trimmedPart, "-", "_")
	trimmedPart = keyInvalidCharsPattern.ReplaceAllString(trimmedPart, "_")
	trimmedPart = strings.Trim(trimmedPart, "_")
	if trimmedPart == "" {
		return "empty"
	}
	return trimmedPart
}

// filterPluginBundle applies namespace and key filters to one plugin-owned bundle.
func (l ResourceLoader) filterPluginBundle(ctx context.Context, pluginID string, source map[string]string) map[string]string {
	return l.filterCatalog(ctx, pluginID, source)
}

// filterCatalog applies the loader's plugin scope and key filter.
func (l ResourceLoader) filterCatalog(ctx context.Context, pluginID string, source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	target := make(map[string]string, len(source))
	namespacePrefix := ""
	if l.pluginScope() == PluginScopeRestrictedToPluginNamespace && strings.TrimSpace(pluginID) != "" {
		namespacePrefix = "plugins." + SanitizeKeyPart(pluginID) + "."
	}
	for key, value := range source {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		if namespacePrefix != "" && !strings.HasPrefix(trimmedKey, namespacePrefix) {
			logger.Warningf(ctx, "ignore i18n resource key outside plugin namespace plugin=%s key=%s", pluginID, trimmedKey)
			continue
		}
		if l.KeyFilter != nil && !l.KeyFilter(trimmedKey) {
			logger.Warningf(ctx, "ignore i18n resource key rejected by loader filter plugin=%s key=%s", pluginID, trimmedKey)
			continue
		}
		target[trimmedKey] = value
	}
	return target
}

// flattenCatalogValue flattens one nested JSON value into dotted keys.
func flattenCatalogValue(prefix string, value interface{}, output map[string]string, valueMode ValueMode) error {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(typedValue))
		for key := range typedValue {
			if strings.TrimSpace(key) == "" {
				continue
			}
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			left := strings.TrimSpace(keys[i])
			right := strings.TrimSpace(keys[j])
			if left == right {
				return keys[i] < keys[j]
			}
			return left < right
		})
		for _, key := range keys {
			trimmedKey := strings.TrimSpace(key)
			nextPrefix := trimmedKey
			if prefix != "" {
				nextPrefix = prefix + "." + trimmedKey
			}
			if err := flattenCatalogValue(nextPrefix, typedValue[key], output, valueMode); err != nil {
				return err
			}
		}
	case string:
		if prefix != "" {
			output[prefix] = typedValue
		}
	default:
		if prefix == "" {
			return nil
		}
		if valueMode == ValueModeStringOnly {
			return gerror.New("i18n resource values must be strings or objects")
		}
		output[prefix] = gconv.String(typedValue)
	}
	return nil
}

// mergeCatalog merges source values into target values.
func mergeCatalog(target map[string]string, source map[string]string) {
	for key, value := range source {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		target[trimmedKey] = value
	}
}
