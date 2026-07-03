// This file implements static host configuration forwarding for plugin-visible
// HostConfig services while leaving governed sys_config behavior to the owner adapter.
package hostconfig

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	capabilityhostconfigcap "lina-core/pkg/plugin/capability/hostconfigcap"
)

// Get returns a static host configuration value.
func (s *hostConfigCapabilityService) Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error) {
	if s == nil || s.static == nil {
		return gvar.New(defaultValue), nil
	}
	return s.static.Get(ctx, key, defaultValue)
}

// Exists reports whether a static host configuration key exists.
func (s *hostConfigCapabilityService) Exists(ctx context.Context, key string) (bool, error) {
	if s == nil || s.static == nil {
		return false, nil
	}
	return s.static.Exists(ctx, key)
}

// String reads a static host configuration string.
func (s *hostConfigCapabilityService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	if s == nil || s.static == nil {
		return defaultValue, nil
	}
	return s.static.String(ctx, key, defaultValue)
}

// Bool reads a static host configuration boolean.
func (s *hostConfigCapabilityService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	if s == nil || s.static == nil {
		return defaultValue, nil
	}
	return s.static.Bool(ctx, key, defaultValue)
}

// Int reads a static host configuration integer.
func (s *hostConfigCapabilityService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	if s == nil || s.static == nil {
		return defaultValue, nil
	}
	return s.static.Int(ctx, key, defaultValue)
}

// Duration reads a static host configuration duration.
func (s *hostConfigCapabilityService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	if s == nil || s.static == nil {
		return defaultValue, nil
	}
	return s.static.Duration(ctx, key, defaultValue)
}

// SysConfig returns governed sys_config operations.
func (s *hostConfigCapabilityService) SysConfig() capabilityhostconfigcap.SysConfigService {
	if s == nil || s.sysConfig == nil {
		if s != nil && s.static != nil {
			return s.static.SysConfig()
		}
		return sysConfigUnavailableService{}
	}
	return s.sysConfig
}
