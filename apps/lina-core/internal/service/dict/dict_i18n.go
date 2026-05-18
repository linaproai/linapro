// This file localizes dictionary type and dictionary data display fields.

package dict

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/bizerr"
)

// dictI18nTranslator defines the narrow translation capability dict needs.
type dictI18nTranslator interface {
	// ResolveLocale resolves one explicit locale override against the current request locale.
	ResolveLocale(ctx context.Context, locale string) string
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// runtimeTextItem defines one source-owned runtime message key and English
// fallback pair used for batch translation.
type runtimeTextItem struct {
	Key      string
	Fallback string
}

// runtimeText returns one localized runtime message after applying named
// parameters to the translated template or English fallback.
func (s *serviceImpl) runtimeText(ctx context.Context, key string, fallback string, params ...bizerr.Param) string {
	template := fallback
	if s != nil && s.i18nSvc != nil {
		template = s.i18nSvc.Translate(ctx, key, fallback)
	}
	return bizerr.Format(template, runtimeParamMap(params))
}

// runtimeTexts translates a small ordered batch of runtime text items.
func (s *serviceImpl) runtimeTexts(ctx context.Context, items []runtimeTextItem) []string {
	texts := make([]string, 0, len(items))
	for _, item := range items {
		texts = append(texts, s.runtimeText(ctx, item.Key, item.Fallback))
	}
	return texts
}

// runtimeParamMap converts named runtime parameters to the map required by the
// formatter.
func runtimeParamMap(params []bizerr.Param) map[string]any {
	values := make(map[string]any, len(params))
	for _, param := range params {
		name := strings.TrimSpace(param.Name)
		if name == "" {
			continue
		}
		values[name] = param.Value
	}
	return values
}

// dictStatusText returns the localized label for a normal/disabled status.
func (s *serviceImpl) dictStatusText(ctx context.Context, status int) string {
	if status == 0 {
		return s.runtimeText(ctx, "dict.sys_normal_disable.0.label", "Disabled")
	}
	return s.runtimeText(ctx, "dict.sys_normal_disable.1.label", "Enabled")
}

// dictTypeSheetName returns the localized worksheet name for dictionary types.
func (s *serviceImpl) dictTypeSheetName(ctx context.Context) string {
	return s.runtimeText(ctx, "artifact.dict.sheet.type", "Dictionary Types")
}

// dictDataSheetName returns the localized worksheet name for dictionary data.
func (s *serviceImpl) dictDataSheetName(ctx context.Context) string {
	return s.runtimeText(ctx, "artifact.dict.sheet.data", "Dictionary Data")
}

// isDictDisabledStatusInput reports whether one import cell requests the
// disabled status in the current locale or by stable numeric value.
func (s *serviceImpl) isDictDisabledStatusInput(ctx context.Context, value string) bool {
	trimmedValue := strings.TrimSpace(value)
	return trimmedValue == "0" || strings.EqualFold(trimmedValue, s.dictStatusText(ctx, 0))
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
