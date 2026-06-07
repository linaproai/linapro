// This file defines plugin-owned manifest/ files as read-only raw resources for
// source and dynamic plugins. Config, SQL, and i18n files remain governed by
// their dedicated lifecycle pipelines when they need to take effect.
package manifestcap

import (
	"io/fs"
	"strings"
)

// EmbeddedFilesResolver returns embedded manifest resources for one source
// plugin. It is injected by the host runtime so this public capability package
// does not depend on pluginhost.
type EmbeddedFilesResolver func(pluginID string) fs.FS

// serviceAdapter reads raw resources under one plugin manifest root.
type serviceAdapter struct {
	pluginID          string
	developmentRoot   string
	embeddedResolver  EmbeddedFilesResolver
	embeddedFiles     fs.FS
	artifactResources map[string][]byte
}

// NewFactory creates a manifest service factory.
func NewFactory(developmentRoot string, embeddedResolvers ...EmbeddedFilesResolver) ServiceFactory {
	var resolver EmbeddedFilesResolver
	if len(embeddedResolvers) > 0 {
		resolver = embeddedResolvers[0]
	}
	return &serviceAdapter{
		developmentRoot:  strings.TrimSpace(developmentRoot),
		embeddedResolver: resolver,
	}
}

// ForPlugin returns a manifest reader scoped to pluginID.
func (s *serviceAdapter) ForPlugin(pluginID string) Service {
	clone := s.clone()
	clone.pluginID = strings.TrimSpace(pluginID)
	if clone.embeddedResolver != nil {
		clone.embeddedFiles = clone.embeddedResolver(clone.pluginID)
	}
	return clone
}

// WithArtifactResources returns a factory clone carrying release-bound manifest
// resources for pluginID. Resource paths are relative to manifest/.
func (s *serviceAdapter) WithArtifactResources(pluginID string, resources map[string][]byte) ServiceFactory {
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
		pluginID:         s.pluginID,
		developmentRoot:  s.developmentRoot,
		embeddedResolver: s.embeddedResolver,
		embeddedFiles:    s.embeddedFiles,
	}
	if len(s.artifactResources) > 0 {
		clone.artifactResources = make(map[string][]byte, len(s.artifactResources))
		for key, content := range s.artifactResources {
			clone.artifactResources[key] = append([]byte(nil), content...)
		}
	}
	return clone
}
