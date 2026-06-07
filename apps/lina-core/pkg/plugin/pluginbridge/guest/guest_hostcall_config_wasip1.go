//go:build wasip1

// This file provides guest-side helpers for the read-only config host service.

package guest

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/errors/gerror"
)

// configHostService is the default guest-side config host-service client.
type configHostService struct{}

// defaultConfigHostService stores the singleton config host-service client.
var defaultConfigHostService ConfigHostService = &configHostService{}

// pluginConfig returns the read-only config host service guest client.
func pluginConfig() ConfigHostService {
	return defaultConfigHostService
}

// Get reads one plugin-scoped configuration value as JSON.
func (*configHostService) Get(key string) (string, bool, error) {
	return configValue(protocol.HostServiceMethodConfigGet, key)
}

// Exists reports whether one configuration key exists.
func (*configHostService) Exists(key string) (bool, error) {
	_, found, err := configValue(protocol.HostServiceMethodConfigGet, key)
	return found, err
}

// String reads one configuration value as a string.
func (*configHostService) String(key string) (string, bool, error) {
	value, found, err := configValue(protocol.HostServiceMethodConfigGet, key)
	if err != nil || !found {
		return "", found, err
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	return strings.Trim(value, `"`), true, nil
}

// Bool reads one configuration value as a bool.
func (*configHostService) Bool(key string) (bool, bool, error) {
	value, found, err := configValue(protocol.HostServiceMethodConfigGet, key)
	if err != nil || !found {
		return false, found, err
	}
	var decoded bool
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, true, gerror.Wrapf(err, "parse config %s bool failed", key)
	}
	return parsed, true, nil
}

// Int reads one configuration value as an int.
func (*configHostService) Int(key string) (int, bool, error) {
	value, found, err := configValue(protocol.HostServiceMethodConfigGet, key)
	if err != nil || !found {
		return 0, found, err
	}
	var decoded int
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse config %s int failed", key)
	}
	return parsed, true, nil
}

// Duration reads one configuration value as a duration.
func (*configHostService) Duration(key string) (time.Duration, bool, error) {
	value, found, err := configValue(protocol.HostServiceMethodConfigGet, key)
	if err != nil || !found {
		return 0, found, err
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		value = decoded
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse config %s duration failed", key)
	}
	return parsed, true, nil
}

// configValue invokes one config host-service method and decodes the common response.
func configValue(method string, key string) (string, bool, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceConfig,
		method,
		"",
		"",
		protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{Key: key}),
	)
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	response, err := protocol.UnmarshalHostServiceConfigValueResponse(payload)
	if err != nil {
		return "", false, err
	}
	return response.Value, response.Found, nil
}
