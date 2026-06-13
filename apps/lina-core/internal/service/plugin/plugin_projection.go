// This file owns the root plugin projection builder used by list, summary,
// detail, and dependency snapshot read paths.

package plugin

import (
	"context"
	"sort"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/management"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/startupstats"
)

// projectionMode selects the field and side-effect boundary for one plugin
// projection build. Summary mode intentionally omits detail-only governance
// payloads and dependency checks.
type projectionMode string

const (
	projectionModeList               projectionMode = "list"
	projectionModeSummary            projectionMode = "summary"
	projectionModeDetail             projectionMode = "detail"
	projectionModeDependencySnapshot projectionMode = "dependency_snapshot"
)

// pluginProjectionInput defines the stable shape of plugin read-model builds.
type pluginProjectionInput struct {
	mode      projectionMode
	pluginID  string
	candidate *catalog.Manifest
	sync      bool
}

// pluginProjectionOutput carries the current read model plus the shared
// snapshots built while projecting it.
type pluginProjectionOutput struct {
	ctx                 context.Context
	list                *ListOutput
	item                *PluginItem
	dependencySnapshots []*plugindep.PluginSnapshot
}

// buildPluginProjection builds plugin management projections using one
// manifest scan, one registry snapshot, and one dependency snapshot context.
func (s *serviceImpl) buildPluginProjection(
	ctx context.Context,
	input pluginProjectionInput,
) (*pluginProjectionOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	manifests, err := s.projectionManifests(ctx, input.sync)
	if err != nil {
		return nil, err
	}

	readCtx, err := s.projectionSnapshotContext(ctx, manifests, input.sync)
	if err != nil {
		return nil, err
	}
	out := &pluginProjectionOutput{ctx: readCtx}
	if input.mode == projectionModeDependencySnapshot {
		out.dependencySnapshots, err = s.buildDependencySnapshotsForProjection(readCtx, input.candidate)
		return out, err
	}
	if input.mode == projectionModeDetail {
		item, err := s.buildDetailProjection(readCtx, manifests, strings.TrimSpace(input.pluginID))
		if err != nil {
			return nil, err
		}
		out.item = item
		return out, nil
	}

	items, err := s.buildListProjection(readCtx, manifests, input)
	if err != nil {
		return nil, err
	}
	sortPluginProjectionItems(items)
	out.list = &ListOutput{List: items, Total: len(items)}
	return out, nil
}

// projectionManifests returns the current manifest snapshot, scanning only when
// the caller has not already attached one to the projection context.
func (s *serviceImpl) projectionManifests(ctx context.Context, sync bool) ([]*catalog.Manifest, error) {
	if manifests := management.ManifestSnapshotFromContext(ctx); manifests != nil && !sync {
		return manifests, nil
	}
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	startupstats.Add(ctx, startupstats.CounterPluginScans, 1)
	startupstats.Add(ctx, startupstats.CounterPluginScanItems, len(manifests))
	return manifests, nil
}

// projectionSnapshotContext attaches store, manifest, integration, and
// dependency snapshots used by all plugin projection modes.
func (s *serviceImpl) projectionSnapshotContext(
	ctx context.Context,
	manifests []*catalog.Manifest,
	sync bool,
) (context.Context, error) {
	readCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return ctx, err
	}
	if sync {
		readCtx, err = s.integrationSvc.WithStartupDataSnapshot(readCtx)
		if err != nil {
			return ctx, err
		}
	}
	readCtx = management.WithManifestSnapshot(readCtx, manifests)
	readCtx = s.WithDependencySnapshotCache(readCtx)
	return readCtx, nil
}

// buildListProjection builds list or summary projections from the shared
// manifest and registry snapshots.
func (s *serviceImpl) buildListProjection(
	ctx context.Context,
	manifests []*catalog.Manifest,
	input pluginProjectionInput,
) ([]*PluginItem, error) {
	registries, err := s.syncOrListRegistries(ctx, manifests, input.sync)
	if err != nil {
		return nil, err
	}
	registryByPluginID := registryByPluginID(registries)
	covered := make(map[string]struct{}, len(manifests))
	items := make([]*PluginItem, 0, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		covered[manifest.ID] = struct{}{}
		if item := s.buildManifestProjectionItem(ctx, manifest, registryByPluginID[manifest.ID], input.mode); item != nil {
			items = append(items, item)
		}
	}

	runtimeItems, err := s.buildRuntimeProjectionItems(ctx, covered, input)
	if err != nil {
		return nil, err
	}
	return append(items, runtimeItems...), nil
}

// registryByPluginID indexes registry rows by plugin ID for read-only list projection.
func registryByPluginID(registries []*store.PluginRecord) map[string]*store.PluginRecord {
	result := make(map[string]*store.PluginRecord, len(registries))
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		result[registry.PluginId] = registry
	}
	return result
}

// sortPluginProjectionItems sorts facade plugin projections by plugin ID.
func sortPluginProjectionItems(items []*PluginItem) {
	sort.Slice(items, func(i int, j int) bool {
		if items[i] == nil {
			return false
		}
		if items[j] == nil {
			return true
		}
		return items[i].Id < items[j].Id
	})
}

// syncOrListRegistries either synchronizes discovered manifests or reads one
// batched registry snapshot for read-only projection.
func (s *serviceImpl) syncOrListRegistries(
	ctx context.Context,
	manifests []*catalog.Manifest,
	sync bool,
) ([]*store.PluginRecord, error) {
	if !sync {
		return s.storeSvc.ListAllRegistries(ctx)
	}
	registries := make([]*store.PluginRecord, 0, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		registry, err := s.storeSvc.SyncManifest(ctx, manifest)
		if err != nil {
			return nil, err
		}
		registries = append(registries, registry)
	}
	return registries, nil
}

// buildManifestProjectionItem wraps one manifest-backed runtime projection for
// the selected mode.
func (s *serviceImpl) buildManifestProjectionItem(
	ctx context.Context,
	manifest *catalog.Manifest,
	registry *store.PluginRecord,
	mode projectionMode,
) *PluginItem {
	if mode == projectionModeSummary {
		return s.buildServicePluginSummaryItem(ctx, s.runtimeSvc.BuildPluginSummaryItem(ctx, manifest, registry))
	}
	return s.buildServicePluginItem(ctx, s.runtimeSvc.BuildPluginItem(ctx, manifest, registry))
}

// buildRuntimeProjectionItems appends registry-only dynamic projections that
// were not covered by the manifest scan.
func (s *serviceImpl) buildRuntimeProjectionItems(
	ctx context.Context,
	covered map[string]struct{},
	input pluginProjectionInput,
) ([]*PluginItem, error) {
	var (
		runtimeItems []*runtime.PluginItem
		err          error
	)
	switch {
	case input.mode == projectionModeSummary:
		runtimeItems, err = s.runtimeSvc.BuildRuntimeSummaryItemsReadOnly(ctx, covered)
	case input.sync:
		runtimeItems, err = s.runtimeSvc.BuildRuntimeItems(ctx, covered)
	default:
		runtimeItems, err = s.runtimeSvc.BuildRuntimeItemsReadOnly(ctx, covered)
	}
	if err != nil {
		return nil, err
	}
	if input.mode == projectionModeSummary {
		return s.buildServicePluginSummaryItems(ctx, runtimeItems), nil
	}
	return s.buildServicePluginItems(ctx, runtimeItems), nil
}

// buildDetailProjection returns one detail projection for an exact plugin ID
// using the shared manifest snapshot and a single registry lookup.
func (s *serviceImpl) buildDetailProjection(
	ctx context.Context,
	manifests []*catalog.Manifest,
	pluginID string,
) (*PluginItem, error) {
	var manifest *catalog.Manifest
	for _, item := range manifests {
		if item != nil && strings.TrimSpace(item.ID) == pluginID {
			manifest = item
			break
		}
	}
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if manifest == nil && registry == nil {
		return nil, nil
	}
	return s.buildServicePluginItem(
		ctx,
		s.runtimeSvc.BuildPluginItemReadOnly(ctx, manifest, registry),
	), nil
}
