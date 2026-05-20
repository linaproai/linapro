// This file implements the agents.prompts.unlink command which removes
// managed directory symlinks pointing at canonical .agents/prompts/...
// source directories. It never removes real directories or files and
// never touches symlinks pointing at foreign targets.

package main

import (
	"context"
	"fmt"

	"linactl/internal/agents/common"
	"linactl/internal/agents/prompts"
)

// runAgentsPromptsUnlink dispatches agents.prompts.unlink command
// invocations.
func runAgentsPromptsUnlink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := common.ParseSelectors(selectorRaw)

	if len(selectors) == 0 && common.IsInteractiveTerminal(stdinAsFile(a)) {
		return runAgentsPromptsUnlinkInteractive(a)
	}
	if len(selectors) == 0 {
		return fmt.Errorf("agent=<name|all|csv> is required for agents.prompts.unlink in non-interactive mode")
	}

	return executeAgentsPromptsUnlink(a, selectors)
}

// runAgentsPromptsUnlinkInteractive walks the user through a numbered
// selection of currently managed links to remove.
func runAgentsPromptsUnlinkInteractive(a *app) error {
	candidates := prompts.UnlinkCandidates(a.root)
	if len(candidates) == 0 {
		return writeLine(a.stdout, "No managed prompts symlinks were found. Nothing to unlink.")
	}
	names, err := common.PromptSelection(a.stdin, a.stdout, "Select agents to unlink (prompts):", candidates)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	return executeAgentsPromptsUnlink(a, names)
}

// executeAgentsPromptsUnlink applies the unlink request and renders
// results.
func executeAgentsPromptsUnlink(a *app, selectors []string) error {
	results, err := prompts.ApplyUnlink(a.root, prompts.UnlinkRequest{Selectors: selectors})
	if err != nil {
		return err
	}
	if err = common.Render(a.stdout, results, promptsExtraColumns()...); err != nil {
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
