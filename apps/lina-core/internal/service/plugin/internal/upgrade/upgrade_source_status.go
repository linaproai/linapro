// This file owns source-plugin upgrade status discovery and version-drift
// projection for the unified upgrade component.

package upgrade

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
)

// SourceUpgradeStatus describes one source plugin's effective version,
// discovered source version, and pending-upgrade state.
type SourceUpgradeStatus struct {
	// PluginID is the immutable plugin identifier.
	PluginID string
	// Name is the human-readable plugin display name.
	Name string
	// EffectiveVersion is the current effective version stored in sys_plugin.
	EffectiveVersion string
	// DiscoveredVersion is the version currently discovered from plugin.yaml.
	DiscoveredVersion string
	// Installed reports whether the plugin is already installed.
	Installed int
	// Enabled reports whether the plugin is currently enabled.
	Enabled int
	// NeedsUpgrade reports whether an installed plugin discovered a newer source version.
	NeedsUpgrade bool
	// DowngradeDetected reports whether the discovered source version is lower
	// than the current effective version, which is unsupported in this iteration.
	DowngradeDetected bool
}

// sourceUpgradeCandidate keeps the discovered manifest, current registry row,
// and flattened upgrade status together during one planning/execution cycle.
type sourceUpgradeCandidate struct {
	// manifest is the discovered source-plugin manifest.
	manifest *catalog.Manifest
	// registry is the synchronized registry row for the plugin.
	registry *store.PluginRecord
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
	registryByPluginID := map[string]*store.PluginRecord{}
	if !synchronize {
		registries, registryErr := s.storeSvc.ListAllRegistries(ctx)
		if registryErr != nil {
			return nil, registryErr
		}
		registryByPluginID = buildRegistryByPluginID(registries)
	}

	items := make([]*sourceUpgradeCandidate, 0)
	for _, manifest := range manifests {
		if manifest == nil || plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeSource {
			continue
		}

		var registry *store.PluginRecord
		if synchronize {
			registry, err = s.storeSvc.SyncManifest(ctx, manifest)
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
func buildRegistryByPluginID(registries []*store.PluginRecord) map[string]*store.PluginRecord {
	result := make(map[string]*store.PluginRecord, len(registries))
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
	registry *store.PluginRecord,
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
	if status.Installed != plugintypes.InstalledYes {
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
	return plugintypes.CompareSemanticVersions(effective, discovered)
}
