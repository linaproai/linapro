// This file verifies that distributed cache-sync cron jobs stay disabled in
// single-node mode while their startup warm-up still succeeds.

package cron

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	rolesvc "lina-core/internal/service/role"
)

// TestSingleNodeModeSkipsDistributedSyncCrons verifies startup warm-up does not
// register distributed watcher crons when cluster mode is disabled.
func TestSingleNodeModeSkipsDistributedSyncCrons(t *testing.T) {
	ctx := context.Background()
	gcron.Remove(CronRuntimeParamSync)
	gcron.Remove(CronAccessTopologySync)
	t.Cleanup(func() {
		gcron.Remove(CronRuntimeParamSync)
		gcron.Remove(CronAccessTopologySync)
	})

	svc := &serviceImpl{
		configSvc:  hostconfig.New(),
		roleSvc:    rolesvc.New(),
		clusterSvc: cluster.New(&hostconfig.ClusterConfig{Enabled: false}),
	}
	svc.runtimeParamSyncJob = newRuntimeParamSnapshotSyncJob(false, svc.configSvc)
	svc.accessTopologySyncJob = newAccessTopologyRevisionSyncJob(false, svc.roleSvc)

	svc.startRuntimeParamSnapshotSync(ctx)
	svc.startAccessTopologyRevisionSync(ctx)

	if entry := gcron.Search(CronRuntimeParamSync); entry != nil {
		t.Fatalf("expected runtime param sync cron to stay disabled in single-node mode, got %#v", entry)
	}
	if entry := gcron.Search(CronAccessTopologySync); entry != nil {
		t.Fatalf("expected access topology sync cron to stay disabled in single-node mode, got %#v", entry)
	}
}

// TestClusterModeRegistersDistributedSyncCrons verifies clustered startup
// registers both distributed watcher crons.
func TestClusterModeRegistersDistributedSyncCrons(t *testing.T) {
	ctx := context.Background()
	gcron.Remove(CronRuntimeParamSync)
	gcron.Remove(CronAccessTopologySync)
	t.Cleanup(func() {
		gcron.Remove(CronRuntimeParamSync)
		gcron.Remove(CronAccessTopologySync)
	})

	svc := &serviceImpl{
		configSvc:  hostconfig.New(),
		roleSvc:    rolesvc.New(),
		clusterSvc: cluster.New(&hostconfig.ClusterConfig{Enabled: true}),
	}
	svc.runtimeParamSyncJob = newRuntimeParamSnapshotSyncJob(true, svc.configSvc)
	svc.accessTopologySyncJob = newAccessTopologyRevisionSyncJob(true, svc.roleSvc)

	svc.startRuntimeParamSnapshotSync(ctx)
	svc.startAccessTopologyRevisionSync(ctx)

	if entry := gcron.Search(CronRuntimeParamSync); entry == nil {
		t.Fatal("expected runtime param sync cron to be registered in cluster mode")
	}
	if entry := gcron.Search(CronAccessTopologySync); entry == nil {
		t.Fatal("expected access topology sync cron to be registered in cluster mode")
	}
}
