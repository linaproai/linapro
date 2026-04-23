package plugin

import (
	pluginapi "lina-core/api/plugin"
	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
)

// ControllerV1 is the plugin controller.
type ControllerV1 struct {
	pluginSvc pluginsvc.Service // plugin service
	bizCtxSvc bizctx.Service    // business context service
	configSvc configsvc.Service // config service
	roleSvc   role.Service      // role service
}

// NewV1 creates and returns a new plugin controller instance.
// Pass a non-nil topology for cluster-aware plugin orchestration; pass nil to
// use the default single-node plugin topology.
func NewV1(topology pluginsvc.Topology) pluginapi.IPluginV1 {
	pluginSvc := pluginsvc.New(topology)
	return &ControllerV1{
		pluginSvc: pluginSvc,
		bizCtxSvc: bizctx.New(),
		configSvc: configsvc.New(),
		roleSvc:   role.New(pluginSvc),
	}
}
