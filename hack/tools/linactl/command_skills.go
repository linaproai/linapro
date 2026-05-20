// This file implements the skills aggregate command. When invoked from a
// terminal it presents a small action menu that dispatches to skills.link
// or skills.unlink interactive flows; in non-TTY contexts it prints usage
// guidance pointing callers at the explicit subcommands.

package main

import (
	"context"
	"fmt"
	"io"

	"linactl/internal/skilllink"
)

// writeLine prints a single line to the writer, wrapping any write error so
// linactl never silently drops stdout failures (e.g. a broken pipe). The
// helper is intentionally local to skills commands; existing linactl
// commands keep their previous fmt.Fprintln-based style.
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

// runSkills dispatches the skills aggregate menu.
func runSkills(ctx context.Context, a *app, input commandInput) error {
	if !skilllink.IsInteractiveTerminal(stdinAsFile(a)) {
		return writeLines(a.stdout,
			"Usage: linactl skills.link | linactl skills.unlink",
			"       make skills.link [AGENT=<name|all|csv>] [FORCE=1]",
			"       make skills.unlink [AGENT=<name|all|csv>]",
			"",
			"Hint: run `make skills` from an interactive terminal to use the action menu.",
		)
	}
	return runSkillsInteractiveMenu(ctx, a, input)
}

// runSkillsInteractiveMenu renders the action menu, reads the user's choice,
// and dispatches to the matching skills.link or skills.unlink interactive
// flow. Cancellation (`q`, `quit`, blank line) returns nil.
func runSkillsInteractiveMenu(_ context.Context, a *app, input commandInput) error {
	if err := writeLines(a.stdout,
		"What do you want to do?",
		"  [1] link    Create symlinks from agent project paths to .agents/skills",
		"  [2] unlink  Remove managed symlinks",
		"  [q] quit",
	); err != nil {
		return err
	}
	if _, err := fmt.Fprint(a.stdout, "> "); err != nil {
		return fmt.Errorf("write prompt: %w", err)
	}

	choice, err := skilllink.ReadLine(a.stdin)
	if err != nil {
		return err
	}
	switch choice {
	case "", "q", "quit":
		return writeLine(a.stdout, "Cancelled.")
	case "1", "link", "l":
		force, _ := input.Bool("force", false)
		return runSkillsLinkInteractive(a, force)
	case "2", "unlink", "u":
		return runSkillsUnlinkInteractive(a)
	default:
		return fmt.Errorf("invalid choice %q: expected 1, 2 or q", choice)
	}
}
