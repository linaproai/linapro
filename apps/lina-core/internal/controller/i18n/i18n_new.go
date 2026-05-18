// This file defines the runtime i18n controller dependencies and constructor.

package i18n

import (
	i18napi "lina-core/api/i18n"
	i18nsvc "lina-core/internal/service/i18n"
)

// ControllerV1 implements the runtime i18n API controller.
type ControllerV1 struct {
	localeResolver i18nsvc.LocaleResolver // localeResolver resolves request and explicit locales.
	bundleProvider i18nsvc.BundleProvider // bundleProvider serves runtime locales and messages.
	maintainer     i18nsvc.Maintainer     // maintainer handles message diagnostics and export.
}

// NewV1 creates and returns a new runtime i18n controller.
func NewV1(i18nSvc i18nsvc.Service) i18napi.II18NV1 {
	return &ControllerV1{
		localeResolver: i18nSvc,
		bundleProvider: i18nSvc,
		maintainer:     i18nSvc,
	}
}
