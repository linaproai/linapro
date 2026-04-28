// This file provides runtime i18n helpers for user import/export artifacts and
// business errors.

package user

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
)

// userI18nTranslator defines the narrow translation capability user service
// needs for request-scoped runtime copy.
type userI18nTranslator interface {
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

// userSexText returns the localized label for one user sex value.
func (s *serviceImpl) userSexText(ctx context.Context, sex int) string {
	switch sex {
	case 1:
		return s.runtimeText(ctx, "dict.sys_user_sex.1.label", "Male")
	case 2:
		return s.runtimeText(ctx, "dict.sys_user_sex.2.label", "Female")
	default:
		return s.runtimeText(ctx, "dict.sys_user_sex.0.label", "Unknown")
	}
}

// userStatusText returns the localized label for one user status value.
func (s *serviceImpl) userStatusText(ctx context.Context, status Status) string {
	if status == StatusDisabled {
		return s.runtimeText(ctx, "dict.sys_normal_disable.0.label", "Disabled")
	}
	return s.runtimeText(ctx, "dict.sys_normal_disable.1.label", "Enabled")
}

// isUserSexInput reports whether one import cell matches the current-locale
// label or stable numeric value for the expected sex.
func (s *serviceImpl) isUserSexInput(ctx context.Context, value string, sex int) bool {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return false
	}
	return trimmedValue == userSexNumericValue(sex) || strings.EqualFold(trimmedValue, s.userSexText(ctx, sex))
}

// userSexNumericValue returns the stable numeric import token for one sex
// value.
func userSexNumericValue(sex int) string {
	switch sex {
	case 1:
		return "1"
	case 2:
		return "2"
	default:
		return "0"
	}
}

// isUserDisabledStatusInput reports whether one import cell requests the
// disabled user status in the current locale or by stable numeric value.
func (s *serviceImpl) isUserDisabledStatusInput(ctx context.Context, value string) bool {
	trimmedValue := strings.TrimSpace(value)
	return trimmedValue == "0" || strings.EqualFold(trimmedValue, s.userStatusText(ctx, StatusDisabled))
}
