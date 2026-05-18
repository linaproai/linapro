// This file implements host-side i18n maintenance capabilities such as export,
// missing-translation checks, and source diagnostics.

package i18n

import (
	"context"
	"sort"
	"strings"
)

// MessageSourceDescriptor describes the effective source of one runtime message.
type MessageSourceDescriptor struct {
	Type      string // Type is the source layer, such as host_file or plugin_file.
	ScopeType string // ScopeType is the logical owning scope, such as host or plugin.
	ScopeKey  string // ScopeKey is the owning scope identifier, such as core or plugin ID.
}

// MessageExportOutput describes one exported runtime message bundle.
type MessageExportOutput struct {
	Locale        string            // Locale is the exported locale.
	DefaultLocale string            // DefaultLocale is the current runtime default locale.
	Mode          string            // Mode is the export mode, such as effective or raw.
	Messages      map[string]string // Messages contains exported flat messages.
}

// MissingMessageItem describes one translation key that is missing in a target locale.
type MissingMessageItem struct {
	Key          string                  // Key is the missing translation key.
	DefaultValue string                  // DefaultValue is the fallback value from the default locale.
	Source       MessageSourceDescriptor // Source identifies where the default value currently comes from.
}

// MessageDiagnosticItem describes the effective resolution result for one message key.
type MessageDiagnosticItem struct {
	Key             string                  // Key is the translation key.
	Value           string                  // Value is the effective translation value.
	RequestedLocale string                  // RequestedLocale is the locale requested by the caller.
	EffectiveLocale string                  // EffectiveLocale is the locale that actually supplied the value.
	FromFallback    bool                    // FromFallback reports whether the default locale supplied the value.
	Source          MessageSourceDescriptor // Source identifies the resolved source layer.
}

// ExportMessages exports flat runtime messages for one locale.
func (s *serviceImpl) ExportMessages(ctx context.Context, locale string, raw bool) MessageExportOutput {
	resolvedLocale := s.ResolveLocale(ctx, locale)
	defaultLocale := s.getDefaultRuntimeLocale(ctx)
	mode := "effective"
	if raw {
		mode = "raw"
	}
	// Both effective and raw exports return the same merged catalog: the cache
	// already deduplicates host/plugin/dynamic sectors. The "raw" flag is
	// retained for API compatibility but no longer carries different semantics.
	messages := cloneFlatMessageMap(s.snapshotMergedCatalog(ctx, resolvedLocale))
	return MessageExportOutput{
		Locale:        resolvedLocale,
		DefaultLocale: defaultLocale,
		Mode:          mode,
		Messages:      messages,
	}
}

// CheckMissingMessages returns translation keys missing from one locale compared with the default locale.
func (s *serviceImpl) CheckMissingMessages(ctx context.Context, locale string, keyPrefix string) []MissingMessageItem {
	resolvedLocale := s.ResolveLocale(ctx, locale)
	defaultLocale := s.getDefaultRuntimeLocale(ctx)
	if resolvedLocale == defaultLocale {
		return []MissingMessageItem{}
	}

	defaultBundle, defaultSources := s.loadRawLocaleBundleWithSources(ctx, defaultLocale)
	targetBundle := cloneFlatMessageMap(s.snapshotMergedCatalog(ctx, resolvedLocale))
	trimmedPrefix := strings.TrimSpace(keyPrefix)

	keys := make([]string, 0, len(defaultBundle))
	for key := range defaultBundle {
		if trimmedPrefix != "" && !strings.HasPrefix(key, trimmedPrefix) {
			continue
		}
		if shouldSkipMissingMessage(key) {
			continue
		}
		if _, ok := targetBundle[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]MissingMessageItem, 0, len(keys))
	for _, key := range keys {
		items = append(items, MissingMessageItem{
			Key:          key,
			DefaultValue: defaultBundle[key],
			Source:       defaultSources[key],
		})
	}
	return items
}

// shouldSkipMissingMessage reports whether one default-locale key is not
// required in the target locale because a registered source-text namespace
// supplies the target-language fallback.
func shouldSkipMissingMessage(key string) bool {
	return isSourceTextBackedRuntimeKey(key)
}

// isSourceTextBackedRuntimeKey reports whether the key is backed by source
// metadata rather than an en-US runtime JSON entry.
func isSourceTextBackedRuntimeKey(key string) bool {
	_, ok := SourceTextNamespaceReason(key)
	return ok
}

// DiagnoseMessages returns effective source diagnostics for one locale.
func (s *serviceImpl) DiagnoseMessages(ctx context.Context, locale string, keyPrefix string) []MessageDiagnosticItem {
	resolvedLocale := s.ResolveLocale(ctx, locale)
	defaultLocale := s.getDefaultRuntimeLocale(ctx)
	requestedBundle, requestedSources := s.loadRawLocaleBundleWithSources(ctx, resolvedLocale)
	defaultBundle, defaultSources := s.loadRawLocaleBundleWithSources(ctx, defaultLocale)
	trimmedPrefix := strings.TrimSpace(keyPrefix)

	keysSet := make(map[string]struct{}, len(defaultBundle)+len(requestedBundle))
	for key := range defaultBundle {
		if trimmedPrefix == "" || strings.HasPrefix(key, trimmedPrefix) {
			keysSet[key] = struct{}{}
		}
	}
	for key := range requestedBundle {
		if trimmedPrefix == "" || strings.HasPrefix(key, trimmedPrefix) {
			keysSet[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(keysSet))
	for key := range keysSet {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]MessageDiagnosticItem, 0, len(keys))
	for _, key := range keys {
		item := MessageDiagnosticItem{
			Key:             key,
			RequestedLocale: resolvedLocale,
			EffectiveLocale: resolvedLocale,
		}
		if value, ok := requestedBundle[key]; ok {
			item.Value = value
			item.Source = requestedSources[key]
		} else {
			item.Value = defaultBundle[key]
			item.Source = defaultSources[key]
			item.EffectiveLocale = defaultLocale
			item.FromFallback = true
		}
		items = append(items, item)
	}
	return items
}
