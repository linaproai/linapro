// This file contains source-plugin upgrade status discovery and version-drift
// projection helpers.

package sourceupgrade

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
)

// sourceUpgradeCandidate keeps the discovered manifest, current registry row,
// and flattened upgrade status together during one planning/execution cycle.
type sourceUpgradeCandidate struct {
	// manifest is the discovered source-plugin manifest.
	manifest *catalog.Manifest
	// registry is the synchronized registry row for the plugin.
	registry *entity.SysPlugin
	// status is the flattened upgrade status projection used by callers.
	status *SourceUpgradeStatus
}

// ListSourceUpgradeStatuses scans source manifests and returns one
// effective-versus-discovered upgrade-status item per source plugin.
func (s *serviceImpl) ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error) {
	candidates, err := s.listSourceUpgradeCandidates(ctx, false)
	if err != nil {
		return nil, err
	}

	items := make([]*SourceUpgradeStatus, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil || candidate.status == nil {
			continue
		}
		items = append(items, candidate.status)
	}
	return items, nil
}

// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
// without failing on pending upgrades.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	_, err := s.ListSourceUpgradeStatuses(ctx)
	return err
}

// listSourceUpgradeCandidates scans all discovered source manifests and returns
// their upgrade-governance view in stable plugin-ID order.
func (s *serviceImpl) listSourceUpgradeCandidates(
	ctx context.Context,
	synchronize bool,
) ([]*sourceUpgradeCandidate, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	registryByPluginID := map[string]*entity.SysPlugin{}
	if !synchronize {
		registries, registryErr := s.catalogSvc.ListAllRegistries(ctx)
		if registryErr != nil {
			return nil, registryErr
		}
		registryByPluginID = buildRegistryByPluginID(registries)
	}

	items := make([]*sourceUpgradeCandidate, 0)
	for _, manifest := range manifests {
		if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeSource {
			continue
		}

		var registry *entity.SysPlugin
		if synchronize {
			registry, err = s.catalogSvc.SyncManifest(ctx, manifest)
			if err != nil {
				return nil, err
			}
		} else {
			registry = registryByPluginID[strings.TrimSpace(manifest.ID)]
		}
		status, err := buildSourceUpgradeStatus(manifest, registry)
		if err != nil {
			return nil, err
		}
		items = append(items, &sourceUpgradeCandidate{
			manifest: manifest,
			registry: registry,
			status:   status,
		})
	}
	return items, nil
}

// buildRegistryByPluginID maps registry rows by stable plugin ID for
// source-upgrade planning so startup validation does not read one row per plugin.
func buildRegistryByPluginID(registries []*entity.SysPlugin) map[string]*entity.SysPlugin {
	result := make(map[string]*entity.SysPlugin, len(registries))
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		result[strings.TrimSpace(registry.PluginId)] = registry
	}
	return result
}

// findSourceUpgradeCandidate returns the synchronized upgrade candidate for the
// requested source plugin identifier.
func (s *serviceImpl) findSourceUpgradeCandidate(ctx context.Context, pluginID string) (*sourceUpgradeCandidate, error) {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return nil, bizerr.NewCode(CodePluginSourceUpgradePluginIDRequired)
	}

	candidates, err := s.listSourceUpgradeCandidates(ctx, false)
	if err != nil {
		return nil, err
	}
	for _, candidate := range candidates {
		if candidate == nil || candidate.status == nil {
			continue
		}
		if candidate.status.PluginID == normalizedID {
			return candidate, nil
		}
	}
	return nil, bizerr.NewCode(CodePluginSourceUpgradePluginNotFound, bizerr.P("pluginId", normalizedID))
}

// buildSourceUpgradeStatus flattens the manifest and registry state of one
// source plugin into an operator-facing upgrade-status projection.
func buildSourceUpgradeStatus(
	manifest *catalog.Manifest,
	registry *entity.SysPlugin,
) (*SourceUpgradeStatus, error) {
	if manifest == nil {
		return nil, bizerr.NewCode(CodePluginSourceUpgradeManifestRequired)
	}

	status := &SourceUpgradeStatus{
		PluginID:          strings.TrimSpace(manifest.ID),
		Name:              strings.TrimSpace(manifest.Name),
		DiscoveredVersion: strings.TrimSpace(manifest.Version),
	}
	if registry != nil {
		if strings.TrimSpace(registry.PluginId) != "" {
			status.PluginID = strings.TrimSpace(registry.PluginId)
		}
		if strings.TrimSpace(registry.Name) != "" {
			status.Name = strings.TrimSpace(registry.Name)
		}
		status.EffectiveVersion = strings.TrimSpace(registry.Version)
		status.Installed = registry.Installed
		status.Enabled = registry.Status
	}
	if status.Installed != catalog.InstalledYes {
		return status, nil
	}

	versionCompare, err := compareSourceUpgradeVersions(status.EffectiveVersion, status.DiscoveredVersion)
	if err != nil {
		return nil, err
	}
	status.NeedsUpgrade = versionCompare < 0
	status.DowngradeDetected = versionCompare > 0
	return status, nil
}

// compareSourceUpgradeVersions compares the current effective version and the
// currently discovered source version for one source plugin.
func compareSourceUpgradeVersions(effectiveVersion string, discoveredVersion string) (int, error) {
	effective := strings.TrimSpace(effectiveVersion)
	discovered := strings.TrimSpace(discoveredVersion)
	if effective == "" || discovered == "" {
		return 0, nil
	}
	return catalog.CompareSemanticVersions(effective, discovered)
}
