// This file keeps runtime artifact cache instrumentation scoped to tests.

package catalog

import (
	"path/filepath"
	"strings"
)

// runtimeArtifactParseCount returns how many times this service fully parsed an artifact path.
func (s *serviceImpl) runtimeArtifactParseCount(artifactPath string) int {
	if s == nil {
		return 0
	}
	key := filepath.Clean(strings.TrimSpace(artifactPath))
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.parseCounts[key]
}
