// This file verifies side-effect-free plugin dependency graph resolution.

package dependency_test

import (
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
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

// TestCheckInstallBuildsDependencyFirstAutoPlan verifies missing automatic
// hard dependencies are planned in deterministic topology order.
func TestCheckInstallBuildsDependencyFirstAutoPlan(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-b", ">=0.1.0", true, catalog.DependencyInstallModeAuto.String()),
				pluginDependency("dep-a", ">=0.1.0", true, catalog.DependencyInstallModeAuto.String()),
			)),
			pluginSnapshot("dep-a", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-c", ">=0.1.0", true, catalog.DependencyInstallModeAuto.String()),
			)),
			pluginSnapshot("dep-b", "v0.1.0", false, nil),
			pluginSnapshot("dep-c", "v0.1.0", false, nil),
		},
	})

	if len(result.Blockers) != 0 {
		t.Fatalf("expected no blockers, got %#v", result.Blockers)
	}
	planIDs := collectPlanIDs(result.AutoInstallPlan)
	want := []string{"dep-c", "dep-a", "dep-b"}
	if !stringSlicesEqual(planIDs, want) {
		t.Fatalf("expected auto plan %v, got %v", want, planIDs)
	}
}

// TestCheckInstallBlocksManualDependency verifies missing manual hard
// dependencies require operator action.
func TestCheckInstallBlocksManualDependency(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("dep-manual", ">=0.1.0", true, catalog.DependencyInstallModeManual.String()),
			)),
			pluginSnapshot("dep-manual", "v0.1.0", false, nil),
		},
	})

	if len(result.ManualInstallRequired) != 1 {
		t.Fatalf("expected manual dependency, got %#v", result.ManualInstallRequired)
	}
	if len(result.Blockers) != 1 || result.Blockers[0].Code != plugindep.BlockerDependencyManualInstallRequired {
		t.Fatalf("expected manual dependency blocker, got %#v", result.Blockers)
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
				pluginDependency("dep-missing", ">=0.1.0", true, catalog.DependencyInstallModeAuto.String()),
				pluginDependency("dep-old", ">=0.2.0", true, catalog.DependencyInstallModeAuto.String()),
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

// TestCheckInstallKeepsSoftDependenciesNonBlocking verifies optional
// dependencies are reported but do not block install.
func TestCheckInstallKeepsSoftDependenciesNonBlocking(t *testing.T) {
	resolver := plugindep.New()

	result := resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         "target",
		FrameworkVersion: "v0.1.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("target", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("optional-analytics", ">=0.1.0", false, catalog.DependencyInstallModeAuto.String()),
			)),
		},
	})

	if len(result.Blockers) != 0 {
		t.Fatalf("expected no blockers for soft dependency, got %#v", result.Blockers)
	}
	if len(result.SoftUnsatisfied) != 1 || result.SoftUnsatisfied[0].DependencyID != "optional-analytics" {
		t.Fatalf("expected soft dependency notice, got %#v", result.SoftUnsatisfied)
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
				pluginDependency("b", "", true, catalog.DependencyInstallModeAuto.String()),
			)),
			pluginSnapshot("b", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("c", "", true, catalog.DependencyInstallModeAuto.String()),
			)),
			pluginSnapshot("c", "v0.1.0", false, dependenciesWithPlugins(
				pluginDependency("a", "", true, catalog.DependencyInstallModeAuto.String()),
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
				pluginDependency("base", ">=0.1.0", true, catalog.DependencyInstallModeManual.String()),
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
				pluginDependency("base", "<0.3.0", true, catalog.DependencyInstallModeManual.String()),
			)),
		},
	})

	if len(result.Blockers) != 1 || result.Blockers[0].Code != plugindep.BlockerReverseDependencyVersion {
		t.Fatalf("expected reverse dependency version blocker, got %#v", result.Blockers)
	}
}

// TestCheckReverseConservativelyBlocksUnknownSnapshot verifies missing
// installed dependency snapshots block destructive lifecycle operations.
func TestCheckReverseConservativelyBlocksUnknownSnapshot(t *testing.T) {
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

func pluginSnapshot(id string, version string, installed bool, dependencies *catalog.DependencySpec) *plugindep.PluginSnapshot {
	return &plugindep.PluginSnapshot{
		ID:           id,
		Name:         id,
		Version:      version,
		Installed:    installed,
		Dependencies: dependencies,
	}
}

func dependenciesWithFramework(version string) *catalog.DependencySpec {
	return &catalog.DependencySpec{
		Framework: &catalog.FrameworkDependencySpec{Version: version},
	}
}

func dependenciesWithPlugins(plugins ...*catalog.PluginDependencySpec) *catalog.DependencySpec {
	return &catalog.DependencySpec{Plugins: plugins}
}

func pluginDependency(id string, version string, required bool, install string) *catalog.PluginDependencySpec {
	return &catalog.PluginDependencySpec{
		ID:       id,
		Version:  version,
		Required: &required,
		Install:  install,
	}
}

func collectPlanIDs(items []*plugindep.AutoInstallPlanItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil {
			out = append(out, item.PluginID)
		}
	}
	return out
}

func hasBlocker(blockers []*plugindep.Blocker, code plugindep.BlockerCode, dependencyID string) bool {
	for _, blocker := range blockers {
		if blocker != nil && blocker.Code == code && blocker.DependencyID == dependencyID {
			return true
		}
	}
	return false
}

func stringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
