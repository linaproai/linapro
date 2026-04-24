// This file loads database-backed i18n message overrides from sys_i18n_message.

package i18n

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
)

// loadDatabaseLocaleBundle loads enabled message overrides for one locale from
// sys_i18n_message and returns them as a flat message catalog.
func (s *serviceImpl) loadDatabaseLocaleBundle(ctx context.Context, locale string) map[string]string {
	bundle, _ := s.loadDatabaseLocaleBundleWithSources(ctx, locale)
	return bundle
}

// loadDatabaseLocaleBundleWithSources loads enabled message overrides for one locale
// together with the source descriptors recorded in sys_i18n_message.
func (s *serviceImpl) loadDatabaseLocaleBundleWithSources(ctx context.Context, locale string) (map[string]string, map[string]MessageSourceDescriptor) {
	normalizedLocale := s.ResolveLocale(ctx, locale)

	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Warningf(ctx, "load runtime i18n database overrides panic locale=%s err=%v", normalizedLocale, recovered)
		}
	}()

	var rows []*entity.SysI18NMessage
	err := dao.SysI18NMessage.Ctx(ctx).
		Where(do.SysI18NMessage{
			Locale: normalizedLocale,
			Status: int(messageStatusEnabled),
		}).
		OrderAsc(dao.SysI18NMessage.Columns().MessageKey).
		Scan(&rows)
	if err != nil {
		logger.Warningf(ctx, "load runtime i18n database overrides failed locale=%s err=%v", normalizedLocale, err)
		return map[string]string{}, map[string]MessageSourceDescriptor{}
	}

	bundle := make(map[string]string, len(rows))
	sources := make(map[string]MessageSourceDescriptor, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		trimmedKey := strings.TrimSpace(row.MessageKey)
		if trimmedKey == "" {
			continue
		}
		bundle[trimmedKey] = row.MessageValue
		sources[trimmedKey] = MessageSourceDescriptor{
			Type:      string(messageOriginTypeDatabase),
			ScopeType: strings.TrimSpace(row.ScopeType),
			ScopeKey:  strings.TrimSpace(row.ScopeKey),
		}
	}
	return bundle, sources
}
