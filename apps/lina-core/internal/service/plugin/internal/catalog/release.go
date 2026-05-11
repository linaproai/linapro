// This file synchronizes release-level plugin metadata snapshots into the
// governance tables used by the host management and review workflows.

package catalog

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	"gopkg.in/yaml.v3"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginbridge"
)

// LoadReleaseManifest loads the dynamic plugin manifest from a persisted release artifact.
// The package path stored in the release row is resolved to an absolute host path before parsing.
func (s *serviceImpl) LoadReleaseManifest(ctx context.Context, release *entity.SysPluginRelease) (*Manifest, error) {
	if release == nil {
		return nil, gerror.New("plugin release cannot be nil")
	}
	packagePath := strings.TrimSpace(release.PackagePath)
	if packagePath == "" {
		return nil, gerror.Newf("plugin release is missing package_path: %s@%s", release.PluginId, release.ReleaseVersion)
	}
	absolutePath, err := s.resolveReleasePackagePath(ctx, packagePath)
	if err != nil {
		return nil, err
	}
	manifest, err := s.loadRuntimeManifestFromArtifact(absolutePath)
	if err != nil {
		return nil, err
	}
	if err = s.applyReleaseAuthorizedHostServices(manifest, release); err != nil {
		return nil, err
	}
	return manifest, nil
}

// resolveReleasePackagePath converts a release package_path (possibly relative) to an
// absolute host path. Relative paths are anchored at the runtime storage directory.
func (s *serviceImpl) resolveReleasePackagePath(ctx context.Context, packagePath string) (string, error) {
	if filepath.IsAbs(packagePath) {
		return filepath.Clean(packagePath), nil
	}
	storageDir, err := s.resolveRuntimeStorageDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(storageDir, filepath.FromSlash(packagePath))), nil
}

// GetRelease returns the sys_plugin_release row for a plugin ID + version pair.
func (s *serviceImpl) GetRelease(ctx context.Context, pluginID string, version string) (*entity.SysPluginRelease, error) {
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.releaseByPluginVersion(pluginID, version), nil
	}
	return s.getReleaseFromDB(ctx, pluginID, version)
}

// GetReleaseByID returns the sys_plugin_release row with the given primary key.
func (s *serviceImpl) GetReleaseByID(ctx context.Context, releaseID int) (*entity.SysPluginRelease, error) {
	if releaseID <= 0 {
		return nil, nil
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.releaseByID(releaseID), nil
	}
	return s.getReleaseByIDFromDB(ctx, releaseID)
}

// GetRegistryRelease returns the active release row for a registry entry, preferring
// the ReleaseId pointer and falling back to a version lookup.
func (s *serviceImpl) GetRegistryRelease(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPluginRelease, error) {
	if registry == nil {
		return nil, nil
	}
	if registry.ReleaseId > 0 {
		release, err := s.GetReleaseByID(ctx, registry.ReleaseId)
		if err != nil {
			return nil, err
		}
		if release != nil {
			return release, nil
		}
	}
	if strings.TrimSpace(registry.Version) == "" {
		return nil, nil
	}
	return s.GetRelease(ctx, registry.PluginId, registry.Version)
}

// GetActiveRelease returns the currently active release row for one plugin.
func (s *serviceImpl) GetActiveRelease(ctx context.Context, pluginID string) (*entity.SysPluginRelease, error) {
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId: pluginID,
			Status:   ReleaseStatusActive.String(),
		}).
		OrderDesc(dao.SysPluginRelease.Columns().Id).
		Scan(&release)
	if isCatalogNoRows(err) {
		return nil, nil
	}
	return release, err
}

// UpdateReleaseState transitions a release row to the given status and optionally
// updates its package path.
func (s *serviceImpl) UpdateReleaseState(ctx context.Context, releaseID int, status ReleaseStatus, packagePath string) error {
	if releaseID <= 0 {
		return nil
	}

	data := do.SysPluginRelease{
		Status: status.String(),
	}
	if strings.TrimSpace(packagePath) != "" {
		data.PackagePath = filepath.ToSlash(strings.TrimSpace(packagePath))
	}

	_, err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: releaseID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	_, err = s.RefreshStartupReleaseByID(ctx, releaseID)
	return err
}

// syncReleaseMetadata upserts the manifest snapshot into sys_plugin_release.
func (s *serviceImpl) syncReleaseMetadata(ctx context.Context, manifest *Manifest, registry *entity.SysPlugin) error {
	if manifest == nil || registry == nil {
		return nil
	}

	existing, err := s.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	snapshot, err := s.buildManifestSnapshot(manifest, existing)
	if err != nil {
		return err
	}

	releaseID := 0
	if existing != nil {
		releaseID = existing.Id
	}
	releaseStatus := s.buildReleaseStatusForManifest(manifest, registry, releaseID)
	data := do.SysPluginRelease{
		PluginId:         manifest.ID,
		ReleaseVersion:   manifest.Version,
		Type:             manifest.Type,
		RuntimeKind:      buildDynamicKind(manifest),
		Status:           releaseStatus.String(),
		ManifestPath:     s.buildReleaseManifestPath(manifest),
		PackagePath:      s.buildReleasePackagePathForSync(manifest, existing),
		Checksum:         s.BuildRegistryChecksum(manifest),
		ManifestSnapshot: snapshot,
	}

	if existing == nil {
		insertID, insertErr := dao.SysPluginRelease.Ctx(ctx).Data(data).InsertAndGetId()
		err = insertErr
		if err != nil {
			if !dialect.IsUniqueConstraintViolation(err) {
				return err
			}
			existing, err = s.refreshStartupRelease(ctx, manifest.ID, manifest.Version)
			if err != nil {
				return err
			}
			if existing == nil {
				return insertErr
			}
			if pluginReleaseMetadataMatches(existing, data) {
				startupstats.Add(ctx, startupstats.CounterPluginSyncNoop, 1)
				return nil
			}
			_, err = dao.SysPluginRelease.Ctx(ctx).
				Where(do.SysPluginRelease{Id: existing.Id}).
				Data(data).
				Update()
			if err != nil {
				return err
			}
			startupstats.Add(ctx, startupstats.CounterPluginSyncChanged, 1)
			if updateStartupRelease(ctx, existing, data) != nil {
				return nil
			}
			_, err = s.refreshStartupRelease(ctx, manifest.ID, manifest.Version)
			return err
		}
		startupstats.Add(ctx, startupstats.CounterPluginSyncChanged, 1)
		if insertStartupRelease(ctx, int(insertID), data) != nil {
			return nil
		}
		_, err = s.refreshStartupRelease(ctx, manifest.ID, manifest.Version)
		return err
	}
	if pluginReleaseMetadataMatches(existing, data) {
		startupstats.Add(ctx, startupstats.CounterPluginSyncNoop, 1)
		return nil
	}
	_, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: existing.Id}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	startupstats.Add(ctx, startupstats.CounterPluginSyncChanged, 1)
	if updateStartupRelease(ctx, existing, data) != nil {
		return nil
	}
	_, err = s.refreshStartupRelease(ctx, manifest.ID, manifest.Version)
	return err
}

// pluginReleaseMetadataMatches reports whether a release row already matches
// the manifest metadata projection produced during startup reconciliation.
func pluginReleaseMetadataMatches(existing *entity.SysPluginRelease, data do.SysPluginRelease) bool {
	if existing == nil {
		return false
	}
	return existing.PluginId == dataString(data.PluginId) &&
		existing.ReleaseVersion == dataString(data.ReleaseVersion) &&
		existing.Type == dataString(data.Type) &&
		existing.RuntimeKind == dataString(data.RuntimeKind) &&
		existing.Status == dataString(data.Status) &&
		existing.ManifestPath == dataString(data.ManifestPath) &&
		existing.PackagePath == dataString(data.PackagePath) &&
		existing.Checksum == dataString(data.Checksum) &&
		existing.ManifestSnapshot == dataString(data.ManifestSnapshot)
}

// dataString normalizes a DO field into its persisted string value.
func dataString(value any) string {
	return gconv.String(value)
}

// SyncReleaseMetadata is the exported form of syncReleaseMetadata for runtime callers.
func (s *serviceImpl) SyncReleaseMetadata(ctx context.Context, manifest *Manifest, registry *entity.SysPlugin) error {
	return s.syncReleaseMetadata(ctx, manifest, registry)
}

// BuildManifestSnapshot is the exported form of buildManifestSnapshot for cross-package access.
func (s *serviceImpl) BuildManifestSnapshot(manifest *Manifest) (string, error) {
	return s.buildManifestSnapshot(manifest, nil)
}

// buildManifestSnapshot marshals the review-oriented manifest fields into a YAML string.
func (s *serviceImpl) buildManifestSnapshot(manifest *Manifest, existing *entity.SysPluginRelease) (string, error) {
	snapshot, err := s.buildManifestSnapshotModel(manifest)
	if err != nil {
		return "", err
	}
	if snapshot == nil {
		return "", gerror.New("plugin manifest cannot be nil")
	}
	if existing != nil {
		existingSnapshot, parseErr := s.ParseManifestSnapshot(existing.ManifestSnapshot)
		if parseErr != nil {
			return "", parseErr
		}
		if existingSnapshot != nil {
			authorizedHostServices, normalizeErr := pluginbridge.NormalizeHostServiceSpecs(existingSnapshot.AuthorizedHostServices)
			if normalizeErr != nil {
				return "", normalizeErr
			}
			snapshot.AuthorizedHostServices = authorizedHostServices
			snapshot.HostServiceAuthConfirmed = existingSnapshot.HostServiceAuthConfirmed
			snapshot.UninstallPurgeStorageData = existingSnapshot.UninstallPurgeStorageData
		}
	}
	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", gerror.Wrap(err, "failed to build plugin manifest snapshot")
	}
	return string(content), nil
}

// buildManifestSnapshotModel converts one manifest into the review-oriented
// release snapshot model persisted in sys_plugin_release.
func (s *serviceImpl) buildManifestSnapshotModel(manifest *Manifest) (*ManifestSnapshot, error) {
	if manifest == nil {
		return nil, nil
	}

	requestedHostServices, err := pluginbridge.NormalizeHostServiceSpecs(manifest.HostServices)
	if err != nil {
		return nil, err
	}

	snapshot := &ManifestSnapshot{
		ID:                        manifest.ID,
		Name:                      manifest.Name,
		Version:                   manifest.Version,
		Type:                      manifest.Type,
		ScopeNature:               manifest.ScopeNature,
		SupportsMultiTenant:       manifest.SupportsTenantGovernance(),
		DefaultInstallMode:        manifest.DefaultInstallMode,
		Description:               manifest.Description,
		Author:                    manifest.Author,
		Homepage:                  manifest.Homepage,
		License:                   manifest.License,
		RuntimeKind:               buildDynamicKind(manifest),
		RuntimeABIVersion:         buildDynamicABIVersion(manifest),
		ManifestDeclared:          s.isManifestDeclared(manifest),
		InstallSQLCount:           s.countSQLAssets(manifest, MigrationDirectionInstall),
		UninstallSQLCount:         s.countSQLAssets(manifest, MigrationDirectionUninstall),
		MockSQLCount:              s.countSQLAssets(manifest, MigrationDirectionMock),
		FrontendPageCount:         s.buildFrontendPageCount(manifest),
		FrontendSlotCount:         s.buildFrontendSlotCount(manifest),
		MenuCount:                 len(manifest.Menus),
		BackendHookCount:          len(manifest.Hooks),
		ResourceSpecCount:         len(manifest.BackendResources),
		RouteCount:                len(manifest.Routes),
		RouteExecutionEnabled:     buildDynamicRouteExecutionEnabled(manifest),
		RouteRequestCodec:         buildDynamicRouteRequestCodec(manifest),
		RouteResponseCodec:        buildDynamicRouteResponseCodec(manifest),
		RuntimeFrontendAssetCount: buildDynamicFrontendAssetCount(manifest),
		RuntimeSQLAssetCount:      buildDynamicSQLAssetCount(manifest),
		RequestedHostServices:     requestedHostServices,
		HostServiceAuthRequired:   HasResourceScopedHostServices(manifest.HostServices),
	}
	if !snapshot.HostServiceAuthRequired {
		authorizedHostServices, normalizeErr := pluginbridge.NormalizeHostServiceSpecs(snapshot.RequestedHostServices)
		if normalizeErr != nil {
			return nil, normalizeErr
		}
		snapshot.AuthorizedHostServices = authorizedHostServices
	}
	return snapshot, nil
}

// PersistReleaseUninstallPurgePolicy writes one host-confirmed uninstall cleanup
// policy snapshot into the given dynamic-plugin release row.
func (s *serviceImpl) PersistReleaseUninstallPurgePolicy(
	ctx context.Context,
	release *entity.SysPluginRelease,
	purgeStorageData bool,
) (*ManifestSnapshot, error) {
	if release == nil {
		return nil, gerror.New("plugin release cannot be nil")
	}

	snapshot, err := s.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		manifest, loadErr := s.LoadReleaseManifest(ctx, release)
		if loadErr != nil {
			return nil, loadErr
		}
		snapshot, err = s.buildManifestSnapshotModel(manifest)
		if err != nil {
			return nil, err
		}
	}
	if snapshot == nil {
		return nil, gerror.New("plugin release manifest snapshot cannot be nil")
	}

	purgeValue := purgeStorageData
	snapshot.UninstallPurgeStorageData = &purgeValue

	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return nil, gerror.Wrap(err, "build plugin uninstall policy snapshot failed")
	}
	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: release.Id}).
		Data(do.SysPluginRelease{ManifestSnapshot: string(content)}).
		Update(); err != nil {
		return nil, err
	}
	if _, err = s.RefreshStartupReleaseByID(ctx, release.Id); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// applyReleaseAuthorizedHostServices restores the host-confirmed service
// snapshot from the persisted release metadata onto the active manifest.
func (s *serviceImpl) applyReleaseAuthorizedHostServices(manifest *Manifest, release *entity.SysPluginRelease) error {
	if manifest == nil || release == nil {
		return nil
	}
	snapshot, err := s.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return err
	}
	if snapshot == nil {
		return nil
	}
	if !snapshot.HostServiceAuthRequired && len(snapshot.AuthorizedHostServices) == 0 {
		return nil
	}
	if !snapshot.HostServiceAuthConfirmed && snapshot.HostServiceAuthRequired {
		return nil
	}
	hostServices, err := pluginbridge.NormalizeHostServiceSpecs(snapshot.AuthorizedHostServices)
	if err != nil {
		return err
	}
	manifest.HostServices = hostServices
	manifest.HostCapabilities = pluginbridge.CapabilityMapFromHostServices(manifest.HostServices)
	return nil
}

// BuildPackagePath returns the canonical package path for a manifest used in release rows.
func (s *serviceImpl) BuildPackagePath(manifest *Manifest) string {
	if manifest == nil {
		return ""
	}
	if HasSourcePluginEmbeddedFiles(manifest) {
		return "embedded/source-plugins/" + manifest.ID
	}
	if manifest.RuntimeArtifact != nil && strings.TrimSpace(manifest.RuntimeArtifact.Path) != "" {
		normalizedPath := filepath.ToSlash(filepath.Clean(manifest.RuntimeArtifact.Path))
		if marker := "/releases/"; strings.Contains(normalizedPath, marker) {
			return strings.TrimPrefix(normalizedPath[strings.LastIndex(normalizedPath, marker):], "/")
		}
		return filepath.ToSlash(filepath.Base(normalizedPath))
	}
	return filepath.ToSlash(manifest.RootDir)
}

// buildReleasePackagePathForSync keeps archived dynamic-release package paths stable.
func (s *serviceImpl) buildReleasePackagePathForSync(manifest *Manifest, existing *entity.SysPluginRelease) string {
	if existing != nil {
		existingPath := filepath.ToSlash(strings.TrimSpace(existing.PackagePath))
		if shouldPreserveArchivedPackagePath(manifest, existingPath) {
			return existingPath
		}
	}
	return s.BuildPackagePath(manifest)
}

// shouldPreserveArchivedPackagePath returns true when a release's package path already
// points to an archived location and should not be overwritten by the mutable staging artifact.
func shouldPreserveArchivedPackagePath(manifest *Manifest, packagePath string) bool {
	if manifest == nil || NormalizeType(manifest.Type) != TypeDynamic {
		return false
	}
	normalizedPath := filepath.ToSlash(strings.TrimSpace(packagePath))
	if normalizedPath == "" {
		return false
	}
	normalizedPath = strings.TrimPrefix(filepath.Clean("/"+normalizedPath), "/")
	return strings.Contains("/"+normalizedPath, "/releases/")
}

// buildReleaseStatusForManifest determines the appropriate release status from
// registry state and whether the release row matches the active registry pointer.
func (s *serviceImpl) buildReleaseStatusForManifest(manifest *Manifest, registry *entity.SysPlugin, releaseID int) ReleaseStatus {
	if manifest == nil || registry == nil {
		return ReleaseStatusPrepared
	}
	if NormalizeType(manifest.Type) == TypeSource {
		if strings.TrimSpace(registry.Version) == strings.TrimSpace(manifest.Version) {
			return BuildReleaseStatus(registry.Installed, registry.Status)
		}
		if registry.Installed != InstalledYes {
			return BuildReleaseStatus(registry.Installed, registry.Status)
		}
		return ReleaseStatusPrepared
	}
	if NormalizeType(manifest.Type) != TypeDynamic {
		return BuildReleaseStatus(registry.Installed, registry.Status)
	}
	if registry.ReleaseId > 0 && releaseID == registry.ReleaseId {
		return BuildReleaseStatus(registry.Installed, registry.Status)
	}
	if strings.TrimSpace(registry.Version) == strings.TrimSpace(manifest.Version) && registry.ReleaseId <= 0 {
		return BuildReleaseStatus(registry.Installed, registry.Status)
	}
	return ReleaseStatusPrepared
}

// buildReleaseManifestPath returns the manifest path to store in the release row.
func (s *serviceImpl) buildReleaseManifestPath(manifest *Manifest) string {
	if manifest == nil || NormalizeType(manifest.Type) == TypeDynamic {
		return ""
	}
	if HasSourcePluginEmbeddedFiles(manifest) {
		return path.Clean(strings.ReplaceAll(manifest.ManifestPath, "\\", "/"))
	}
	return filepath.ToSlash(filepath.Base(manifest.ManifestPath))
}

// isManifestDeclared reports whether the manifest has a valid manifest path or embedded manifest.
func (s *serviceImpl) isManifestDeclared(manifest *Manifest) bool {
	if manifest == nil {
		return false
	}
	if strings.TrimSpace(manifest.ManifestPath) != "" {
		return true
	}
	return manifest.RuntimeArtifact != nil && manifest.RuntimeArtifact.Manifest != nil
}

// countSQLAssets counts SQL migration steps for the given direction from manifest metadata.
func (s *serviceImpl) countSQLAssets(manifest *Manifest, direction MigrationDirection) int {
	if manifest == nil {
		return 0
	}
	if manifest.RuntimeArtifact != nil {
		switch direction {
		case MigrationDirectionInstall:
			return len(manifest.RuntimeArtifact.InstallSQLAssets)
		case MigrationDirectionMock:
			return len(manifest.RuntimeArtifact.MockSQLAssets)
		default:
			return len(manifest.RuntimeArtifact.UninstallSQLAssets)
		}
	}
	switch direction {
	case MigrationDirectionInstall:
		return len(s.ListInstallSQLPaths(manifest))
	case MigrationDirectionMock:
		return len(s.ListMockSQLPaths(manifest))
	default:
		return len(s.ListUninstallSQLPaths(manifest))
	}
}

// buildFrontendPageCount counts discovered source-plugin frontend pages.
func (s *serviceImpl) buildFrontendPageCount(manifest *Manifest) int {
	if manifest == nil || NormalizeType(manifest.Type) != TypeSource {
		return 0
	}
	return len(s.ListFrontendPagePaths(manifest))
}

// buildFrontendSlotCount counts discovered source-plugin frontend slot entries.
func (s *serviceImpl) buildFrontendSlotCount(manifest *Manifest) int {
	if manifest == nil || NormalizeType(manifest.Type) != TypeSource {
		return 0
	}
	return len(s.ListFrontendSlotPaths(manifest))
}

// buildDynamicKind returns the runtime kind recorded in the embedded artifact.
func buildDynamicKind(manifest *Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.RuntimeKind
}

// buildDynamicABIVersion returns the ABI version recorded in the embedded artifact.
func buildDynamicABIVersion(manifest *Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.ABIVersion
}

// buildDynamicFrontendAssetCount returns the embedded frontend asset count.
func buildDynamicFrontendAssetCount(manifest *Manifest) int {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return 0
	}
	return manifest.RuntimeArtifact.FrontendAssetCount
}

// buildDynamicSQLAssetCount returns the embedded SQL asset count.
func buildDynamicSQLAssetCount(manifest *Manifest) int {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return 0
	}
	return manifest.RuntimeArtifact.SQLAssetCount
}

// buildDynamicRouteExecutionEnabled reports whether the runtime bridge allows route execution.
func buildDynamicRouteExecutionEnabled(manifest *Manifest) bool {
	if manifest == nil || manifest.BridgeSpec == nil {
		return false
	}
	return manifest.BridgeSpec.RouteExecution
}

// buildDynamicRouteRequestCodec returns the declared bridge request codec.
func buildDynamicRouteRequestCodec(manifest *Manifest) string {
	if manifest == nil || manifest.BridgeSpec == nil {
		return ""
	}
	return manifest.BridgeSpec.RequestCodec
}

// buildDynamicRouteResponseCodec returns the declared bridge response codec.
func buildDynamicRouteResponseCodec(manifest *Manifest) string {
	if manifest == nil || manifest.BridgeSpec == nil {
		return ""
	}
	return manifest.BridgeSpec.ResponseCodec
}
