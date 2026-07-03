// Package hostconfig adapts static host configuration reads and governed
// sys_config rows to plugin-visible HostConfig capability contracts.
package hostconfig

import (
	"lina-core/internal/service/cachecoord"
	capabilityhostconfigcap "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// hostConfigCapabilityService combines static host config reads with governed
// sys_config operations under the unified hostconfigcap.Service.
type hostConfigCapabilityService struct {
	static    capabilityhostconfigcap.Service
	sysConfig capabilityhostconfigcap.SysConfigService
}

var _ capabilityhostconfigcap.Service = (*hostConfigCapabilityService)(nil)

// NewCapabilityService combines static host config reads with sys_config operations.
func NewCapabilityService(
	static capabilityhostconfigcap.Service,
	sysConfig capabilityhostconfigcap.SysConfigService,
) capabilityhostconfigcap.Service {
	return &hostConfigCapabilityService{static: static, sysConfig: sysConfig}
}

// sysConfigCapabilityAdapter exposes governed sys_config projections.
type sysConfigCapabilityAdapter struct {
	tenantFilter tenantcap.FilterService
	cacheCoord   cachecoord.Service
}

// sysConfigUnavailableService reports that no sys_config backend was injected.
type sysConfigUnavailableService struct{}

// NewSysConfigCapabilityAdapter creates the host-owned sys_config capability adapter.
func NewSysConfigCapabilityAdapter(
	tenantFilter tenantcap.FilterService,
	cacheCoord cachecoord.Service,
) capabilityhostconfigcap.SysConfigService {
	return &sysConfigCapabilityAdapter{tenantFilter: tenantFilter, cacheCoord: cacheCoord}
}
