// This file synchronizes release-level plugin metadata snapshots into the
// governance tables used by the host management and review workflows.

package catalog

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginbridge"
)

// LoadReleaseManifest loads the dynamic plugin manifest from a persisted release artifact.
// The package path stored in the release row is resolved to an absolute host path before parsing.
func (s *serviceImpl) LoadReleaseManifest(ctx context.Context, release *entity.SysPluginRelease) (*Manifest, error) {
	if release == nil {
		return nil, gerror.New("插件 release 不能为空")
	}
	packagePath := strings.TrimSpace(release.PackagePath)
	if packagePath == "" {
		return nil, gerror.Newf("插件 release 缺少 package_path: %s@%s", release.PluginId, release.ReleaseVersion)
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
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId:       pluginID,
			ReleaseVersion: version,
		}).
		Scan(&release)
	return release, err
}

// GetReleaseByID returns the sys_plugin_release row with the given primary key.
func (s *serviceImpl) GetReleaseByID(ctx context.Context, releaseID int) (*entity.SysPluginRelease, error) {
	if releaseID <= 0 {
		return nil, nil
	}
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: releaseID}).
		Scan(&release)
	return release, err
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
		_, err = dao.SysPluginRelease.Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: existing.Id}).
		Data(data).
		Update()
	return err
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
	snapshot := s.buildManifestSnapshotModel(manifest)
	if snapshot == nil {
		return "", gerror.New("plugin manifest cannot be nil")
	}
	if existing != nil {
		existingSnapshot, err := s.ParseManifestSnapshot(existing.ManifestSnapshot)
		if err != nil {
			return "", err
		}
		if existingSnapshot != nil {
			snapshot.AuthorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(existingSnapshot.AuthorizedHostServices)
			snapshot.HostServiceAuthConfirmed = existingSnapshot.HostServiceAuthConfirmed
		}
	}
	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", gerror.Wrap(err, "failed to build plugin manifest snapshot")
	}
	return string(content), nil
}

func (s *serviceImpl) buildManifestSnapshotModel(manifest *Manifest) *ManifestSnapshot {
	if manifest == nil {
		return nil
	}

	snapshot := &ManifestSnapshot{
		ID:                        manifest.ID,
		Name:                      manifest.Name,
		Version:                   manifest.Version,
		Type:                      manifest.Type,
		Description:               manifest.Description,
		Author:                    manifest.Author,
		Homepage:                  manifest.Homepage,
		License:                   manifest.License,
		RuntimeKind:               buildDynamicKind(manifest),
		RuntimeABIVersion:         buildDynamicABIVersion(manifest),
		ManifestDeclared:          s.isManifestDeclared(manifest),
		InstallSQLCount:           s.countSQLAssets(manifest, MigrationDirectionInstall),
		UninstallSQLCount:         s.countSQLAssets(manifest, MigrationDirectionUninstall),
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
		RequestedHostServices:     pluginbridge.NormalizeHostServiceSpecs(manifest.HostServices),
		HostServiceAuthRequired:   HasResourceScopedHostServices(manifest.HostServices),
	}
	if !snapshot.HostServiceAuthRequired {
		snapshot.AuthorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.RequestedHostServices)
	}
	return snapshot
}

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
	manifest.HostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.AuthorizedHostServices)
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
		if direction == MigrationDirectionInstall {
			return len(manifest.RuntimeArtifact.InstallSQLAssets)
		}
		return len(manifest.RuntimeArtifact.UninstallSQLAssets)
	}
	if direction == MigrationDirectionInstall {
		return len(s.ListInstallSQLPaths(manifest))
	}
	return len(s.ListUninstallSQLPaths(manifest))
}

func (s *serviceImpl) buildFrontendPageCount(manifest *Manifest) int {
	if manifest == nil || NormalizeType(manifest.Type) != TypeSource {
		return 0
	}
	return len(s.ListFrontendPagePaths(manifest))
}

func (s *serviceImpl) buildFrontendSlotCount(manifest *Manifest) int {
	if manifest == nil || NormalizeType(manifest.Type) != TypeSource {
		return 0
	}
	return len(s.ListFrontendSlotPaths(manifest))
}

func buildDynamicKind(manifest *Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.RuntimeKind
}

func buildDynamicABIVersion(manifest *Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.ABIVersion
}

func buildDynamicFrontendAssetCount(manifest *Manifest) int {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return 0
	}
	return manifest.RuntimeArtifact.FrontendAssetCount
}

func buildDynamicSQLAssetCount(manifest *Manifest) int {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return 0
	}
	return manifest.RuntimeArtifact.SQLAssetCount
}

func buildDynamicRouteExecutionEnabled(manifest *Manifest) bool {
	if manifest == nil || manifest.BridgeSpec == nil {
		return false
	}
	return manifest.BridgeSpec.RouteExecution
}

func buildDynamicRouteRequestCodec(manifest *Manifest) string {
	if manifest == nil || manifest.BridgeSpec == nil {
		return ""
	}
	return manifest.BridgeSpec.RequestCodec
}

func buildDynamicRouteResponseCodec(manifest *Manifest) string {
	if manifest == nil || manifest.BridgeSpec == nil {
		return ""
	}
	return manifest.BridgeSpec.ResponseCodec
}
