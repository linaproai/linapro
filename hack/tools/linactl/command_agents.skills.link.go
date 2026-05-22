// This file implements the agents.skills.link command which manages
// repository-local directory symlinks from supported agents' project
// skills paths to .agents/skills. It delegates planning/apply logic to
// the internal/agents/skills subpackage and offers an interactive
// selection flow when invoked from a TTY without an explicit agent
// argument.

package main

import (
	"context"
	"fmt"

	"linactl/internal/agents/common"
	"linactl/internal/agents/skills"
)

// runAgentsSkillsLink dispatches agents.skills.link command invocations.
func runAgentsSkillsLink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := common.ParseSelectors(selectorRaw)
	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}

	// Interactive mode triggers only when the caller did not specify any
	// agent on the command line and stdin is attached to a real terminal.
	// CI and piped contexts retain the read-only listing behavior so
	// existing automations are not disrupted.
	if len(selectors) == 0 && common.IsInteractiveTerminal(stdinAsFile(a)) {
		return runAgentsSkillsLinkInteractive(a, force)
	}

	if len(selectors) == 0 {
		results := skills.PlanList(a.root)
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

	return executeAgentsSkillsLink(a, selectors, force)
}

// runAgentsSkillsLinkInteractive walks the user through a numbered
// selection of link-class agents and optionally enables FORCE for
// mismatched rebuilds. Before showing the selection grid it renders a
// full status overview that includes native and rootCollision agents,
// so users can see every agent's current state — including agents that
// read .agents/skills directly and need no symlink — even though the
// grid only lets them pick link-class agents to act on.
func runAgentsSkillsLinkInteractive(a *app, force bool) error {
	overview := skills.PlanList(a.root)
	if err := common.Render(a.stdout, overview); err != nil {
		return err
	}
	if err := writeLine(a.stdout, ""); err != nil {
		return err
	}
	candidates := skills.LinkCandidates(a.root)
	names, err := common.PromptSelection(a.stdin, a.stdout, "Select agents to link (skills):", candidates)
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
	return executeAgentsSkillsLink(a, names, force)
}

// executeAgentsSkillsLink applies the link request and renders results.
func executeAgentsSkillsLink(a *app, selectors []string, force bool) error {
	results, err := skills.ApplyLink(a.root, skills.LinkRequest{
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
