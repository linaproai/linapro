// This file implements the agents.prompts.link command which manages
// repository-local directory symlinks from supported agents' project
// commands/prompts roots to .agents/prompts. The status table includes a
// SOURCE column because each agent declares its own source path.

package main

import (
	"context"
	"fmt"

	"linactl/internal/agents/common"
	"linactl/internal/agents/prompts"
)

// promptsExtraColumns returns the extra column specs used by the prompts
// resource: a SOURCE column showing the per-agent source path so users
// can verify which source each binding points at.
func promptsExtraColumns() []common.ColumnSpec {
	return []common.ColumnSpec{
		{
			Header: "SOURCE",
			Value:  func(r common.Result) string { return r.Spec.SpecSourcePath() },
		},
	}
}

// runAgentsPromptsLink dispatches agents.prompts.link command invocations.
func runAgentsPromptsLink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := common.ParseSelectors(selectorRaw)
	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}

	if len(selectors) == 0 && common.IsInteractiveTerminal(stdinAsFile(a)) {
		return runAgentsPromptsLinkInteractive(a, force)
	}

	if len(selectors) == 0 {
		results := prompts.PlanList(a.root)
		if err = common.Render(a.stdout, results, promptsExtraColumns()...); err != nil {
			return err
		}
		if err = writeLine(a.stdout, ""); err != nil {
			return err
		}
		if err = writeLine(a.stdout, "Hint: pass agent=<name|all|csv> to create or rebuild links."); err != nil {
			return err
		}
		return common.EmitHints(a.stdout, results)
	}

	return executeAgentsPromptsLink(a, selectors, force)
}

// runAgentsPromptsLinkInteractive walks the user through a numbered
// selection of link-class agents and optionally enables FORCE for
// mismatched rebuilds.
func runAgentsPromptsLinkInteractive(a *app, force bool) error {
	candidates := prompts.LinkCandidates(a.root)
	names, err := common.PromptSelection(a.stdin, a.stdout, "Select agents to link (prompts):", candidates)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	if !force {
		hasMismatch := false
		for _, entry := range candidates {
			if entry.CurrentStatus == common.StatusMismatch {
				for _, picked := range names {
					if picked == entry.Spec.SpecName() {
						hasMismatch = true
						break
					}
				}
			}
			if hasMismatch {
				break
			}
		}
		if hasMismatch {
			confirmed, confirmErr := common.PromptYesNo(a.stdin, a.stdout,
				"One or more selected agents have mismatched links. Rebuild with FORCE=1?", false)
			if confirmErr != nil {
				return confirmErr
			}
			force = confirmed
		}
	}
	return executeAgentsPromptsLink(a, names, force)
}

// executeAgentsPromptsLink applies the link request and renders results.
func executeAgentsPromptsLink(a *app, selectors []string, force bool) error {
	results, err := prompts.ApplyLink(a.root, prompts.LinkRequest{
		Selectors: selectors,
		Force:     force,
	})
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
