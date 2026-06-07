// This file owns the root plugin management list read-model cache, including
// startup prewarm, locale-sensitive cache keys, and runtime revision invalidation.

package plugin

import (
	"context"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/management"
)

// PrewarmManagementList builds the lightweight plugin management summary read
// model so the first administrator request can reuse hot discovery projections.
// Failures are returned to foreground callers and logged by
// asynchronous startup callers.
func (s *serviceImpl) PrewarmManagementList(ctx context.Context) error {
	if _, err := s.managementSummaryList(ctx); err != nil {
		return err
	}
	return nil
}

// managementSummaryList returns the unfiltered plugin management summary read model.
func (s *serviceImpl) managementSummaryList(ctx context.Context) (*ListOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	cacheKey, err := s.managementListCacheKey(ctx)
	if err != nil {
		return nil, err
	}
	out, err := s.managementListCache.LoadOrBuild(cacheKey, func() (*ListOutput, error) {
		return s.buildManagementSummaryList(ctx)
	})
	if err != nil {
		return nil, err
	}
	latestKey, err := s.managementListCacheKey(ctx)
	if err != nil {
		return nil, err
	}
	if latestKey.String() != cacheKey.String() {
		s.managementListCache.Store(latestKey, out)
	}
	return out, nil
}

// InvalidateManagementListCache clears this process-local read model. Cluster
// peers observe the same plugin-runtime revision and invalidate through the
// root runtime-cache refresh callback.
func (s *serviceImpl) InvalidateManagementListCache(_ context.Context, _ string) {
	if s == nil || s.managementListCache == nil {
		return
	}
	s.managementListCache.Invalidate()
}

// managementListCacheKey returns the current cache partition because plugin
// display metadata is localized during projection and can change when the
// runtime translation bundle version or plugin-runtime revision changes.
func (s *serviceImpl) managementListCacheKey(ctx context.Context) (management.ListCacheKey, error) {
	if s == nil || s.i18nSvc == nil {
		return management.ListCacheKey{Locale: i18nsvc.DefaultLocale}, nil
	}
	locale := normalizeManagementListCacheLocale(s.i18nSvc.GetLocale(ctx))
	runtimeRevision := int64(0)
	if s.runtimeCacheRevisionCtrl != nil {
		revision, err := s.runtimeCacheRevisionCtrl.CurrentRevision(ctx)
		if err != nil {
			return management.ListCacheKey{}, err
		}
		runtimeRevision = revision
	}
	return management.ListCacheKey{
		Locale:               locale,
		RuntimeBundleVersion: s.i18nSvc.BundleVersion(locale),
		RuntimeRevision:      runtimeRevision,
	}, nil
}

// normalizeManagementListCacheLocale keeps cache keys stable for detached
// startup contexts and tests that do not carry business locale metadata.
func normalizeManagementListCacheLocale(locale string) string {
	if locale == "" {
		return i18nsvc.DefaultLocale
	}
	return locale
}
