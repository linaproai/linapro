// This file localizes dictionary type and dictionary data display fields.

package dict

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
)

// localizeDictTypeEntities localizes one dictionary-type entity list in place.
func (s *serviceImpl) localizeDictTypeEntities(ctx context.Context, items []*entity.SysDictType) {
	for _, item := range items {
		s.localizeDictTypeEntity(ctx, item)
	}
}

// localizeDictTypeEntity localizes one dictionary-type entity in place.
func (s *serviceImpl) localizeDictTypeEntity(ctx context.Context, item *entity.SysDictType) {
	if s == nil || s.i18nSvc == nil || item == nil {
		return
	}
	trimmedType := strings.TrimSpace(item.Type)
	if trimmedType == "" {
		return
	}
	// The default locale is edited directly through dictionary management.
	if s.i18nSvc.ResolveLocale(ctx, "") == i18nsvc.DefaultLocale {
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
	if s == nil || s.i18nSvc == nil || item == nil {
		return
	}
	trimmedType := strings.TrimSpace(item.DictType)
	trimmedValue := strings.TrimSpace(item.Value)
	if trimmedType == "" || trimmedValue == "" {
		return
	}
	// The default locale is edited directly through dictionary management.
	if s.i18nSvc.ResolveLocale(ctx, "") == i18nsvc.DefaultLocale {
		return
	}
	prefix := "dict." + trimmedType + "." + trimmedValue
	item.Label = s.i18nSvc.Translate(ctx, prefix+".label", item.Label)
	item.Remark = s.i18nSvc.Translate(ctx, prefix+".remark", item.Remark)
}
