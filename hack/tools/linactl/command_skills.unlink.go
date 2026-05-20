// This file implements the skills.unlink command which removes managed
// symlinks pointing at .agents/skills. It never removes real directories
// or files and never touches symlinks pointing at foreign targets. When
// invoked on a TTY without an explicit agent argument, it offers an
// interactive selection of currently linked agents.

package main

import (
	"context"
	"fmt"

	"linactl/internal/skilllink"
)

// runSkillsUnlink dispatches skills.unlink command invocations.
func runSkillsUnlink(_ context.Context, a *app, input commandInput) error {
	selectorRaw := input.GetDefault("agent", "")
	selectors := skilllink.ParseSelectors(selectorRaw)

	if len(selectors) == 0 && skilllink.IsInteractiveTerminal(stdinAsFile(a)) {
		return runSkillsUnlinkInteractive(a)
	}
	if len(selectors) == 0 {
		return fmt.Errorf("agent=<name|all|csv> is required for skills.unlink in non-interactive mode")
	}

	return executeSkillsUnlink(a, selectors)
}

// runSkillsUnlinkInteractive walks the user through a numbered selection
// of currently managed links to remove.
func runSkillsUnlinkInteractive(a *app) error {
	candidates := skilllink.UnlinkCandidates(a.root)
	if len(candidates) == 0 {
		return writeLine(a.stdout, "No managed agent skill symlinks were found. Nothing to unlink.")
	}
	names, err := skilllink.PromptSelection(a.stdin, a.stdout, "Select agents to unlink:", candidates)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	return executeSkillsUnlink(a, names)
}

// executeSkillsUnlink applies the unlink request and renders results.
func executeSkillsUnlink(a *app, selectors []string) error {
	results, err := skilllink.ApplyUnlink(a.root, skilllink.UnlinkRequest{Selectors: selectors})
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
