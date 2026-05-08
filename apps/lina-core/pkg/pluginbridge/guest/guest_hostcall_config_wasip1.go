//go:build wasip1

// This file provides guest-side helpers for the read-only config host service.

package guest

import (
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ConfigHostService exposes guest-side helpers for the read-only config host service.
type ConfigHostService interface {
	// Get reads one configuration value as JSON. Empty key and "." read the full config snapshot.
	Get(key string) (string, bool, error)
	// Exists reports whether one configuration key exists.
	Exists(key string) (bool, error)
	// String reads one configuration value as a string.
	String(key string) (string, bool, error)
	// Bool reads one configuration value as a bool.
	Bool(key string) (bool, bool, error)
	// Int reads one configuration value as an int.
	Int(key string) (int, bool, error)
	// Duration reads one configuration value as a duration.
	Duration(key string) (time.Duration, bool, error)
}

// configHostService is the default guest-side config host-service client.
type configHostService struct{}

// defaultConfigHostService stores the singleton config host-service client.
var defaultConfigHostService ConfigHostService = &configHostService{}

// Config returns the read-only config host service guest client.
func Config() ConfigHostService {
	return defaultConfigHostService
}

// Get reads one configuration value as JSON. Empty key and "." read the full config snapshot.
func (*configHostService) Get(key string) (string, bool, error) {
	return configValue(HostServiceMethodConfigGet, key)
}

// Exists reports whether one configuration key exists.
func (*configHostService) Exists(key string) (bool, error) {
	_, found, err := configValue(HostServiceMethodConfigExists, key)
	return found, err
}

// String reads one configuration value as a string.
func (*configHostService) String(key string) (string, bool, error) {
	return configValue(HostServiceMethodConfigString, key)
}

// Bool reads one configuration value as a bool.
func (*configHostService) Bool(key string) (bool, bool, error) {
	value, found, err := configValue(HostServiceMethodConfigBool, key)
	if err != nil || !found {
		return false, found, err
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, true, gerror.Wrapf(err, "parse config %s bool failed", key)
	}
	return parsed, true, nil
}

// Int reads one configuration value as an int.
func (*configHostService) Int(key string) (int, bool, error) {
	value, found, err := configValue(HostServiceMethodConfigInt, key)
	if err != nil || !found {
		return 0, found, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse config %s int failed", key)
	}
	return parsed, true, nil
}

// Duration reads one configuration value as a duration.
func (*configHostService) Duration(key string) (time.Duration, bool, error) {
	value, found, err := configValue(HostServiceMethodConfigDuration, key)
	if err != nil || !found {
		return 0, found, err
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse config %s duration failed", key)
	}
	return parsed, true, nil
}

// configValue invokes one config host-service method and decodes the common response.
func configValue(method string, key string) (string, bool, error) {
	payload, err := invokeHostService(
		HostServiceConfig,
		method,
		"",
		"",
		MarshalHostServiceConfigKeyRequest(&HostServiceConfigKeyRequest{Key: key}),
	)
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	response, err := UnmarshalHostServiceConfigValueResponse(payload)
	if err != nil {
		return "", false, err
	}
	return response.Value, response.Found, nil
}
