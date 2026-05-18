// This file contains the concrete GoFrame configuration adapter operations for
// source plugins. It keeps lookup, scan, and typed conversion behavior outside
// the package entrypoint while preserving the published config contract.

package config

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Get returns the raw configuration value for the given key.
func (s *serviceAdapter) Get(ctx context.Context, key string) (*gvar.Var, error) {
	normalizedKey := normalizeConfigKey(key)
	value, err := g.Cfg().Get(ctx, normalizedKey)
	if err != nil {
		return nil, gerror.Wrapf(err, "read plugin config %s failed", normalizedKey)
	}
	return value, nil
}

// Exists reports whether the given configuration key exists.
func (s *serviceAdapter) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return !isMissing(value), nil
}

// Scan scans the configuration section into target.
func (s *serviceAdapter) Scan(ctx context.Context, key string, target any) error {
	if target == nil {
		return gerror.New("plugin config scan target cannot be nil")
	}

	value, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	if isMissing(value) {
		return nil
	}
	if err := value.Scan(target); err != nil {
		return gerror.Wrapf(err, "scan plugin config %s failed", key)
	}
	return nil
}

// String reads a string value or returns defaultValue when the key is absent or blank.
func (s *serviceAdapter) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if isMissing(value) {
		return defaultValue, nil
	}

	raw := value.String()
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	return raw, nil
}

// Bool reads a bool value or returns defaultValue when the key is absent.
func (s *serviceAdapter) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if isMissing(value) {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

// Int reads an int value or returns defaultValue when the key is absent.
func (s *serviceAdapter) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if isMissing(value) {
		return defaultValue, nil
	}
	return value.Int(), nil
}

// Duration reads a time.Duration value from a duration string or returns defaultValue when the key is absent or blank.
func (s *serviceAdapter) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if isMissing(value) {
		return defaultValue, nil
	}

	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, gerror.Wrapf(err, "parse plugin config %s duration failed", key)
	}
	return duration, nil
}

// isMissing reports whether a GoFrame config lookup returned no concrete value.
func isMissing(value *gvar.Var) bool {
	return value == nil || value.IsNil()
}

// normalizeConfigKey maps blank keys to GoFrame's full-config lookup pattern.
func normalizeConfigKey(key string) string {
	if strings.TrimSpace(key) == "" {
		return "."
	}
	return key
}
