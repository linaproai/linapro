// This file contains source-plugin effective release promotion helpers.

package sourceupgrade

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
)

// applySourcePluginUpgradedRelease promotes the discovered release to the
// current effective registry version without changing installed/enabled flags.
func (s *serviceImpl) applySourcePluginUpgradedRelease(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
) error {
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceUpgradeRegistryRequired)
	}
	if manifest == nil {
		return bizerr.NewCode(CodePluginSourceUpgradeManifestRequired)
	}
	if release == nil {
		return bizerr.NewCode(CodePluginSourceUpgradeTargetReleaseRequired)
	}

	stableState := catalog.DeriveHostState(registry.Installed, registry.Status)
	data := do.SysPlugin{
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		ReleaseId:    release.Id,
		Installed:    registry.Installed,
		Status:       registry.Status,
		DesiredState: stableState,
		CurrentState: stableState,
		ManifestPath: manifest.ManifestPath,
		Checksum:     s.catalogSvc.BuildRegistryChecksum(manifest),
		Remark:       manifest.Description,
	}
	if registry.Generation <= 0 {
		data.Generation = int64(1)
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	return err
}
