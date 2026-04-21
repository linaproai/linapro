// This file verifies that distributed cache-sync cron jobs stay disabled in
// single-node mode while their startup warm-up still succeeds.

package cron

import (
	"context"
	"testing"

	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	rolesvc "lina-core/internal/service/role"
)

// TestSingleNodeModeSkipsDistributedSyncCrons verifies single-node mode keeps
// distributed watcher jobs out of the projected builtin job list.
func TestSingleNodeModeSkipsDistributedSyncCrons(t *testing.T) {
	ctx := context.Background()

	svc := &serviceImpl{
		configSvc:  hostconfig.New(),
		roleSvc:    rolesvc.New(nil),
		clusterSvc: cluster.New(&hostconfig.ClusterConfig{Enabled: false}),
	}
	svc.runtimeParamSyncJob = newRuntimeParamSnapshotSyncJob(false, svc.configSvc)
	svc.accessTopologySyncJob = newAccessTopologyRevisionSyncJob(false, svc.roleSvc)

	svc.startRuntimeParamSnapshotSync(ctx)
	svc.startAccessTopologyRevisionSync(ctx)

	for _, item := range svc.buildHostBuiltinJobs() {
		if item.HandlerRef == "host:runtime-param-sync" {
			t.Fatalf("expected runtime param sync watcher to stay hidden in single-node mode, got %#v", item)
		}
		if item.HandlerRef == "host:access-topology-sync" {
			t.Fatalf("expected access topology sync watcher to stay hidden in single-node mode, got %#v", item)
		}
	}
}

// TestClusterModeRegistersDistributedSyncCrons verifies clustered startup
// projects both distributed watcher jobs into the builtin job set.
func TestClusterModeRegistersDistributedSyncCrons(t *testing.T) {
	ctx := context.Background()

	svc := &serviceImpl{
		configSvc:  hostconfig.New(),
		roleSvc:    rolesvc.New(nil),
		clusterSvc: cluster.New(&hostconfig.ClusterConfig{Enabled: true}),
	}
	svc.runtimeParamSyncJob = newRuntimeParamSnapshotSyncJob(true, svc.configSvc)
	svc.accessTopologySyncJob = newAccessTopologyRevisionSyncJob(true, svc.roleSvc)

	svc.startRuntimeParamSnapshotSync(ctx)
	svc.startAccessTopologyRevisionSync(ctx)

	var (
		hasRuntimeSync bool
		hasAccessSync  bool
	)
	for _, item := range svc.buildHostBuiltinJobs() {
		if item.HandlerRef == "host:runtime-param-sync" {
			hasRuntimeSync = true
		}
		if item.HandlerRef == "host:access-topology-sync" {
			hasAccessSync = true
		}
	}
	if !hasRuntimeSync {
		t.Fatal("expected runtime param sync watcher to be projected in cluster mode")
	}
	if !hasAccessSync {
		t.Fatal("expected access topology sync watcher to be projected in cluster mode")
	}
}
