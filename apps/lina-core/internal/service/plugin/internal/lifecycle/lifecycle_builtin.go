// This file owns startup reconciliation for built-in source plugins. It keeps
// builtin lifecycle convergence separate from plugin.autoEnable while reusing
// the same lifecycle side effects, dependency checks, cluster waiting, and
// enabled-snapshot publication paths.

package lifecycle

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
)

// BootstrapBuiltinOptions carries startup inputs for builtin plugin convergence.
type BootstrapBuiltinOptions struct {
	// Manifests are the already-synchronized desired plugin manifests.
	Manifests []*catalog.Manifest
	// FrameworkVersion is the current LinaPro framework version used for
	// dependency compatibility checks.
	FrameworkVersion string
	// Upgrade executes a startup-safe upgrade for an already installed builtin
	// plugin. It is called only by the primary node or single-node startup path.
	Upgrade func(ctx context.Context, pluginID string) error
}

// BootstrapBuiltinPlugins installs, upgrades, and enables project built-in
// source plugins during startup. It never loads mock-data SQL and does not
// consult plugin.autoEnable.
func (s *serviceImpl) BootstrapBuiltinPlugins(ctx context.Context, options BootstrapBuiltinOptions) error {
	if s == nil {
		return nil
	}
	builtinManifests, err := orderBuiltinManifests(options.Manifests)
	if err != nil {
		return err
	}
	if len(builtinManifests) == 0 {
		return nil
	}
	for _, manifest := range builtinManifests {
		if err = s.bootstrapBuiltinPlugin(ctx, manifest, options); err != nil {
			return err
		}
	}
	if err = s.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		return bizerr.WrapCode(err, CodePluginEnabledSnapshotRefreshFailed)
	}
	return nil
}

// bootstrapBuiltinPlugin reconciles one builtin source plugin to a normal,
// installed, and enabled startup state.
func (s *serviceImpl) bootstrapBuiltinPlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	options BootstrapBuiltinOptions,
) error {
	if manifest == nil {
		return nil
	}
	pluginID := strings.TrimSpace(manifest.ID)
	if pluginID == "" {
		return nil
	}
	return s.ensurePluginStateDuringStartupUnwrapped(ctx, pluginID, func(registry *store.PluginRecord) bool {
		return s.isBuiltinStartupReady(ctx, registry, manifest)
	}, func() error {
		if err := s.installBuiltinPluginIfNeeded(ctx, manifest, options.FrameworkVersion); err != nil {
			return err
		}
		if err := s.upgradeBuiltinPluginIfNeeded(ctx, manifest, options); err != nil {
			return err
		}
		if err := s.enableBuiltinPluginIfNeeded(ctx, manifest, options.FrameworkVersion); err != nil {
			return err
		}
		return nil
	})
}

// installBuiltinPluginIfNeeded installs one builtin source plugin without
// loading optional mock-data SQL.
func (s *serviceImpl) installBuiltinPluginIfNeeded(
	ctx context.Context,
	manifest *catalog.Manifest,
	frameworkVersion string,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return bizerr.WrapCode(err, CodePluginRegistryReadFailed, bizerr.P("pluginId", manifest.ID))
	}
	if registry != nil && registry.Installed == plugintypes.InstalledYes {
		return nil
	}
	if _, err = s.Install(ctx, manifest.ID, InstallOptions{
		InstallMockData:  false,
		FrameworkVersion: frameworkVersion,
	}); err != nil {
		return bizerr.WrapCode(err, CodePluginSourceInstallFailed)
	}
	return nil
}

// upgradeBuiltinPluginIfNeeded uses the startup-safe upgrade callback when the
// desired manifest is newer than the effective registry release.
func (s *serviceImpl) upgradeBuiltinPluginIfNeeded(
	ctx context.Context,
	manifest *catalog.Manifest,
	options BootstrapBuiltinOptions,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return bizerr.WrapCode(err, CodePluginRegistryReadFailed, bizerr.P("pluginId", manifest.ID))
	}
	projection, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return err
	}
	switch projection.State {
	case plugintypes.RuntimeUpgradeStateNormal:
		return nil
	default:
		if options.Upgrade == nil {
			return bizerr.NewCode(CodePluginAutoEnableSharedExecutorMissing, bizerr.P("pluginId", manifest.ID))
		}
		return options.Upgrade(ctx, manifest.ID)
	}
}

// enableBuiltinPluginIfNeeded enables one installed builtin source plugin.
func (s *serviceImpl) enableBuiltinPluginIfNeeded(
	ctx context.Context,
	manifest *catalog.Manifest,
	frameworkVersion string,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return bizerr.WrapCode(err, CodePluginRegistryReadFailed, bizerr.P("pluginId", manifest.ID))
	}
	if registry != nil && registry.Status == plugintypes.StatusEnabled {
		return nil
	}
	if err = s.UpdateStatus(ctx, manifest.ID, plugintypes.StatusEnabled, UpdateStatusOptions{
		FrameworkVersion: frameworkVersion,
	}); err != nil {
		return bizerr.WrapCode(err, CodePluginSourceEnableFailed)
	}
	return nil
}

// isBuiltinStartupReady reports whether one builtin plugin has fully converged
// for startup route, frontend, and cron registration.
func (s *serviceImpl) isBuiltinStartupReady(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
) bool {
	if !isPluginStartupEnabled(registry) {
		return false
	}
	if strings.TrimSpace(registry.CurrentState) != plugintypes.HostStateEnabled.String() {
		return false
	}
	projection, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return false
	}
	return projection.State == plugintypes.RuntimeUpgradeStateNormal
}

// orderBuiltinManifests returns builtin source manifests in dependency-first
// order. Dependencies outside the builtin set are left to lifecycle dependency
// checks so this sorter only controls execution order within the builtin set.
func orderBuiltinManifests(manifests []*catalog.Manifest) ([]*catalog.Manifest, error) {
	builtinByID := make(map[string]*catalog.Manifest, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
			continue
		}
		if plugintypes.NormalizeDistribution(manifest.Distribution) != plugintypes.DistributionBuiltin {
			continue
		}
		if plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeSource {
			continue
		}
		builtinByID[strings.TrimSpace(manifest.ID)] = manifest
	}
	if len(builtinByID) == 0 {
		return nil, nil
	}

	var (
		ordered  = make([]*catalog.Manifest, 0, len(builtinByID))
		visiting = make(map[string]bool, len(builtinByID))
		visited  = make(map[string]bool, len(builtinByID))
		visit    func(string) error
	)
	visit = func(pluginID string) error {
		if visited[pluginID] {
			return nil
		}
		if visiting[pluginID] {
			return bizerr.NewCode(
				CodePluginDependencyBlocked,
				bizerr.P("pluginId", pluginID),
				bizerr.P("dependencyId", pluginID),
				bizerr.P("blockers", "builtin dependency cycle"),
			)
		}
		manifest := builtinByID[pluginID]
		if manifest == nil {
			return nil
		}
		visiting[pluginID] = true
		for _, dependency := range builtinDependencyEdges(manifest) {
			if _, ok := builtinByID[dependency]; ok {
				if err := visit(dependency); err != nil {
					return err
				}
			}
		}
		visiting[pluginID] = false
		visited[pluginID] = true
		ordered = append(ordered, manifest)
		return nil
	}

	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if _, ok := builtinByID[pluginID]; !ok {
			continue
		}
		if err := visit(pluginID); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// builtinDependencyEdges returns direct plugin dependency IDs in manifest order.
func builtinDependencyEdges(manifest *catalog.Manifest) []string {
	if manifest == nil || manifest.Dependencies == nil {
		return nil
	}
	out := make([]string, 0, len(manifest.Dependencies.Plugins))
	for _, dependency := range manifest.Dependencies.Plugins {
		if dependency == nil {
			continue
		}
		dependencyID := strings.TrimSpace(dependency.ID)
		if dependencyID != "" {
			out = append(out, dependencyID)
		}
	}
	return out
}
