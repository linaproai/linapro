// This file contains tests for the resource-agnostic common subpackage.
// Tests cover the generic ResolveTargets selector resolver and the
// shared StatusGlyph mapping; per-resource Inspect / ApplyOneLink /
// ApplyOneUnlink behavior is exercised through each resource subpackage's
// own test suite.

package common

import (
	"testing"
)

// fakeSpec is a minimal SpecLike implementation used by ResolveTargets
// tests. It carries just enough information to exercise the all-expansion
// policy filters and missing-name lookup paths.
type fakeSpec struct {
	name     string
	category Category
}

func (s fakeSpec) SpecName() string         { return s.name }
func (s fakeSpec) SpecDisplayName() string  { return s.name }
func (s fakeSpec) SpecCategory() Category   { return s.category }
func (s fakeSpec) SpecSourcePath() string   { return ".source" }
func (s fakeSpec) SpecProjectPath() string  { return ".project" }
func (s fakeSpec) SpecKind() Kind           { return KindDir }

func makeFakeRegistry() []fakeSpec {
	return []fakeSpec{
		{name: "a-native", category: CategoryNative},
		{name: "b-link-1", category: CategoryLink},
		{name: "c-link-2", category: CategoryLink},
		{name: "d-root", category: CategoryRootCollision},
	}
}

func TestResolveTargetsAllExpansionExcludesNativeAndRootByDefault(t *testing.T) {
	registry := makeFakeRegistry()
	got, err := ResolveTargets([]string{SelectorAll}, registry, TargetPolicy{})
	if err != nil {
		t.Fatalf("ResolveTargets(all): %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected link-class agents in default expansion")
	}
	for _, spec := range got {
		if spec.SpecCategory() != CategoryLink {
			t.Fatalf("agent %s leaked into default all expansion (category=%s)",
				spec.SpecName(), spec.SpecCategory())
		}
	}
}

func TestResolveTargetsAllExpansionIncludesPolicyOptIns(t *testing.T) {
	registry := makeFakeRegistry()
	got, err := ResolveTargets([]string{SelectorAll}, registry, TargetPolicy{
		IncludeNative:        true,
		IncludeRootCollision: true,
	})
	if err != nil {
		t.Fatalf("ResolveTargets(all,full): %v", err)
	}
	if len(got) != len(registry) {
		t.Fatalf("expected full expansion to cover %d agents, got %d", len(registry), len(got))
	}
}

func TestResolveTargetsUnknownAgentReturnsError(t *testing.T) {
	registry := makeFakeRegistry()
	if _, err := ResolveTargets([]string{"no-such-agent"}, registry, TargetPolicy{}); err == nil {
		t.Fatalf("expected unknown agent error")
	}
}

func TestResolveTargetsExplicitNamesReturnsRequestedSpecs(t *testing.T) {
	registry := makeFakeRegistry()
	got, err := ResolveTargets([]string{"b-link-1", "a-native"}, registry, TargetPolicy{})
	if err != nil {
		t.Fatalf("ResolveTargets explicit: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(got))
	}
	// Output is sorted by SpecName, so a-native should precede b-link-1.
	if got[0].SpecName() != "a-native" || got[1].SpecName() != "b-link-1" {
		t.Fatalf("expected sorted output, got %v", got)
	}
}

func TestResolveTargetsEmptySelectorsReturnsNil(t *testing.T) {
	registry := makeFakeRegistry()
	got, err := ResolveTargets([]string{}, registry, TargetPolicy{})
	if err != nil {
		t.Fatalf("ResolveTargets empty: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty selectors, got %v", got)
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
		if got := StatusGlyph(testCase.status); got != testCase.want {
			t.Fatalf("StatusGlyph(%s) got=%s want=%s", testCase.status, got, testCase.want)
		}
	}
}

func TestHasErrorReports(t *testing.T) {
	if HasError([]Result{{Status: StatusOK}}) {
		t.Fatalf("expected false")
	}
	if !HasError([]Result{{Status: StatusError}}) {
		t.Fatalf("expected true")
	}
}
