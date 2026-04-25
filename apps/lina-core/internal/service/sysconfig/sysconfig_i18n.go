// This file localizes config-management display metadata using stable config keys.

package sysconfig

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
)

// localizeConfigEntities localizes one config-entity list in place.
func (s *serviceImpl) localizeConfigEntities(ctx context.Context, items []*entity.SysConfig) {
	for _, item := range items {
		s.localizeConfigEntity(ctx, item)
	}
}

// localizeConfigEntity localizes one config entity in place.
func (s *serviceImpl) localizeConfigEntity(ctx context.Context, item *entity.SysConfig) {
	if s == nil || s.i18nSvc == nil || item == nil {
		return
	}
	trimmedKey := strings.TrimSpace(item.Key)
	if trimmedKey == "" {
		return
	}
	item.Name = s.localizedConfigName(ctx, trimmedKey, item.Name)
	item.Remark = s.localizedConfigRemark(ctx, trimmedKey, item.Remark)
}

// localizedConfigName returns one localized config display name.
func (s *serviceImpl) localizedConfigName(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil {
		return fallback
	}
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, "config."+trimmedKey+".name", fallback)
}

// localizedConfigRemark returns one localized config display remark.
func (s *serviceImpl) localizedConfigRemark(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil {
		return fallback
	}
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, "config."+trimmedKey+".remark", fallback)
}

// buildLocalizedImportTemplateHeaders returns localized config-template headers.
func (s *serviceImpl) buildLocalizedImportTemplateHeaders(ctx context.Context) []string {
	return []string{
		s.localizedConfigFieldLabel(ctx, "name", "参数名称"),
		s.localizedConfigFieldLabel(ctx, "key", "参数键名"),
		s.localizedConfigFieldLabel(ctx, "value", "参数键值"),
		s.localizedConfigFieldLabel(ctx, "remark", "备注"),
	}
}

// buildLocalizedExportHeaders returns localized config-export headers.
func (s *serviceImpl) buildLocalizedExportHeaders(ctx context.Context) []string {
	return []string{
		s.localizedConfigFieldLabel(ctx, "name", "参数名称"),
		s.localizedConfigFieldLabel(ctx, "key", "参数键名"),
		s.localizedConfigFieldLabel(ctx, "value", "参数键值"),
		s.localizedConfigFieldLabel(ctx, "remark", "备注"),
		s.localizedConfigFieldLabel(ctx, "createdAt", "创建时间"),
		s.localizedConfigFieldLabel(ctx, "updatedAt", "修改时间"),
	}
}

// localizedConfigFieldLabel returns one localized config field label.
func (s *serviceImpl) localizedConfigFieldLabel(ctx context.Context, field string, fallback string) string {
	var (
		englishLabels = map[string]string{
			"createdAt": "Created At",
			"key":       "Parameter Key",
			"name":      "Parameter Name",
			"remark":    "Remark",
			"updatedAt": "Updated At",
			"value":     "Parameter Value",
		}
		chineseLabels = map[string]string{
			"createdAt": "创建时间",
			"key":       "参数键名",
			"name":      "参数名称",
			"remark":    "备注",
			"updatedAt": "修改时间",
			"value":     "参数键值",
		}
	)

	trimmedField := strings.TrimSpace(field)
	if trimmedField == "" {
		return fallback
	}
	if s != nil && s.i18nSvc != nil && s.i18nSvc.ResolveLocale(ctx, "") == "en-US" {
		if label, ok := englishLabels[trimmedField]; ok {
			return label
		}
	}
	if label, ok := chineseLabels[trimmedField]; ok {
		return label
	}
	return fallback
}
