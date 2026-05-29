// Package manifest exposes plugin-owned manifest/ files as read-only raw
// resources to source and dynamic plugins. Config, SQL, and i18n files remain
// governed by their dedicated lifecycle pipelines when they need to take effect.
package manifest

import (
	"io/fs"
	"strings"

	"lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// serviceAdapter reads raw resources under one plugin manifest root.
type serviceAdapter struct {
	pluginID          string
	developmentRoot   string
	embeddedFiles     fs.FS
	artifactResources map[string][]byte
}

// NewFactory creates a manifest service factory.
func NewFactory(developmentRoot string) contract.ManifestServiceFactory {
	return &serviceAdapter{developmentRoot: strings.TrimSpace(developmentRoot)}
}

// ForPlugin returns a manifest reader scoped to pluginID.
func (s *serviceAdapter) ForPlugin(pluginID string) contract.ManifestService {
	clone := s.clone()
	clone.pluginID = strings.TrimSpace(pluginID)
	if sourcePlugin, ok := pluginhost.GetSourcePlugin(clone.pluginID); ok && sourcePlugin != nil {
		clone.embeddedFiles = sourcePlugin.GetEmbeddedFiles()
	}
	return clone
}

// WithArtifactResources returns a factory clone carrying release-bound manifest
// resources for pluginID. Resource paths are relative to manifest/.
func (s *serviceAdapter) WithArtifactResources(pluginID string, resources map[string][]byte) contract.ManifestServiceFactory {
	clone := s.clone()
	if strings.TrimSpace(pluginID) == "" || len(resources) == 0 {
		return clone
	}
	if clone.artifactResources == nil {
		clone.artifactResources = make(map[string][]byte)
	}
	for path, content := range resources {
		clone.artifactResources[strings.TrimSpace(pluginID)+"\x00"+path] = append([]byte(nil), content...)
	}
	return clone
}

// clone returns a detached adapter copy.
func (s *serviceAdapter) clone() *serviceAdapter {
	if s == nil {
		return &serviceAdapter{}
	}
	clone := &serviceAdapter{
		pluginID:        s.pluginID,
		developmentRoot: s.developmentRoot,
		embeddedFiles:   s.embeddedFiles,
	}
	if len(s.artifactResources) > 0 {
		clone.artifactResources = make(map[string][]byte, len(s.artifactResources))
		for key, content := range s.artifactResources {
			clone.artifactResources[key] = append([]byte(nil), content...)
		}
	}
	return clone
}
