// This file provides cached sys_i18n_content lookup helpers for business
// modules that need multilingual titles, summaries, or body content.

package i18n

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
)

// ContentLookupInput describes one business-content lookup request.
type ContentLookupInput struct {
	BusinessType   string // BusinessType is the stable module or aggregate identifier.
	BusinessID     string // BusinessID is the stable business primary key or anchor id.
	DefaultContent string // DefaultContent is the caller-owned fallback when no locale variant exists.
	Field          string // Field is the stable business field name, for example title or body.
	Locale         string // Locale optionally overrides the request locale for this lookup.
}

// ContentLookupOutput describes the resolved business-content value.
type ContentLookupOutput struct {
	BusinessType     string      // BusinessType is the stable module or aggregate identifier.
	BusinessID       string      // BusinessID is the stable business primary key or anchor id.
	Content          string      // Content is the resolved multilingual content value.
	ContentType      ContentType // ContentType is the stored payload format.
	Defaulted        bool        // Defaulted reports whether DefaultContent supplied the final value.
	EffectiveLocale  string      // EffectiveLocale is the locale that actually supplied the content.
	Field            string      // Field is the stable business field name.
	Found            bool        // Found reports whether sys_i18n_content supplied the content.
	RequestedLocale  string      // RequestedLocale is the normalized requested locale.
	ResolvedFallback bool        // ResolvedFallback reports whether the default locale supplied the content.
}

// ContentVariant describes one stored locale variant of a business-content anchor.
type ContentVariant struct {
	Content     string      // Content is the stored multilingual content value.
	ContentType ContentType // ContentType is the stored payload format.
	Locale      string      // Locale is the canonical locale code for this variant.
}

// runtimeContentCache stores per-anchor locale variants loaded from sys_i18n_content.
var runtimeContentCache = struct {
	sync.RWMutex
	variants map[string]map[string]ContentVariant
}{
	variants: make(map[string]map[string]ContentVariant),
}

// GetContent resolves one business-content translation from sys_i18n_content.
func (s *serviceImpl) GetContent(ctx context.Context, input ContentLookupInput) (ContentLookupOutput, error) {
	anchorKey, businessType, businessID, field, err := normalizeContentAnchorInput(input.BusinessType, input.BusinessID, input.Field)
	if err != nil {
		return ContentLookupOutput{}, err
	}

	requestedLocale := s.ResolveLocale(ctx, input.Locale)
	defaultLocale := s.getDefaultRuntimeLocale(ctx)
	output := ContentLookupOutput{
		BusinessType:    businessType,
		BusinessID:      businessID,
		ContentType:     ContentTypePlain,
		EffectiveLocale: requestedLocale,
		Field:           field,
		RequestedLocale: requestedLocale,
	}

	variantMap, err := s.loadContentVariantMap(ctx, anchorKey, businessType, businessID, field)
	if err != nil {
		return output, err
	}
	if variant, ok := variantMap[requestedLocale]; ok {
		output.Content = variant.Content
		output.ContentType = normalizeContentType(variant.ContentType)
		output.Found = true
		return output, nil
	}
	if requestedLocale != defaultLocale {
		if variant, ok := variantMap[defaultLocale]; ok {
			output.Content = variant.Content
			output.ContentType = normalizeContentType(variant.ContentType)
			output.EffectiveLocale = defaultLocale
			output.Found = true
			output.ResolvedFallback = true
			return output, nil
		}
	}

	output.Content = strings.TrimSpace(input.DefaultContent)
	output.Defaulted = output.Content != ""
	return output, nil
}

// ListContentVariants lists all enabled locale variants for one business-content anchor.
func (s *serviceImpl) ListContentVariants(ctx context.Context, businessType string, businessID string, field string) ([]ContentVariant, error) {
	anchorKey, normalizedBusinessType, normalizedBusinessID, normalizedField, err := normalizeContentAnchorInput(businessType, businessID, field)
	if err != nil {
		return nil, err
	}

	variantMap, err := s.loadContentVariantMap(ctx, anchorKey, normalizedBusinessType, normalizedBusinessID, normalizedField)
	if err != nil {
		return nil, err
	}

	locales := make([]string, 0, len(variantMap))
	for locale := range variantMap {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	items := make([]ContentVariant, 0, len(locales))
	for _, locale := range locales {
		variant := variantMap[locale]
		variant.Locale = locale
		variant.ContentType = normalizeContentType(variant.ContentType)
		items = append(items, variant)
	}
	return items, nil
}

// loadContentVariantMap loads all enabled locale variants for one business-content anchor.
func (s *serviceImpl) loadContentVariantMap(ctx context.Context, anchorKey string, businessType string, businessID string, field string) (map[string]ContentVariant, error) {
	runtimeContentCache.RLock()
	cachedVariants, ok := runtimeContentCache.variants[anchorKey]
	runtimeContentCache.RUnlock()
	if ok {
		return cloneContentVariantMap(cachedVariants), nil
	}

	var rows []*entity.SysI18NContent
	err := dao.SysI18NContent.Ctx(ctx).
		Where(do.SysI18NContent{
			BusinessType: businessType,
			BusinessId:   businessID,
			Field:        field,
			Status:       int(contentStatusEnabled),
		}).
		OrderAsc(dao.SysI18NContent.Columns().Locale).
		Scan(&rows)
	if err != nil {
		logger.Warningf(ctx, "load business i18n content failed businessType=%s businessID=%s field=%s err=%v", businessType, businessID, field, err)
		return nil, err
	}

	variants := make(map[string]ContentVariant, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		locale := normalizeLocale(row.Locale)
		if locale == "" {
			continue
		}
		variants[locale] = ContentVariant{
			Content:     row.Content,
			ContentType: normalizeContentType(ContentType(strings.TrimSpace(row.ContentType))),
			Locale:      locale,
		}
	}

	runtimeContentCache.Lock()
	runtimeContentCache.variants[anchorKey] = cloneContentVariantMap(variants)
	runtimeContentCache.Unlock()
	return variants, nil
}

// normalizeContentAnchorInput validates and normalizes one business-content anchor.
func normalizeContentAnchorInput(businessType string, businessID string, field string) (string, string, string, string, error) {
	normalizedBusinessType := strings.TrimSpace(businessType)
	normalizedBusinessID := strings.TrimSpace(businessID)
	normalizedField := strings.TrimSpace(field)
	if normalizedBusinessType == "" || normalizedBusinessID == "" || normalizedField == "" {
		return "", "", "", "", gerror.New("i18n business content requires business_type, business_id, and field")
	}
	return buildContentAnchorKey(normalizedBusinessType, normalizedBusinessID, normalizedField), normalizedBusinessType, normalizedBusinessID, normalizedField, nil
}

// buildContentAnchorKey returns the in-memory cache key for one business-content anchor.
func buildContentAnchorKey(businessType string, businessID string, field string) string {
	return businessType + "\x00" + businessID + "\x00" + field
}

// cloneContentVariantMap copies one locale-variant map so callers can mutate it safely.
func cloneContentVariantMap(src map[string]ContentVariant) map[string]ContentVariant {
	if len(src) == 0 {
		return map[string]ContentVariant{}
	}
	dst := make(map[string]ContentVariant, len(src))
	for locale, variant := range src {
		dst[locale] = variant
	}
	return dst
}

// normalizeContentType sanitizes persisted content_type values.
func normalizeContentType(value ContentType) ContentType {
	switch value {
	case ContentTypeMarkdown, ContentTypeHTML, ContentTypeJSON:
		return value
	default:
		return ContentTypePlain
	}
}
