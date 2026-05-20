// This file implements the skills resource thin wrappers around the
// common engine: Inspect/PlanList/ApplyLink/ApplyUnlink dispatch to
// common functions while constraining selector resolution to the skills
// agent registry. Selector parsing, target rendering, hint emission and
// interactive helpers are re-exported aliases of the common implementations
// so callers can keep importing this subpackage without churn.

package skills

import (
	"errors"

	"linactl/internal/agents/common"
)

// Status aliases common.Status. New code should reference common.Status
// directly; the alias exists so existing call sites keep working.
type Status = common.Status

// Re-exported status constants. Same alias rationale as Status above.
const (
	StatusNative               = common.StatusNative
	StatusOK                   = common.StatusOK
	StatusCreated              = common.StatusCreated
	StatusRebuilt              = common.StatusRebuilt
	StatusMismatch             = common.StatusMismatch
	StatusConflict             = common.StatusConflict
	StatusSkippedRootCollision = common.StatusSkippedRootCollision
	StatusRemoved              = common.StatusRemoved
	StatusSkippedForeignTarget = common.StatusSkippedForeignTarget
	StatusSkippedNotManaged    = common.StatusSkippedNotManaged
	StatusAbsent               = common.StatusAbsent
	StatusError                = common.StatusError
)

// Result aliases common.Result.
type Result = common.Result

// SelectableEntry aliases common.SelectableEntry.
type SelectableEntry = common.SelectableEntry

// SelectorAll aliases common.SelectorAll.
const SelectorAll = common.SelectorAll

// LinkRequest captures one skills.link invocation parameters.
type LinkRequest struct {
	// Selectors is the list of agent names provided by the caller. An
	// empty list means "no selection" and the command should print status
	// only. A list containing "all" expands to every link-class agent.
	Selectors []string
	// Force enables rebuilding mismatched symlinks and creating
	// rootCollision agents. It never affects real directories or files.
	Force bool
}

// UnlinkRequest captures one skills.unlink invocation parameters.
type UnlinkRequest struct {
	// Selectors mirrors LinkRequest.Selectors but applies to unlink flow.
	// Empty selectors means "no selection" and the command should refuse
	// to perform any change.
	Selectors []string
}

// Inspect returns the current Status and Detail for an agent without any
// filesystem mutation. It is used by the default no-selector listing flow.
func Inspect(repoRoot string, spec AgentSpec) Result {
	return common.Inspect(repoRoot, spec)
}

// PlanList returns inspection results for every agent in the registry.
func PlanList(repoRoot string) []Result {
	out := make([]Result, 0, len(agents))
	for _, spec := range agents {
		out = append(out, common.Inspect(repoRoot, spec))
	}
	return out
}

// ApplyLink executes the link request and returns one Result per resolved
// target. native agents are reported via StatusNative and skipped from any
// filesystem mutation.
func ApplyLink(repoRoot string, request LinkRequest) ([]Result, error) {
	if len(request.Selectors) == 0 {
		return nil, errors.New("no agent selected; pass agent=<name|all|csv>")
	}
	policy := common.TargetPolicy{
		IncludeNative:        true,
		IncludeRootCollision: request.Force,
	}
	targets, err := common.ResolveTargets(request.Selectors, agents, policy)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.New("no agent selected")
	}
	results := make([]Result, 0, len(targets))
	for _, spec := range targets {
		results = append(results, common.ApplyOneLink(repoRoot, spec, request.Force))
	}
	return results, nil
}

// ApplyUnlink executes the unlink request and returns one Result per
// resolved target.
func ApplyUnlink(repoRoot string, request UnlinkRequest) ([]Result, error) {
	if len(request.Selectors) == 0 {
		return nil, errors.New("no agent selected; pass agent=<name|all|csv>")
	}
	// Unlink never implicitly touches native or rootCollision paths.
	policy := common.TargetPolicy{
		IncludeNative:        false,
		IncludeRootCollision: false,
	}
	targets, err := common.ResolveTargets(request.Selectors, agents, policy)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.New("no agent selected")
	}
	results := make([]Result, 0, len(targets))
	for _, spec := range targets {
		results = append(results, common.ApplyOneUnlink(repoRoot, spec))
	}
	return results, nil
}

// LinkCandidates returns selectable entries for skills.link interactive
// mode. native agents are excluded because they require no action;
// rootCollision agents are excluded because they require explicit FORCE=1.
func LinkCandidates(repoRoot string) []SelectableEntry {
	out := make([]SelectableEntry, 0)
	for _, spec := range agents {
		if spec.Category != common.CategoryLink {
			continue
		}
		result := common.Inspect(repoRoot, spec)
		out = append(out, SelectableEntry{
			Spec:          spec,
			CurrentStatus: result.Status,
			Detail:        result.Detail,
		})
	}
	return out
}

// UnlinkCandidates returns selectable entries for skills.unlink
// interactive mode. Only agents whose project path is currently a managed
// symlink (i.e. pointing at SourceDir) are returned.
func UnlinkCandidates(repoRoot string) []SelectableEntry {
	out := make([]SelectableEntry, 0)
	for _, spec := range agents {
		if spec.Category == common.CategoryNative {
			continue
		}
		result := common.Inspect(repoRoot, spec)
		if result.Status != common.StatusOK {
			continue
		}
		out = append(out, SelectableEntry{
			Spec:          spec,
			CurrentStatus: result.Status,
			Detail:        result.Detail,
		})
	}
	return out
}

// ParseSelectors aliases common.ParseSelectors so existing call sites keep
// working.
func ParseSelectors(value string) []string {
	return common.ParseSelectors(value)
}
