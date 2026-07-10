// This file verifies side-effect-free runtime upgrade planning helpers.

package upgrade

import (
	"testing"

	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestBuildHostServicesDiffKeepsOwnerAwareServiceIdentity verifies same-name
// plugin-owned services from different owners are compared independently.
func TestBuildHostServicesDiffKeepsOwnerAwareServiceIdentity(t *testing.T) {
	fromSnapshot := &store.ManifestSnapshot{
		ID:      "linapro-demo-dynamic",
		Version: "v0.1.0",
		RequestedHostServices: []*protocol.HostServiceSpec{
			{
				Owner:   "linapro-ai-core",
				Service: "ai",
				Version: "v1",
				Methods: []string{"text.generate"},
				Resources: []*protocol.HostServiceResourceSpec{
					{Ref: "purpose:summary"},
				},
			},
			{
				Owner:   "other-ai-core",
				Service: "ai",
				Version: "v1",
				Methods: []string{"text.generate"},
			},
		},
	}
	toSnapshot := &store.ManifestSnapshot{
		ID:      "linapro-demo-dynamic",
		Version: "v0.2.0",
		RequestedHostServices: []*protocol.HostServiceSpec{
			{
				Owner:   "linapro-ai-core",
				Service: "ai",
				Version: "v1",
				Methods: []string{
					"text.generate",
					"text.method_status.get",
				},
				Resources: []*protocol.HostServiceResourceSpec{
					{Ref: "purpose:rewrite"},
					{Ref: "purpose:summary"},
				},
			},
			{
				Owner:   "other-ai-core",
				Service: "ai",
				Version: "v1",
				Methods: []string{"text.generate"},
			},
		},
	}

	diff, err := buildHostServicesDiff(fromSnapshot, toSnapshot)
	if err != nil {
		t.Fatalf("expected owner-aware host service diff to build, got %v", err)
	}
	if len(diff.Changed) != 1 {
		t.Fatalf("expected exactly one owner-aware host service change, got %#v", diff)
	}
	change := diff.Changed[0]
	if change.Owner != "linapro-ai-core" || change.Service != "ai" || change.Version != "v1" {
		t.Fatalf("expected changed owner-aware identity to be preserved, got %#v", change)
	}
	if len(change.ToMethods) != 2 {
		t.Fatalf("expected method change to stay scoped to linapro-ai-core, got %#v", change)
	}
	if change.FromResourceCount != 1 || change.ToResourceCount != 2 {
		t.Fatalf("expected resource count change to be preserved, got %#v", change)
	}
	if len(change.FromResourceRefs) != 1 || change.FromResourceRefs[0] != "purpose:summary" {
		t.Fatalf("expected from resource refs to be preserved, got %#v", change.FromResourceRefs)
	}
	if len(change.ToResourceRefs) != 2 ||
		change.ToResourceRefs[0] != "purpose:rewrite" ||
		change.ToResourceRefs[1] != "purpose:summary" {
		t.Fatalf("expected sorted to resource refs to be preserved, got %#v", change.ToResourceRefs)
	}
}
