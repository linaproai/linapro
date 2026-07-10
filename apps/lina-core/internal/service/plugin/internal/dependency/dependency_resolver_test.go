// This file verifies side-effect-free plugin dependency graph resolution.

package dependency_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	pluginv1 "lina-core/api/plugin/v1"
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
	if result.Blockers[0].Code != pluginv1.BlockerCodeFrameworkVersionUnsatisfied {
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

	if len(result.Blockers) != 1 || !hasBlocker(result.Blockers, pluginv1.BlockerCodeDependencyMissing, "dep-a") {
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

	if !hasBlocker(result.Blockers, pluginv1.BlockerCodeDependencyMissing, "dep-c") {
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
	if !hasBlocker(result.Blockers, pluginv1.BlockerCodeDependencyMissing, "dep-missing") {
		t.Fatalf("expected missing dependency blocker, got %#v", result.Blockers)
	}
	if !hasBlocker(result.Blockers, pluginv1.BlockerCodeDependencyVersionUnsatisfied, "dep-old") {
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

	if len(result.Blockers) != 1 || !hasBlocker(result.Blockers, pluginv1.BlockerCodeDependencyMissing, "analytics") {
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
	if !hasBlocker(result.Blockers, pluginv1.BlockerCodeDependencyCycle, "a") {
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
	if len(result.Blockers) != 1 || result.Blockers[0].Code != pluginv1.BlockerCodeReverseDependency {
		t.Fatalf("expected reverse dependency blocker, got %#v", result.Blockers)
	}
}

// TestCheckReverseIncludesOwnerHostServiceSummaries verifies reverse
// dependency diagnostics preserve owner-aware host service methods for the
// target owner only.
func TestCheckReverseIncludesOwnerHostServiceSummaries(t *testing.T) {
	resolver := plugindep.New()
	consumer := pluginSnapshot("consumer", "v0.1.0", true, dependenciesWithPlugins(
		pluginDependency("owner-core", "<0.2.0"),
	))
	consumer.OwnerHostServices = []*plugindep.OwnerHostServiceSummary{
		{
			Owner:   "unrelated-owner",
			Service: "ai",
			Version: "v1",
			Methods: []string{
				"text.generate",
			},
		},
		{
			Owner:   "owner-core",
			Service: "ai",
			Version: "v1",
			Methods: []string{
				"text.method_status.get",
				"text.generate",
			},
		},
	}

	result := resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID:         "owner-core",
		CandidateVersion: "v0.2.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("owner-core", "v0.2.0", true, nil),
			consumer,
		},
	})

	if len(result.Dependents) != 1 || result.Dependents[0].PluginID != "consumer" {
		t.Fatalf("expected consumer dependent, got %#v", result.Dependents)
	}
	hostServices := result.Dependents[0].OwnerHostServices
	if len(hostServices) != 1 {
		t.Fatalf("expected one target owner host service summary, got %#v", hostServices)
	}
	if hostServices[0].Owner != "owner-core" ||
		hostServices[0].Service != "ai" ||
		hostServices[0].Version != "v1" {
		t.Fatalf("expected owner-core ai.v1 summary, got %#v", hostServices[0])
	}
	if !reflect.DeepEqual(hostServices[0].Methods, []string{"text.generate", "text.method_status.get"}) {
		t.Fatalf("expected sorted owner methods, got %#v", hostServices[0].Methods)
	}
	if len(result.Blockers) != 1 || result.Blockers[0].Code != pluginv1.BlockerCodeReverseDependencyVersion {
		t.Fatalf("expected reverse dependency version blocker, got %#v", result.Blockers)
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

	if len(result.Blockers) != 1 || result.Blockers[0].Code != pluginv1.BlockerCodeReverseDependencyVersion {
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

	if len(result.Blockers) != 1 || result.Blockers[0].Code != pluginv1.BlockerCodeDependencySnapshotUnknown {
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

	if len(result.Blockers) != 1 || result.Blockers[0].Code != pluginv1.BlockerCodeDependencySnapshotUnknown {
		t.Fatalf("expected discovered hard dependency with unknown snapshot to block, got %#v", result.Blockers)
	}
}

// TestCheckReverseMatchesLegacyTraversal verifies the indexed reverse check
// keeps the same externally visible result shape as the former full-scan logic.
func TestCheckReverseMatchesLegacyTraversal(t *testing.T) {
	input := plugindep.ReverseCheckInput{
		TargetID:         "base",
		CandidateVersion: "v0.3.0",
		Plugins: []*plugindep.PluginSnapshot{
			pluginSnapshot("base", "v0.3.0", true, nil),
			pluginSnapshot("consumer-a", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("base", "<0.3.0"),
			)),
			pluginSnapshot("consumer-b", "v0.1.0", true, dependenciesWithPlugins(
				pluginDependency("unrelated", ">=0.1.0"),
			)),
			{
				ID:                        "consumer-c",
				Name:                      "consumer-c",
				Version:                   "v0.1.0",
				Installed:                 true,
				Dependencies:              dependenciesWithPlugins(pluginDependency("base", ">=0.1.0")),
				DependencySnapshotUnknown: true,
			},
			{
				ID:                        "unknown-with-manifest",
				Name:                      "unknown-with-manifest",
				Version:                   "v0.1.0",
				Installed:                 true,
				Manifest:                  &catalog.Manifest{ID: "unknown-with-manifest"},
				DependencySnapshotUnknown: true,
			},
			{
				ID:                        "unknown-without-manifest",
				Name:                      "unknown-without-manifest",
				Version:                   "v0.1.0",
				Installed:                 true,
				DependencySnapshotUnknown: true,
			},
		},
	}

	actual := plugindep.New().CheckReverse(input)
	expected := legacyCheckReverseForTest(input)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected indexed reverse check to match legacy traversal\nactual:   %#v\nexpected: %#v", actual, expected)
	}
}

// TestCheckReverseThirtyPluginFixtureReturnsOnlyTargetDependents verifies a
// target lookup is bounded to indexed downstream edges rather than reporting
// unrelated plugins from a larger snapshot set.
func TestCheckReverseThirtyPluginFixtureReturnsOnlyTargetDependents(t *testing.T) {
	plugins := make([]*plugindep.PluginSnapshot, 0, 31)
	plugins = append(plugins, pluginSnapshot("target-07", "v0.1.0", true, nil))
	for i := 0; i < 30; i++ {
		pluginID := "consumer-" + twoDigit(i)
		dependencyID := "unrelated-" + twoDigit(i)
		if i == 3 || i == 17 || i == 29 {
			dependencyID = "target-07"
		}
		plugins = append(plugins, pluginSnapshot(pluginID, "v0.1.0", true, dependenciesWithPlugins(
			pluginDependency(dependencyID, ">=0.1.0"),
		)))
	}

	result := plugindep.New().CheckReverse(plugindep.ReverseCheckInput{
		TargetID: "target-07",
		Plugins:  plugins,
	})

	if len(result.Dependents) != 3 {
		t.Fatalf("expected three target dependents, got %#v", result.Dependents)
	}
	dependentIDs := []string{
		result.Dependents[0].PluginID,
		result.Dependents[1].PluginID,
		result.Dependents[2].PluginID,
	}
	expectedIDs := []string{"consumer-03", "consumer-17", "consumer-29"}
	if !reflect.DeepEqual(dependentIDs, expectedIDs) {
		t.Fatalf("expected target dependents %v, got %v", expectedIDs, dependentIDs)
	}
	if len(result.Blockers) != 3 {
		t.Fatalf("expected uninstall blockers for target dependents, got %#v", result.Blockers)
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

func legacyCheckReverseForTest(input plugindep.ReverseCheckInput) *plugindep.ReverseCheckResult {
	targetID := strings.TrimSpace(input.TargetID)
	result := &plugindep.ReverseCheckResult{
		TargetID:         targetID,
		CandidateVersion: strings.TrimSpace(input.CandidateVersion),
	}
	for _, plugin := range legacySortedSnapshotsForTest(input.Plugins) {
		if plugin == nil || !plugin.Installed || strings.TrimSpace(plugin.ID) == targetID {
			continue
		}
		if plugin.DependencySnapshotUnknown {
			if legacyUnknownSnapshotRequiresReverseBlockForTest(plugin, targetID, result.CandidateVersion == "") {
				result.Blockers = append(result.Blockers, &plugindep.Blocker{
					Code:     pluginv1.BlockerCodeDependencySnapshotUnknown,
					PluginID: strings.TrimSpace(plugin.ID),
					Detail:   "installed plugin dependency snapshot is unavailable",
				})
			}
			continue
		}
		for _, declaredDependency := range legacyNormalizedPluginDependenciesForTest(plugin.Dependencies) {
			if declaredDependency == nil || strings.TrimSpace(declaredDependency.ID) != targetID {
				continue
			}
			dependent := &plugindep.ReverseDependent{
				PluginID:        strings.TrimSpace(plugin.ID),
				Name:            strings.TrimSpace(plugin.Name),
				Version:         strings.TrimSpace(plugin.Version),
				RequiredVersion: strings.TrimSpace(declaredDependency.Version),
			}
			result.Dependents = append(result.Dependents, dependent)

			if result.CandidateVersion == "" {
				result.Blockers = append(result.Blockers, &plugindep.Blocker{
					Code:            pluginv1.BlockerCodeReverseDependency,
					PluginID:        dependent.PluginID,
					DependencyID:    targetID,
					RequiredVersion: dependent.RequiredVersion,
					CurrentVersion:  result.CandidateVersion,
					Chain:           []string{dependent.PluginID, targetID},
					Detail:          "installed plugin depends on target plugin",
				})
				continue
			}
			if dependent.RequiredVersion == "" {
				continue
			}
			matches, err := plugintypes.MatchesSemanticVersionRange(result.CandidateVersion, dependent.RequiredVersion)
			if err != nil || !matches {
				result.Blockers = append(result.Blockers, &plugindep.Blocker{
					Code:            pluginv1.BlockerCodeReverseDependencyVersion,
					PluginID:        dependent.PluginID,
					DependencyID:    targetID,
					RequiredVersion: dependent.RequiredVersion,
					CurrentVersion:  result.CandidateVersion,
					Chain:           []string{dependent.PluginID, targetID},
					Detail:          "candidate version does not satisfy downstream dependency range",
				})
			}
		}
	}
	return result
}

func legacyUnknownSnapshotRequiresReverseBlockForTest(plugin *plugindep.PluginSnapshot, targetID string, uninstall bool) bool {
	dependencies := legacyNormalizedPluginDependenciesForTest(plugin.Dependencies)
	if len(dependencies) == 0 {
		return uninstall && plugin.Manifest == nil
	}
	for _, declaredDependency := range dependencies {
		if declaredDependency == nil || strings.TrimSpace(declaredDependency.ID) != targetID {
			continue
		}
		return true
	}
	return false
}

func legacyNormalizedPluginDependenciesForTest(spec *plugintypes.DependencySpec) []*plugintypes.PluginDependencySpec {
	if spec == nil || len(spec.Plugins) == 0 {
		return nil
	}
	normalized := plugintypes.CloneDependencySpec(spec)
	plugintypes.NormalizeDependencySpec(normalized)
	dependencies := make([]*plugintypes.PluginDependencySpec, 0, len(normalized.Plugins))
	for _, dependency := range normalized.Plugins {
		if dependency != nil {
			dependencies = append(dependencies, dependency)
		}
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.TrimSpace(dependencies[i].ID) < strings.TrimSpace(dependencies[j].ID)
	})
	return dependencies
}

func legacySortedSnapshotsForTest(plugins []*plugindep.PluginSnapshot) []*plugindep.PluginSnapshot {
	sorted := make([]*plugindep.PluginSnapshot, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin != nil {
			sorted = append(sorted, plugin)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return strings.TrimSpace(sorted[i].ID) < strings.TrimSpace(sorted[j].ID)
	})
	return sorted
}

func twoDigit(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return string(rune('0'+value/10)) + string(rune('0'+value%10))
}
