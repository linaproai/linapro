// This file defines business-neutral read-only plugin configuration access. It
// resolves config.yaml inside the current plugin scope instead of reading the
// host-wide GoFrame configuration tree.
package plugincap

import (
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/os/gcfg"
)

const (
	// RuntimeConfigFileName is the only plugin runtime config filename read by
	// the generic service.
	RuntimeConfigFileName = "config.yaml"
	// TemplateConfigFileName is the plugin config template filename. The service
	// deliberately never reads it as runtime defaults.
	TemplateConfigFileName = "config.example.yaml"
)

// serviceAdapter resolves one plugin-scoped config view from ordered sources.
type serviceAdapter struct {
	pluginID        string
	productionRoot  string
	developmentRoot string
	artifactConfigs map[string][]byte
}

// NewConfig creates and returns the published plugin config service adapter.
func NewConfig() ConfigService {
	return &serviceAdapter{}
}

// NewConfigFactory creates a config service factory with optional root overrides.
func NewConfigFactory(productionRoot string, developmentRoot string) ConfigServiceFactory {
	return &serviceAdapter{
		productionRoot:  strings.TrimSpace(productionRoot),
		developmentRoot: strings.TrimSpace(developmentRoot),
	}
}

// ForPlugin returns a service scoped to pluginID.
func (s *serviceAdapter) ForPlugin(pluginID string) ConfigService {
	clone := s.clone()
	clone.pluginID = strings.TrimSpace(pluginID)
	return clone
}

// WithArtifactConfig returns a factory clone with a release-bound default
// config snapshot for pluginID.
func (s *serviceAdapter) WithArtifactConfig(pluginID string, artifactContent []byte) ConfigServiceFactory {
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
