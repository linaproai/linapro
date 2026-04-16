package cluster

import (
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
)

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
	time.Sleep(200 * time.Millisecond)

	if !service.IsEnabled() {
		t.Fatal("expected cluster mode to be enabled")
	}
	if !service.IsPrimary() {
		t.Fatal("expected clustered service to become primary when no competitor exists")
	}
}

func cleanupElectionLock(t *testing.T) {
	t.Helper()

	if _, err := g.DB().Model("sys_locker").Where("name", "leader-election").Delete(); err != nil {
		t.Fatalf("failed to cleanup leader-election lock: %v", err)
	}
}
