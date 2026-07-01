// This file declares the plugin controller constructor and its explicit
// service dependencies.

package plugin

import (
	pluginapi "lina-core/api/plugin"
	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
)

// ControllerV1 is the plugin controller.
type ControllerV1 struct {
	pluginSvc pluginsvc.Service // unified plugin service
	bizCtxSvc bizctx.Service    // business context service
	configSvc configsvc.Service // config service
	i18nSvc   i18nsvc.Service   // i18n error localization service
	roleSvc   role.Service      // role service
}

// NewV1 creates and returns a new plugin controller instance.
func NewV1(pluginSvc pluginsvc.Service, bizCtxSvc bizctx.Service, configSvc configsvc.Service, i18nSvc i18nsvc.Service, roleSvc role.Service) pluginapi.IPluginV1 {
	return &ControllerV1{
		pluginSvc: pluginSvc,
		bizCtxSvc: bizCtxSvc,
		configSvc: configSvc,
		i18nSvc:   i18nSvc,
		roleSvc:   roleSvc,
	}
}
