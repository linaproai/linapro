// This file adapts the dynamic guest manifest host-call client to the shared
// manifestcap.Service contract used by source and dynamic plugins.

package pluginbridge

import (
	"context"

	"lina-core/pkg/plugin/capability/manifestcap"
)

// manifestCapabilityService adapts manifest.get transport calls to the
// plugin-scoped manifest resource capability contract.
type manifestCapabilityService struct {
	client ManifestHostService
}

var _ manifestcap.Service = (*manifestCapabilityService)(nil)

// manifestCapability returns the process-default manifest resource capability
// client.
func manifestCapability() manifestcap.Service {
	return &manifestCapabilityService{client: Manifest()}
}

// Get returns one raw resource under the current plugin manifest root.
func (s *manifestCapabilityService) Get(_ context.Context, path string) ([]byte, error) {
	content, _, err := s.client.Get(path)
	return content, err
}

// Exists reports whether one allowed manifest resource exists.
func (s *manifestCapabilityService) Exists(_ context.Context, path string) (bool, error) {
	_, found, err := s.client.Get(path)
	return found, err
}

// Scan unmarshals the selected YAML resource, or the nested key inside it, into target.
func (s *manifestCapabilityService) Scan(_ context.Context, path string, key string, target any) error {
	_, err := s.client.Scan(path, key, target)
	return err
}
