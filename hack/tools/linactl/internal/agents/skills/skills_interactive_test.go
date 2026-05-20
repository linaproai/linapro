// This file contains unit tests for the skills resource's interactive
// candidate selectors (LinkCandidates, UnlinkCandidates). The
// resource-agnostic prompt/grid behavior is tested in the common
// subpackage; here we verify only that the skills registry filters
// produce the right candidate set.

package skills

import (
	"testing"

	"linactl/internal/agents/common"
)

func TestLinkCandidatesExcludeNativeAndRoot(t *testing.T) {
	root := newRepoFixture(t)
	candidates := LinkCandidates(root)
	for _, entry := range candidates {
		if entry.Spec.SpecCategory() != common.CategoryLink {
			t.Fatalf("LinkCandidates leaked %s (category=%s)",
				entry.Spec.SpecName(), entry.Spec.SpecCategory())
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
		names[entry.Spec.SpecName()] = true
		if entry.CurrentStatus != common.StatusOK {
			t.Fatalf("UnlinkCandidates included non-managed entry %s status=%s",
				entry.Spec.SpecName(), entry.CurrentStatus)
		}
	}
	if !names["claude-code"] || !names["codebuddy"] {
		t.Fatalf("expected seeded agents in candidates, got %v", names)
	}
}
