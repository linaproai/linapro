// This file implements the skills.link command which manages repository-local
// symlinks from supported agents' project skill paths to .agents/skills.
// It delegates planning/apply logic to the internal/skilllink subcomponent
// and offers an interactive selection flow when invoked from a TTY without
// an explicit agent argument.

package main

import (
	"context"
	"fmt"
	"os"

	"linactl/internal/skilllink"
)

// runSkillsLink dispatches skills.link command invocations.
func runSkillsLink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := skilllink.ParseSelectors(selectorRaw)
	force, err := input.Bool("force", false)
	if err != nil {
		return err
	}

	// Interactive mode triggers only when the caller did not specify any
	// agent on the command line and stdin is attached to a real terminal.
	// CI and piped contexts retain the read-only listing behavior so
	// existing automations are not disrupted.
	if len(selectors) == 0 && skilllink.IsInteractiveTerminal(stdinAsFile(a)) {
		return runSkillsLinkInteractive(a, force)
	}

	if len(selectors) == 0 {
		results := skilllink.PlanList(a.root)
		if err = skilllink.Render(a.stdout, results); err != nil {
			return err
		}
		if err = writeLine(a.stdout, ""); err != nil {
			return err
		}
		if err = writeLine(a.stdout, "Hint: pass agent=<name|all|csv> to create or rebuild links."); err != nil {
			return err
		}
		return skilllink.EmitHints(a.stdout, results)
	}

	return executeSkillsLink(a, selectors, force)
}

// runSkillsLinkInteractive walks the user through a numbered selection of
// link-class agents and optionally enables FORCE for mismatched rebuilds.
func runSkillsLinkInteractive(a *app, force bool) error {
	candidates := skilllink.LinkCandidates(a.root)
	names, err := skilllink.PromptSelection(a.stdin, a.stdout, "Select agents to link:", candidates)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	if !force {
		hasMismatch := false
		for _, entry := range candidates {
			if entry.CurrentStatus == skilllink.StatusMismatch {
				for _, picked := range names {
					if picked == entry.Spec.Name {
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
			confirmed, confirmErr := skilllink.PromptYesNo(a.stdin, a.stdout,
				"One or more selected agents have mismatched links. Rebuild with FORCE=1?", false)
			if confirmErr != nil {
				return confirmErr
			}
			force = confirmed
		}
	}
	return executeSkillsLink(a, names, force)
}

// executeSkillsLink applies the link request and renders results.
func executeSkillsLink(a *app, selectors []string, force bool) error {
	results, err := skilllink.ApplyLink(a.root, skilllink.LinkRequest{
		Selectors: selectors,
		Force:     force,
	})
	if err != nil {
		return err
	}
	if err = skilllink.Render(a.stdout, results); err != nil {
		return err
	}
	if err = writeLine(a.stdout, ""); err != nil {
		return err
	}
	if err = skilllink.EmitHints(a.stdout, results); err != nil {
		return err
	}
	if skilllink.HasError(results) {
		return fmt.Errorf("one or more agents failed; see DETAIL column")
	}
	return nil
}

// stdinAsFile returns the *os.File backing the app's stdin when available,
// or nil when the test harness wired an in-memory reader.
func stdinAsFile(a *app) *os.File {
	if file, ok := a.stdin.(*os.File); ok {
		return file
	}
	return nil
}
