// This file implements deterministic, side-effect-free dependency resolution
// for plugin lifecycle planning.

package dependency

import (
	"lina-core/internal/service/plugin/internal/plugintypes"
	"sort"
	"strings"
)

// Resolver evaluates plugin dependency declarations against discovered and
// installed plugin snapshots.
type Resolver struct{}

// New creates a dependency resolver.
func New() *Resolver {
	return &Resolver{}
}

// CheckInstall evaluates whether target can be installed with all declared
// plugin dependencies already installed and version-compatible.
func (r *Resolver) CheckInstall(input InstallCheckInput) *InstallCheckResult {
	targetID := strings.TrimSpace(input.TargetID)
	result := &InstallCheckResult{
		TargetID: targetID,
	}
	plugins := buildPluginMap(input.Plugins)
	target := plugins[targetID]
	if target == nil {
		result.Blockers = append(result.Blockers, &Blocker{
			Code:     BlockerDependencyMissing,
			PluginID: targetID,
			Detail:   "target plugin snapshot is missing",
		})
		return result
	}

	result.Framework = r.checkFramework(target, input.FrameworkVersion)
	if result.Framework.Status == FrameworkStatusUnsatisfied {
		result.Blockers = append(result.Blockers, &Blocker{
			Code:            BlockerFrameworkVersionUnsatisfied,
			PluginID:        targetID,
			RequiredVersion: result.Framework.RequiredVersion,
			CurrentVersion:  result.Framework.CurrentVersion,
			Chain:           []string{targetID},
			Detail:          "framework version does not satisfy plugin dependency range",
		})
	}

	visited := map[string]bool{}
	visiting := map[string]int{}
	r.walkDependencies(walkState{
		owner:    target,
		plugins:  plugins,
		chain:    []string{targetID},
		visited:  visited,
		visiting: visiting,
		result:   result,
	})
	sortInstallResult(result)
	return result
}

// CheckReverse evaluates whether uninstalling or upgrading target would break
// installed downstream hard dependencies.
func (r *Resolver) CheckReverse(input ReverseCheckInput) *ReverseCheckResult {
	targetID := strings.TrimSpace(input.TargetID)
	result := &ReverseCheckResult{
		TargetID:         targetID,
		CandidateVersion: strings.TrimSpace(input.CandidateVersion),
	}
	for _, plugin := range sortedSnapshots(input.Plugins) {
		if plugin == nil || !plugin.Installed || strings.TrimSpace(plugin.ID) == targetID {
			continue
		}
		if plugin.DependencySnapshotUnknown {
			if unknownSnapshotRequiresReverseBlock(plugin, targetID, result.CandidateVersion == "") {
				result.Blockers = append(result.Blockers, &Blocker{
					Code:     BlockerDependencySnapshotUnknown,
					PluginID: strings.TrimSpace(plugin.ID),
					Detail:   "installed plugin dependency snapshot is unavailable",
				})
			}
			continue
		}
		for _, declaredDependency := range normalizedPluginDependencies(plugin.Dependencies) {
			if declaredDependency == nil || strings.TrimSpace(declaredDependency.ID) != targetID {
				continue
			}
			dependent := &ReverseDependent{
				PluginID:        strings.TrimSpace(plugin.ID),
				Name:            strings.TrimSpace(plugin.Name),
				Version:         strings.TrimSpace(plugin.Version),
				RequiredVersion: strings.TrimSpace(declaredDependency.Version),
			}
			result.Dependents = append(result.Dependents, dependent)

			if result.CandidateVersion == "" {
				result.Blockers = append(result.Blockers, &Blocker{
					Code:            BlockerReverseDependency,
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
				result.Blockers = append(result.Blockers, &Blocker{
					Code:            BlockerReverseDependencyVersion,
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

// unknownSnapshotRequiresReverseBlock keeps destructive lifecycle checks
// conservative while allowing unrelated installed plugins with a current
// discovered manifest to avoid blocking every target plugin operation.
func unknownSnapshotRequiresReverseBlock(plugin *PluginSnapshot, targetID string, uninstall bool) bool {
	dependencies := normalizedPluginDependencies(plugin.Dependencies)
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

// walkState carries mutable traversal state for install dependency resolution.
type walkState struct {
	owner    *PluginSnapshot
	plugins  map[string]*PluginSnapshot
	chain    []string
	visited  map[string]bool
	visiting map[string]int
	result   *InstallCheckResult
}

// checkFramework evaluates the target plugin framework-version declaration.
func (r *Resolver) checkFramework(plugin *PluginSnapshot, frameworkVersion string) FrameworkCheck {
	check := FrameworkCheck{
		CurrentVersion: strings.TrimSpace(frameworkVersion),
		Status:         FrameworkStatusNotDeclared,
	}
	dependencies := pluginDependencies(plugin)
	if dependencies == nil || dependencies.Framework == nil || strings.TrimSpace(dependencies.Framework.Version) == "" {
		return check
	}
	check.RequiredVersion = strings.TrimSpace(dependencies.Framework.Version)
	matches, err := plugintypes.MatchesSemanticVersionRange(check.CurrentVersion, check.RequiredVersion)
	if err != nil || !matches {
		check.Status = FrameworkStatusUnsatisfied
		return check
	}
	check.Status = FrameworkStatusSatisfied
	return check
}

// walkDependencies traverses hard dependency edges in deterministic order.
func (r *Resolver) walkDependencies(state walkState) {
	if state.owner == nil {
		return
	}
	ownerID := strings.TrimSpace(state.owner.ID)
	if ownerID == "" {
		return
	}
	if index, ok := state.visiting[ownerID]; ok {
		cycle := append([]string(nil), state.chain[index:]...)
		cycle = append(cycle, ownerID)
		state.result.Cycle = cycle
		state.result.Blockers = append(state.result.Blockers, &Blocker{
			Code:         BlockerDependencyCycle,
			PluginID:     ownerID,
			DependencyID: ownerID,
			Chain:        cycle,
			Detail:       "plugin dependency cycle detected",
		})
		return
	}
	if state.visited[ownerID] {
		return
	}

	state.visiting[ownerID] = len(state.chain) - 1
	dependencies := normalizedPluginDependencies(pluginDependencies(state.owner))
	for _, declaredDependency := range dependencies {
		dependencyID := strings.TrimSpace(declaredDependency.ID)
		if index, ok := state.visiting[dependencyID]; ok {
			cycle := append([]string(nil), state.chain[index:]...)
			cycle = append(cycle, dependencyID)
			state.result.Cycle = cycle
			state.result.Blockers = append(state.result.Blockers, &Blocker{
				Code:         BlockerDependencyCycle,
				PluginID:     ownerID,
				DependencyID: dependencyID,
				Chain:        cycle,
				Detail:       "plugin dependency cycle detected",
			})
			continue
		}
		check := r.evaluateDependency(state.owner, declaredDependency, state.plugins, state.chain)
		state.result.Dependencies = append(state.result.Dependencies, check)
		if check == nil {
			continue
		}
		r.recordDependencyOutcome(state, check)
		if check.Discovered && check.Status == DependencyStatusSatisfied {
			next := state.plugins[check.DependencyID]
			if next != nil {
				nextChain := append(append([]string(nil), state.chain...), check.DependencyID)
				r.walkDependencies(walkState{
					owner:    next,
					plugins:  state.plugins,
					chain:    nextChain,
					visited:  state.visited,
					visiting: state.visiting,
					result:   state.result,
				})
			}
		}
	}
	delete(state.visiting, ownerID)
	state.visited[ownerID] = true
}

// evaluateDependency classifies one declared dependency edge.
func (r *Resolver) evaluateDependency(
	owner *PluginSnapshot,
	declaredDependency *plugintypes.PluginDependencySpec,
	plugins map[string]*PluginSnapshot,
	chain []string,
) *PluginDependencyCheck {
	dependencyID := strings.TrimSpace(declaredDependency.ID)
	dependency := plugins[dependencyID]
	check := &PluginDependencyCheck{
		OwnerID:         strings.TrimSpace(owner.ID),
		DependencyID:    dependencyID,
		RequiredVersion: strings.TrimSpace(declaredDependency.Version),
		Chain:           append(append([]string(nil), chain...), dependencyID),
	}
	if dependency == nil {
		check.Status = DependencyStatusMissing
		return check
	}

	check.DependencyName = strings.TrimSpace(dependency.Name)
	check.CurrentVersion = strings.TrimSpace(dependency.Version)
	check.Installed = dependency.Installed
	check.Discovered = true
	if check.RequiredVersion != "" {
		matches, err := plugintypes.MatchesSemanticVersionRange(check.CurrentVersion, check.RequiredVersion)
		if err != nil || !matches {
			check.Status = DependencyStatusVersionUnsatisfied
			return check
		}
	}
	if dependency.Installed {
		check.Status = DependencyStatusSatisfied
		return check
	}
	check.Status = DependencyStatusMissing
	return check
}

// recordDependencyOutcome updates result collections for one dependency edge.
func (r *Resolver) recordDependencyOutcome(state walkState, check *PluginDependencyCheck) {
	switch check.Status {
	case DependencyStatusSatisfied:
		return
	case DependencyStatusMissing:
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, BlockerDependencyMissing, check)
	case DependencyStatusVersionUnsatisfied:
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, BlockerDependencyVersionUnsatisfied, check)
	}
}

// appendDependencyBlocker builds one blocker from a dependency check.
func appendDependencyBlocker(blockers []*Blocker, code BlockerCode, check *PluginDependencyCheck) []*Blocker {
	return append(blockers, &Blocker{
		Code:            code,
		PluginID:        check.OwnerID,
		DependencyID:    check.DependencyID,
		RequiredVersion: check.RequiredVersion,
		CurrentVersion:  check.CurrentVersion,
		Chain:           append([]string(nil), check.Chain...),
		Detail:          string(check.Status),
	})
}

// buildPluginMap normalizes plugin snapshots into a deterministic lookup map.
func buildPluginMap(plugins []*PluginSnapshot) map[string]*PluginSnapshot {
	result := make(map[string]*PluginSnapshot, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil || strings.TrimSpace(plugin.ID) == "" {
			continue
		}
		normalized := *plugin
		normalized.ID = strings.TrimSpace(plugin.ID)
		normalized.Name = strings.TrimSpace(plugin.Name)
		normalized.Version = strings.TrimSpace(plugin.Version)
		result[normalized.ID] = &normalized
	}
	return result
}

// pluginDependencies returns the effective dependency declaration for a snapshot.
func pluginDependencies(plugin *PluginSnapshot) *plugintypes.DependencySpec {
	if plugin == nil {
		return nil
	}
	if plugin.Dependencies != nil {
		return plugin.Dependencies
	}
	if plugin.Manifest != nil {
		return plugin.Manifest.Dependencies
	}
	return nil
}

// normalizedPluginDependencies returns sorted plugin dependency edges.
func normalizedPluginDependencies(spec *plugintypes.DependencySpec) []*plugintypes.PluginDependencySpec {
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

// sortedSnapshots returns plugin snapshots in plugin-ID order.
func sortedSnapshots(plugins []*PluginSnapshot) []*PluginSnapshot {
	sorted := make([]*PluginSnapshot, 0, len(plugins))
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

// sortInstallResult keeps check outputs deterministic across platforms.
func sortInstallResult(result *InstallCheckResult) {
	sort.SliceStable(result.Dependencies, func(i, j int) bool {
		return dependencyCheckSortKey(result.Dependencies[i]) < dependencyCheckSortKey(result.Dependencies[j])
	})
	sort.SliceStable(result.Blockers, func(i, j int) bool {
		return blockerSortKey(result.Blockers[i]) < blockerSortKey(result.Blockers[j])
	})
}

// dependencyCheckSortKey builds a stable key for dependency checks.
func dependencyCheckSortKey(check *PluginDependencyCheck) string {
	if check == nil {
		return ""
	}
	return strings.Join(check.Chain, "/") + "|" + check.OwnerID + "|" + check.DependencyID
}

// blockerSortKey builds a stable key for blockers.
func blockerSortKey(blocker *Blocker) string {
	if blocker == nil {
		return ""
	}
	return string(blocker.Code) + "|" + strings.Join(blocker.Chain, "/") + "|" + blocker.PluginID + "|" + blocker.DependencyID
}
