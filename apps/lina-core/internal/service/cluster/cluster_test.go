// This file tests top-level cluster service behavior in single-node and
// clustered modes.

package cluster

import (
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
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
	cleanupElectionLock(t)

	service := New(&config.ClusterConfig{
		Enabled: true,
		Election: config.ElectionConfig{
			Lease:         30 * time.Second,
			RenewInterval: 1 * time.Second,
		},
	})

	t.Cleanup(func() {
		service.Stop(ctx)
		cleanupElectionLock(t)
	})

	service.Start(ctx)

	if !service.IsEnabled() {
		t.Fatal("expected cluster mode to be enabled")
	}
	if !waitForPrimaryState(service, true, electionStateWait) {
		t.Fatal("expected clustered service to become primary when no competitor exists")
	}
}

// cleanupElectionLock removes the leader-election row used by cluster service
// integration tests.
func cleanupElectionLock(t *testing.T) {
	t.Helper()

	if _, err := g.DB().Model("sys_locker").Where("name", "leader-election").Delete(); err != nil {
		t.Fatalf("failed to cleanup leader-election lock: %v", err)
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
