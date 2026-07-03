// This file verifies startup warm-up behavior and managed watcher projection
// for single-node and clustered cron deployments.

package cron

import (
	"context"
	"testing"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	rolesvc "lina-core/internal/service/role"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestSingleNodeModeSkipsDistributedSyncCrons verifies single-node mode keeps
// distributed watcher jobs out of the projected builtin job list.
func TestSingleNodeModeSkipsDistributedSyncCrons(t *testing.T) {
	ctx := context.Background()

	svc := &serviceImpl{
		configSvc:  hostconfig.New(),
		roleSvc:    newCronRoleTestService(),
		clusterSvc: cluster.New(&hostconfig.ClusterConfig{Enabled: false}),
	}

	svc.startRuntimeParamSnapshotSync(ctx)
	svc.startAccessTopologyRevisionSync(ctx)

	jobs, err := svc.buildHostBuiltinJobs(ctx)
	if err != nil {
		t.Fatalf("build host builtin jobs: %v", err)
	}
	for _, item := range jobs {
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
		roleSvc:    newCronRoleTestService(),
		clusterSvc: cluster.New(&hostconfig.ClusterConfig{Enabled: true}),
	}

	svc.startRuntimeParamSnapshotSync(ctx)
	svc.startAccessTopologyRevisionSync(ctx)

	var (
		hasRuntimeSync bool
		hasAccessSync  bool
	)
	jobs, err := svc.buildHostBuiltinJobs(ctx)
	if err != nil {
		t.Fatalf("build host builtin jobs: %v", err)
	}
	for _, item := range jobs {
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

// newCronRoleTestService builds the explicit role dependency used by cron
// startup projection tests.
func newCronRoleTestService() rolesvc.Service {
	var (
		bizCtxSvc = bizctx.New()
		configSvc = hostconfig.New()
		i18nSvc   = i18nsvc.New(bizCtxSvc, configSvc, cachecoord.Default(nil))
		orgCapSvc = orgspi.New(nil, nil, nil)
		tenantSvc = tenantspi.New(nil, nil, nil, bizCtxSvc)
		roleSvc   = rolesvc.New(nil, bizCtxSvc, configSvc, i18nSvc, nil, tenantSvc)
	)
	roleSvc.SetDataScopeService(datascope.New(bizCtxSvc, roleSvc, orgCapSvc.Scope()))
	return roleSvc
}
