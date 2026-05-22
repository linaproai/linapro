// This file implements the agents aggregate command. The command takes
// an agent-first stance: users either pass `agent=<name>` for a one-shot
// non-interactive setup, or run from a TTY to walk a two-step menu
// (pick agent -> pick link/unlink) that automatically applies to every
// resource type (skills / prompts / md) the chosen agent participates
// in. Resources where the agent is native or unregistered are skipped
// with an explicit reason in the final summary.
//
// This file also owns the writeLine / writeLines stdout helpers shared
// by every agents.* command, and the stdinAsFile helper used by the
// resource subcommands' TTY-detection branches.
//
// The previous three-level menu (resource -> action -> agent) and the
// `agents.<resource>.<action>` six-target Makefile expansion remain
// available as advanced entry points; only the aggregate `agents`
// command is reshaped here.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"linactl/internal/agents/common"
	"linactl/internal/agents/md"
	"linactl/internal/agents/prompts"
	"linactl/internal/agents/skills"
)

// agentSetupAction enumerates the actions the aggregate command can
// dispatch. Only link and unlink are supported today; the type exists
// so the dispatcher can be extended without leaking string literals
// into business logic.
type agentSetupAction string

const (
	actionLink   agentSetupAction = "link"
	actionUnlink agentSetupAction = "unlink"
)

// resourceKind tags a resource entry inside the cross-resource agent
// universe with a stable identifier matching the registry it came from.
type resourceKind string

const (
	resourceSkills  resourceKind = "skills"
	resourcePrompts resourceKind = "prompts"
	resourceMd      resourceKind = "md"
)

// selectableAgent describes one agent's role across the three resource
// registries (skills / prompts / md). It is built once at command entry
// by collectAgentUniverse and consumed by both the interactive picker
// (huh option labels) and the dispatch loop (which resources to act on).
type selectableAgent struct {
	// Name is the canonical agent identifier shared across registries.
	Name string
	// DisplayName is the human-readable label aggregated from whichever
	// registry first carries it.
	DisplayName string
	// Roles records the agent's category in each resource it appears
	// in. A resource not present in the map means the agent is not
	// registered there.
	Roles map[resourceKind]common.Category
}

// optionLabel returns the compact label shown in the aggregate TTY agent
// picker. The picker intentionally hides internal IDs, resource paths and
// status summaries; those remain available in the per-resource commands.
func (s selectableAgent) optionLabel() string {
	if label := strings.TrimSpace(s.DisplayName); label != "" {
		return label
	}
	return s.Name
}

// hasLinkRole reports whether the agent has at least one resource where
// it is link-class. Only such agents are eligible for interactive
// selection (native-only agents have nothing to link/unlink).
func (s selectableAgent) hasLinkRole() bool {
	for _, category := range s.Roles {
		if category == common.CategoryLink {
			return true
		}
	}
	return false
}

// writeLine prints a single line to the writer, wrapping any write error
// so linactl never silently drops stdout failures (e.g. a broken pipe).
// The helper is shared by every agents.* command.
func writeLine(out io.Writer, line string) error {
	if _, err := fmt.Fprintln(out, line); err != nil {
		return fmt.Errorf("write output line: %w", err)
	}
	return nil
}

// writeLines prints multiple lines, returning the first write error.
func writeLines(out io.Writer, lines ...string) error {
	for _, line := range lines {
		if err := writeLine(out, line); err != nil {
			return err
		}
	}
	return nil
}

// stdinAsFile returns the *os.File backing the app's stdin when available,
// or nil when the test harness wired an in-memory reader. It is shared by
// every agents.* command's TTY-detection branch.
func stdinAsFile(a *app) *os.File {
	if file, ok := a.stdin.(*os.File); ok {
		return file
	}
	return nil
}

// runAgents dispatches the aggregate agents command. Behaviour:
//
//   - agent=<name> [action=link|unlink] [force=1] : one-shot setup; the
//     selected action runs against every resource type the agent is
//     link-class in.
//   - no agent + TTY                              : two-step menu (pick
//     agent -> pick action) backed by huh.
//   - no agent + non-TTY                          : print usage and
//     return successfully.
//
// Unknown agents, the literal string "all", and comma-separated lists
// are explicitly rejected to keep the one-shot path safe; multi-agent
// batch flows remain available through the per-resource
// `agents.<resource>.<action>` subcommands.
func runAgents(_ context.Context, a *app, input commandInput) error {
	rawAgent := strings.TrimSpace(input.GetDefault("agent", ""))
	rawAction := strings.TrimSpace(strings.ToLower(input.GetDefault("action", "")))
	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}

	universe := collectAgentUniverse(a.root)

	if rawAgent == "" {
		if !common.IsInteractiveTerminal(stdinAsFile(a)) {
			return printAgentsUsage(a.stdout)
		}
		return runAgentInteractiveMenu(a, universe, force)
	}

	agentName, err := validateSingleAgentName(rawAgent, universe)
	if err != nil {
		return err
	}
	action, err := parseAgentSetupAction(rawAction, actionLink)
	if err != nil {
		return err
	}
	return dispatchAgentSetup(a, agentName, action, force, universe)
}

// printAgentsUsage emits the non-interactive usage hint pointing
// callers at the one-shot mode and the advanced subcommands.
func printAgentsUsage(out io.Writer) error {
	return writeLines(out,
		"Usage: linactl agents [agent=<name>] [action=link|unlink] [force=1]",
		"",
		"One-shot mode (works in any environment):",
		"  make agents agent=<name> [action=link|unlink] [force=1]",
		"  - agent must name a single supported agent (no 'all', no csv).",
		"  - action defaults to 'link'.",
		"  - Upper-case AGENT/ACTION/FORCE aliases remain accepted for compatibility.",
		"  - The selected action runs against every resource type the agent supports.",
		"",
		"Interactive mode (TTY only):",
		"  make agents",
		"  - Step 1: arrow-key pick the agent.",
		"  - Step 2: arrow-key pick link or unlink.",
		"",
		"Advanced per-resource entry points (still available):",
		"  make agents.skills.link  | agents.skills.unlink",
		"  make agents.prompts.link | agents.prompts.unlink",
		"  make agents.md.link      | agents.md.unlink",
	)
}

// parseAgentSetupAction normalizes an action string (link/unlink). An
// empty value falls back to fallback. Any other value yields an error
// so typos are caught at the CLI boundary.
func parseAgentSetupAction(raw string, fallback agentSetupAction) (agentSetupAction, error) {
	switch raw {
	case "":
		return fallback, nil
	case string(actionLink):
		return actionLink, nil
	case string(actionUnlink):
		return actionUnlink, nil
	default:
		return "", fmt.Errorf("invalid action %q: expected 'link' or 'unlink'", raw)
	}
}

// validateSingleAgentName enforces the one-shot mode contract: exactly
// one supported agent name, no "all", no comma list. The candidate
// listing in error messages is limited to the link-class universe so
// users see only agents the aggregate command can actually drive.
func validateSingleAgentName(raw string, universe []selectableAgent) (string, error) {
	normalized := common.NormalizeAgentName(raw)
	if normalized == "" {
		return "", fmt.Errorf("agent= must be set; pass a single supported agent name")
	}
	if strings.Contains(raw, ",") {
		return "", fmt.Errorf("agent=%q: comma-separated lists are not supported by `linactl agents`; use the per-resource subcommands for batch operations", raw)
	}
	if normalized == common.SelectorAll {
		return "", fmt.Errorf("agent=all is not supported by `linactl agents` (safety guard); pass a specific agent name")
	}
	for _, candidate := range universe {
		if candidate.Name == normalized {
			return candidate.Name, nil
		}
	}
	return "", fmt.Errorf("unknown agent %q; supported agents: %s", raw, joinAgentNames(universe))
}

// joinAgentNames flattens the universe slice into a comma-separated
// candidate listing for error messages.
func joinAgentNames(universe []selectableAgent) string {
	names := make([]string, 0, len(universe))
	for _, agent := range universe {
		names = append(names, agent.Name)
	}
	return strings.Join(names, ", ")
}

// agentDisplayPriority lists the most commonly used agents in the order
// they should appear at the top of the interactive picker. The first
// entry takes precedence over the second, and so on. Agents not present
// in this list fall through to alphabetical order. The list reflects
// broader community/usage signals so the picker surfaces the agents
// most users actually need without scrolling; project-internal entries
// (e.g. codebuddy) intentionally fall back to the alphabetical tail to
// avoid biasing the default surface. Names that are not currently
// registered (e.g. cline) are kept in the priority list intentionally —
// they cost nothing when absent and keep the ordering stable when new
// registrations land.
var agentDisplayPriority = []string{
	"claude-code",
	"codex",
	"cursor",
	"gemini-cli",
	"windsurf",
	"qwen-code",
	"continue",
	"cline",
	"trae",
	"roo",
	"kiro-cli",
	"aider-desk",
	"augment",
}

// agentPriorityRank returns the configured display rank for the given
// agent name. Agents not present in agentDisplayPriority receive a rank
// strictly larger than any priority entry so they sort after every
// priority agent. The boolean reports whether the name was found inside
// the priority list, which lets callers distinguish "alphabetical tail"
// from "explicit priority" without relying on sentinel values.
func agentPriorityRank(name string) (int, bool) {
	for index, candidate := range agentDisplayPriority {
		if candidate == name {
			return index, true
		}
	}
	return len(agentDisplayPriority), false
}

// collectAgentUniverse merges the three resource registries and returns
// every agent that is link-class in at least one resource. Native-only
// agents are excluded because the aggregate command never has work to
// do on them. The returned slice is sorted so that agents listed in
// agentDisplayPriority appear first in the configured order, with the
// remaining agents falling back to alphabetical order for stable
// output.
//
// The repoRoot parameter is intentionally unused. It remains in the
// signature because callers already pass it and future registries may need
// repository context, but the aggregate picker does not inspect runtime
// link state while building its compact labels.
func collectAgentUniverse(_ string) []selectableAgent {
	universe := make(map[string]*selectableAgent)

	upsert := func(name, display string, kind resourceKind, category common.Category) {
		entry, exists := universe[name]
		if !exists {
			entry = &selectableAgent{
				Name:        name,
				DisplayName: display,
				Roles:       map[resourceKind]common.Category{},
			}
			universe[name] = entry
		}
		if entry.DisplayName == "" {
			entry.DisplayName = display
		}
		entry.Roles[kind] = category
	}

	for _, spec := range skills.Agents() {
		upsert(spec.Name, spec.DisplayName, resourceSkills, spec.Category)
	}
	for _, spec := range prompts.Agents() {
		upsert(spec.Name, spec.DisplayName, resourcePrompts, spec.Category)
	}
	for _, spec := range md.Agents() {
		upsert(spec.Name, spec.DisplayName, resourceMd, spec.Category)
	}

	out := make([]selectableAgent, 0, len(universe))
	for _, entry := range universe {
		if !entry.hasLinkRole() {
			continue
		}
		out = append(out, *entry)
	}
	sort.Slice(out, func(left, right int) bool {
		leftRank, leftPriority := agentPriorityRank(out[left].Name)
		rightRank, rightPriority := agentPriorityRank(out[right].Name)
		switch {
		case leftPriority && rightPriority:
			return leftRank < rightRank
		case leftPriority:
			return true
		case rightPriority:
			return false
		default:
			return out[left].Name < out[right].Name
		}
	})
	return out
}

// runAgentInteractiveMenu drives the two-step TTY flow:
//  1. select one agent from the link-class universe (huh single-select);
//  2. select link or unlink (huh single-select).
//
// Cancellation at either step returns nil after printing "Cancelled.".
func runAgentInteractiveMenu(a *app, universe []selectableAgent, force bool) error {
	if len(universe) == 0 {
		return writeLine(a.stdout, "No link-class agents are registered; nothing to set up.")
	}

	options := make([]common.SingleOption, 0, len(universe))
	for _, agent := range universe {
		options = append(options, common.SingleOption{Value: agent.Name, Label: agent.optionLabel()})
	}

	agentName, err := common.PromptSingleSelection(a.stdin, a.stdout, "Select an agent to configure:", options)
	if err != nil {
		return err
	}
	if agentName == "" {
		return writeLine(a.stdout, "Cancelled.")
	}

	actionChoice, err := common.PromptSingleSelection(a.stdin, a.stdout,
		fmt.Sprintf("What should we do for %s?", agentName), []common.SingleOption{
			{Value: string(actionLink), Label: "link    Create or rebuild symlinks for this agent"},
			{Value: string(actionUnlink), Label: "unlink  Remove managed symlinks for this agent"},
		})
	if err != nil {
		return err
	}
	if actionChoice == "" {
		return writeLine(a.stdout, "Cancelled.")
	}

	action, err := parseAgentSetupAction(actionChoice, actionLink)
	if err != nil {
		return err
	}
	return dispatchAgentSetup(a, agentName, action, force, universe)
}

// dispatchAgentSetup executes the chosen action across every resource
// type the agent participates in and renders a compact resource-level
// summary. Per-resource commands still own the verbose path/category
// tables; the aggregate command stays focused on the high-level outcome
// for the selected agent.
func dispatchAgentSetup(a *app, agentName string, action agentSetupAction, force bool, universe []selectableAgent) error {
	target, ok := lookupAgent(universe, agentName)
	if !ok {
		// Should not happen: validateSingleAgentName already rejected
		// unknown names. Defensive guard so callers that bypass the
		// validator (future internal callers) still get a clear error.
		return fmt.Errorf("agent %q not found in the cross-resource registry", agentName)
	}

	if err := writeLines(a.stdout,
		fmt.Sprintf("Agent: %s", target.optionLabel()),
		fmt.Sprintf("Action: %s", action),
		"",
	); err != nil {
		return err
	}

	outcomes := make([]aggregateResourceOutcome, 0, 3)
	var firstErr error

	for _, kind := range []resourceKind{resourceSkills, resourcePrompts, resourceMd} {
		category, present := target.Roles[kind]
		if !present {
			outcomes = append(outcomes, aggregateResourceOutcome{
				kind:   kind,
				status: aggregateStatusSkipped,
				detail: "not registered",
			})
			continue
		}
		if category != common.CategoryLink {
			outcomes = append(outcomes, aggregateResourceOutcome{
				kind:   kind,
				status: aggregateStatusSkipped,
				detail: fmt.Sprintf("%s (no symlink work)", category),
			})
			continue
		}
		results, err := runAggregateResourceAction(a, kind, agentName, action, force)
		if err != nil {
			outcomes = append(outcomes, aggregateResourceOutcome{
				kind:   kind,
				status: aggregateStatusFailed,
				detail: err.Error(),
			})
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		outcome := aggregateOutcomeFromResults(kind, results)
		outcomes = append(outcomes, outcome)
		if outcome.status == aggregateStatusFailed && firstErr == nil {
			firstErr = fmt.Errorf("one or more agents failed; see result table")
		}
	}

	if err := renderAggregateOutcomes(a.stdout, outcomes); err != nil {
		return err
	}

	return firstErr
}

// aggregateResourceOutcome is one row in the aggregate agents result
// table. status intentionally uses coarse values so the table stays
// scannable; the original per-resource Status is included in detail.
type aggregateResourceOutcome struct {
	kind   resourceKind
	status string
	detail string
}

const (
	aggregateStatusApplied = "applied"
	aggregateStatusSkipped = "skipped"
	aggregateStatusFailed  = "failed"
)

// runAggregateResourceAction executes one resource without rendering the
// verbose per-resource table. The aggregate command renders its own
// compact table after all resources have been processed.
func runAggregateResourceAction(a *app, kind resourceKind, agentName string, action agentSetupAction, force bool) ([]common.Result, error) {
	selectors := []string{agentName}
	switch kind {
	case resourceSkills:
		if action == actionLink {
			return skills.ApplyLink(a.root, skills.LinkRequest{Selectors: selectors, Force: force})
		}
		return skills.ApplyUnlink(a.root, skills.UnlinkRequest{Selectors: selectors})
	case resourcePrompts:
		if action == actionLink {
			return prompts.ApplyLink(a.root, prompts.LinkRequest{Selectors: selectors, Force: force})
		}
		return prompts.ApplyUnlink(a.root, prompts.UnlinkRequest{Selectors: selectors})
	case resourceMd:
		if action == actionLink {
			return md.ApplyLink(a.root, md.LinkRequest{Selectors: selectors, Force: force})
		}
		return md.ApplyUnlink(a.root, md.UnlinkRequest{Selectors: selectors})
	default:
		return nil, fmt.Errorf("unsupported resource %q", kind)
	}
}

// aggregateOutcomeFromResults converts a per-resource result list into
// one coarse summary row. The aggregate command always resolves a single
// agent, but the loop keeps the function robust if a future caller passes
// more than one result.
func aggregateOutcomeFromResults(kind resourceKind, results []common.Result) aggregateResourceOutcome {
	if len(results) == 0 {
		return aggregateResourceOutcome{kind: kind, status: aggregateStatusSkipped, detail: "no result"}
	}
	status := aggregateStatusApplied
	parts := make([]string, 0, len(results))
	for _, result := range results {
		if result.Status == common.StatusError {
			status = aggregateStatusFailed
		} else if status != aggregateStatusFailed && !aggregateResultApplied(result.Status) {
			status = aggregateStatusSkipped
		}
		detail := string(result.Status)
		if trimmed := strings.TrimSpace(result.Detail); trimmed != "" {
			detail = fmt.Sprintf("%s: %s", result.Status, trimmed)
		}
		parts = append(parts, detail)
	}
	return aggregateResourceOutcome{
		kind:   kind,
		status: status,
		detail: strings.Join(parts, "; "),
	}
}

// aggregateResultApplied reports whether a per-resource status represents
// successful application or an already-satisfied binding.
func aggregateResultApplied(status common.Status) bool {
	switch status {
	case common.StatusOK, common.StatusCreated, common.StatusRebuilt, common.StatusRemoved:
		return true
	default:
		return false
	}
}

// renderAggregateOutcomes writes the compact aggregate result table.
func renderAggregateOutcomes(out io.Writer, outcomes []aggregateResourceOutcome) error {
	const (
		columnResource = "RESOURCE"
		columnStatus   = "STATUS"
		columnDetail   = "DETAIL"
	)
	maxResource := len(columnResource)
	maxStatus := len(columnStatus)
	for _, outcome := range outcomes {
		if width := len(string(outcome.kind)); width > maxResource {
			maxResource = width
		}
		if width := len(outcome.status); width > maxStatus {
			maxStatus = width
		}
	}
	if _, err := fmt.Fprintf(out, "%-*s  %-*s  %s\n", maxResource, columnResource, maxStatus, columnStatus, columnDetail); err != nil {
		return fmt.Errorf("write aggregate header: %w", err)
	}
	for _, outcome := range outcomes {
		if _, err := fmt.Fprintf(out, "%-*s  %-*s  %s\n", maxResource, outcome.kind, maxStatus, outcome.status, outcome.detail); err != nil {
			return fmt.Errorf("write aggregate row: %w", err)
		}
	}
	return nil
}

// lookupAgent finds a selectableAgent by name in the universe slice.
// Used by dispatchAgentSetup to resolve role information once name
// validation has succeeded.
func lookupAgent(universe []selectableAgent, name string) (selectableAgent, bool) {
	for _, agent := range universe {
		if agent.Name == name {
			return agent, true
		}
	}
	return selectableAgent{}, false
}
