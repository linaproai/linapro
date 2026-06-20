// This file verifies source-plugin plugin config capability wiring uses the
// startup-injected config factory instead of constructing an isolated one.

package capabilityhost

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/plugincap"
)

// TestServicesForPluginUsesInjectedPluginConfigFactory verifies scoped source
// plugin services read config through the startup-provided factory.
func TestServicesForPluginUsesInjectedPluginConfigFactory(t *testing.T) {
	factory := &recordingConfigFactory{service: recordingConfigService{
		values: map[string]any{"storage.endpoint": "host"},
	}}
	services := &directory{
		plugins: newPluginCapabilityAdapter(factory, nil, nil, nil, nil, nil),
	}

	config := capability.ServicesForPlugin(services, "source-plugin-a").Plugins().Config()
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
func (f *recordingConfigFactory) WithArtifactConfig(string, []byte) plugincap.ConfigServiceFactory {
	return f
}

// recordingConfigService returns deterministic values for config calls.
type recordingConfigService struct {
	values map[string]any
}

// Get returns one deterministic raw config value.
func (s recordingConfigService) Get(_ context.Context, key string) (*gvar.Var, error) {
	value, ok := s.values[key]
	if !ok {
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
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads a deterministic bool value.
func (s recordingConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads a deterministic integer value.
func (s recordingConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration reads a deterministic duration string value.
func (s recordingConfigService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return time.ParseDuration(value.String())
}
