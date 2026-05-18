// This file implements deterministic, side-effect-free dependency resolution
// for plugin lifecycle planning.

package dependency

import (
	"sort"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
)

// Resolver evaluates plugin dependency declarations against discovered and
// installed plugin snapshots.
type Resolver struct{}

// New creates a dependency resolver.
func New() *Resolver {
	return &Resolver{}
}

// CheckInstall evaluates whether target can be installed and which automatic
// dependency installs must run first.
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
	autoPlanSeen := map[string]bool{}
	r.walkDependencies(walkState{
		owner:        target,
		plugins:      plugins,
		chain:        []string{targetID},
		visited:      visited,
		visiting:     visiting,
		autoPlanSeen: autoPlanSeen,
		result:       result,
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
			result.Blockers = append(result.Blockers, &Blocker{
				Code:     BlockerDependencySnapshotUnknown,
				PluginID: strings.TrimSpace(plugin.ID),
				Detail:   "installed plugin dependency snapshot is unavailable",
			})
			continue
		}
		for _, declaredDependency := range normalizedPluginDependencies(plugin.Dependencies) {
			if declaredDependency == nil || strings.TrimSpace(declaredDependency.ID) != targetID {
				continue
			}
			if declaredDependency.Required != nil && !*declaredDependency.Required {
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
			matches, err := catalog.MatchesSemanticVersionRange(result.CandidateVersion, dependent.RequiredVersion)
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

// walkState carries mutable traversal state for install dependency resolution.
type walkState struct {
	owner        *PluginSnapshot
	plugins      map[string]*PluginSnapshot
	chain        []string
	visited      map[string]bool
	visiting     map[string]int
	autoPlanSeen map[string]bool
	result       *InstallCheckResult
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
	matches, err := catalog.MatchesSemanticVersionRange(check.CurrentVersion, check.RequiredVersion)
	if err != nil || !matches {
		check.Status = FrameworkStatusUnsatisfied
		return check
	}
	check.Status = FrameworkStatusSatisfied
	return check
}

// walkDependencies traverses hard dependency edges in deterministic order while
// collecting soft-dependency notices and automatic install plans.
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
		check := r.evaluateDependency(state.owner, declaredDependency, state.plugins, state.chain)
		state.result.Dependencies = append(state.result.Dependencies, check)
		if check == nil {
			continue
		}
		r.recordDependencyOutcome(state, check)
		if check.Required && check.Discovered &&
			(check.Status == DependencyStatusSatisfied || check.Status == DependencyStatusAutoInstallPlanned) {
			next := state.plugins[check.DependencyID]
			if next != nil {
				nextChain := append(append([]string(nil), state.chain...), check.DependencyID)
				r.walkDependencies(walkState{
					owner:        next,
					plugins:      state.plugins,
					chain:        nextChain,
					visited:      state.visited,
					visiting:     state.visiting,
					autoPlanSeen: state.autoPlanSeen,
					result:       state.result,
				})
			}
		}
		if check.Status == DependencyStatusAutoInstallPlanned {
			appendAutoInstallPlan(state, check)
		}
	}
	delete(state.visiting, ownerID)
	state.visited[ownerID] = true
}

// evaluateDependency classifies one declared dependency edge.
func (r *Resolver) evaluateDependency(
	owner *PluginSnapshot,
	declaredDependency *catalog.PluginDependencySpec,
	plugins map[string]*PluginSnapshot,
	chain []string,
) *PluginDependencyCheck {
	required := true
	if declaredDependency.Required != nil {
		required = *declaredDependency.Required
	}
	installMode := catalog.NormalizeDependencyInstallMode(declaredDependency.Install)
	dependencyID := strings.TrimSpace(declaredDependency.ID)
	dependency := plugins[dependencyID]
	check := &PluginDependencyCheck{
		OwnerID:         strings.TrimSpace(owner.ID),
		DependencyID:    dependencyID,
		RequiredVersion: strings.TrimSpace(declaredDependency.Version),
		Required:        required,
		InstallMode:     installMode,
		Chain:           append(append([]string(nil), chain...), dependencyID),
	}
	if dependency == nil {
		check.Status = DependencyStatusMissing
		if !required {
			check.Status = DependencyStatusSoftUnsatisfied
		}
		return check
	}

	check.DependencyName = strings.TrimSpace(dependency.Name)
	check.CurrentVersion = strings.TrimSpace(dependency.Version)
	check.Installed = dependency.Installed
	check.Discovered = true
	if check.RequiredVersion != "" {
		matches, err := catalog.MatchesSemanticVersionRange(check.CurrentVersion, check.RequiredVersion)
		if err != nil || !matches {
			check.Status = DependencyStatusVersionUnsatisfied
			if !required {
				check.Status = DependencyStatusSoftUnsatisfied
			}
			return check
		}
	}
	if dependency.Installed {
		check.Status = DependencyStatusSatisfied
		return check
	}
	if !required {
		check.Status = DependencyStatusSoftUnsatisfied
		return check
	}
	if installMode == catalog.DependencyInstallModeAuto {
		check.Status = DependencyStatusAutoInstallPlanned
		return check
	}
	check.Status = DependencyStatusManualInstallRequired
	return check
}

// recordDependencyOutcome updates result collections for one dependency edge.
func (r *Resolver) recordDependencyOutcome(state walkState, check *PluginDependencyCheck) {
	switch check.Status {
	case DependencyStatusSatisfied:
		return
	case DependencyStatusAutoInstallPlanned:
		return
	case DependencyStatusManualInstallRequired:
		state.result.ManualInstallRequired = append(state.result.ManualInstallRequired, check)
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, BlockerDependencyManualInstallRequired, check)
	case DependencyStatusMissing:
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, BlockerDependencyMissing, check)
	case DependencyStatusVersionUnsatisfied:
		state.result.Blockers = appendDependencyBlocker(state.result.Blockers, BlockerDependencyVersionUnsatisfied, check)
	case DependencyStatusSoftUnsatisfied:
		state.result.SoftUnsatisfied = append(state.result.SoftUnsatisfied, check)
	}
}

// appendAutoInstallPlan adds an auto-install item after traversing its own
// dependencies, producing dependency-first topological order.
func appendAutoInstallPlan(state walkState, check *PluginDependencyCheck) {
	if state.autoPlanSeen[check.DependencyID] || check.DependencyID == state.result.TargetID {
		return
	}
	state.autoPlanSeen[check.DependencyID] = true
	state.result.AutoInstallPlan = append(state.result.AutoInstallPlan, &AutoInstallPlanItem{
		PluginID:   check.DependencyID,
		Name:       check.DependencyName,
		Version:    check.CurrentVersion,
		RequiredBy: check.OwnerID,
		Chain:      append([]string(nil), check.Chain...),
	})
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
func pluginDependencies(plugin *PluginSnapshot) *catalog.DependencySpec {
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
func normalizedPluginDependencies(spec *catalog.DependencySpec) []*catalog.PluginDependencySpec {
	if spec == nil || len(spec.Plugins) == 0 {
		return nil
	}
	normalized := catalog.CloneDependencySpec(spec)
	catalog.NormalizeDependencySpec(normalized)
	dependencies := make([]*catalog.PluginDependencySpec, 0, len(normalized.Plugins))
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
	sort.SliceStable(result.ManualInstallRequired, func(i, j int) bool {
		return dependencyCheckSortKey(result.ManualInstallRequired[i]) < dependencyCheckSortKey(result.ManualInstallRequired[j])
	})
	sort.SliceStable(result.SoftUnsatisfied, func(i, j int) bool {
		return dependencyCheckSortKey(result.SoftUnsatisfied[i]) < dependencyCheckSortKey(result.SoftUnsatisfied[j])
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
