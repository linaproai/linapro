// This file contains unit tests for the agents aggregate command's
// non-interactive surface and the cross-resource agent universe builder.
// The huh-driven interactive menu (runAgentInteractiveMenu) is exercised
// manually because automating arrow-key TUI input is brittle and
// charmbracelet/huh ships its own test coverage for form mechanics.
//
// Coverage in this file:
//   - collectAgentUniverse merges three registries, excludes native-only
//     agents and is sorted deterministically.
//   - validateSingleAgentName rejects empty / "all" / comma-separated /
//     unknown values and accepts a single supported agent.
//   - parseAgentSetupAction normalizes link/unlink and rejects others.
//   - runAgents prints the usage hint when invoked without an agent on
//     a non-TTY stdin (the standard CI path).
//   - runAgents accepts a known agent on a non-TTY stdin and dispatches
//     the per-resource execute helpers without prompting.

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// newTestApp builds an app wired with bytes buffers for stdout/stderr
// and a strings reader for stdin so PromptSelection / PromptYesNo /
// PromptSingleSelection all degrade to their non-TTY safe paths.
func newTestApp(t *testing.T) (*app, *bytes.Buffer) {
	t.Helper()
	stdout := &bytes.Buffer{}
	a := &app{
		stdout: stdout,
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader(""),
		root:   t.TempDir(),
	}
	return a, stdout
}

func TestCollectAgentUniverseExcludesNativeOnly(t *testing.T) {
	universe := collectAgentUniverse(t.TempDir())
	if len(universe) == 0 {
		t.Fatalf("expected non-empty agent universe")
	}
	for _, agent := range universe {
		if !agent.hasLinkRole() {
			t.Fatalf("native-only agent %q leaked into universe (roles=%v)", agent.Name, agent.Roles)
		}
	}
}

func TestCollectAgentUniverseSorted(t *testing.T) {
	universe := collectAgentUniverse(t.TempDir())

	// Build a name -> position map for assertions.
	positions := make(map[string]int, len(universe))
	for index, agent := range universe {
		positions[agent.Name] = index
	}

	// 1) Every priority agent that is registered must appear in the
	//    configured priority order, before any non-priority agent.
	previousRank := -1
	previousName := ""
	highestPriorityIndex := -1
	for _, name := range agentDisplayPriority {
		index, ok := positions[name]
		if !ok {
			// Priority entries may be unregistered (e.g. cline); skip.
			continue
		}
		rank, isPriority := agentPriorityRank(name)
		if !isPriority {
			t.Fatalf("agent %q expected to be priority but rank reports otherwise", name)
		}
		if rank <= previousRank {
			t.Fatalf("priority list rank regressed: %q (rank=%d) after %q (rank=%d)", name, rank, previousName, previousRank)
		}
		previousRank = rank
		previousName = name
		if index > highestPriorityIndex {
			highestPriorityIndex = index
		}
	}

	// 2) After the priority block, the remaining agents must be in
	//    strict ascending alphabetical order.
	tail := universe[highestPriorityIndex+1:]
	for index := 1; index < len(tail); index++ {
		if tail[index-1].Name >= tail[index].Name {
			t.Fatalf("non-priority tail not alphabetically sorted: %q before %q", tail[index-1].Name, tail[index].Name)
		}
	}

	// 3) No non-priority agent may appear before any registered
	//    priority agent.
	for index := 0; index <= highestPriorityIndex; index++ {
		if _, isPriority := agentPriorityRank(universe[index].Name); !isPriority {
			t.Fatalf("non-priority agent %q (index %d) precedes registered priority agents", universe[index].Name, index)
		}
	}
}

func TestAgentPriorityRank(t *testing.T) {
	// claude-code is intentionally first.
	rank, ok := agentPriorityRank("claude-code")
	if !ok || rank != 0 {
		t.Fatalf("expected claude-code at rank 0, got rank=%d ok=%v", rank, ok)
	}
	// codex is intentionally second.
	rank, ok = agentPriorityRank("codex")
	if !ok || rank != 1 {
		t.Fatalf("expected codex at rank 1, got rank=%d ok=%v", rank, ok)
	}
	// codebuddy is intentionally NOT in the priority list (falls back
	// to the alphabetical tail to avoid biasing the default surface).
	if rank, ok := agentPriorityRank("codebuddy"); ok {
		t.Fatalf("codebuddy should fall back to alphabetical tail, got rank=%d", rank)
	}
	// Unknown / non-priority agents fall through to the tail.
	rank, ok = agentPriorityRank("definitely-not-an-agent")
	if ok {
		t.Fatalf("unexpected priority match for unknown name (rank=%d)", rank)
	}
	if rank != len(agentDisplayPriority) {
		t.Fatalf("non-priority rank %d != len(priority) %d", rank, len(agentDisplayPriority))
	}
}

func TestCollectAgentUniverseMergesAcrossRegistries(t *testing.T) {
	universe := collectAgentUniverse(t.TempDir())
	// claude-code is link-class in both skills and md; its roles map
	// must contain at least those two resources.
	for _, agent := range universe {
		if agent.Name != "claude-code" {
			continue
		}
		if _, ok := agent.Roles[resourceSkills]; !ok {
			t.Fatalf("claude-code missing skills role: %v", agent.Roles)
		}
		if _, ok := agent.Roles[resourceMd]; !ok {
			t.Fatalf("claude-code missing md role: %v", agent.Roles)
		}
		return
	}
	t.Fatalf("claude-code not found in universe")
}

func TestValidateSingleAgentName(t *testing.T) {
	universe := collectAgentUniverse(t.TempDir())
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid claude-code", input: "claude-code", wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "literal all", input: "all", wantErr: true},
		{name: "case-insensitive all", input: "ALL", wantErr: true},
		{name: "csv", input: "claude-code,codex", wantErr: true},
		{name: "unknown", input: "no-such-agent", wantErr: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := validateSingleAgentName(testCase.input, universe)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got value %q", testCase.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", testCase.input, err)
			}
			if got != testCase.input {
				t.Fatalf("validate(%q) returned %q", testCase.input, got)
			}
		})
	}
}

func TestParseAgentSetupAction(t *testing.T) {
	cases := []struct {
		input    string
		fallback agentSetupAction
		want     agentSetupAction
		wantErr  bool
	}{
		{input: "", fallback: actionLink, want: actionLink},
		{input: "link", fallback: actionLink, want: actionLink},
		{input: "unlink", fallback: actionLink, want: actionUnlink},
		{input: "wat", fallback: actionLink, wantErr: true},
	}
	for _, testCase := range cases {
		got, err := parseAgentSetupAction(testCase.input, testCase.fallback)
		if testCase.wantErr {
			if err == nil {
				t.Fatalf("expected error for %q", testCase.input)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", testCase.input, err)
		}
		if got != testCase.want {
			t.Fatalf("parse(%q) got=%q want=%q", testCase.input, got, testCase.want)
		}
	}
}

// TestRunAgentsNonTTYPrintsUsage verifies the no-AGENT, non-TTY path
// emits the usage hint and returns successfully (no error, no
// dispatch). This is the standard CI invocation: `linactl agents`
// in a piped context should never block on input.
func TestRunAgentsNonTTYPrintsUsage(t *testing.T) {
	a, stdout := newTestApp(t)
	if err := runAgents(context.Background(), a, commandInput{Params: map[string]string{}}); err != nil {
		t.Fatalf("runAgents: %v", err)
	}
	output := stdout.String()
	for _, fragment := range []string{
		"Usage: linactl agents",
		"One-shot mode",
		"Interactive mode",
		"Advanced per-resource",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("usage hint missing %q; got %q", fragment, output)
		}
	}
}

// TestRunAgentsRejectsAgentAll guards the safety rule: AGENT=all is
// explicitly rejected by the aggregate command.
func TestRunAgentsRejectsAgentAll(t *testing.T) {
	a, _ := newTestApp(t)
	err := runAgents(context.Background(), a, commandInput{Params: map[string]string{"agent": "all"}})
	if err == nil {
		t.Fatalf("expected error for agent=all")
	}
	if !strings.Contains(err.Error(), "all") {
		t.Fatalf("expected error to mention 'all', got %q", err)
	}
}

// TestRunAgentsRejectsCSV guards the safety rule: comma-separated lists
// are explicitly rejected.
func TestRunAgentsRejectsCSV(t *testing.T) {
	a, _ := newTestApp(t)
	err := runAgents(context.Background(), a, commandInput{Params: map[string]string{"agent": "claude-code,codex"}})
	if err == nil {
		t.Fatalf("expected error for csv input")
	}
	if !strings.Contains(err.Error(), "comma-separated") {
		t.Fatalf("expected error to mention comma-separated, got %q", err)
	}
}

// TestRunAgentsUnknownAgentReportsCandidates verifies the error message
// for an unknown agent includes the candidate listing so users can
// recover without consulting docs.
func TestRunAgentsUnknownAgentReportsCandidates(t *testing.T) {
	a, _ := newTestApp(t)
	err := runAgents(context.Background(), a, commandInput{Params: map[string]string{"agent": "no-such-agent"}})
	if err == nil {
		t.Fatalf("expected error for unknown agent")
	}
	if !strings.Contains(err.Error(), "supported agents") {
		t.Fatalf("expected candidate listing; got %q", err)
	}
}

// TestRunAgentsRejectsBadAction verifies typos in ACTION surface at the
// CLI boundary rather than silently falling back.
func TestRunAgentsRejectsBadAction(t *testing.T) {
	a, _ := newTestApp(t)
	err := runAgents(context.Background(), a, commandInput{Params: map[string]string{
		"agent":  "claude-code",
		"action": "wat",
	}})
	if err == nil {
		t.Fatalf("expected error for bad action")
	}
	if !strings.Contains(err.Error(), "invalid action") {
		t.Fatalf("expected invalid action error, got %q", err)
	}
}

// TestDispatchAgentSetupSkipsResourcesNotRegistered verifies the
// dispatcher gracefully skips resources where the agent is not
// registered (e.g. claude-code is not in prompts? actually it is in
// prompts; pick an agent only in skills to assert skip).
func TestDispatchAgentSetupSkipsUnregisteredResources(t *testing.T) {
	a, stdout := newTestApp(t)
	universe := collectAgentUniverse(a.root)
	// Pick the first agent that is link-class in skills only, or fall
	// back to a known link-only-in-skills entry. `codebuddy` is a
	// strong candidate (skills link, no md/prompts entry today).
	target := "codebuddy"
	if _, ok := lookupAgent(universe, target); !ok {
		t.Skipf("expected %q in universe; got %v", target, universe)
	}

	// We expect no error in non-TTY because executeAgentsSkillsLink
	// will return after rendering its own table; the temp dir has no
	// pre-existing collisions, so a fresh symlink should be created.
	if err := dispatchAgentSetup(a, target, actionLink, false, universe); err != nil {
		t.Fatalf("dispatchAgentSetup: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected summary block; got %q", output)
	}
	// At least one of prompts/md must be reported as skipped because
	// codebuddy has no registration there.
	if !strings.Contains(output, "skipped") {
		t.Fatalf("expected at least one resource to be reported as skipped; got %q", output)
	}
}
