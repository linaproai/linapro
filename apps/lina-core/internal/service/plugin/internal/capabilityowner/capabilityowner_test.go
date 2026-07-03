// This file verifies source-plugin plugin config capability wiring uses the
// startup-injected config factory instead of constructing an isolated one.

package capabilityowner

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
)

// TestServicesForPluginUsesInjectedPluginConfigFactory verifies scoped source
// plugin services read config through the startup-provided factory.
func TestServicesForPluginUsesInjectedPluginConfigFactory(t *testing.T) {
	factory := &recordingConfigFactory{service: recordingConfigService{
		values: map[string]any{"storage.endpoint": "host"},
	}}
	plugins := NewCapabilityAdapter(factory, nil, nil, nil, nil)

	config := plugins.ForPlugin("source-plugin-a").Config()
	value, err := config.String(context.Background(), "storage.endpoint", "")
	if err != nil {
		t.Fatalf("read scoped plugin config: %v", err)
	}
	if value != "host" {
		t.Fatalf("expected injected config factory value, got %q", value)
	}
	if factory.lastPluginID != "source-plugin-a" {
		t.Fatalf("expected factory to be scoped to source-plugin-a, got %q", factory.lastPluginID)
	}
}

// TestSetTenantPluginEnabledRequiresCacheCoord verifies tenant enablement writes
// fail before touching plugin state when cachecoord is not injected.
func TestSetTenantPluginEnabledRequiresCacheCoord(t *testing.T) {
	adapter := &pluginCapabilityAdapter{}
	err := adapter.setTenantPluginEnabled(context.Background(), 1, "source-plugin-a", true, false)
	if !bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
		t.Fatalf("expected cachecoord unavailable error, got %v", err)
	}
}

// recordingConfigFactory records plugin scopes requested by source-plugin
// service wiring.
type recordingConfigFactory struct {
	service      recordingConfigService
	lastPluginID string
}

// ForPlugin returns the configured recording service.
func (f *recordingConfigFactory) ForPlugin(pluginID string) plugincap.ConfigService {
	f.lastPluginID = strings.TrimSpace(pluginID)
	return f.service
}

// WithArtifactConfig records no behavior because source-plugin wiring does
// not bind artifact defaults.
func (f *recordingConfigFactory) WithArtifactConfig(string, []byte) pluginconfig.Factory {
	return f
}

// recordingConfigService returns deterministic values for config calls.
type recordingConfigService struct {
	values map[string]any
}

// Get returns one deterministic raw config value.
func (s recordingConfigService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value, ok := s.values[key]
	if !ok {
		if defaultValue != nil {
			return gvar.New(defaultValue), nil
		}
		return nil, nil
	}
	return gvar.New(value), nil
}

// Exists reports whether a deterministic config value is configured.
func (s recordingConfigService) Exists(_ context.Context, key string) (bool, error) {
	_, ok := s.values[key]
	return ok, nil
}

// Scan is unused by this wiring test.
func (s recordingConfigService) Scan(context.Context, string, any) error { return nil }

// String reads a deterministic string value.
func (s recordingConfigService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads a deterministic bool value.
func (s recordingConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads a deterministic integer value.
func (s recordingConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration reads a deterministic duration string value.
func (s recordingConfigService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return time.ParseDuration(value.String())
}
