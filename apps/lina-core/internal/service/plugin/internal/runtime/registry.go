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
	"lina-core/pkg/pluginbridge"
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
	// StatusKey is the host config key used by the public shell.
	StatusKey string
	// UpdatedAt is the ISO timestamp of the last registry update.
	UpdatedAt string
	// AuthorizationRequired reports whether any resource-scoped host services need confirmation.
	AuthorizationRequired bool
	// AuthorizationStatus identifies whether host-service authorization is pending or already confirmed.
	AuthorizationStatus AuthorizationStatus
	// RequestedHostServices is the current requested host service snapshot.
	RequestedHostServices []*pluginbridge.HostServiceSpec
	// AuthorizedHostServices is the host-confirmed host service snapshot.
	AuthorizedHostServices []*pluginbridge.HostServiceSpec
}

// listRuntimeRegistries returns all dynamic-type plugin registry rows.
func (s *serviceImpl) listRuntimeRegistries(ctx context.Context) ([]*entity.SysPlugin, error) {
	var list []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{Type: catalog.TypeDynamic.String()}).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// buildPluginItem returns a PluginItem projection combining manifest and registry data.
func (s *serviceImpl) buildPluginItem(ctx context.Context, manifest *catalog.Manifest, registry *entity.SysPlugin) *PluginItem {
	if manifest == nil && registry == nil {
		return nil
	}

	var (
		id          string
		name        string
		version     string
		pluginType  string
		description string
		installed   int
		enabled     int
		installedAt string
		updatedAt   string
		release     *entity.SysPluginRelease
		snapshot    *catalog.ManifestSnapshot
		err         error
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

	var (
		requestedHostServices  = pluginbridge.NormalizeHostServiceSpecs(nil)
		authorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(nil)
		authorizationRequired  bool
		authorizationStatus    = AuthorizationStatusNotRequired
	)

	if snapshot != nil {
		requestedHostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.RequestedHostServices)
		authorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.AuthorizedHostServices)
		authorizationRequired = snapshot.HostServiceAuthRequired
		authorizationStatus = buildAuthorizationStatus(snapshot.HostServiceAuthRequired, snapshot.HostServiceAuthConfirmed)
	} else if manifest != nil {
		requestedHostServices = pluginbridge.NormalizeHostServiceSpecs(manifest.HostServices)
		authorizationRequired = catalog.HasResourceScopedHostServices(manifest.HostServices)
		if authorizationRequired {
			authorizationStatus = AuthorizationStatusPending
		} else {
			authorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(manifest.HostServices)
		}
	}

	return &PluginItem{
		Id:                     id,
		Name:                   name,
		Version:                version,
		Type:                   pluginType,
		Description:            description,
		Installed:              installed,
		InstalledAt:            installedAt,
		Enabled:                enabled,
		StatusKey:              s.catalogSvc.BuildPluginStatusKey(id),
		UpdatedAt:              updatedAt,
		AuthorizationRequired:  authorizationRequired,
		AuthorizationStatus:    authorizationStatus,
		RequestedHostServices:  requestedHostServices,
		AuthorizedHostServices: authorizedHostServices,
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

	s.invalidateFrontendBundle(ctx, registry.PluginId, "runtime_artifact_missing")

	updated, err := s.catalogSvc.GetRegistry(ctx, registry.PluginId)
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

func buildAuthorizationStatus(required bool, confirmed bool) AuthorizationStatus {
	if !required {
		return AuthorizationStatusNotRequired
	}
	if confirmed {
		return AuthorizationStatusConfirmed
	}
	return AuthorizationStatusPending
}
