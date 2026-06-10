// This file adapts the dynamic guest host-config host-call client to the
// shared hostconfigcap.Service contract used by source and dynamic plugins.

package pluginbridge

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// hostConfigCapabilityService adapts hostConfig.get transport calls to the
// read-only host configuration capability contract.
type hostConfigCapabilityService struct {
	client HostConfigHostService
}

var _ hostconfigcap.Service = (*hostConfigCapabilityService)(nil)

// hostConfigCapability returns the process-default host configuration
// capability client.
func hostConfigCapability() hostconfigcap.Service {
	return &hostConfigCapabilityService{client: HostConfig()}
}

// Get returns the raw host config value for the requested key.
func (s *hostConfigCapabilityService) Get(_ context.Context, key string) (*gvar.Var, error) {
	value, found, err := s.client.Get(key)
	if err != nil || !found {
		return nil, err
	}
	var decoded any
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return gvar.New(decoded), nil
	}
	return gvar.New(value), nil
}

// Exists reports whether a host config key is available.
func (s *hostConfigCapabilityService) Exists(_ context.Context, key string) (bool, error) {
	_, found, err := s.client.Get(key)
	return found, err
}

// String reads a host config string value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) String(_ context.Context, key string, defaultValue string) (string, error) {
	value, found, err := s.client.String(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// Bool reads a host config bool value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	value, found, err := s.client.Bool(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// Int reads a host config int value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) Int(_ context.Context, key string, defaultValue int) (int, error) {
	value, found, err := s.client.Int(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// Duration reads a host config duration value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) Duration(_ context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, found, err := s.client.Duration(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}
