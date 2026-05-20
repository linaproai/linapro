// This file contains unit tests for the interactive selection helpers.
// Tests use bytes.Buffer as both input and output so they exercise the
// parser without requiring a real TTY.

package skilllink

import (
	"bytes"
	"strings"
	"testing"
)

func makeCandidates() []SelectableEntry {
	return []SelectableEntry{
		{Spec: AgentSpec{Name: "claude-code", ProjectPath: ".claude/skills", Category: CategoryLink}, CurrentStatus: StatusAbsent},
		{Spec: AgentSpec{Name: "codebuddy", ProjectPath: ".codebuddy/skills", Category: CategoryLink}, CurrentStatus: StatusOK},
		{Spec: AgentSpec{Name: "qoder", ProjectPath: ".qoder/skills", Category: CategoryLink}, CurrentStatus: StatusMismatch, Detail: "-> elsewhere"},
	}
}

func TestPromptSelectionAcceptsCommaList(t *testing.T) {
	in := bytes.NewBufferString("1,3\n")
	out := &bytes.Buffer{}
	got, err := PromptSelection(in, out, "Select:", makeCandidates())
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(got) != 2 || got[0] != "claude-code" || got[1] != "qoder" {
		t.Fatalf("unexpected selection: %v", got)
	}
	if !strings.Contains(out.String(), "claude-code") {
		t.Fatalf("expected candidate listing in output: %q", out.String())
	}
}

func TestPromptSelectionAll(t *testing.T) {
	in := bytes.NewBufferString("all\n")
	got, err := PromptSelection(in, &bytes.Buffer{}, "Select:", makeCandidates())
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected all 3 selected, got %v", got)
	}
}

func TestPromptSelectionCancel(t *testing.T) {
	for _, line := range []string{"\n", "q\n", "  \n"} {
		in := bytes.NewBufferString(line)
		got, err := PromptSelection(in, &bytes.Buffer{}, "Select:", makeCandidates())
		if err != nil {
			t.Fatalf("prompt(%q): %v", line, err)
		}
		if len(got) != 0 {
			t.Fatalf("prompt(%q) expected no selection, got %v", line, got)
		}
	}
}

func TestPromptSelectionEmptyCandidates(t *testing.T) {
	out := &bytes.Buffer{}
	got, err := PromptSelection(bytes.NewBufferString(""), out, "Select:", nil)
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty selection")
	}
	if !strings.Contains(out.String(), "no candidates") {
		t.Fatalf("expected no-candidates message: %q", out.String())
	}
}

func TestPromptSelectionRejectsOutOfRange(t *testing.T) {
	in := bytes.NewBufferString("99\n")
	if _, err := PromptSelection(in, &bytes.Buffer{}, "Select:", makeCandidates()); err == nil {
		t.Fatalf("expected out-of-range error")
	}
	in2 := bytes.NewBufferString("abc\n")
	if _, err := PromptSelection(in2, &bytes.Buffer{}, "Select:", makeCandidates()); err == nil {
		t.Fatalf("expected non-numeric error")
	}
}

func TestPromptSelectionDeduplicates(t *testing.T) {
	in := bytes.NewBufferString("1,1,2,2,3\n")
	got, err := PromptSelection(in, &bytes.Buffer{}, "Select:", makeCandidates())
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 unique selections, got %v", got)
	}
}

func TestPromptYesNoDefaults(t *testing.T) {
	got, err := PromptYesNo(bytes.NewBufferString("\n"), &bytes.Buffer{}, "OK?", false)
	if err != nil || got {
		t.Fatalf("default no: got=%v err=%v", got, err)
	}
	got, err = PromptYesNo(bytes.NewBufferString("\n"), &bytes.Buffer{}, "OK?", true)
	if err != nil || !got {
		t.Fatalf("default yes: got=%v err=%v", got, err)
	}
}

func TestPromptYesNoExplicit(t *testing.T) {
	for _, line := range []string{"y\n", "Y\n", "yes\n", "YES\n"} {
		got, err := PromptYesNo(bytes.NewBufferString(line), &bytes.Buffer{}, "OK?", false)
		if err != nil || !got {
			t.Fatalf("expected true for %q got=%v err=%v", line, got, err)
		}
	}
	for _, line := range []string{"n\n", "no\n"} {
		got, err := PromptYesNo(bytes.NewBufferString(line), &bytes.Buffer{}, "OK?", true)
		if err != nil || got {
			t.Fatalf("expected false for %q got=%v err=%v", line, got, err)
		}
	}
}

func TestPromptYesNoInvalid(t *testing.T) {
	if _, err := PromptYesNo(bytes.NewBufferString("maybe\n"), &bytes.Buffer{}, "OK?", false); err == nil {
		t.Fatalf("expected error for invalid yes/no")
	}
}

func TestLinkCandidatesExcludeNativeAndRoot(t *testing.T) {
	root := newRepoFixture(t)
	candidates := LinkCandidates(root)
	for _, entry := range candidates {
		if entry.Spec.Category != CategoryLink {
			t.Fatalf("LinkCandidates leaked %s (category=%s)", entry.Spec.Name, entry.Spec.Category)
		}
	}
}

func TestUnlinkCandidatesOnlyManagedLinks(t *testing.T) {
	root := newRepoFixture(t)
	if _, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code", "codebuddy"}}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	candidates := UnlinkCandidates(root)
	names := make(map[string]bool, len(candidates))
	for _, entry := range candidates {
		names[entry.Spec.Name] = true
		if entry.CurrentStatus != StatusOK {
			t.Fatalf("UnlinkCandidates included non-managed entry %s status=%s", entry.Spec.Name, entry.CurrentStatus)
		}
	}
	if !names["claude-code"] || !names["codebuddy"] {
		t.Fatalf("expected seeded agents in candidates, got %v", names)
	}
}

func TestIsInteractiveTerminalNilFile(t *testing.T) {
	if IsInteractiveTerminal(nil) {
		t.Fatalf("nil file must not be a terminal")
	}
}

func TestReadLineTrimsAndLowercases(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{input: "Link\n", want: "link"},
		{input: "  q  \n", want: "q"},
		{input: "\n", want: ""},
		{input: "", want: ""},
	}
	for _, testCase := range cases {
		got, err := ReadLine(bytes.NewBufferString(testCase.input))
		if err != nil {
			t.Fatalf("ReadLine(%q): %v", testCase.input, err)
		}
		if got != testCase.want {
			t.Fatalf("ReadLine(%q) got=%q want=%q", testCase.input, got, testCase.want)
		}
	}
}

func TestStatusGlyphCoversAllStatuses(t *testing.T) {
	cases := []struct {
		status Status
		want   string
	}{
		{StatusOK, "[+]"},
		{StatusCreated, "[+]"},
		{StatusRebuilt, "[+]"},
		{StatusRemoved, "[+]"},
		{StatusMismatch, "[~]"},
		{StatusSkippedForeignTarget, "[~]"},
		{StatusSkippedNotManaged, "[~]"},
		{StatusAbsent, "[.]"},
		{StatusNative, "[.]"},
		{StatusConflict, "[!]"},
		{StatusSkippedRootCollision, "[*]"},
		{StatusError, "[?]"},
		{Status("unknown"), "[?]"},
	}
	for _, testCase := range cases {
		if got := statusGlyph(testCase.status); got != testCase.want {
			t.Fatalf("statusGlyph(%s) got=%s want=%s", testCase.status, got, testCase.want)
		}
	}
}

func TestPromptSelectionRendersThreeColumnGrid(t *testing.T) {
	in := bytes.NewBufferString("q\n")
	out := &bytes.Buffer{}
	candidates := makeCandidates()
	if _, err := PromptSelection(in, out, "Select:", candidates); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Legend:") {
		t.Fatalf("expected legend line; got %q", output)
	}
	for _, fragment := range []string{"[+]", "[~]", "[.]", "[!]"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("legend missing glyph %s; got %q", fragment, output)
		}
	}
	// All three candidates share a single row in the 3-column layout.
	gridLines := strings.Split(output, "\n")
	rowCount := 0
	for _, line := range gridLines {
		if strings.HasPrefix(line, "  [") {
			rowCount++
		}
	}
	if rowCount != 1 {
		t.Fatalf("expected exactly 1 grid row for 3 candidates in 3 columns; got %d rows in %q", rowCount, output)
	}
	// All candidate names must appear on that single row.
	for _, name := range []string{"claude-code", "codebuddy", "qoder"} {
		if !strings.Contains(output, name) {
			t.Fatalf("expected candidate %s in grid; got %q", name, output)
		}
	}
}
