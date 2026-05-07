// This file provides request-scoped startup snapshots for small plugin
// governance tables so startup reconciliation can avoid per-plugin lookups.

package catalog

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/startupstats"
)

// startupDataSnapshotContextKey stores catalog startup snapshots in context.
type startupDataSnapshotContextKey struct{}

// startupDataSnapshot contains full-table snapshots for small plugin catalog
// tables used repeatedly during startup reconciliation.
type startupDataSnapshot struct {
	registriesByPluginID     map[string]*entity.SysPlugin
	releasesByPluginVersion  map[string]*entity.SysPluginRelease
	releasesByID             map[int]*entity.SysPluginRelease
	startupSnapshotAvailable bool
}

// WithStartupDataSnapshot returns a child context containing full-table
// snapshots for plugin registry and release rows.
func (s *serviceImpl) WithStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if startupDataSnapshotFromContext(ctx) != nil {
		return ctx, nil
	}

	snapshot, err := s.buildStartupDataSnapshot(ctx)
	if err != nil {
		return ctx, err
	}
	startupstats.Add(ctx, startupstats.CounterCatalogSnapshotBuilds, 1)
	return context.WithValue(ctx, startupDataSnapshotContextKey{}, snapshot), nil
}

// buildStartupDataSnapshot loads startup catalog rows in bulk.
func (s *serviceImpl) buildStartupDataSnapshot(ctx context.Context) (*startupDataSnapshot, error) {
	registries, err := s.listAllRegistriesFromDB(ctx)
	if err != nil {
		return nil, err
	}

	var releases []*entity.SysPluginRelease
	if err = dao.SysPluginRelease.Ctx(ctx).Scan(&releases); err != nil {
		return nil, err
	}

	snapshot := &startupDataSnapshot{
		registriesByPluginID:     make(map[string]*entity.SysPlugin, len(registries)),
		releasesByPluginVersion:  make(map[string]*entity.SysPluginRelease, len(releases)),
		releasesByID:             make(map[int]*entity.SysPluginRelease, len(releases)),
		startupSnapshotAvailable: true,
	}
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		snapshot.registriesByPluginID[strings.TrimSpace(registry.PluginId)] = registry
	}
	for _, release := range releases {
		if release == nil {
			continue
		}
		snapshot.releasesByID[release.Id] = release
		snapshot.releasesByPluginVersion[releaseKey(release.PluginId, release.ReleaseVersion)] = release
	}
	return snapshot, nil
}

// startupDataSnapshotFromContext returns the catalog startup snapshot stored
// on the context, if the caller is running inside a startup reconciliation pass.
func startupDataSnapshotFromContext(ctx context.Context) *startupDataSnapshot {
	if ctx == nil {
		return nil
	}
	snapshot, ok := ctx.Value(startupDataSnapshotContextKey{}).(*startupDataSnapshot)
	if !ok || snapshot == nil || !snapshot.startupSnapshotAvailable {
		return nil
	}
	return snapshot
}

// registry returns one registry row from the startup snapshot.
func (s *startupDataSnapshot) registry(pluginID string) *entity.SysPlugin {
	if s == nil {
		return nil
	}
	return s.registriesByPluginID[strings.TrimSpace(pluginID)]
}

// storeRegistry records a refreshed registry row in the startup snapshot.
func (s *startupDataSnapshot) storeRegistry(registry *entity.SysPlugin) {
	if s == nil || registry == nil || strings.TrimSpace(registry.PluginId) == "" {
		return
	}
	s.registriesByPluginID[strings.TrimSpace(registry.PluginId)] = registry
}

// releaseByPluginVersion returns one release row from the startup snapshot.
func (s *startupDataSnapshot) releaseByPluginVersion(pluginID string, version string) *entity.SysPluginRelease {
	if s == nil {
		return nil
	}
	return s.releasesByPluginVersion[releaseKey(pluginID, version)]
}

// releaseByID returns one release row by primary key from the startup snapshot.
func (s *startupDataSnapshot) releaseByID(releaseID int) *entity.SysPluginRelease {
	if s == nil {
		return nil
	}
	return s.releasesByID[releaseID]
}

// storeRelease records a refreshed release row in both startup snapshot indexes.
func (s *startupDataSnapshot) storeRelease(release *entity.SysPluginRelease) {
	if s == nil || release == nil {
		return
	}
	s.releasesByID[release.Id] = release
	s.releasesByPluginVersion[releaseKey(release.PluginId, release.ReleaseVersion)] = release
}

// updateRegistry applies a partial registry update to the startup snapshot and
// returns the updated row when the snapshot carries the target registry.
func updateStartupRegistry(ctx context.Context, pluginID string, data do.SysPlugin) *entity.SysPlugin {
	snapshot := startupDataSnapshotFromContext(ctx)
	if snapshot == nil {
		return nil
	}
	existing := snapshot.registry(pluginID)
	if existing == nil {
		return nil
	}
	updated := clonePluginRegistry(existing)
	applyPluginRegistryData(updated, data)
	snapshot.storeRegistry(updated)
	return updated
}

// insertStartupRegistry stores a newly inserted registry projection when the
// startup snapshot is present.
func insertStartupRegistry(ctx context.Context, registryID int, data do.SysPlugin) *entity.SysPlugin {
	snapshot := startupDataSnapshotFromContext(ctx)
	if snapshot == nil {
		return nil
	}
	registry := buildPluginRegistryEntity(registryID, data)
	snapshot.storeRegistry(registry)
	return registry
}

// updateStartupRelease applies a partial release update to the startup snapshot
// and returns the updated row when the snapshot carries the target release.
func updateStartupRelease(ctx context.Context, existing *entity.SysPluginRelease, data do.SysPluginRelease) *entity.SysPluginRelease {
	snapshot := startupDataSnapshotFromContext(ctx)
	if snapshot == nil || existing == nil {
		return nil
	}
	updated := clonePluginRelease(existing)
	applyPluginReleaseData(updated, data)
	snapshot.storeRelease(updated)
	return updated
}

// insertStartupRelease stores a newly inserted release projection when the
// startup snapshot is present.
func insertStartupRelease(ctx context.Context, releaseID int, data do.SysPluginRelease) *entity.SysPluginRelease {
	snapshot := startupDataSnapshotFromContext(ctx)
	if snapshot == nil {
		return nil
	}
	release := buildPluginReleaseEntity(releaseID, data)
	snapshot.storeRelease(release)
	return release
}

// buildPluginRegistryEntity creates a startup projection from registry DO data.
func buildPluginRegistryEntity(registryID int, data do.SysPlugin) *entity.SysPlugin {
	registry := &entity.SysPlugin{Id: registryID}
	applyPluginRegistryData(registry, data)
	return registry
}

// applyPluginRegistryData overlays non-nil DO fields on one registry entity.
func applyPluginRegistryData(registry *entity.SysPlugin, data do.SysPlugin) {
	if registry == nil {
		return
	}
	if data.PluginId != nil {
		registry.PluginId = strings.TrimSpace(gconv.String(data.PluginId))
	}
	if data.Name != nil {
		registry.Name = gconv.String(data.Name)
	}
	if data.Version != nil {
		registry.Version = gconv.String(data.Version)
	}
	if data.Type != nil {
		registry.Type = gconv.String(data.Type)
	}
	if data.Installed != nil {
		registry.Installed = gconv.Int(data.Installed)
	}
	if data.Status != nil {
		registry.Status = gconv.Int(data.Status)
	}
	if data.DesiredState != nil {
		registry.DesiredState = gconv.String(data.DesiredState)
	}
	if data.CurrentState != nil {
		registry.CurrentState = gconv.String(data.CurrentState)
	}
	if data.Generation != nil {
		registry.Generation = gconv.Int64(data.Generation)
	}
	if data.ReleaseId != nil {
		registry.ReleaseId = gconv.Int(data.ReleaseId)
	}
	if data.ManifestPath != nil {
		registry.ManifestPath = gconv.String(data.ManifestPath)
	}
	if data.Checksum != nil {
		registry.Checksum = gconv.String(data.Checksum)
	}
	if data.InstalledAt != nil {
		registry.InstalledAt = data.InstalledAt
	}
	if data.EnabledAt != nil {
		registry.EnabledAt = data.EnabledAt
	}
	if data.DisabledAt != nil {
		registry.DisabledAt = data.DisabledAt
	}
	if data.Remark != nil {
		registry.Remark = gconv.String(data.Remark)
	}
}

// clonePluginRegistry copies one registry entity before snapshot mutation.
func clonePluginRegistry(existing *entity.SysPlugin) *entity.SysPlugin {
	if existing == nil {
		return nil
	}
	clone := *existing
	return &clone
}

// buildPluginReleaseEntity creates a startup projection from release DO data.
func buildPluginReleaseEntity(releaseID int, data do.SysPluginRelease) *entity.SysPluginRelease {
	release := &entity.SysPluginRelease{Id: releaseID}
	applyPluginReleaseData(release, data)
	return release
}

// applyPluginReleaseData overlays non-nil DO fields on one release entity.
func applyPluginReleaseData(release *entity.SysPluginRelease, data do.SysPluginRelease) {
	if release == nil {
		return
	}
	if data.PluginId != nil {
		release.PluginId = strings.TrimSpace(gconv.String(data.PluginId))
	}
	if data.ReleaseVersion != nil {
		release.ReleaseVersion = gconv.String(data.ReleaseVersion)
	}
	if data.Type != nil {
		release.Type = gconv.String(data.Type)
	}
	if data.RuntimeKind != nil {
		release.RuntimeKind = gconv.String(data.RuntimeKind)
	}
	if data.SchemaVersion != nil {
		release.SchemaVersion = gconv.String(data.SchemaVersion)
	}
	if data.MinHostVersion != nil {
		release.MinHostVersion = gconv.String(data.MinHostVersion)
	}
	if data.MaxHostVersion != nil {
		release.MaxHostVersion = gconv.String(data.MaxHostVersion)
	}
	if data.Status != nil {
		release.Status = gconv.String(data.Status)
	}
	if data.ManifestPath != nil {
		release.ManifestPath = gconv.String(data.ManifestPath)
	}
	if data.PackagePath != nil {
		release.PackagePath = gconv.String(data.PackagePath)
	}
	if data.Checksum != nil {
		release.Checksum = gconv.String(data.Checksum)
	}
	if data.ManifestSnapshot != nil {
		release.ManifestSnapshot = gconv.String(data.ManifestSnapshot)
	}
}

// clonePluginRelease copies one release entity before snapshot mutation.
func clonePluginRelease(existing *entity.SysPluginRelease) *entity.SysPluginRelease {
	if existing == nil {
		return nil
	}
	clone := *existing
	return &clone
}

// listRegistries returns all startup registry rows ordered by plugin_id.
func (s *startupDataSnapshot) listRegistries() []*entity.SysPlugin {
	if s == nil {
		return nil
	}
	items := make([]*entity.SysPlugin, 0, len(s.registriesByPluginID))
	for _, registry := range s.registriesByPluginID {
		if registry == nil {
			continue
		}
		items = append(items, registry)
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.TrimSpace(items[i].PluginId) < strings.TrimSpace(items[j].PluginId)
	})
	return items
}

// releaseKey builds the composite release identity used by the startup snapshot.
func releaseKey(pluginID string, version string) string {
	return strings.TrimSpace(pluginID) + "\x00" + strings.TrimSpace(version)
}

// refreshStartupRegistry reloads one registry row after a startup write and
// updates the request-scoped snapshot when present.
func (s *serviceImpl) refreshStartupRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	registry, err := s.getRegistryFromDB(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		snapshot.storeRegistry(registry)
	}
	return registry, nil
}

// RefreshStartupRegistry reloads one registry row from the database and
// refreshes the startup snapshot when present.
func (s *serviceImpl) RefreshStartupRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	return s.refreshStartupRegistry(ctx, pluginID)
}

// refreshStartupRelease reloads one release row after a startup write and
// updates the request-scoped snapshot when present.
func (s *serviceImpl) refreshStartupRelease(
	ctx context.Context,
	pluginID string,
	version string,
) (*entity.SysPluginRelease, error) {
	release, err := s.getReleaseFromDB(ctx, pluginID, version)
	if err != nil {
		return nil, err
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		snapshot.storeRelease(release)
	}
	return release, nil
}

// RefreshStartupReleaseByID reloads one release row from the database and
// refreshes the startup snapshot when present.
func (s *serviceImpl) RefreshStartupReleaseByID(ctx context.Context, releaseID int) (*entity.SysPluginRelease, error) {
	release, err := s.getReleaseByIDFromDB(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		snapshot.storeRelease(release)
	}
	return release, nil
}

// getRegistryFromDB reads one registry row without consulting the startup snapshot.
func (s *serviceImpl) getRegistryFromDB(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return nil, nil
	}

	var plugin *entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: normalizedID}).
		Scan(&plugin)
	return plugin, err
}

// getReleaseFromDB reads one release row without consulting the startup snapshot.
func (s *serviceImpl) getReleaseFromDB(ctx context.Context, pluginID string, version string) (*entity.SysPluginRelease, error) {
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId:       pluginID,
			ReleaseVersion: version,
		}).
		Scan(&release)
	return release, err
}

// getReleaseByIDFromDB reads one release row by primary key without consulting
// the startup snapshot.
func (s *serviceImpl) getReleaseByIDFromDB(ctx context.Context, releaseID int) (*entity.SysPluginRelease, error) {
	if releaseID <= 0 {
		return nil, nil
	}
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: releaseID}).
		Scan(&release)
	return release, err
}

// listAllRegistriesFromDB reads all registry rows without consulting the
// startup snapshot.
func (s *serviceImpl) listAllRegistriesFromDB(ctx context.Context) ([]*entity.SysPlugin, error) {
	var list []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
