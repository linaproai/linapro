// This file exposes raw host configuration reads for trusted internal
// adapters that need business-neutral access without expanding Service.

package config

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// GetRaw returns the raw host configuration value for key. Empty key and "."
// follow GoFrame semantics and return the full static configuration snapshot.
// Managed sys_config keys return their runtime-effective values.
func (s *serviceImpl) GetRaw(ctx context.Context, key string) (*gvar.Var, error) {
	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == RuntimeParamKeyLogRetentionDays {
		value, err := s.getRequiredLogRetentionDaysValue(ctx)
		if err != nil {
			return nil, err
		}
		return gvar.New(value), nil
	}
	if IsManagedSysConfigKey(normalizedKey) {
		value, err := s.getProtectedConfigValueOrDefault(ctx, normalizedKey)
		if err != nil {
			return nil, err
		}
		return gvar.New(value), nil
	}
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		return nil, gerror.Wrapf(err, "read host config key failed key=%s", key)
	}
	return value, nil
}
