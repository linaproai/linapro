// This file contains unit tests for agents.* command interactive flows
// that are owned by the main package. The resource-agnostic prompt and
// engine logic are tested in internal/agents/common; the per-resource
// candidate filters are tested in internal/agents/{skills,prompts,md}.
// Here we verify only the main package wiring — specifically that
// agents.skills.link and agents.md.link interactive entry points render
// a full status overview (including native-class agents) before showing
// the link-class selection grid, so users can see every agent's current
// state in the TTY menu reachable via `make agents`.

package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestAgentsSkillsLinkInteractiveShowsNativeAgents verifies that the
// skills interactive link flow renders native-class agents (e.g.
// `cursor`, `gemini-cli`, `codex`) in the status overview emitted before
// the candidate grid. The user enters `q` to cancel before any apply
// runs, so the test is filesystem-safe and does not depend on a
// pre-seeded fixture.
func TestAgentsSkillsLinkInteractiveShowsNativeAgents(t *testing.T) {
	repoRoot := t.TempDir()
	var stdout bytes.Buffer
	a := &app{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("q\n"),
		root:   repoRoot,
	}

	if err := runAgentsSkillsLinkInteractive(a, false); err != nil {
		t.Fatalf("runAgentsSkillsLinkInteractive: %v", err)
	}

	output := stdout.String()
	// Native-class agents must appear in the overview table emitted
	// before the candidate grid. The overview row for a native agent
	// shows its name, an empty PROJECT PATH column and the literal
	// `native` status word (rendered twice: once as category, once as
	// status) — we assert on the agent name plus the `native` keyword
	// to stay resilient to minor column-width changes.
	for _, name := range []string{"cursor", "gemini-cli", "codex"} {
		if !strings.Contains(output, name) {
			t.Fatalf("native agent %q missing from skills interactive overview; output=%q", name, output)
		}
	}
	if !strings.Contains(output, "native") {
		t.Fatalf("expected native status keyword in skills interactive overview; output=%q", output)
	}
	// The candidate grid prompt must still be rendered after the
	// overview, proving that the overview did not replace the grid.
	if !strings.Contains(output, "Select agents to link (skills)") {
		t.Fatalf("expected candidate selection prompt after overview; output=%q", output)
	}
}

// TestAgentsMdLinkInteractiveShowsNativeAgents verifies the md
// interactive link flow renders native-class agents (e.g. `codex`,
// `cursor`, `amp`) — agents that read AGENTS.md natively and need no
// symlink — in the status overview emitted before the candidate grid.
func TestAgentsMdLinkInteractiveShowsNativeAgents(t *testing.T) {
	repoRoot := t.TempDir()
	var stdout bytes.Buffer
	a := &app{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("q\n"),
		root:   repoRoot,
	}

	if err := runAgentsMdLinkInteractive(a, false); err != nil {
		t.Fatalf("runAgentsMdLinkInteractive: %v", err)
	}

	output := stdout.String()
	// Sample three diverse md-native agents. `amp` only exists in the
	// md registry; `codex` and `cursor` exist in multiple registries
	// and must still appear here because the md PlanList covers the md
	// registry only.
	for _, name := range []string{"codex", "cursor", "amp"} {
		if !strings.Contains(output, name) {
			t.Fatalf("native agent %q missing from md interactive overview; output=%q", name, output)
		}
	}
	if !strings.Contains(output, "native") {
		t.Fatalf("expected native status keyword in md interactive overview; output=%q", output)
	}
	if !strings.Contains(output, "Select agents to link (md)") {
		t.Fatalf("expected candidate selection prompt after overview; output=%q", output)
	}
}
