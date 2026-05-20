// This file implements the agents.skills.unlink command which removes
// managed directory symlinks pointing at .agents/skills. It never removes
// real directories or files and never touches symlinks pointing at
// foreign targets. When invoked on a TTY without an explicit agent
// argument, it offers an interactive selection of currently linked
// agents.

package main

import (
	"context"
	"fmt"

	"linactl/internal/agents/common"
	"linactl/internal/agents/skills"
)

// runAgentsSkillsUnlink dispatches agents.skills.unlink command invocations.
func runAgentsSkillsUnlink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := common.ParseSelectors(selectorRaw)

	if len(selectors) == 0 && common.IsInteractiveTerminal(stdinAsFile(a)) {
		return runAgentsSkillsUnlinkInteractive(a)
	}
	if len(selectors) == 0 {
		return fmt.Errorf("agent=<name|all|csv> is required for agents.skills.unlink in non-interactive mode")
	}

	return executeAgentsSkillsUnlink(a, selectors)
}

// runAgentsSkillsUnlinkInteractive walks the user through a numbered
// selection of currently managed links to remove.
func runAgentsSkillsUnlinkInteractive(a *app) error {
	candidates := skills.UnlinkCandidates(a.root)
	if len(candidates) == 0 {
		return writeLine(a.stdout, "No managed agent skill symlinks were found. Nothing to unlink.")
	}
	names, err := common.PromptSelection(a.stdin, a.stdout, "Select agents to unlink (skills):", candidates)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	return executeAgentsSkillsUnlink(a, names)
}

// executeAgentsSkillsUnlink applies the unlink request and renders results.
func executeAgentsSkillsUnlink(a *app, selectors []string) error {
	results, err := skills.ApplyUnlink(a.root, skills.UnlinkRequest{Selectors: selectors})
	if err != nil {
		return err
	}
	if err = common.Render(a.stdout, results); err != nil {
		return err
	}
	if err = writeLine(a.stdout, ""); err != nil {
		return err
	}
	if err = common.EmitHints(a.stdout, results); err != nil {
		return err
	}
	if common.HasError(results) {
		return fmt.Errorf("one or more agents failed; see DETAIL column")
	}
	return nil
}
