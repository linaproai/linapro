// This file provides request-scoped startup snapshots for small plugin
// governance tables so startup reconciliation can avoid per-plugin lookups.

package catalog

import (
	"context"
	"sort"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
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

	snapshot, err := s.buildStartupDataSnapshot(ctx)
	if err != nil {
		return ctx, err
	}
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
