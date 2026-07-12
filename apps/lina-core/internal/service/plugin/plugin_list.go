// This file exposes root-facade list and manifest synchronization methods.

package plugin

import (
	"context"
	"strings"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/management"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/statusflag"
)

// SyncSourcePluginsStrict synchronizes source plugins discovered by the
// running host. Tooling is responsible for official submodule preflight before
// plugin-full operations reach the runtime API.
func (s *serviceImpl) SyncSourcePluginsStrict(ctx context.Context) (*ListOutput, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	out, err := s.syncAndList(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.publishPluginChange(ctx, pluginChangePublishInput{reason: "source_plugins_synced"}); err != nil {
		return nil, err
	}
	return out, nil
}

// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
// returns the combined list of source and dynamic plugin items.
func (s *serviceImpl) SyncAndList(ctx context.Context) (*ListOutput, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	out, err := s.syncAndList(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.publishPluginChange(ctx, pluginChangePublishInput{reason: "plugins_synced_and_listed"}); err != nil {
		return nil, err
	}
	return out, nil
}

// syncAndList scans plugin manifests and mutates plugin governance tables for
// trusted startup or already-guarded platform management paths.
func (s *serviceImpl) syncAndList(ctx context.Context) (*ListOutput, error) {
	out, readCtx, err := s.buildPluginProjection(ctx, pluginProjectionInput{
		mode: projectionModeList,
		sync: true,
	})
	if err != nil {
		return nil, err
	}
	if err = s.integrationSvc.RefreshEnabledSnapshot(readCtx); err != nil {
		return nil, err
	}
	return out.list, nil
}

// List returns the paginated read-only plugin summary list with optional
// in-memory filtering applied to the lightweight summary read model.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.managementSummaryList(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]*PluginItem, 0, len(out.List))
	for _, item := range out.List {
		if item == nil {
			continue
		}
		if in.ID != "" && !strings.Contains(item.Id, in.ID) {
			continue
		}
		if in.Name != "" && !strings.Contains(item.Name, in.Name) {
			continue
		}
		if in.Type != "" && !matchesPluginType(item.Type, in.Type) {
			continue
		}
		if in.Status != nil && item.Enabled != *in.Status {
			continue
		}
		if in.Installed != nil && item.Installed != *in.Installed {
			continue
		}
		// Builtin plugins remain visible in ordinary management lists. Write
		// operations stay blocked by lifecycle governance. IncludeBuiltin is
		// retained only for request binding compatibility and is not applied.
		filtered = append(filtered, item)
	}
	page, total := paginatePluginItems(filtered, in.PageNum, in.PageSize)
	return &ListOutput{List: page, Total: total}, nil
}

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
	runtimeBundleRevision, err := s.i18nSvc.BundleRevision(ctx, locale)
	if err != nil {
		return management.ListCacheKey{}, err
	}
	return management.ListCacheKey{
		Locale:               locale,
		RuntimeBundleVersion: runtimeBundleRevision.Version,
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

const (
	// defaultListPageNum is used when callers omit the page number.
	defaultListPageNum = 1
	// defaultListPageSize is used when callers omit the page size.
	defaultListPageSize = 20
	// maxListPageSize bounds the plugin management summary list response size.
	maxListPageSize = 100
)

// normalizeListPage applies default and maximum pagination bounds.
func normalizeListPage(pageNum int, pageSize int) (int, int) {
	if pageNum < defaultListPageNum {
		pageNum = defaultListPageNum
	}
	if pageSize <= 0 {
		pageSize = defaultListPageSize
	}
	if pageSize > maxListPageSize {
		pageSize = maxListPageSize
	}
	return pageNum, pageSize
}

// paginatePluginItems returns the requested page and the total item count.
func paginatePluginItems(items []*PluginItem, pageNum int, pageSize int) ([]*PluginItem, int) {
	pageNum, pageSize = normalizeListPage(pageNum, pageSize)
	total := len(items)
	start := (pageNum - 1) * pageSize
	if start >= total {
		return []*PluginItem{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

// matchesPluginType compares normalized plugin types so list filtering accepts
// user input that differs only by case or alias formatting.
func matchesPluginType(actual string, expected string) bool {
	actualType := plugintypes.NormalizeType(actual)
	expectedType := plugintypes.NormalizeType(expected)
	if expectedType == "" {
		return true
	}
	return actualType == expectedType
}

// Get returns one read-only plugin detail projection by exact plugin ID.
func (s *serviceImpl) Get(ctx context.Context, pluginID string) (*PluginItem, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}
	item, err := s.buildManagementDetail(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
}

// ReadOnlyList scans plugin manifests and projects current registry state
// without synchronizing governance tables.
func (s *serviceImpl) ReadOnlyList(ctx context.Context) (*ListOutput, error) {
	return s.buildManagementList(ctx)
}

// buildManagementList scans plugin manifests and projects current registry
// state with complete governance detail, without synchronizing governance tables.
func (s *serviceImpl) buildManagementList(ctx context.Context) (*ListOutput, error) {
	out, _, err := s.buildPluginProjection(ctx, pluginProjectionInput{mode: projectionModeList})
	if err != nil {
		return nil, err
	}
	return out.list, nil
}

// buildManagementSummaryList scans plugin manifests and projects current
// registry state without detail-only dependency, host-service, route, or cron
// payloads.
func (s *serviceImpl) buildManagementSummaryList(ctx context.Context) (*ListOutput, error) {
	out, _, err := s.buildPluginProjection(ctx, pluginProjectionInput{mode: projectionModeSummary})
	if err != nil {
		return nil, err
	}
	return out.list, nil
}

// buildManagementDetail scans plugin manifests once and projects complete
// governance detail only for the requested plugin ID.
func (s *serviceImpl) buildManagementDetail(ctx context.Context, pluginID string) (*PluginItem, error) {
	out, _, err := s.buildPluginProjection(ctx, pluginProjectionInput{
		mode:     projectionModeDetail,
		pluginID: pluginID,
	})
	if err != nil {
		return nil, err
	}
	return out.item, nil
}

// buildServicePluginItems wraps runtime projections with facade-level metadata.
func (s *serviceImpl) buildServicePluginItems(ctx context.Context, items []*runtime.PluginItem) []*PluginItem {
	out := make([]*PluginItem, 0, len(items))
	for _, item := range items {
		if wrapped := s.buildServicePluginItem(ctx, item); wrapped != nil {
			out = append(out, wrapped)
		}
	}
	return out
}

// buildServicePluginSummaryItems wraps runtime summary projections without
// attaching dependency checks or detail-only governance payloads.
func (s *serviceImpl) buildServicePluginSummaryItems(ctx context.Context, items []*runtime.PluginItem) []*PluginItem {
	out := make([]*PluginItem, 0, len(items))
	for _, item := range items {
		if wrapped := s.buildServicePluginSummaryItem(ctx, item); wrapped != nil {
			out = append(out, wrapped)
		}
	}
	return out
}

// buildServicePluginItem wraps one runtime projection and attaches dependency status.
func (s *serviceImpl) buildServicePluginItem(ctx context.Context, item *runtime.PluginItem) *PluginItem {
	if item == nil {
		return nil
	}
	out := &PluginItem{PluginItem: *item}
	if dependencyCheck, err := s.CheckPluginDependencies(ctx, item.Id); err == nil {
		out.DependencyCheck = dependencyCheck
	}
	return out
}

// buildServicePluginSummaryItem wraps one runtime summary projection without
// computing dependency status for list rendering.
func (s *serviceImpl) buildServicePluginSummaryItem(_ context.Context, item *runtime.PluginItem) *PluginItem {
	if item == nil {
		return nil
	}
	return &PluginItem{PluginItem: *item}
}

// ListEnabledPluginIDs returns the IDs of plugins that are currently
// installed and enabled.
func (s *serviceImpl) ListEnabledPluginIDs(ctx context.Context) ([]string, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	registries, err := s.storeSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}

	pluginIDs := make([]string, 0, len(registries))
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		if registry.Installed != statusflag.Installed.Int() || registry.Status != statusflag.EnabledValue.Int() {
			continue
		}
		pluginIDs = append(pluginIDs, strings.TrimSpace(registry.PluginId))
	}
	return pluginIDs, nil
}
