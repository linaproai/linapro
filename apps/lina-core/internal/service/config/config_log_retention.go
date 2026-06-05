// This file exposes global log-retention runtime settings backed by protected
// sys_config parameters.

package config

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
)

// GetLogRetentionDays returns the runtime-effective maximum log retention
// period in days from the required sys_config seed row.
func (s *serviceImpl) GetLogRetentionDays(ctx context.Context) (int64, error) {
	value, err := s.getRequiredLogRetentionDaysValue(ctx)
	if err != nil {
		return 0, err
	}
	return validatePositiveInt64Value(RuntimeParamKeyLogRetentionDays, value)
}

// getRequiredLogRetentionDaysValue reads the raw sys_config value for the
// log-retention seed row. The row is part of delivery SQL and must exist.
func (s *serviceImpl) getRequiredLogRetentionDaysValue(ctx context.Context) (string, error) {
	value, ok, err := s.lookupRuntimeParamValue(ctx, RuntimeParamKeyLogRetentionDays)
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return "", bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", RuntimeParamKeyLogRetentionDays))
	}
	return value, nil
}
