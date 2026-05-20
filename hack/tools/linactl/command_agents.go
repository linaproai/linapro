// This file implements the agents aggregate command. When invoked from a
// terminal it presents a three-level menu: first choose the resource
// type (skills / prompts / md), then the action (link / unlink), then
// hand off to the matching agents.<resource>.<action> interactive flow.
// In non-TTY contexts it prints usage guidance pointing callers at the
// six explicit subcommands.
//
// This file also owns the writeLine / writeLines stdout helpers shared
// by every agents.* command.

package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"linactl/internal/agents/common"
)

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

// runAgents dispatches the agents aggregate menu.
func runAgents(ctx context.Context, a *app, input commandInput) error {
	if !common.IsInteractiveTerminal(stdinAsFile(a)) {
		return writeLines(a.stdout,
			"Usage: linactl agents.<resource>.<action>",
			"  resource: skills | prompts | md",
			"  action:   link | unlink",
			"",
			"Examples:",
			"  make agents.skills.link [AGENT=<name|all|csv>] [FORCE=1]",
			"  make agents.skills.unlink [AGENT=<name|all|csv>]",
			"  make agents.prompts.link [AGENT=<name|all|csv>] [FORCE=1]",
			"  make agents.prompts.unlink [AGENT=<name|all|csv>]",
			"  make agents.md.link [AGENT=<name|all|csv>] [FORCE=1]",
			"  make agents.md.unlink [AGENT=<name|all|csv>]",
			"",
			"Hint: run `make agents` from an interactive terminal to use the menu.",
		)
	}
	return runAgentsInteractiveMenu(ctx, a, input)
}

// runAgentsInteractiveMenu renders the resource-selection menu (level 1)
// and dispatches to the action menu (level 2). Cancellation (`q`,
// `quit`, blank line) at any level returns nil.
func runAgentsInteractiveMenu(ctx context.Context, a *app, input commandInput) error {
	for {
		if err := writeLines(a.stdout,
			"What resource do you want to manage?",
			"  [1] skills    Agent skills directory bridge (.<tool>/skills -> .agents/skills)",
			"  [2] prompts   Agent commands/prompts directory bridge (.<tool>/.../opsx -> .agents/prompts/opsx)",
			"  [3] md        AGENTS.md project guide file bridge (.<tool>.md -> AGENTS.md)",
			"  [q] quit",
		); err != nil {
			return err
		}
		if _, err := fmt.Fprint(a.stdout, "> "); err != nil {
			return fmt.Errorf("write prompt: %w", err)
		}

		choice, err := common.ReadLine(a.stdin)
		if err != nil {
			return err
		}
		switch choice {
		case "", "q", "quit":
			return writeLine(a.stdout, "Cancelled.")
		case "1", "skills", "s":
			if err = runAgentsActionMenu(ctx, a, input, "skills"); err != nil {
				return err
			}
			return nil
		case "2", "prompts", "p":
			if err = runAgentsActionMenu(ctx, a, input, "prompts"); err != nil {
				return err
			}
			return nil
		case "3", "md", "m":
			if err = runAgentsActionMenu(ctx, a, input, "md"); err != nil {
				return err
			}
			return nil
		default:
			return fmt.Errorf("invalid choice %q: expected 1, 2, 3 or q", choice)
		}
	}
}

// runAgentsActionMenu renders the action-selection menu (level 2) and
// dispatches to the matching agents.<resource>.<action> interactive
// flow. `q` / `back` returns to the resource menu via nil.
func runAgentsActionMenu(_ context.Context, a *app, input commandInput, resource string) error {
	if err := writeLines(a.stdout,
		"",
		"What do you want to do?",
		"  [1] link    Create symlinks for the chosen resource",
		"  [2] unlink  Remove managed symlinks for the chosen resource",
		"  [q] back",
	); err != nil {
		return err
	}
	if _, err := fmt.Fprint(a.stdout, "> "); err != nil {
		return fmt.Errorf("write prompt: %w", err)
	}

	choice, err := common.ReadLine(a.stdin)
	if err != nil {
		return err
	}
	switch choice {
	case "", "q", "quit", "back", "b":
		return writeLine(a.stdout, "Cancelled.")
	case "1", "link", "l":
		force, _ := input.Bool("force", false)
		switch resource {
		case "skills":
			return runAgentsSkillsLinkInteractive(a, force)
		case "prompts":
			return runAgentsPromptsLinkInteractive(a, force)
		case "md":
			return runAgentsMdLinkInteractive(a, force)
		}
	case "2", "unlink", "u":
		switch resource {
		case "skills":
			return runAgentsSkillsUnlinkInteractive(a)
		case "prompts":
			return runAgentsPromptsUnlinkInteractive(a)
		case "md":
			return runAgentsMdUnlinkInteractive(a)
		}
	default:
		return fmt.Errorf("invalid choice %q: expected 1, 2 or q", choice)
	}
	// Unreachable: every resource is handled above.
	return fmt.Errorf("unsupported resource %q", resource)
}
