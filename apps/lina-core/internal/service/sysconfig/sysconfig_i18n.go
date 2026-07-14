// This file localizes config-management display metadata using stable config keys.

package sysconfig

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
)

// publicFrontendConfigValueMessageKeys maps protected sys_config keys to the
// runtime public-frontend message key used for read-only value projection.
var publicFrontendConfigValueMessageKeys = map[string]string{
	hostconfig.PublicFrontendSettingKeyAppName:            "publicFrontend.app.name",
	hostconfig.PublicFrontendSettingKeyAuthPageTitle:      "publicFrontend.auth.pageTitle",
	hostconfig.PublicFrontendSettingKeyAuthPageDesc:       "publicFrontend.auth.pageDesc",
	hostconfig.PublicFrontendSettingKeyAuthLoginSubtitle:  "publicFrontend.auth.loginSubtitle",
	hostconfig.PublicFrontendSettingKeyAuthPrivacyPolicy:  "publicFrontend.auth.privacyPolicy",
	hostconfig.PublicFrontendSettingKeyAuthTermsOfService: "publicFrontend.auth.termsOfService",
	hostconfig.PublicFrontendSettingKeyUIWatermarkContent: "publicFrontend.ui.watermarkContent",
}

// localizeConfigEntities localizes one config-entity list in place.
func (s *serviceImpl) localizeConfigEntities(ctx context.Context, items []*entity.SysConfig) {
	for _, item := range items {
		s.localizeConfigEntity(ctx, item)
	}
}

// localizeConfigEntity localizes one config entity in place for list and
// key-based reads, including optional public-frontend default value projection.
func (s *serviceImpl) localizeConfigEntity(ctx context.Context, item *entity.SysConfig) {
	if item == nil {
		return
	}
	s.localizeConfigEntityMetadata(ctx, item)
	trimmedKey := strings.TrimSpace(item.Key)
	if trimmedKey == "" {
		return
	}
	item.Value = s.localizedConfigDisplayValue(ctx, trimmedKey, item.Value)
}

// localizeConfigEntityMetadata localizes name and remark for display without
// projecting value. Edit/detail APIs use this so stored parameter values stay
// authoritative for form backfill and save.
func (s *serviceImpl) localizeConfigEntityMetadata(ctx context.Context, item *entity.SysConfig) {
	if s == nil || s.i18nSvc == nil || item == nil {
		return
	}
	trimmedKey := strings.TrimSpace(item.Key)
	if trimmedKey == "" {
		return
	}
	item.Name = s.i18nSvc.Translate(ctx, "config."+trimmedKey+".name", item.Name)
	item.Remark = s.i18nSvc.Translate(ctx, "config."+trimmedKey+".remark", item.Remark)
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

// localizedConfigDisplayValue localizes read-only list values for protected
// public-frontend text settings that still match the built-in source text.
// Custom runtime values remain raw because sys_config does not store per-locale
// overrides.
func (s *serviceImpl) localizedConfigDisplayValue(ctx context.Context, key string, current string) string {
	if s == nil || s.i18nSvc == nil {
		return current
	}
	messageKey, ok := publicFrontendConfigValueMessageKeys[strings.TrimSpace(key)]
	if !ok {
		return current
	}
	spec, ok := hostconfig.LookupPublicFrontendSettingSpec(key)
	if !ok {
		return current
	}
	trimmedCurrent := strings.TrimSpace(current)
	if trimmedCurrent == "" {
		return current
	}
	defaultLocaleValue := strings.TrimSpace(
		s.i18nSvc.Translate(context.Background(), messageKey, spec.DefaultValue),
	)
	if trimmedCurrent != strings.TrimSpace(spec.DefaultValue) && trimmedCurrent != defaultLocaleValue {
		return current
	}
	return s.i18nSvc.Translate(ctx, messageKey, current)
}

// buildLocalizedImportTemplateHeaders returns localized config-template headers.
func (s *serviceImpl) buildLocalizedImportTemplateHeaders(ctx context.Context) []string {
	return []string{
		s.localizedConfigFieldLabel(ctx, "name", "Parameter Name"),
		s.localizedConfigFieldLabel(ctx, "key", "Parameter Key"),
		s.localizedConfigFieldLabel(ctx, "value", "Parameter Value"),
		s.localizedConfigFieldLabel(ctx, "valueType", "Value Type"),
		s.localizedConfigFieldLabel(ctx, "options", "Options"),
		s.localizedConfigFieldLabel(ctx, "remark", "Remark"),
	}
}

// buildLocalizedExportHeaders returns localized config-export headers.
func (s *serviceImpl) buildLocalizedExportHeaders(ctx context.Context) []string {
	return []string{
		s.localizedConfigFieldLabel(ctx, "name", "Parameter Name"),
		s.localizedConfigFieldLabel(ctx, "key", "Parameter Key"),
		s.localizedConfigFieldLabel(ctx, "value", "Parameter Value"),
		s.localizedConfigFieldLabel(ctx, "valueType", "Value Type"),
		s.localizedConfigFieldLabel(ctx, "options", "Options"),
		s.localizedConfigFieldLabel(ctx, "remark", "Remark"),
		s.localizedConfigFieldLabel(ctx, "createdAt", "Created At"),
		s.localizedConfigFieldLabel(ctx, "updatedAt", "Updated At"),
	}
}

// localizedConfigFieldLabel returns one localized config field label.
func (s *serviceImpl) localizedConfigFieldLabel(ctx context.Context, field string, fallback string) string {
	trimmedField := strings.TrimSpace(field)
	if trimmedField == "" {
		return fallback
	}
	if s == nil || s.i18nSvc == nil {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, "config.field."+trimmedField, fallback)
}

// localizedConfigImportFailure returns one localized config import failure reason.
func (s *serviceImpl) localizedConfigImportFailure(ctx context.Context, key string, fallback string) string {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" || s == nil || s.i18nSvc == nil {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, "artifact.config.import.failure."+trimmedKey, fallback)
}

// localizedConfigImportError renders one import-row error in the request locale.
func (s *serviceImpl) localizedConfigImportError(ctx context.Context, err error) string {
	if err == nil {
		return ""
	}
	if s == nil || s.i18nSvc == nil {
		return err.Error()
	}
	if localized := s.i18nSvc.LocalizeError(ctx, err); localized != "" {
		return localized
	}
	return err.Error()
}
