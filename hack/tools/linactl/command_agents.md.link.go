// This file implements the agents.md.link command which manages
// repository-local single-file symlinks from supported agents' private
// project guide files (e.g. CLAUDE.md, GEMINI.md, .junie/guidelines.md)
// to the repo-root AGENTS.md project specification.

package main

import (
	"context"
	"fmt"

	"linactl/internal/agents/common"
	"linactl/internal/agents/md"
)

// runAgentsMdLink dispatches agents.md.link command invocations.
func runAgentsMdLink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := common.ParseSelectors(selectorRaw)
	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}

	if len(selectors) == 0 && common.IsInteractiveTerminal(stdinAsFile(a)) {
		return runAgentsMdLinkInteractive(a, force)
	}

	if len(selectors) == 0 {
		results := md.PlanList(a.root)
		if err = common.Render(a.stdout, results); err != nil {
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

	return executeAgentsMdLink(a, selectors, force)
}

// runAgentsMdLinkInteractive walks the user through a numbered selection
// of link-class agents and optionally enables FORCE for mismatched
// rebuilds. Before showing the selection grid it renders a full status
// overview that includes native-class agents, so users can see every
// agent's current state — including agents that read AGENTS.md
// natively and need no symlink — even though the grid only lets them
// pick link-class agents to act on.
func runAgentsMdLinkInteractive(a *app, force bool) error {
	overview := md.PlanList(a.root)
	if err := common.Render(a.stdout, overview); err != nil {
		return err
	}
	if err := writeLine(a.stdout, ""); err != nil {
		return err
	}
	candidates := md.LinkCandidates(a.root)
	names, err := common.PromptSelection(a.stdin, a.stdout, "Select agents to link (md):", candidates)
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
	return executeAgentsMdLink(a, names, force)
}

// executeAgentsMdLink applies the link request and renders results.
func executeAgentsMdLink(a *app, selectors []string, force bool) error {
	results, err := md.ApplyLink(a.root, md.LinkRequest{
		Selectors: selectors,
		Force:     force,
	})
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
