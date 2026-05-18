// This file coordinates runtime i18n cache freshness checks against the shared
// plugin-runtime cache revision controller before translation bundles are read.

package i18n

import (
	"context"

	"lina-core/pkg/logger"
)

// EnsureRuntimeBundleCacheFresh synchronizes clustered plugin-runtime cache
// revisions before callers make HTTP cache decisions.
func (s *serviceImpl) EnsureRuntimeBundleCacheFresh(ctx context.Context) error {
	if s == nil || s.runtimeCacheRevisionCtl == nil {
		return nil
	}
	return s.runtimeCacheRevisionCtl.EnsureFresh(ctx)
}

// ensureRuntimeBundleCacheFreshBestEffort keeps translation read paths
// available while still surfacing cluster revision failures in logs.
func (s *serviceImpl) ensureRuntimeBundleCacheFreshBestEffort(ctx context.Context) {
	if err := s.EnsureRuntimeBundleCacheFresh(ctx); err != nil {
		logger.Warningf(ctx, "refresh runtime i18n cache failed: %v", err)
	}
}
