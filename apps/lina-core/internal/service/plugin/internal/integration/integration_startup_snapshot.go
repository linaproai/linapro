// This file provides request-scoped startup snapshots for small plugin
// integration tables so startup reconciliation can avoid per-plugin lookups.

package integration

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/startupstats"
)

// startupDataSnapshotContextKey stores integration startup snapshots in context.
type startupDataSnapshotContextKey struct{}

// startupDataSnapshot contains full-table snapshots for plugin-owned menu and
// governance resource rows used during startup reconciliation.
type startupDataSnapshot struct {
	menusByKey              map[string]*entity.SysMenu
	resourceRefsByID        map[int]*entity.SysPluginResourceRef
	resourceRefsByPluginRel map[string][]*entity.SysPluginResourceRef
	startupSnapshotLoaded   bool
}

// WithStartupDataSnapshot returns a child context containing full-table
// snapshots for plugin menu and resource-reference rows.
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
	startupstats.Add(ctx, startupstats.CounterIntegrationSnapshotBuilds, 1)
	return context.WithValue(ctx, startupDataSnapshotContextKey{}, snapshot), nil
}

// buildStartupDataSnapshot loads startup integration rows in bulk.
func (s *serviceImpl) buildStartupDataSnapshot(ctx context.Context) (*startupDataSnapshot, error) {
	var menus []*entity.SysMenu
	if err := dao.SysMenu.Ctx(ctx).
		Unscoped().
		OrderAsc(dao.SysMenu.Columns().Id).
		Scan(&menus); err != nil {
		return nil, err
	}

	var refs []*entity.SysPluginResourceRef
	if err := dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Scan(&refs); err != nil {
		return nil, err
	}

	snapshot := &startupDataSnapshot{
		menusByKey:              make(map[string]*entity.SysMenu, len(menus)),
		resourceRefsByID:        make(map[int]*entity.SysPluginResourceRef, len(refs)),
		resourceRefsByPluginRel: make(map[string][]*entity.SysPluginResourceRef),
		startupSnapshotLoaded:   true,
	}
	for _, menu := range menus {
		if menu == nil || strings.TrimSpace(menu.MenuKey) == "" {
			continue
		}
		snapshot.menusByKey[strings.TrimSpace(menu.MenuKey)] = menu
	}
	for _, ref := range refs {
		snapshot.storeResourceRef(ref)
	}
	return snapshot, nil
}

// startupDataSnapshotFromContext returns the integration startup snapshot
// stored on the context, if present.
func startupDataSnapshotFromContext(ctx context.Context) *startupDataSnapshot {
	if ctx == nil {
		return nil
	}
	snapshot, ok := ctx.Value(startupDataSnapshotContextKey{}).(*startupDataSnapshot)
	if !ok || snapshot == nil || !snapshot.startupSnapshotLoaded {
		return nil
	}
	return snapshot
}

// pluginMenus returns plugin-owned menu rows from the startup snapshot.
func (s *startupDataSnapshot) pluginMenus(pluginID string) []*entity.SysMenu {
	if s == nil {
		return nil
	}
	prefix := catalog.MenuKeyPrefix + strings.TrimSpace(pluginID) + ":"
	items := make([]*entity.SysMenu, 0)
	for key, menu := range s.menusByKey {
		if menu == nil || !strings.HasPrefix(key, prefix) {
			continue
		}
		items = append(items, menu)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Id < items[j].Id
	})
	return items
}

// menusByKeys resolves menu rows by key from the startup snapshot.
func (s *startupDataSnapshot) menusByKeys(menuKeys []string, unscoped bool) map[string]*entity.SysMenu {
	result := make(map[string]*entity.SysMenu, len(menuKeys))
	if s == nil {
		return result
	}
	for _, key := range menuKeys {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			continue
		}
		menu := s.menusByKey[normalizedKey]
		if menu == nil {
			continue
		}
		if !unscoped && menu.DeletedAt != nil {
			continue
		}
		result[normalizedKey] = menu
	}
	return result
}

// storeMenu records a menu row in the startup snapshot.
func (s *startupDataSnapshot) storeMenu(menu *entity.SysMenu) {
	if s == nil || menu == nil || strings.TrimSpace(menu.MenuKey) == "" {
		return
	}
	s.menusByKey[strings.TrimSpace(menu.MenuKey)] = menu
}

// deleteMenus removes menu keys from the startup snapshot.
func (s *startupDataSnapshot) deleteMenus(menuKeys []string) {
	if s == nil {
		return
	}
	for _, key := range menuKeys {
		delete(s.menusByKey, strings.TrimSpace(key))
	}
}

// resourceRefs returns release-scoped resource refs from the startup snapshot.
func (s *startupDataSnapshot) resourceRefs(pluginID string, releaseID int) []*entity.SysPluginResourceRef {
	if s == nil {
		return nil
	}
	items := append([]*entity.SysPluginResourceRef(nil), s.resourceRefsByPluginRel[resourceRefKey(pluginID, releaseID)]...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Id < items[j].Id
	})
	return items
}

// storeResourceRef records a resource-reference row in the startup snapshot.
func (s *startupDataSnapshot) storeResourceRef(ref *entity.SysPluginResourceRef) {
	if s == nil || ref == nil {
		return
	}
	if existing := s.resourceRefsByID[ref.Id]; existing != nil {
		oldKey := resourceRefKey(existing.PluginId, existing.ReleaseId)
		s.resourceRefsByPluginRel[oldKey] = removeResourceRefByID(s.resourceRefsByPluginRel[oldKey], ref.Id)
	}
	s.resourceRefsByID[ref.Id] = ref
	key := resourceRefKey(ref.PluginId, ref.ReleaseId)
	s.resourceRefsByPluginRel[key] = append(s.resourceRefsByPluginRel[key], ref)
}

// deleteResourceRef removes one resource-reference row from the startup snapshot.
func (s *startupDataSnapshot) deleteResourceRef(refID int) {
	if s == nil || refID == 0 {
		return
	}
	existing := s.resourceRefsByID[refID]
	if existing == nil {
		return
	}
	delete(s.resourceRefsByID, refID)
	key := resourceRefKey(existing.PluginId, existing.ReleaseId)
	s.resourceRefsByPluginRel[key] = removeResourceRefByID(s.resourceRefsByPluginRel[key], refID)
}

// resourceRefKey builds the release-scoped resource-reference snapshot key.
func resourceRefKey(pluginID string, releaseID int) string {
	return strings.TrimSpace(pluginID) + "\x00" + strconv.Itoa(releaseID)
}

// removeResourceRefByID removes one row from a resource-reference slice.
func removeResourceRefByID(items []*entity.SysPluginResourceRef, refID int) []*entity.SysPluginResourceRef {
	result := items[:0]
	for _, item := range items {
		if item == nil || item.Id == refID {
			continue
		}
		result = append(result, item)
	}
	return result
}
