// This file contains unit tests for the resource-agnostic interactive
// primitives that survived the migration to charmbracelet/huh.
//
// We test only the parts that are testable without a real terminal:
//   - IsInteractiveTerminal and ReadLine (legacy primitives that still
//     drive non-TTY paths and unit tests).
//   - The non-TTY degrade behaviour of PromptSelection / PromptSingleSelection
//     / PromptYesNo. When stdin is not a *os.File pointing at a character
//     device the helpers must return safe zero values rather than spawn a
//     huh form. This protects test harnesses that wire bytes.Buffer streams
//     and CI invocations that should never reach the interactive path.
//
// The actual huh-driven TUI rendering and key handling is exercised by
// charmbracelet/huh's own test suite and verified manually; we do not
// attempt to assert on its output bytes here because they include
// terminal control sequences whose layout is intentionally not part of
// our public contract.

package common

import (
	"bytes"
	"strings"
	"testing"
)

// fakeSelectables produces a tiny SelectableEntry slice for tests in
// this file. It mirrors makeFakeRegistry/fakeSpec from common_test.go
// but composes the SelectableEntry rows directly.
func fakeSelectables() []SelectableEntry {
	return []SelectableEntry{
		{Spec: fakeSpec{name: "claude-code", category: CategoryLink}, CurrentStatus: StatusAbsent},
		{Spec: fakeSpec{name: "codebuddy", category: CategoryLink}, CurrentStatus: StatusOK},
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

// TestPromptSelectionNonTTYReturnsEmpty verifies that PromptSelection
// gracefully degrades to (nil, nil) when stdin is not a *os.File. This
// keeps mock-based command tests from accidentally spawning a huh form.
func TestPromptSelectionNonTTYReturnsEmpty(t *testing.T) {
	in := bytes.NewBufferString("")
	out := &bytes.Buffer{}
	got, err := PromptSelection(in, out, "Select:", fakeSelectables())
	if err != nil {
		t.Fatalf("PromptSelection: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty selection on non-TTY, got %v", got)
	}
}

// TestPromptSelectionEmptyCandidates verifies the explicit empty-list
// message that runs even on non-TTY callers, since it does not require
// terminal IO.
func TestPromptSelectionEmptyCandidates(t *testing.T) {
	out := &bytes.Buffer{}
	got, err := PromptSelection(bytes.NewBufferString(""), out, "Select:", nil)
	if err != nil {
		t.Fatalf("PromptSelection: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty selection")
	}
	if !strings.Contains(out.String(), "no candidates") {
		t.Fatalf("expected no-candidates message: %q", out.String())
	}
}

// TestPromptSingleSelectionNonTTYReturnsEmpty mirrors the multi-select
// degrade path for the single-select helper.
func TestPromptSingleSelectionNonTTYReturnsEmpty(t *testing.T) {
	got, err := PromptSingleSelection(bytes.NewBufferString(""), &bytes.Buffer{}, "Pick:", []SingleOption{
		{Value: "a", Label: "A"},
		{Value: "b", Label: "B"},
	})
	if err != nil {
		t.Fatalf("PromptSingleSelection: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty value on non-TTY, got %q", got)
	}
}

// TestPromptSingleSelectionEmptyOptionsErrors guards against passing an
// empty options slice, which is a programmer error.
func TestPromptSingleSelectionEmptyOptionsErrors(t *testing.T) {
	if _, err := PromptSingleSelection(bytes.NewBufferString(""), &bytes.Buffer{}, "Pick:", nil); err == nil {
		t.Fatalf("expected error for empty options")
	}
}

// TestPromptYesNoNonTTYReturnsDefault verifies the helper returns the
// caller-supplied default when stdin is not a TTY, instead of spawning
// a form or erroring.
func TestPromptYesNoNonTTYReturnsDefault(t *testing.T) {
	got, err := PromptYesNo(bytes.NewBufferString(""), &bytes.Buffer{}, "OK?", true)
	if err != nil || !got {
		t.Fatalf("default yes path: got=%v err=%v", got, err)
	}
	got, err = PromptYesNo(bytes.NewBufferString(""), &bytes.Buffer{}, "OK?", false)
	if err != nil || got {
		t.Fatalf("default no path: got=%v err=%v", got, err)
	}
}

// TestFormatCandidateLabelEmbedsGlyphAndStatus exercises the label
// builder used to compose huh option text. We verify glyph + name +
// status descriptor are all present.
func TestFormatCandidateLabelEmbedsGlyphAndStatus(t *testing.T) {
	entry := SelectableEntry{
		Spec:          fakeSpec{name: "claude-code", category: CategoryLink},
		CurrentStatus: StatusMismatch,
		Detail:        "-> /elsewhere",
	}
	got := formatCandidateLabel(entry)
	for _, want := range []string{"[~]", "claude-code", "-> /elsewhere"} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatCandidateLabel missing %q in %q", want, got)
		}
	}
}
