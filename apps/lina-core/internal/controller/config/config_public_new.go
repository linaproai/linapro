// This file declares the public config controller used by unauthenticated
// frontend pages to load safe display settings.

package config

import (
	publicconfigapi "lina-core/api/publicconfig"
	hostconfig "lina-core/internal/service/config"
)

// PublicControllerV1 implements the public config API controller.
type PublicControllerV1 struct {
	configSvc hostconfig.Service
}

// NewPublicV1 creates and returns a new public config controller.
func NewPublicV1() publicconfigapi.IConfigPublicV1 {
	return &PublicControllerV1{
		configSvc: hostconfig.New(),
	}
}
