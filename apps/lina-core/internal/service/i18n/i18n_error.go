// This file translates validation and business errors into the effective
// request locale using the host runtime catalog plus GoFrame's gi18n fallback.

package i18n

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/util/gvalid"

	"lina-core/pkg/bizerr"
)

// LocalizeError translates one request-scoped error into the effective locale.
func (s *serviceImpl) LocalizeError(ctx context.Context, err error) string {
	if err == nil {
		return ""
	}

	if validationErr, ok := err.(gvalid.Error); ok {
		return s.localizeValidationError(ctx, validationErr)
	}

	if messageErr, ok := bizerr.As(err); ok {
		return s.localizeRuntimeMessageError(ctx, messageErr)
	}

	var textArgs gerror.ITextArgs
	if errors.As(err, &textArgs) {
		format := s.localizeText(ctx, textArgs.Text())
		if len(textArgs.Args()) > 0 {
			return fmt.Sprintf(format, textArgs.Args()...)
		}
		return format
	}

	return s.localizeText(ctx, err.Error())
}

// localizeRuntimeMessageError renders one structured runtime-message error
// using the current request locale and named parameters.
func (s *serviceImpl) localizeRuntimeMessageError(ctx context.Context, err *bizerr.Error) string {
	if err == nil {
		return ""
	}

	template := s.Translate(ctx, err.MessageKey(), err.Fallback())
	if template == "" {
		template = err.Error()
	}
	return bizerr.Format(template, err.Params())
}

// localizeValidationError translates each validation item independently so flat
// validation keys and gi18n-backed rule messages both render in request locale.
func (s *serviceImpl) localizeValidationError(ctx context.Context, err gvalid.Error) string {
	if err == nil {
		return ""
	}

	items := err.Strings()
	if len(items) == 0 {
		return s.localizeText(ctx, err.Error())
	}

	localized := make([]string, 0, len(items))
	for _, item := range items {
		localized = append(localized, s.localizeText(ctx, item))
	}
	return strings.Join(localized, "; ")
}

// localizeText resolves one plain text or translation key using the runtime
// bundle first, then falls back to GoFrame's gi18n manager for builtin rules.
func (s *serviceImpl) localizeText(ctx context.Context, text string) string {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return ""
	}

	if translated := s.Translate(ctx, trimmedText, ""); translated != "" {
		return translated
	}

	translated := gi18n.Translate(s.ensureLanguageCtx(ctx), trimmedText)
	if translated != trimmedText {
		return translated
	}

	return trimmedText
}

// ensureLanguageCtx guarantees that gi18n fallback translation always sees the
// resolved request locale even when the caller passes a nil or raw context.
func (s *serviceImpl) ensureLanguageCtx(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return gi18n.WithLanguage(ctx, s.GetLocale(ctx))
}
