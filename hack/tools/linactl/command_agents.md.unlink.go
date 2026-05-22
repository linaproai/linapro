// This file implements the agents.md.unlink command which removes
// managed single-file symlinks pointing at the repo-root AGENTS.md. It
// never removes real files and never touches symlinks pointing at
// foreign targets.

package main

import (
	"context"
	"fmt"

	"linactl/internal/agents/common"
	"linactl/internal/agents/md"
)

// runAgentsMdUnlink dispatches agents.md.unlink command invocations.
func runAgentsMdUnlink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := common.ParseSelectors(selectorRaw)

	if len(selectors) == 0 && common.IsInteractiveTerminal(stdinAsFile(a)) {
		return runAgentsMdUnlinkInteractive(a)
	}
	if len(selectors) == 0 {
		return fmt.Errorf("agent=<name|all|csv> is required for agents.md.unlink in non-interactive mode")
	}

	return executeAgentsMdUnlink(a, selectors)
}

// runAgentsMdUnlinkInteractive walks the user through a numbered
// selection of currently managed links to remove.
func runAgentsMdUnlinkInteractive(a *app) error {
	candidates := md.UnlinkCandidates(a.root)
	if len(candidates) == 0 {
		return writeLine(a.stdout, "No managed AGENTS.md symlinks were found. Nothing to unlink.")
	}
	names, err := common.PromptSelection(a.stdin, a.stdout, "Select agents to unlink (md):", candidates)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	return executeAgentsMdUnlink(a, names)
}

// executeAgentsMdUnlink applies the unlink request and renders results.
func executeAgentsMdUnlink(a *app, selectors []string) error {
	results, err := md.ApplyUnlink(a.root, md.UnlinkRequest{Selectors: selectors})
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
