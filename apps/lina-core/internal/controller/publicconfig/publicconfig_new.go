// This file defines the public frontend-config controller dependencies and constructor.

package publicconfig

import (
	"lina-core/api/publicconfig"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
)

// ControllerV1 implements the public frontend-config API controller.
type ControllerV1 struct {
	configSvc hostconfig.Service // configSvc loads the public-safe frontend config snapshot.
	i18nSvc   i18nsvc.Translator // i18nSvc localizes public-facing frontend copy.
}

// NewV1 creates and returns a new public frontend-config controller.
func NewV1() publicconfig.IPublicconfigV1 {
	return &ControllerV1{
		configSvc: hostconfig.New(),
		i18nSvc:   i18nsvc.New(),
	}
}
