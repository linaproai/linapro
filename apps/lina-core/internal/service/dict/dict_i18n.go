// This file localizes dictionary type and dictionary data display fields.

package dict

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
)

// dictI18nTranslator defines the narrow translation capability dict needs.
type dictI18nTranslator interface {
	// ResolveLocale resolves one explicit locale override against the current request locale.
	ResolveLocale(ctx context.Context, locale string) string
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// localizeDictTypeEntities localizes one dictionary-type entity list in place.
func (s *serviceImpl) localizeDictTypeEntities(ctx context.Context, items []*entity.SysDictType) {
	for _, item := range items {
		s.localizeDictTypeEntity(ctx, item)
	}
}

// localizeDictTypeEntity localizes one dictionary-type entity in place.
func (s *serviceImpl) localizeDictTypeEntity(ctx context.Context, item *entity.SysDictType) {
	if s == nil || s.i18nSvc == nil || item == nil || s.shouldKeepEditableDefaultLocale(ctx) {
		return
	}
	trimmedType := strings.TrimSpace(item.Type)
	if trimmedType == "" {
		return
	}
	item.Name = s.i18nSvc.Translate(ctx, "dict."+trimmedType+".name", item.Name)
	item.Remark = s.i18nSvc.Translate(ctx, "dict."+trimmedType+".remark", item.Remark)
}

// localizeDictDataEntities localizes one dictionary-data entity list in place.
func (s *serviceImpl) localizeDictDataEntities(ctx context.Context, items []*entity.SysDictData) {
	for _, item := range items {
		s.localizeDictDataEntity(ctx, item)
	}
}

// localizeDictDataEntity localizes one dictionary-data entity in place.
func (s *serviceImpl) localizeDictDataEntity(ctx context.Context, item *entity.SysDictData) {
	if s == nil || s.i18nSvc == nil || item == nil || s.shouldKeepEditableDefaultLocale(ctx) {
		return
	}
	trimmedType := strings.TrimSpace(item.DictType)
	trimmedValue := strings.TrimSpace(item.Value)
	if trimmedType == "" || trimmedValue == "" {
		return
	}
	prefix := "dict." + trimmedType + "." + trimmedValue
	item.Label = s.i18nSvc.Translate(ctx, prefix+".label", item.Label)
	item.Remark = s.i18nSvc.Translate(ctx, prefix+".remark", item.Remark)
}

// shouldKeepEditableDefaultLocale reports whether editable dictionary values
// should remain as stored for the host default locale.
func (s *serviceImpl) shouldKeepEditableDefaultLocale(ctx context.Context) bool {
	if s == nil || s.i18nSvc == nil {
		return true
	}
	return s.i18nSvc.ResolveLocale(ctx, "") == i18nsvc.DefaultLocale
}
