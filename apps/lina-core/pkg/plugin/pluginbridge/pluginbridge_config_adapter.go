// This file adapts the dynamic-plugin config host-service client to the shared
// plugincap.ConfigService contract used by source and dynamic plugins.

package pluginbridge

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/plugincap"
)

// pluginConfigCapability returns the plugin-scoped config capability client.
func pluginConfigCapability() plugincap.ConfigService {
	return pluginConfigCapabilityService{}
}

// pluginConfigCapabilityService adapts plugins.config.get host calls to plugincap.ConfigService.
type pluginConfigCapabilityService struct{}

// Get returns the raw configuration value for the given key.
func (s pluginConfigCapabilityService) Get(_ context.Context, key string) (*gvar.Var, error) {
	value, found, err := configValue(key)
	if err != nil || !found {
		return nil, err
	}
	var decoded any
	if decodeErr := json.Unmarshal([]byte(value), &decoded); decodeErr != nil {
		decoded = value
	}
	return gvar.New(decoded), nil
}

// Exists reports whether the given configuration key exists.
func (s pluginConfigCapabilityService) Exists(_ context.Context, key string) (bool, error) {
	_, found, err := configValue(key)
	return found, err
}

// Scan scans the configuration section into target.
func (s pluginConfigCapabilityService) Scan(ctx context.Context, key string, target any) error {
	if target == nil {
		return gerror.New("plugin config scan target cannot be nil")
	}
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return err
	}
	return value.Scan(target)
}

// String reads a string value or returns defaultValue when the key is absent or blank.
func (s pluginConfigCapabilityService) String(_ context.Context, key string, defaultValue string) (string, error) {
	value, found, err := configValue(key)
	if err != nil {
		return "", err
	}
	if !found || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, nil
	}
	value = strings.Trim(value, `"`)
	return value, nil
}

// Bool reads a bool value or returns defaultValue when the key is absent.
func (s pluginConfigCapabilityService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	value, found, err := configValue(key)
	if err != nil {
		return false, err
	}
	if !found {
		return defaultValue, nil
	}
	var decoded bool
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, gerror.Wrapf(err, "parse config %s bool failed", key)
	}
	return parsed, nil
}

// Int reads an int value or returns defaultValue when the key is absent.
func (s pluginConfigCapabilityService) Int(_ context.Context, key string, defaultValue int) (int, error) {
	value, found, err := configValue(key)
	if err != nil {
		return 0, err
	}
	if !found {
		return defaultValue, nil
	}
	var decoded int
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, gerror.Wrapf(err, "parse config %s int failed", key)
	}
	return parsed, nil
}

// Duration reads a time.Duration value or returns defaultValue when the key is absent or blank.
func (s pluginConfigCapabilityService) Duration(_ context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, found, err := configValue(key)
	if err != nil {
		return 0, err
	}
	if !found {
		return defaultValue, nil
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		value = decoded
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, gerror.Wrapf(err, "parse config %s duration failed", key)
	}
	return parsed, nil
}

var _ plugincap.ConfigService = (*pluginConfigCapabilityService)(nil)
