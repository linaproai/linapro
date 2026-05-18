// This file builds side-effect-free plugin runtime upgrade previews.

package plugin

import (
	"context"
	"sort"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
)

const (
	// RuntimeUpgradeRiskHintUpgradeSQLRequiresReview warns that upgrade SQL
	// should be reviewed before the user confirms runtime side effects.
	RuntimeUpgradeRiskHintUpgradeSQLRequiresReview = "plugin.runtimeUpgrade.risk.upgradeSqlRequiresReview"
	// RuntimeUpgradeRiskHintMockSQLExcluded warns that mock SQL is never loaded by upgrade.
	RuntimeUpgradeRiskHintMockSQLExcluded = "plugin.runtimeUpgrade.risk.mockSqlExcluded"
	// RuntimeUpgradeRiskHintHostServiceAuthorizationChanged warns that hostServices changed.
	RuntimeUpgradeRiskHintHostServiceAuthorizationChanged = "plugin.runtimeUpgrade.risk.hostServiceAuthorizationChanged"
	// RuntimeUpgradeRiskHintDependencyBlockers warns that dependency checks found hard blockers.
	RuntimeUpgradeRiskHintDependencyBlockers = "plugin.runtimeUpgrade.risk.dependencyBlockers"
)

// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one
// plugin currently marked as pending or failed runtime upgrade.
func (s *serviceImpl) PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*RuntimeUpgradePreview, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}

	targetManifest, err := s.loadDesiredManifestForPreview(normalizedPluginID)
	if err != nil {
		return nil, err
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	if registry == nil {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}

	projection, err := s.catalogSvc.BuildRuntimeUpgradeState(ctx, registry, targetManifest)
	if err != nil {
		return nil, err
	}
	if !runtimeUpgradeCanExecute(projection.State) {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradePreviewUnavailable,
			bizerr.P("pluginId", normalizedPluginID),
			bizerr.P("runtimeState", projection.State.String()),
		)
	}

	fromSnapshot, err := s.loadEffectiveManifestSnapshot(ctx, registry)
	if err != nil {
		return nil, err
	}
	toSnapshot, err := s.buildTargetManifestSnapshot(targetManifest)
	if err != nil {
		return nil, err
	}
	dependencyCheck, err := s.buildRuntimeUpgradeDependencyCheck(ctx, targetManifest)
	if err != nil {
		return nil, err
	}
	hostServicesDiff, err := buildRuntimeUpgradeHostServicesDiff(fromSnapshot, toSnapshot)
	if err != nil {
		return nil, err
	}
	sqlSummary := buildRuntimeUpgradeSQLSummary(toSnapshot)

	return &RuntimeUpgradePreview{
		PluginID:          normalizedPluginID,
		RuntimeState:      RuntimeUpgradeState(projection.State),
		EffectiveVersion:  projection.EffectiveVersion,
		DiscoveredVersion: projection.DiscoveredVersion,
		FromManifest:      fromSnapshot,
		ToManifest:        toSnapshot,
		DependencyCheck:   dependencyCheck,
		SQLSummary:        sqlSummary,
		HostServicesDiff:  hostServicesDiff,
		RiskHints:         buildRuntimeUpgradeRiskHints(sqlSummary, hostServicesDiff, dependencyCheck),
	}, nil
}

// loadDesiredManifestForPreview finds the currently discovered manifest while
// preserving catalog scan errors as diagnostics instead of converting every
// failure into not-found.
func (s *serviceImpl) loadDesiredManifestForPreview(pluginID string) (*catalog.Manifest, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		if manifest != nil && strings.TrimSpace(manifest.ID) == pluginID {
			return manifest, nil
		}
	}
	return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", pluginID))
}

// buildRuntimeUpgradeDependencyCheck evaluates target dependencies and
// downstream compatibility against the target upgrade version without side effects.
func (s *serviceImpl) buildRuntimeUpgradeDependencyCheck(
	ctx context.Context,
	targetManifest *catalog.Manifest,
) (*DependencyCheckResult, error) {
	if targetManifest == nil {
		return &DependencyCheckResult{}, nil
	}
	installResult, err := s.resolveInstallDependenciesForManifest(ctx, targetManifest)
	if err != nil {
		return nil, err
	}
	reverseResult, err := s.resolveReverseDependencies(ctx, targetManifest.ID, targetManifest.Version)
	if err != nil {
		return nil, err
	}
	result := toDependencyCheckResult(installResult)
	result.ReverseDependents = toDependencyReverseDependents(reverseResult.Dependents)
	result.ReverseBlockers = toDependencyBlockers(reverseResult.Blockers)
	return result, nil
}

// loadEffectiveManifestSnapshot reads the manifest snapshot tied to the
// database-effective registry release.
func (s *serviceImpl) loadEffectiveManifestSnapshot(
	ctx context.Context,
	registry *entity.SysPlugin,
) (*catalog.ManifestSnapshot, error) {
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release == nil {
		pluginID := ""
		version := ""
		if registry != nil {
			pluginID = registry.PluginId
			version = registry.Version
		}
		return nil, bizerr.NewCode(
			CodePluginReleaseNotFound,
			bizerr.P("pluginId", pluginID),
			bizerr.P("version", version),
		)
	}
	snapshot, err := s.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", registry.PluginId),
			bizerr.P("version", registry.Version),
		)
	}
	if snapshot == nil {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeSnapshotMissing,
			bizerr.P("pluginId", registry.PluginId),
			bizerr.P("version", registry.Version),
		)
	}
	return snapshot, nil
}

// buildTargetManifestSnapshot converts the discovered target manifest into the
// same review snapshot model persisted for releases.
func (s *serviceImpl) buildTargetManifestSnapshot(
	manifest *catalog.Manifest,
) (*catalog.ManifestSnapshot, error) {
	snapshotYAML, err := s.catalogSvc.BuildManifestSnapshot(manifest)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", manifestPluginID(manifest)),
			bizerr.P("version", manifestVersion(manifest)),
		)
	}
	snapshot, err := s.catalogSvc.ParseManifestSnapshot(snapshotYAML)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", manifestPluginID(manifest)),
			bizerr.P("version", manifestVersion(manifest)),
		)
	}
	if snapshot == nil {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeSnapshotMissing,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("version", manifest.Version),
		)
	}
	return snapshot, nil
}

// manifestPluginID safely extracts the plugin ID for diagnostics.
func manifestPluginID(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	return manifest.ID
}

// manifestVersion safely extracts the plugin version for diagnostics.
func manifestVersion(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	return manifest.Version
}

// buildRuntimeUpgradeSQLSummary derives a count-only SQL preview from the target snapshot.
func buildRuntimeUpgradeSQLSummary(snapshot *catalog.ManifestSnapshot) RuntimeUpgradeSQLSummary {
	if snapshot == nil {
		return RuntimeUpgradeSQLSummary{}
	}
	return RuntimeUpgradeSQLSummary{
		InstallSQLCount:      snapshot.InstallSQLCount,
		UninstallSQLCount:    snapshot.UninstallSQLCount,
		MockSQLCount:         snapshot.MockSQLCount,
		RuntimeSQLAssetCount: snapshot.RuntimeSQLAssetCount,
	}
}

// buildRuntimeUpgradeHostServicesDiff compares effective and target requested
// hostServices at service level.
func buildRuntimeUpgradeHostServicesDiff(
	fromSnapshot *catalog.ManifestSnapshot,
	toSnapshot *catalog.ManifestSnapshot,
) (RuntimeUpgradeHostServicesDiff, error) {
	fromServices, err := normalizeRuntimeUpgradeHostServices(fromSnapshot, snapshotRequestedHostServices(fromSnapshot))
	if err != nil {
		return RuntimeUpgradeHostServicesDiff{}, err
	}
	toServices, err := normalizeRuntimeUpgradeHostServices(toSnapshot, snapshotRequestedHostServices(toSnapshot))
	if err != nil {
		return RuntimeUpgradeHostServicesDiff{}, err
	}
	diff := RuntimeUpgradeHostServicesDiff{}
	diff.AuthorizationRequired = toSnapshot != nil && toSnapshot.HostServiceAuthRequired

	serviceSet := make(map[string]struct{}, len(fromServices)+len(toServices))
	for service := range fromServices {
		serviceSet[service] = struct{}{}
	}
	for service := range toServices {
		serviceSet[service] = struct{}{}
	}
	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}
	sort.Strings(services)

	for _, service := range services {
		fromSpec := fromServices[service]
		toSpec := toServices[service]
		change := buildRuntimeUpgradeHostServiceChange(service, fromSpec, toSpec)
		switch {
		case fromSpec == nil && toSpec != nil:
			diff.Added = append(diff.Added, change)
		case fromSpec != nil && toSpec == nil:
			diff.Removed = append(diff.Removed, change)
		case runtimeUpgradeHostServiceChanged(fromSpec, toSpec):
			diff.Changed = append(diff.Changed, change)
		}
	}
	diff.AuthorizationChanged = len(diff.Added) > 0 || len(diff.Removed) > 0 || len(diff.Changed) > 0
	return diff, nil
}

// snapshotRequestedHostServices returns requested hostServices from one snapshot.
func snapshotRequestedHostServices(snapshot *catalog.ManifestSnapshot) []*pluginbridge.HostServiceSpec {
	if snapshot == nil {
		return nil
	}
	return snapshot.RequestedHostServices
}

// normalizeRuntimeUpgradeHostServices normalizes hostServices into one map.
func normalizeRuntimeUpgradeHostServices(
	snapshot *catalog.ManifestSnapshot,
	specs []*pluginbridge.HostServiceSpec,
) (map[string]*pluginbridge.HostServiceSpec, error) {
	normalized, err := pluginbridge.NormalizeHostServiceSpecs(specs)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", snapshotPluginID(snapshot)),
			bizerr.P("version", snapshotVersion(snapshot)),
		)
	}
	result := make(map[string]*pluginbridge.HostServiceSpec, len(normalized))
	for _, spec := range normalized {
		if spec == nil || strings.TrimSpace(spec.Service) == "" {
			continue
		}
		result[strings.TrimSpace(spec.Service)] = spec
	}
	return result, nil
}

// snapshotPluginID safely extracts the plugin ID for diagnostics.
func snapshotPluginID(snapshot *catalog.ManifestSnapshot) string {
	if snapshot == nil {
		return ""
	}
	return snapshot.ID
}

// snapshotVersion safely extracts the plugin version for diagnostics.
func snapshotVersion(snapshot *catalog.ManifestSnapshot) string {
	if snapshot == nil {
		return ""
	}
	return snapshot.Version
}

// buildRuntimeUpgradeHostServiceChange converts two service specs into a
// stable summary for API projection.
func buildRuntimeUpgradeHostServiceChange(
	service string,
	fromSpec *pluginbridge.HostServiceSpec,
	toSpec *pluginbridge.HostServiceSpec,
) *RuntimeUpgradeHostServiceChange {
	return &RuntimeUpgradeHostServiceChange{
		Service:           service,
		FromMethods:       cloneSortedStrings(hostServiceMethods(fromSpec)),
		ToMethods:         cloneSortedStrings(hostServiceMethods(toSpec)),
		FromResourceCount: len(hostServiceResources(fromSpec)),
		ToResourceCount:   len(hostServiceResources(toSpec)),
		FromTables:        cloneSortedStrings(hostServiceTables(fromSpec)),
		ToTables:          cloneSortedStrings(hostServiceTables(toSpec)),
		FromPaths:         cloneSortedStrings(hostServicePaths(fromSpec)),
		ToPaths:           cloneSortedStrings(hostServicePaths(toSpec)),
	}
}

// runtimeUpgradeHostServiceChanged reports whether two normalized specs differ
// in methods or governed targets.
func runtimeUpgradeHostServiceChanged(
	fromSpec *pluginbridge.HostServiceSpec,
	toSpec *pluginbridge.HostServiceSpec,
) bool {
	if fromSpec == nil || toSpec == nil {
		return fromSpec != toSpec
	}
	if !stringSlicesEqual(hostServiceMethods(fromSpec), hostServiceMethods(toSpec)) {
		return true
	}
	if !stringSlicesEqual(hostServiceTables(fromSpec), hostServiceTables(toSpec)) {
		return true
	}
	if !stringSlicesEqual(hostServicePaths(fromSpec), hostServicePaths(toSpec)) {
		return true
	}
	return !stringSlicesEqual(hostServiceResourceRefs(fromSpec), hostServiceResourceRefs(toSpec))
}

// hostServiceMethods returns normalized methods from one spec.
func hostServiceMethods(spec *pluginbridge.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Methods
}

// hostServiceTables returns normalized data tables from one spec.
func hostServiceTables(spec *pluginbridge.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Tables
}

// hostServicePaths returns normalized storage paths from one spec.
func hostServicePaths(spec *pluginbridge.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Paths
}

// hostServiceResources returns governed resources from one spec.
func hostServiceResources(spec *pluginbridge.HostServiceSpec) []*pluginbridge.HostServiceResourceSpec {
	if spec == nil {
		return nil
	}
	return spec.Resources
}

// hostServiceResourceRefs returns normalized resource refs from one spec.
func hostServiceResourceRefs(spec *pluginbridge.HostServiceSpec) []string {
	resources := hostServiceResources(spec)
	refs := make([]string, 0, len(resources))
	for _, resource := range resources {
		if resource == nil || strings.TrimSpace(resource.Ref) == "" {
			continue
		}
		refs = append(refs, strings.TrimSpace(resource.Ref))
	}
	sort.Strings(refs)
	return refs
}

// cloneSortedStrings copies and sorts a string slice before exposing it.
func cloneSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

// stringSlicesEqual compares slices as normalized sorted sets.
func stringSlicesEqual(left []string, right []string) bool {
	leftValues := cloneSortedStrings(left)
	rightValues := cloneSortedStrings(right)
	if len(leftValues) != len(rightValues) {
		return false
	}
	for index := range leftValues {
		if leftValues[index] != rightValues[index] {
			return false
		}
	}
	return true
}

// buildRuntimeUpgradeRiskHints returns stable i18n keys for operator risk hints.
func buildRuntimeUpgradeRiskHints(
	sqlSummary RuntimeUpgradeSQLSummary,
	hostServicesDiff RuntimeUpgradeHostServicesDiff,
	dependencyCheck *DependencyCheckResult,
) []string {
	hints := make([]string, 0, 4)
	if sqlSummary.InstallSQLCount > 0 || sqlSummary.RuntimeSQLAssetCount > 0 {
		hints = append(hints, RuntimeUpgradeRiskHintUpgradeSQLRequiresReview)
	}
	if sqlSummary.MockSQLCount > 0 {
		hints = append(hints, RuntimeUpgradeRiskHintMockSQLExcluded)
	}
	if hostServicesDiff.AuthorizationChanged || hostServicesDiff.AuthorizationRequired {
		hints = append(hints, RuntimeUpgradeRiskHintHostServiceAuthorizationChanged)
	}
	if dependencyCheck != nil && (len(dependencyCheck.Blockers) > 0 || len(dependencyCheck.ReverseBlockers) > 0) {
		hints = append(hints, RuntimeUpgradeRiskHintDependencyBlockers)
	}
	return hints
}
