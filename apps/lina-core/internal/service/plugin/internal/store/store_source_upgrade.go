// This file keeps source-plugin release promotion writes inside the plugin
// governance store instead of leaking registry DO updates to upgrade callers.

package store

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// PromoteSourceRelease switches one installed source plugin to the discovered
// release while preserving installed/enabled flags and stable lifecycle state.
func (s *serviceImpl) PromoteSourceRelease(
	ctx context.Context,
	registry *PluginRecord,
	manifest *catalog.Manifest,
	release *ReleaseRecord,
) (*PluginRecord, error) {
	if registry == nil || manifest == nil || release == nil {
		return registry, nil
	}

	stableState := plugintypes.DeriveHostState(registry.Installed, registry.Status)
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
		Checksum:     s.BuildRegistryChecksum(manifest),
		Remark:       manifest.Description,
	}
	if registry.Generation <= 0 {
		data.Generation = int64(1)
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	if err != nil {
		return nil, err
	}
	return s.RefreshStartupRegistry(ctx, registry.PluginId)
}
