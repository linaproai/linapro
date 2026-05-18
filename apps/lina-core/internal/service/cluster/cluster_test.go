// This file tests top-level cluster service behavior in single-node and
// clustered modes.

package cluster

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
)

// TestServiceDisabledTreatsCurrentNodeAsPrimary verifies single-node mode keeps
// the local node primary without starting election infrastructure.
func TestServiceDisabledTreatsCurrentNodeAsPrimary(t *testing.T) {
	service := New(&config.ClusterConfig{Enabled: false})
	ctx := context.Background()

	if service.IsEnabled() {
		t.Fatal("expected cluster mode to be disabled")
	}
	if !service.IsPrimary() {
		t.Fatal("expected single-node mode to treat current node as primary")
	}

	service.Start(ctx)
	service.Stop(ctx)
}

// TestServiceEnabledStartsPrimaryElection verifies enabling cluster mode starts
// election and promotes the current node when no competitor exists.
func TestServiceEnabledStartsPrimaryElection(t *testing.T) {
	ctx := context.Background()

	service := NewWithCoordination(&config.ClusterConfig{
		Enabled: true,
		Election: config.ElectionConfig{
			Lease:         30 * time.Second,
			RenewInterval: 1 * time.Second,
		},
	}, coordination.NewMemory(nil))

	t.Cleanup(func() {
		service.Stop(ctx)
	})

	service.Start(ctx)

	if !service.IsEnabled() {
		t.Fatal("expected cluster mode to be enabled")
	}
	if !waitForPrimaryState(service, true, electionStateWait) {
		t.Fatal("expected clustered service to become primary when no competitor exists")
	}
}

// waitForPrimaryState polls the cluster service primary projection until the
// expected state becomes visible or the bounded timeout expires.
func waitForPrimaryState(service Service, expected bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if service.IsPrimary() == expected {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return service.IsPrimary() == expected
}
