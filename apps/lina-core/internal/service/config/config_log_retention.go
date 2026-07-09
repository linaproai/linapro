// This file exposes global log-retention runtime settings backed by protected
// sys_config parameters.

package config

import (
	"context"
)

// GetLogRetentionDays returns the runtime-effective maximum log retention
// period in days from sys_config, static config, or host default metadata.
func (s *serviceImpl) GetLogRetentionDays(ctx context.Context) (int64, error) {
	value, err := s.getLogRetentionDaysValue(ctx)
	if err != nil {
		return 0, err
	}
	return validatePositiveInt64Value(RuntimeParamKeyLogRetentionDays, value)
}

// getLogRetentionDaysValue reads the default-aware raw value for log retention.
func (s *serviceImpl) getLogRetentionDaysValue(ctx context.Context) (string, error) {
	value, err := s.getProtectedConfigValueOrDefault(ctx, RuntimeParamKeyLogRetentionDays)
	if err != nil {
		return "", err
	}
	return value, nil
}
