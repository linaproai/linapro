// This file implements host-side i18n maintenance capabilities such as import,
// export, missing-translation checks, and source diagnostics.

package i18n

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// MessageSourceDescriptor describes the effective source of one runtime message.
type MessageSourceDescriptor struct {
	Type      string // Type is the source layer, such as host_file or database.
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

// MessageImportInput defines one import request for database-backed i18n overrides.
type MessageImportInput struct {
	Locale     string            // Locale is the target locale.
	ScopeType  string            // ScopeType is the logical scope receiving imported messages.
	ScopeKey   string            // ScopeKey is the concrete scope identifier receiving imported messages.
	Overwrite  bool              // Overwrite reports whether existing rows should be updated.
	Remark     string            // Remark is written to created or updated rows.
	Messages   map[string]string // Messages contains flat translation keys and values to import.
	SourceType string            // SourceType identifies how imported rows should be tagged.
}

// MessageImportOutput describes the outcome of one import operation.
type MessageImportOutput struct {
	Locale   string // Locale is the imported locale.
	Created  int    // Created is the number of inserted rows.
	Updated  int    // Updated is the number of updated rows.
	Skipped  int    // Skipped is the number of skipped rows.
	Imported int    // Imported is the number of processed flat keys in the request.
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
	// already deduplicates host/plugin/dynamic/db sectors. The "raw" flag is
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
		if shouldSkipMissingMessage(resolvedLocale, key) {
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
// required in the target locale because another source of truth supplies the
// target-language fallback. For example, en-US omits job.handler.* keys because
// built-in scheduled-job handlers keep their English source text in Go code,
// source-plugin cron metadata, or dynamic-plugin cron contracts.
func shouldSkipMissingMessage(locale string, key string) bool {
	if locale != EnglishLocale {
		return false
	}
	return isSourceTextBackedRuntimeKey(key)
}

// isSourceTextBackedRuntimeKey reports whether the key is backed by source
// metadata rather than an en-US runtime JSON entry.
func isSourceTextBackedRuntimeKey(key string) bool {
	trimmedKey := strings.TrimSpace(key)
	return strings.HasPrefix(trimmedKey, "job.handler.") ||
		strings.HasPrefix(trimmedKey, "job.group.default.")
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

// ImportMessages imports flat translation keys into sys_i18n_message.
func (s *serviceImpl) ImportMessages(ctx context.Context, input MessageImportInput) (output MessageImportOutput, err error) {
	output.Locale = s.ResolveLocale(ctx, input.Locale)
	output.Imported = len(input.Messages)
	if len(input.Messages) == 0 {
		return output, nil
	}

	scopeType := strings.TrimSpace(input.ScopeType)
	if scopeType == "" {
		scopeType = string(messageScopeTypeHost)
	}
	scopeKey := strings.TrimSpace(input.ScopeKey)
	if scopeKey == "" {
		scopeKey = hostMessageScopeKey
	}
	sourceType := strings.TrimSpace(input.SourceType)
	if sourceType == "" {
		sourceType = string(messageSourceTypeImport)
	}

	keys := make([]string, 0, len(input.Messages))
	trimmedMessages := make(map[string]string, len(input.Messages))
	for key := range input.Messages {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		trimmedMessages[trimmedKey] = input.Messages[key]
		keys = append(keys, trimmedKey)
	}
	sort.Strings(keys)

	err = dao.SysI18NMessage.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, key := range keys {
			value := trimmedMessages[key]

			var existing *entity.SysI18NMessage
			queryErr := dao.SysI18NMessage.Ctx(ctx).
				Unscoped().
				Where(do.SysI18NMessage{
					Locale:     output.Locale,
					MessageKey: key,
					ScopeType:  scopeType,
					ScopeKey:   scopeKey,
				}).
				Scan(&existing)
			if queryErr != nil {
				return queryErr
			}

			if existing == nil {
				if _, insertErr := dao.SysI18NMessage.Ctx(ctx).Data(do.SysI18NMessage{
					Locale:       output.Locale,
					MessageKey:   key,
					MessageValue: value,
					ScopeType:    scopeType,
					ScopeKey:     scopeKey,
					SourceType:   sourceType,
					Status:       int(messageStatusEnabled),
					Remark:       input.Remark,
				}).Insert(); insertErr != nil {
					return insertErr
				}
				output.Created++
				continue
			}

			if !input.Overwrite {
				output.Skipped++
				continue
			}

			if _, updateErr := dao.SysI18NMessage.Ctx(ctx).
				Unscoped().
				Where(do.SysI18NMessage{Id: existing.Id}).
				Data(do.SysI18NMessage{
					MessageValue: value,
					SourceType:   sourceType,
					Status:       int(messageStatusEnabled),
					Remark:       input.Remark,
				}).
				Update(); updateErr != nil {
				return updateErr
			}
			if existing.DeletedAt != nil {
				if _, restoreErr := dao.SysI18NMessage.Ctx(ctx).
					Unscoped().
					Where(do.SysI18NMessage{Id: existing.Id}).
					Data(dao.SysI18NMessage.Columns().DeletedAt, gdb.Raw("NULL")).
					Update(); restoreErr != nil {
					return restoreErr
				}
			}
			output.Updated++
		}
		return nil
	})
	if err != nil {
		return output, err
	}

	// Database imports only mutate the database sector for the targeted
	// locale; other locales and other sectors stay hot.
	s.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{output.Locale},
		Sectors: []Sector{SectorDatabase},
	})
	return output, nil
}
