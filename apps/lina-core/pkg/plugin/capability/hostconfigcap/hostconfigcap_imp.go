// This file implements read-only host config key access for source plugins.

package hostconfigcap

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
)

// RawConfigReader is implemented by the host-owned config service. It keeps
// this adapter dependent on the startup-injected config instance without
// coupling this public capability package to host internal service packages.
type RawConfigReader interface {
	// GetRaw returns one raw host configuration value or root snapshot.
	GetRaw(ctx context.Context, key string) (*gvar.Var, error)
}

// serviceAdapter reads individual host config keys from the host config service.
type serviceAdapter struct {
	configSvc RawConfigReader
}

// New creates a host config reader backed by the host config service.
func New(configSvc RawConfigReader) Service {
	return &serviceAdapter{configSvc: configSvc}
}

// Get returns the raw host config value for the requested key or root snapshot.
func (s *serviceAdapter) Get(ctx context.Context, key string) (*gvar.Var, error) {
	normalizedKey, err := normalizeHostConfigKey(key)
	if err != nil {
		return nil, err
	}
	value, err := s.valueForKey(ctx, normalizedKey)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return value, nil
}

// Exists reports whether a host config key is available.
func (s *serviceAdapter) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return value != nil && !value.IsNil(), nil
}

// String reads a host config string value or returns defaultValue when the key is absent or blank.
func (s *serviceAdapter) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if value == nil || value.IsNil() || strings.TrimSpace(value.String()) == "" {
		return defaultValue, nil
	}
	return value.String(), nil
}

// Bool reads a host config bool value or returns defaultValue when the key is absent.
func (s *serviceAdapter) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

// Int reads a host config int value or returns defaultValue when the key is absent.
func (s *serviceAdapter) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Int(), nil
}

// Duration reads a host config duration value or returns defaultValue when the key is absent or blank.
func (s *serviceAdapter) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key)
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

// normalizeHostConfigKey normalizes host config lookups without applying any
// source-plugin key restriction.
func normalizeHostConfigKey(key string) (string, error) {
	return strings.TrimSpace(key), nil
}

// valueForKey returns one host config value without applying a key allowlist.
func (s *serviceAdapter) valueForKey(ctx context.Context, key string) (*gvar.Var, error) {
	if s == nil || s.configSvc == nil {
		return nil, gerror.New("host config service is not configured")
	}
	return s.configSvc.GetRaw(ctx, key)
}
