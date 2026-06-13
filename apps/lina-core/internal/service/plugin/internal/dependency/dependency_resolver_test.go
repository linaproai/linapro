// This file verifies side-effect-free plugin dependency graph resolution.

package dependency_test

import (
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// TestCheckInstallBlocksUnsatisfiedFrameworkVersion verifies framework
// dependency ranges are checked before lifecycle side effects.
func TestCheckInstallBlocksUnsatisfiedFrameworkVersion(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithFramework(">=0.2.0")),
		},
	})

	if len(result.Blockers) != 1 {
		t.Fatalf("expected one blocker, got %#v", result.Blockers)
	}
	if result.Blockers[0].Code != plugindep.BlockerFrameworkVersionUnsatisfied {
		t.Fatalf("expected framework blocker, got %#v", result.Blockers[0])
	}
	if result.Framework.CurrentVersion != "v0.1.0" || result.Framework.RequiredVersion != ">=0.2.0" {
		t.Fatalf("unexpected framework result: %#v", result.Framework)
	}
}

// TestCheckInstallBlocksUninstalledDependencies verifies declared plugin
// dependencies are hard blockers and are not installed automatically.
func TestCheckInstallBlocksUninstalledDependencies(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-a", ">=0.1.0"),
			)),
			pluginSnapshot("dep-a", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-c", ">=0.1.0"),
			)),
			pluginSnapshot("dep-c", "v0.1.0", false, nil),
		},
	})

	if len(result.Blockers) != 1 || !hasBlocker(result.Blockers, plugindep.BlockerDependencyMissing, "dep-a") {
		t.Fatalf("expected uninstalled dependency blocker, got %#v", result.Blockers)
	}
	if hasDependency(result.Dependencies, "dep-c") {
		t.Fatalf("expected transitive dependencies to be skipped until dep-a is installed, got %#v", result.Dependencies)
	}
}

// TestCheckInstallChecksTransitiveDependenciesWhenParentIsInstalled verifies
// installed dependencies are traversed so their own hard dependencies block.
func TestCheckInstallChecksTransitiveDependenciesWhenParentIsInstalled(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-a", ">=0.1.0"),
			)),
			pluginSnapshot("dep-a", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("dep-c", ">=0.1.0"),
			)),
			pluginSnapshot("dep-c", "v0.1.0", false, nil),
		},
	})

	if !hasBlocker(result.Blockers, plugindep.BlockerDependencyMissing, "dep-c") {
		t.Fatalf("expected transitive dependency blocker, got %#v", result.Blockers)
	}
}

// TestCheckInstallReportsMissingAndVersionUnsatisfied verifies hard
// dependencies fail when unavailable or outside the version range.
func TestCheckInstallReportsMissingAndVersionUnsatisfied(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-missing", ">=0.1.0"),
				pluginDependency("dep-old", ">=0.2.0"),
			)),
			pluginSnapshot("dep-old", "v0.1.0", true, nil),
		},
	})

	if len(result.Blockers) != 2 {
		t.Fatalf("expected two blockers, got %#v", result.Blockers)
	}
	if !hasBlocker(result.Blockers, plugindep.BlockerDependencyMissing, "dep-missing") {
		t.Fatalf("expected missing dependency blocker, got %#v", result.Blockers)
	}
	if !hasBlocker(result.Blockers, plugindep.BlockerDependencyVersionUnsatisfied, "dep-old") {
		t.Fatalf("expected version dependency blocker, got %#v", result.Blockers)
	}
}

// TestCheckInstallTreatsDeclaredDependenciesAsHard verifies every declared
// plugin dependency blocks lifecycle when unsatisfied.
func TestCheckInstallTreatsDeclaredDependenciesAsHard(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("analytics", ">=0.1.0"),
			)),
		},
	})

	if len(result.Blockers) != 1 || !hasBlocker(result.Blockers, plugindep.BlockerDependencyMissing, "analytics") {
		t.Fatalf("expected hard dependency blocker, got %#v", result.Blockers)
	}
}

// TestCheckInstallBlocksDependencyCycle verifies hard dependency cycles are
// returned as blockers with the cycle chain.
func TestCheckInstallBlocksDependencyCycle(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "a",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("a", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("b", ""),
			)),
			pluginSnapshot("b", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("c", ""),
			)),
			pluginSnapshot("c", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("a", ""),
			)),
		},
	})

	if len(result.Cycle) == 0 {
		t.Fatalf("expected cycle chain, got %#v", result)
	}
	if !hasBlocker(result.Blockers, plugindep.BlockerDependencyCycle, "a") {
		t.Fatalf("expected cycle blocker, got %#v", result.Blockers)
	}
}

// TestCheckReverseBlocksInstalledDependents verifies uninstall protection
// catches installed downstream hard dependencies.
func TestCheckReverseBlocksInstalledDependents(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID: "base",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("base", "v0.1.0", true, nil),
			pluginSnapshot("consumer", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("base", ">=0.1.0"),
			)),
		},
	})

	if len(result.Dependents) != 1 || result.Dependents[0].PluginID != "consumer" {
		t.Fatalf("expected consumer dependent, got %#v", result.Dependents)
	}
	if len(result.Blockers) != 1 || result.Blockers[0].Code != plugindep.BlockerReverseDependency {
		t.Fatalf("expected reverse dependency blocker, got %#v", result.Blockers)
	}
}

// TestCheckReverseBlocksCandidateVersionBreakage verifies upgrades cannot
// break installed downstream hard dependency ranges.
func TestCheckReverseBlocksCandidateVersionBreakage(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID:         "base",
		CandidateVersion: "v0.3.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("base", "v0.3.0", true, nil),
			pluginSnapshot("consumer", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("base", "<0.3.0"),
			)),
		},
	})

	if len(result.Blockers) != 1 || result.Blockers[0].Code != plugindep.BlockerReverseDependencyVersion {
		t.Fatalf("expected reverse dependency version blocker, got %#v", result.Blockers)
	}
}

// TestCheckReverseBlocksUnknownSnapshotWithoutDiscoveredManifest verifies
// destructive lifecycle operations fail closed when neither a release snapshot
// nor a discovered manifest can identify downstream dependencies.
func TestCheckReverseBlocksUnknownSnapshotWithoutDiscoveredManifest(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID: "base",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("base", "v0.1.0", true, nil),
			{
				ID:                        "consumer",
				Name:                      "consumer",
				Version:                   "v0.1.0",
				Installed:                 true,
				DependencySnapshotUnknown: true,
			},
		},
	})

	if len(result.Blockers) != 1 || result.Blockers[0].Code != plugindep.BlockerDependencySnapshotUnknown {
		t.Fatalf("expected unknown snapshot blocker, got %#v", result.Blockers)
	}
}

// TestCheckReverseIgnoresUnknownSnapshotForUnrelatedDiscoveredPlugin verifies
// one stale installed plugin does not block unrelated target lifecycle when the
// current discovered manifest shows it does not depend on the target.
func TestCheckReverseIgnoresUnknownSnapshotForUnrelatedDiscoveredPlugin(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID: "base",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("base", "v0.1.0", true, nil),
			{
				ID:                        "unrelated",
				Name:                      "unrelated",
				Version:                   "v0.1.0",
				Installed:                 true,
				Manifest:                  &catalog.Manifest{ID: "unrelated"},
				DependencySnapshotUnknown: true,
			},
		},
	})

	if len(result.Blockers) != 0 {
		t.Fatalf("expected unrelated unknown snapshot to be non-blocking, got %#v", result.Blockers)
	}
}

// TestCheckReverseBlocksUnknownSnapshotWithDiscoveredTargetDependency verifies
// current discovered manifests still protect the target when the effective
// release snapshot is unavailable but a hard dependency on the target is known.
func TestCheckReverseBlocksUnknownSnapshotWithDiscoveredTargetDependency(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID: "base",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("base", "v0.1.0", true, nil),
			{
				ID:                        "consumer",
				Name:                      "consumer",
				Version:                   "v0.1.0",
				Installed:                 true,
				Manifest:                  &catalog.Manifest{ID: "consumer"},
				Dependencies:              dependenciesWithPlugins(pluginDependency("base", ">=0.1.0")),
				DependencySnapshotUnknown: true,
			},
		},
	})

	if len(result.Blockers) != 1 || result.Blockers[0].Code != plugindep.BlockerDependencySnapshotUnknown {
		t.Fatalf("expected discovered hard dependency with unknown snapshot to block, got %#v", result.Blockers)
	}
}

func pluginSnapshot(id string, version string, installed bool, dependencies *plugintypes.DependencySpec) *plugindep.PluginSnapshot {
	return &plugindep.PluginSnapshot{
		ID:           id,
		Name:         id,
		Version:      version,
		Installed:    installed,
		Dependencies: dependencies,
	}
}

func dependenciesWithFramework(version string) *plugintypes.DependencySpec {
	return &plugintypes.DependencySpec{
		Framework: &plugintypes.FrameworkDependencySpec{Version: version},
	}
}

func dependenciesWithPlugins(plugins ...*plugintypes.PluginDependencySpec) *plugintypes.DependencySpec {
	return &plugintypes.DependencySpec{Plugins: plugins}
}

func pluginDependency(id string, version string) *plugintypes.PluginDependencySpec {
	return &plugintypes.PluginDependencySpec{
		ID:      id,
		Version: version,
	}
}

func hasBlocker(blockers []*plugindep.Blocker, code plugindep.BlockerCode, dependencyID string) bool {
	for _, blocker := range blockers {
		if blocker != nil && blocker.Code == code && blocker.DependencyID == dependencyID {
			return true
		}
	}
	return false
}

func hasDependency(dependencies []*plugindep.PluginDependencyCheck, dependencyID string) bool {
	for _, dependency := range dependencies {
		if dependency != nil && dependency.DependencyID == dependencyID {
			return true
		}
	}
	return false
}
