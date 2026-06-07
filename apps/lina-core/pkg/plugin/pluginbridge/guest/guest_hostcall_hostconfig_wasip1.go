//go:build wasip1

// This file provides guest-side helpers for authorized host config reads.

package guest

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/errors/gerror"
)

// hostConfigHostService is the default guest-side hostConfig client.
type hostConfigHostService struct{}

// defaultHostConfigHostService stores the singleton hostConfig client.
var defaultHostConfigHostService HostConfigHostService = &hostConfigHostService{}

// HostConfig returns the host config guest client.
func HostConfig() HostConfigHostService {
	return defaultHostConfigHostService
}

// Get reads one authorized host config value as JSON.
func (*hostConfigHostService) Get(key string) (string, bool, error) {
	return hostConfigValue(key)
}

// String reads one authorized host config value as a string.
func (*hostConfigHostService) String(key string) (string, bool, error) {
	value, found, err := hostConfigValue(key)
	if err != nil || !found {
		return "", found, err
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	return strings.Trim(value, `"`), true, nil
}

// Bool reads one authorized host config value as a bool.
func (*hostConfigHostService) Bool(key string) (bool, bool, error) {
	value, found, err := hostConfigValue(key)
	if err != nil || !found {
		return false, found, err
	}
	var decoded bool
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, true, gerror.Wrapf(err, "parse host config %s bool failed", key)
	}
	return parsed, true, nil
}

// Int reads one authorized host config value as an int.
func (*hostConfigHostService) Int(key string) (int, bool, error) {
	value, found, err := hostConfigValue(key)
	if err != nil || !found {
		return 0, found, err
	}
	var decoded int
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse host config %s int failed", key)
	}
	return parsed, true, nil
}

// Duration reads one authorized host config value as a duration.
func (*hostConfigHostService) Duration(key string) (time.Duration, bool, error) {
	value, found, err := hostConfigValue(key)
	if err != nil || !found {
		return 0, found, err
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		value = decoded
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse host config %s duration failed", key)
	}
	return parsed, true, nil
}

// hostConfigValue invokes hostConfig.get and decodes the common value response.
func hostConfigValue(key string) (string, bool, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceHostConfig,
		protocol.HostServiceMethodHostConfigGet,
		key,
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
