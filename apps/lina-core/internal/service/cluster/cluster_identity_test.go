// This file tests clustered node identifier resolution.

package cluster

import "testing"

// TestGenerateNodeIdentifierUsesEnvironmentOverride verifies multi-process
// hosts can distinguish nodes even when they run on the same machine.
func TestGenerateNodeIdentifierUsesEnvironmentOverride(t *testing.T) {
	t.Setenv(nodeIDEnvName, "node-e2e-a")

	if nodeID := generateNodeIdentifier(); nodeID != "node-e2e-a" {
		t.Fatalf("expected environment node identifier, got %q", nodeID)
	}
}
