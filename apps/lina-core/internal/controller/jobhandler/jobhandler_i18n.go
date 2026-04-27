// This file localizes scheduled-job handler metadata at the HTTP boundary.

package jobhandler

import (
	"context"
	"strings"

	"lina-core/internal/service/jobmeta"
)

const (
	// handlerNameI18nField identifies the handler display-name i18n field.
	handlerNameI18nField = "name"
	// handlerDescriptionI18nField identifies the handler description i18n field.
	handlerDescriptionI18nField = "description"
)

// localizeHandlerName returns the current-language display name for one handler ref.
func (c *ControllerV1) localizeHandlerName(ctx context.Context, ref string, fallback string) string {
	return c.translateSourceText(ctx, jobmeta.HandlerI18nKey(ref, handlerNameI18nField), fallback)
}

// localizeHandlerDescription returns the current-language description for one handler ref.
func (c *ControllerV1) localizeHandlerDescription(ctx context.Context, ref string, fallback string) string {
	return c.translateSourceText(ctx, jobmeta.HandlerI18nKey(ref, handlerDescriptionI18nField), fallback)
}

// translateSourceText resolves handler display metadata with source-text
// fallback so the selected locale never silently falls back to another language.
func (c *ControllerV1) translateSourceText(ctx context.Context, key string, sourceText string) string {
	if c == nil || c.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return sourceText
	}
	return c.i18nSvc.TranslateSourceText(ctx, key, sourceText)
}
