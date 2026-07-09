// Package pluginconfig creates plugin-scoped configuration readers for source
// plugin capability directories and dynamic-plugin WASM host services.
package pluginconfig

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/plugin/capability/plugincap"
)

const (
	// RuntimeConfigFileName is the only plugin runtime config filename read by
	// the generic service.
	RuntimeConfigFileName = "config.yaml"
	// TemplateConfigFileName is the plugin config template filename. The service
	// deliberately never reads it as runtime defaults.
	TemplateConfigFileName = "config.example.yaml"
)

// Factory creates plugin-scoped configuration service views.
type Factory interface {
	// ForPlugin returns a configuration service scoped to pluginID. A blank
	// pluginID returns a service that rejects reads rather than falling back to
	// host-wide configuration.
	ForPlugin(pluginID string) plugincap.ConfigService
	// WithArtifactConfig returns a new factory view that can use artifactContent
	// as the release-bound default config for pluginID when no external or
	// development config file exists.
	WithArtifactConfig(pluginID string, artifactContent []byte) Factory
}

// rawHostConfigReader is the host static config slice required by plugin config views.
type rawHostConfigReader interface {
	// GetRaw returns one raw host configuration value or root snapshot for key.
	GetRaw(ctx context.Context, key string) (*gvar.Var, error)
}

// serviceAdapter resolves one plugin-scoped config view from ordered sources.
type serviceAdapter struct {
	pluginID        string
	productionRoot  string
	developmentRoot string
	hostStatic      rawHostConfigReader
	artifactConfigs map[string][]byte
}

var (
	_ Factory                 = (*serviceAdapter)(nil)
	_ plugincap.ConfigService = (*serviceAdapter)(nil)
)

// NewFactory creates a config service factory with optional root overrides.
func NewFactory(productionRoot string, developmentRoot string) Factory {
	return &serviceAdapter{
		productionRoot:  strings.TrimSpace(productionRoot),
		developmentRoot: strings.TrimSpace(developmentRoot),
	}
}

// NewFactoryWithHostStaticConfig creates a config service factory that checks
// host static plugin.<plugin-id> sections before file and artifact sources.
func NewFactoryWithHostStaticConfig(productionRoot string, developmentRoot string, hostStatic rawHostConfigReader) Factory {
	return &serviceAdapter{
		productionRoot:  strings.TrimSpace(productionRoot),
		developmentRoot: strings.TrimSpace(developmentRoot),
		hostStatic:      hostStatic,
	}
}

// ForPlugin returns a service scoped to pluginID.
func (s *serviceAdapter) ForPlugin(pluginID string) plugincap.ConfigService {
	clone := s.clone()
	clone.pluginID = strings.TrimSpace(pluginID)
	return clone
}

// WithArtifactConfig returns a factory clone with a release-bound default
// config snapshot for pluginID.
func (s *serviceAdapter) WithArtifactConfig(pluginID string, artifactContent []byte) Factory {
	clone := s.clone()
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" || len(artifactContent) == 0 {
		return clone
	}
	if clone.artifactConfigs == nil {
		clone.artifactConfigs = make(map[string][]byte)
	}
	clone.artifactConfigs[normalizedPluginID] = append([]byte(nil), artifactContent...)
	return clone
}

// clone returns a detached adapter copy so plugin-scoped views do not mutate
// the base factory state.
func (s *serviceAdapter) clone() *serviceAdapter {
	if s == nil {
		return &serviceAdapter{}
	}
	clone := &serviceAdapter{
		pluginID:        s.pluginID,
		productionRoot:  s.productionRoot,
		developmentRoot: s.developmentRoot,
		hostStatic:      s.hostStatic,
	}
	if len(s.artifactConfigs) > 0 {
		clone.artifactConfigs = make(map[string][]byte, len(s.artifactConfigs))
		for pluginID, content := range s.artifactConfigs {
			clone.artifactConfigs[pluginID] = append([]byte(nil), content...)
		}
	}
	return clone
}

// buildConfigFromContent creates a GoFrame config object from YAML content.
func buildConfigFromContent(content []byte) (*gcfg.Config, error) {
	adapter, err := gcfg.NewAdapterContent(string(content))
	if err != nil {
		return nil, err
	}
	return gcfg.NewWithAdapter(adapter), nil
}

// buildConfigFromFile creates a GoFrame config object pinned to one concrete
// config.yaml path.
func buildConfigFromFile(filePath string) (*gcfg.Config, error) {
	adapter, err := gcfg.NewAdapterFile(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	return gcfg.NewWithAdapter(adapter), nil
}

// jsonConfigReader adapts a gjson section to the same read operation used by
// file-backed GoFrame configs.
type jsonConfigReader struct {
	doc *gjson.Json
}

// Get returns one value from the JSON configuration section.
func (r *jsonConfigReader) Get(_ context.Context, key string, def ...any) (*gvar.Var, error) {
	if r == nil || r.doc == nil {
		return nil, nil
	}
	value := r.doc.Get(key, def...)
	return value, nil
}
