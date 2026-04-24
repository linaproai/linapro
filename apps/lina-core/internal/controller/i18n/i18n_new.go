// This file defines the runtime i18n controller dependencies and constructor.

package i18n

import (
	i18napi "lina-core/api/i18n"
	i18nsvc "lina-core/internal/service/i18n"
)

// ControllerV1 implements the runtime i18n API controller.
type ControllerV1 struct {
	i18nSvc i18nsvc.Service // i18nSvc resolves request locale, runtime locales, and runtime messages.
}

// NewV1 creates and returns a new runtime i18n controller.
func NewV1() i18napi.II18NV1 {
	return &ControllerV1{
		i18nSvc: i18nsvc.New(),
	}
}
