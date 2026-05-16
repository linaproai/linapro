// This file contains inbox category normalization and runtime i18n lookup for
// category labels and colors rendered in current-user message views.

package usermsg

import "context"

// resolveCategoryCode normalizes an inbox message category code, falling back
// to the canonical "other" bucket when the sender did not declare one.
func resolveCategoryCode(categoryCode string) string {
	if categoryCode == "" {
		return usermsgCategoryFallbackCode
	}
	return categoryCode
}

// localizeCategoryLabel resolves the localized category display label for the
// given category code. Translation is looked up at
// `notify.category.{code}.label`. If the requested code has no translation
// resource, it falls back to the canonical "other" bucket and finally to a
// safety literal so the inbox never renders an empty category cell.
func (s *serviceImpl) localizeCategoryLabel(ctx context.Context, categoryCode string) string {
	if s == nil || s.i18nSvc == nil {
		return usermsgCategoryDefaultLabel
	}
	code := resolveCategoryCode(categoryCode)
	key := usermsgCategoryI18nNamespace + code + usermsgCategoryLabelI18nSuffix
	if label := s.i18nSvc.Translate(ctx, key, ""); label != "" {
		return label
	}
	if code != usermsgCategoryFallbackCode {
		fallbackKey := usermsgCategoryI18nNamespace + usermsgCategoryFallbackCode + usermsgCategoryLabelI18nSuffix
		if label := s.i18nSvc.Translate(ctx, fallbackKey, ""); label != "" {
			return label
		}
	}
	return usermsgCategoryDefaultLabel
}

// localizeCategoryColor resolves the localized category tag color for the
// given category code. Color is treated as a localizable display attribute so
// senders can override their preferred palette per locale if needed; the
// resolution path mirrors localizeCategoryLabel and falls back to a generic
// neutral color.
func (s *serviceImpl) localizeCategoryColor(ctx context.Context, categoryCode string) string {
	if s == nil || s.i18nSvc == nil {
		return usermsgCategoryDefaultColor
	}
	code := resolveCategoryCode(categoryCode)
	key := usermsgCategoryI18nNamespace + code + usermsgCategoryColorI18nSuffix
	if color := s.i18nSvc.Translate(ctx, key, ""); color != "" {
		return color
	}
	if code != usermsgCategoryFallbackCode {
		fallbackKey := usermsgCategoryI18nNamespace + usermsgCategoryFallbackCode + usermsgCategoryColorI18nSuffix
		if color := s.i18nSvc.Translate(ctx, fallbackKey, ""); color != "" {
			return color
		}
	}
	return usermsgCategoryDefaultColor
}
