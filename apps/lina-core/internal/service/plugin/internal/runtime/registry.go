// This file provides registry-level helpers used by the reconciler and dynamic
// state projections: listing runtime registries, checking artifact file existence,
// and reconciling registry rows when artifacts are missing from storage.

package runtime

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// PluginItem is a flattened, display-ready projection of one plugin entry combining
// manifest fields with the live registry row for management API responses.
type PluginItem struct {
	// Id is the stable plugin identifier.
	Id string
	// Name is the human-readable display name.
	Name string
	// Version is the currently active version string.
	Version string
	// Type is the normalized plugin type (source or dynamic).
	Type string
	// Description is the short plugin description.
	Description string
	// Installed reports whether the plugin has been installed.
	Installed int
	// InstalledAt is the ISO timestamp of first installation.
	InstalledAt string
	// Enabled reports whether the plugin is currently enabled.
	Enabled int
	// AutoEnableForNewTenants reports the platform-owned new-tenant provisioning policy.
	AutoEnableForNewTenants bool
	// ScopeNature is the plugin scope nature persisted in the host registry.
	ScopeNature string
	// InstallMode is the plugin install mode persisted in the host registry.
	InstallMode string
	// StatusKey is the host config key used by the public shell.
	StatusKey string
	// UpdatedAt is the ISO timestamp of the last registry update.
	UpdatedAt string
	// AuthorizationRequired reports whether any resource-scoped host services need confirmation.
	AuthorizationRequired bool
	// AuthorizationStatus identifies whether host-service authorization is pending or already confirmed.
	AuthorizationStatus AuthorizationStatus
	// RequestedHostServices is the current requested host service snapshot.
	RequestedHostServices []*bridgehostservice.HostServiceSpec
	// AuthorizedHostServices is the host-confirmed host service snapshot.
	AuthorizedHostServices []*bridgehostservice.HostServiceSpec
	// DeclaredRoutes is the current release route-declaration snapshot used by
	// install and enable review UIs for dynamic plugins.
	DeclaredRoutes []*bridgecontract.RouteContract
	// HasMockData reports whether the plugin ships any mock-data SQL files
	// under manifest/sql/mock-data/. Used by the management UI to decide
	// whether to render the optional Install mock data checkbox.
	HasMockData bool
}

// listRuntimeRegistries returns all dynamic-type plugin registry rows.
func (s *serviceImpl) listRuntimeRegistries(ctx context.Context) ([]*entity.SysPlugin, error) {
	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*entity.SysPlugin, 0)
	for _, registry := range registries {
		if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
			continue
		}
		list = append(list, registry)
	}
	return list, nil
}

// buildPluginItem returns a PluginItem projection combining manifest and registry data.
func (s *serviceImpl) buildPluginItem(ctx context.Context, manifest *catalog.Manifest, registry *entity.SysPlugin) *PluginItem {
	if manifest == nil && registry == nil {
		return nil
	}

	var (
		id                      string
		name                    string
		version                 string
		pluginType              string
		description             string
		installed               int
		enabled                 int
		installedAt             string
		updatedAt               string
		scopeNature             string
		installMode             string
		autoEnableForNewTenants bool
		release                 *entity.SysPluginRelease
		snapshot                *catalog.ManifestSnapshot
		err                     error
	)

	if manifest != nil {
		id = manifest.ID
		name = manifest.Name
		version = manifest.Version
		pluginType = manifest.Type
		description = manifest.Description
	}
	if registry != nil {
		if registry.PluginId != "" {
			id = registry.PluginId
		}
		if registry.Name != "" {
			name = registry.Name
		}
		if registry.Version != "" {
			version = registry.Version
		}
		if registry.Type != "" {
			pluginType = registry.Type
		}
		if registry.Remark != "" {
			description = registry.Remark
		}
		installed = registry.Installed
		enabled = registry.Status
		scopeNature = registry.ScopeNature
		installMode = registry.InstallMode
		autoEnableForNewTenants = registry.AutoEnableForNewTenants
		if registry.InstalledAt != nil {
			installedAt = registry.InstalledAt.String()
		}
		if registry.UpdatedAt != nil {
			updatedAt = registry.UpdatedAt.String()
		}
		if ctx != nil {
			release, err = s.catalogSvc.GetRegistryRelease(ctx, registry)
			if err != nil {
				logger.Warningf(ctx, "load registry release failed plugin=%s err=%v", registry.PluginId, err)
			}
		}
	}
	if release == nil && manifest != nil && ctx != nil {
		release, err = s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			logger.Warningf(ctx, "load plugin release failed plugin=%s version=%s err=%v", manifest.ID, manifest.Version, err)
		}
	}
	if release != nil {
		snapshot, err = s.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
		if err != nil {
			logger.Warningf(ctx, "parse plugin release manifest snapshot failed plugin=%s releaseID=%d err=%v", id, release.Id, err)
		}
	}

	normalizeHostServices := func(source string, specs []*bridgehostservice.HostServiceSpec) []*bridgehostservice.HostServiceSpec {
		normalized, normalizeErr := bridgehostservice.NormalizeHostServiceSpecs(specs)
		if normalizeErr != nil {
			logger.Warningf(ctx, "normalize plugin host services failed plugin=%s source=%s err=%v", id, source, normalizeErr)
			return []*bridgehostservice.HostServiceSpec{}
		}
		return normalized
	}

	var (
		requestedHostServices  = []*bridgehostservice.HostServiceSpec{}
		authorizedHostServices = []*bridgehostservice.HostServiceSpec{}
		authorizationRequired  bool
		authorizationStatus    = AuthorizationStatusNotRequired
		declaredRoutes         []*bridgecontract.RouteContract
	)

	if snapshot != nil {
		requestedHostServices = normalizeHostServices("snapshot.requested", snapshot.RequestedHostServices)
		authorizedHostServices = normalizeHostServices("snapshot.authorized", snapshot.AuthorizedHostServices)
		authorizationRequired = snapshot.HostServiceAuthRequired
		authorizationStatus = buildAuthorizationStatus(snapshot.HostServiceAuthRequired, snapshot.HostServiceAuthConfirmed)
	} else if manifest != nil {
		requestedHostServices = normalizeHostServices("manifest.requested", manifest.HostServices)
		authorizationRequired = catalog.HasResourceScopedHostServices(manifest.HostServices)
		if authorizationRequired {
			authorizationStatus = AuthorizationStatusPending
		} else {
			authorizedHostServices = normalizeHostServices("manifest.authorized", manifest.HostServices)
		}
	}
	if manifest != nil {
		declaredRoutes = cloneRouteContracts(manifest.Routes)
	}
	name, description = s.localizePluginMetadata(ctx, id, pluginType, name, description)

	hasMockData := false
	if manifest != nil {
		hasMockData = s.catalogSvc.HasMockSQLData(manifest)
	} else if snapshot != nil {
		hasMockData = snapshot.MockSQLCount > 0
	}

	return &PluginItem{
		Id:                      id,
		Name:                    name,
		Version:                 version,
		Type:                    pluginType,
		Description:             description,
		Installed:               installed,
		InstalledAt:             installedAt,
		Enabled:                 enabled,
		AutoEnableForNewTenants: autoEnableForNewTenants,
		ScopeNature:             scopeNature,
		InstallMode:             installMode,
		StatusKey:               s.catalogSvc.BuildPluginStatusKey(id),
		UpdatedAt:               updatedAt,
		AuthorizationRequired:   authorizationRequired,
		AuthorizationStatus:     authorizationStatus,
		RequestedHostServices:   requestedHostServices,
		AuthorizedHostServices:  authorizedHostServices,
		DeclaredRoutes:          declaredRoutes,
		HasMockData:             hasMockData,
	}
}

// hasArtifactStorageFile reports whether the runtime artifact for pluginID exists
// in the configured storage directory.
func (s *serviceImpl) hasArtifactStorageFile(ctx context.Context, pluginID string) (bool, string, error) {
	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return false, "", err
	}

	targetPath := filepath.Join(storageDir, buildArtifactFileName(pluginID))
	if gfile.Exists(targetPath) {
		return true, targetPath, nil
	}

	conflictPath, err := s.findDuplicateArtifactPath(storageDir, pluginID, targetPath)
	if err != nil {
		return false, "", err
	}
	if conflictPath != "" {
		return true, conflictPath, nil
	}
	return false, targetPath, nil
}

// HasArtifactStorageFile is the exported form of hasArtifactStorageFile for cross-package access.
func (s *serviceImpl) HasArtifactStorageFile(ctx context.Context, pluginID string) (bool, string, error) {
	return s.hasArtifactStorageFile(ctx, pluginID)
}

// reconcileRegistryArtifactState resets a dynamic plugin registry row to
// uninstalled when its runtime artifact file can no longer be found on disk.
func (s *serviceImpl) reconcileRegistryArtifactState(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPlugin, error) {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return registry, nil
	}
	if strings.TrimSpace(registry.PluginId) == "" {
		return registry, nil
	}

	exists, _, err := s.hasArtifactStorageFile(ctx, registry.PluginId)
	if err != nil {
		return nil, err
	}
	if exists {
		return registry, nil
	}
	if registry.Installed != catalog.InstalledYes && registry.Status != catalog.StatusEnabled {
		return registry, nil
	}

	data := do.SysPlugin{
		Installed:    catalog.InstalledNo,
		Status:       catalog.StatusDisabled,
		DesiredState: catalog.HostStateUninstalled.String(),
		CurrentState: catalog.HostStateUninstalled.String(),
		ReleaseId:    0,
		Generation:   catalog.NextGeneration(registry),
		DisabledAt:   gtime.Now(),
	}
	if _, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update(); err != nil {
		return nil, err
	}

	s.invalidateRuntimeCaches(ctx, registry.PluginId, "runtime_artifact_missing")
	if err = s.notifyRuntimeCacheChanged(ctx, "runtime_artifact_missing"); err != nil {
		return nil, err
	}
	if err = s.notifyReconcilerChanged(ctx, "runtime_artifact_missing"); err != nil {
		return nil, err
	}

	updated, err := s.catalogSvc.RefreshStartupRegistry(ctx, registry.PluginId)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	if err = s.SyncPluginReleaseRuntimeState(ctx, updated); err != nil {
		return nil, err
	}
	if err = s.SyncPluginNodeState(
		ctx,
		updated.PluginId,
		updated.Version,
		updated.Installed,
		updated.Status,
		"Runtime plugin artifact missing from storage path; host registry reconciled to uninstalled.",
	); err != nil {
		return nil, err
	}
	return updated, nil
}

// projectRegistryArtifactState returns a read-only projection of a dynamic
// registry row when its runtime artifact is missing from storage.
func (s *serviceImpl) projectRegistryArtifactState(ctx context.Context, registry *entity.SysPlugin) *entity.SysPlugin {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return registry
	}
	if strings.TrimSpace(registry.PluginId) == "" {
		return registry
	}

	exists, _, err := s.hasArtifactStorageFile(ctx, registry.PluginId)
	if err != nil || exists {
		return registry
	}
	if registry.Installed != catalog.InstalledYes && registry.Status != catalog.StatusEnabled {
		return registry
	}

	projected := *registry
	projected.Installed = catalog.InstalledNo
	projected.Status = catalog.StatusDisabled
	projected.DesiredState = catalog.HostStateUninstalled.String()
	projected.CurrentState = catalog.HostStateUninstalled.String()
	projected.ReleaseId = 0
	return &projected
}

// SortPluginItems sorts a PluginItem slice by plugin ID ascending.
func SortPluginItems(items []*PluginItem) {
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

// buildAuthorizationStatus maps manifest snapshot flags into one list-facing
// authorization review state.
func buildAuthorizationStatus(required bool, confirmed bool) AuthorizationStatus {
	if !required {
		return AuthorizationStatusNotRequired
	}
	if confirmed {
		return AuthorizationStatusConfirmed
	}
	return AuthorizationStatusPending
}
