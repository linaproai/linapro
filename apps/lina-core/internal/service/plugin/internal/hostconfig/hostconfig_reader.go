// This file adapts the startup-owned raw host configuration reader to the
// plugin-visible HostConfig capability contract. The adapter stays in the
// plugin host internals so the public hostconfigcap package remains a contract
// package instead of a concrete implementation owner.

package hostconfig

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"

	capabilityhostconfigcap "lina-core/pkg/plugin/capability/hostconfigcap"
)

// rawHostConfigReader is the host config slice required by this adapter.
type rawHostConfigReader interface {
	// GetRaw returns one raw host configuration value or root snapshot for key.
	GetRaw(ctx context.Context, key string) (*gvar.Var, error)
}

// staticConfigAdapter reads individual host config keys from the host config service.
type staticConfigAdapter struct {
	configSvc rawHostConfigReader
}

var _ capabilityhostconfigcap.Service = (*staticConfigAdapter)(nil)

// NewStaticCapabilityAdapter creates a host config reader backed by the host config service.
func NewStaticCapabilityAdapter(configSvc rawHostConfigReader) capabilityhostconfigcap.Service {
	return &staticConfigAdapter{configSvc: configSvc}
}

// Get returns the raw host config value for the requested key or root snapshot.
func (s *staticConfigAdapter) Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error) {
	normalizedKey := strings.TrimSpace(key)
	value, err := s.valueForKey(ctx, normalizedKey)
	if err != nil {
		return nil, err
	}
	if value == nil || value.IsNil() {
		if defaultValue != nil {
			return gvar.New(defaultValue), nil
		}
		return nil, nil
	}
	return value, nil
}

// Exists reports whether a host config key is available.
func (s *staticConfigAdapter) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil {
		return false, err
	}
	return value != nil && !value.IsNil(), nil
}

// String reads a host config string value or returns defaultValue when the key is absent or blank.
func (s *staticConfigAdapter) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil {
		return "", err
	}
	if value == nil || value.IsNil() || strings.TrimSpace(value.String()) == "" {
		return defaultValue, nil
	}
	return value.String(), nil
}

// Bool reads a host config bool value or returns defaultValue when the key is absent.
func (s *staticConfigAdapter) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

// Int reads a host config int value or returns defaultValue when the key is absent.
func (s *staticConfigAdapter) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Int(), nil
}

// Duration reads a host config duration value or returns defaultValue when the key is absent or blank.
func (s *staticConfigAdapter) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, gerror.Wrapf(err, "parse host config %s duration failed", key)
	}
	return duration, nil
}

// SysConfig reports that this static config adapter has no sys_config backend.
func (s *staticConfigAdapter) SysConfig() capabilityhostconfigcap.SysConfigService {
	return sysConfigUnavailableService{}
}

// valueForKey returns one host config value without applying a key allowlist.
func (s *staticConfigAdapter) valueForKey(ctx context.Context, key string) (*gvar.Var, error) {
	if s == nil || s.configSvc == nil {
		return nil, gerror.New("host config service is not configured")
	}
	return s.configSvc.GetRaw(ctx, key)
}
