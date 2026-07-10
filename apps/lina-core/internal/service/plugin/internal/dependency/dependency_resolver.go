// This file implements deterministic, side-effect-free dependency resolution
// for plugin lifecycle planning.

package dependency

import (
	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"sort"
	"strings"
)

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
			Code:     pluginv1.BlockerCodeDependencyMissing,
			PluginID: targetID,
			Detail:   "target plugin snapshot is missing",
		})
		return result
	}

	result.Framework = r.checkFramework(target, input.FrameworkVersion)
	if result.Framework.Status == pluginv1.FrameworkStatusUnsatisfied {
		result.Blockers = append(result.Blockers, &Blocker{
			Code:            pluginv1.BlockerCodeFrameworkVersionUnsatisfied,
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
	index := input.ReverseIndex
	if index == nil {
		index = NewReverseDependencyIndex(input.Plugins)
	}
	entries := append([]*reverseDependencyEntry(nil), index.entriesByTarget[targetID]...)
	if result.CandidateVersion == "" && targetID != "" {
		entries = append(entries, index.wildcardUnknown...)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return reverseDependencyEntrySortKey(entries[i]) < reverseDependencyEntrySortKey(entries[j])
	})
	for _, entry := range entries {
		if entry == nil || entry.dependent == nil {
			continue
		}
		dependentID := strings.TrimSpace(entry.dependent.ID)
		if dependentID == targetID {
			continue
		}
		if entry.unknown {
			result.Blockers = append(result.Blockers, &Blocker{
				Code:     pluginv1.BlockerCodeDependencySnapshotUnknown,
				PluginID: dependentID,
				Detail:   "installed plugin dependency snapshot is unavailable",
			})
			continue
		}
		dependent := &ReverseDependent{
			PluginID:          dependentID,
			Name:              strings.TrimSpace(entry.dependent.Name),
			Version:           strings.TrimSpace(entry.dependent.Version),
			RequiredVersion:   strings.TrimSpace(entry.requiredVersion),
			OwnerHostServices: ownerHostServiceSummariesForOwner(entry.dependent.OwnerHostServices, targetID),
		}
		result.Dependents = append(result.Dependents, dependent)

		if result.CandidateVersion == "" {
			result.Blockers = append(result.Blockers, &Blocker{
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
			result.Blockers = append(result.Blockers, &Blocker{
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
	return result
}

// reverseDependencyEntry identifies one installed plugin dependency declaration
// or one conservative unknown-snapshot blocker relevant to a target plugin.
type reverseDependencyEntry struct {
	dependent       *PluginSnapshot
	requiredVersion string
	unknown         bool
}

// NewReverseDependencyIndex indexes installed plugin dependency declarations
// and conservative unknown-snapshot blockers in deterministic plugin-ID order.
func NewReverseDependencyIndex(plugins []*PluginSnapshot) *ReverseDependencyIndex {
	index := &ReverseDependencyIndex{
		entriesByTarget: make(map[string][]*reverseDependencyEntry),
	}
	for _, plugin := range sortedSnapshots(plugins) {
		if plugin == nil {
			continue
		}
		pluginID := strings.TrimSpace(plugin.ID)
		if !plugin.Installed || pluginID == "" {
			continue
		}
		if plugin.DependencySnapshotUnknown {
			indexUnknownReverseDependencies(index, plugin)
			continue
		}
		for _, declaredDependency := range normalizedPluginDependencies(plugin.Dependencies) {
			if declaredDependency == nil {
				continue
			}
			dependencyID := strings.TrimSpace(declaredDependency.ID)
			if dependencyID == "" || dependencyID == pluginID {
				continue
			}
			index.entriesByTarget[dependencyID] = append(index.entriesByTarget[dependencyID], &reverseDependencyEntry{
				dependent:       plugin,
				requiredVersion: strings.TrimSpace(declaredDependency.Version),
			})
		}
	}
	return index
}

// indexUnknownReverseDependencies preserves the previous fail-closed behavior
// for installed plugins whose effective dependency snapshot is unavailable.
func indexUnknownReverseDependencies(index *ReverseDependencyIndex, plugin *PluginSnapshot) {
	if index == nil || plugin == nil {
		return
	}
	dependencies := normalizedPluginDependencies(plugin.Dependencies)
	if len(dependencies) == 0 {
		if plugin.Manifest == nil {
			index.wildcardUnknown = append(index.wildcardUnknown, &reverseDependencyEntry{
				dependent: plugin,
				unknown:   true,
			})
		}
		return
	}
	for _, declaredDependency := range dependencies {
		if declaredDependency == nil {
			continue
		}
		dependencyID := strings.TrimSpace(declaredDependency.ID)
		if dependencyID == "" {
			continue
		}
		index.entriesByTarget[dependencyID] = append(index.entriesByTarget[dependencyID], &reverseDependencyEntry{
			dependent: plugin,
			unknown:   true,
		})
	}
}

// reverseDependencyEntrySortKey matches the previous all-snapshot traversal
// ordering without requiring each query to inspect unrelated dependencies.
func reverseDependencyEntrySortKey(entry *reverseDependencyEntry) string {
	if entry == nil || entry.dependent == nil {
		return ""
	}
	return strings.Join([]string{
		strings.TrimSpace(entry.dependent.ID),
		strings.TrimSpace(entry.requiredVersion),
	}, "|")
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
		Status:         pluginv1.FrameworkStatusNotDeclared,
	}
	dependencies := pluginDependencies(plugin)
	if dependencies == nil || dependencies.Framework == nil || strings.TrimSpace(dependencies.Framework.Version) == "" {
		return check
	}
	check.RequiredVersion = strings.TrimSpace(dependencies.Framework.Version)
	matches, err := plugintypes.MatchesSemanticVersionRange(check.CurrentVersion, check.RequiredVersion)
	if err != nil || !matches {
		check.Status = pluginv1.FrameworkStatusUnsatisfied
		return check
	}
	check.Status = pluginv1.FrameworkStatusSatisfied
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
			Code:         pluginv1.BlockerCodeDependencyCycle,
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
				Code:         pluginv1.BlockerCodeDependencyCycle,
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
		if check.Discovered && check.Status == pluginv1.DependencyStatusSatisfied {
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
		check.Status = pluginv1.DependencyStatusMissing
		return check
	}

	check.DependencyName = strings.TrimSpace(dependency.Name)
	check.CurrentVersion = strings.TrimSpace(dependency.Version)
	check.Installed = dependency.Installed
	check.Discovered = true
	if check.RequiredVersion != "" {
		matches, err := plugintypes.MatchesSemanticVersionRange(check.CurrentVersion, check.RequiredVersion)
		if err != nil || !matches {
			check.Status = pluginv1.DependencyStatusVersionUnsatisfied
			return check
		}
	}
	if dependency.Installed {
		check.Status = pluginv1.DependencyStatusSatisfied
		return check
	}
	check.Status = pluginv1.DependencyStatusMissing
	return check
}

// recordDependencyOutcome updates result collections for one dependency edge.
func (r *Resolver) recordDependencyOutcome(state walkState, check *PluginDependencyCheck) {
	switch check.Status {
	case pluginv1.DependencyStatusSatisfied:
		return
	case pluginv1.DependencyStatusMissing:
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, pluginv1.BlockerCodeDependencyMissing, check)
	case pluginv1.DependencyStatusVersionUnsatisfied:
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, pluginv1.BlockerCodeDependencyVersionUnsatisfied, check)
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
