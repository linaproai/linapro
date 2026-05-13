// This file verifies monitor-server plugin callback wiring helpers.

package backend

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/pluginhost"
	pluginconfig "lina-core/pkg/pluginservice/config"
	plugincontract "lina-core/pkg/pluginservice/contract"
	monitorsvc "lina-plugin-monitor-server/backend/internal/service/monitor"
)

// fakeCronRegistrar provides the node role needed by cleanup callback tests.
type fakeCronRegistrar struct {
	// primary reports whether the current test registrar is the primary node.
	primary bool
	// hostServices exposes fake host-published services to callbacks.
	hostServices pluginhost.HostServices
}

// Add satisfies pluginhost.CronRegistrar for tests.
func (r *fakeCronRegistrar) Add(
	ctx context.Context,
	pattern string,
	name string,
	handler pluginhost.CronJobHandler,
) error {
	return nil
}

// AddWithMetadata satisfies pluginhost.CronRegistrar for tests.
func (r *fakeCronRegistrar) AddWithMetadata(
	ctx context.Context,
	pattern string,
	name string,
	displayName string,
	description string,
	handler pluginhost.CronJobHandler,
) error {
	return nil
}

// IsPrimaryNode reports the configured test node role.
func (r *fakeCronRegistrar) IsPrimaryNode() bool {
	return r.primary
}

// HostServices returns fake host-published services for callback tests.
func (r *fakeCronRegistrar) HostServices() pluginhost.HostServices {
	if r == nil {
		return nil
	}
	return r.hostServices
}

// fakeHostServices publishes only the config dependency used by this test file.
type fakeHostServices struct {
	configSvc plugincontract.ConfigService
}

// APIDoc returns no apidoc service.
func (s fakeHostServices) APIDoc() plugincontract.APIDocService { return nil }

// Auth returns no auth service.
func (s fakeHostServices) Auth() plugincontract.AuthService { return nil }

// BizCtx returns no bizctx service.
func (s fakeHostServices) BizCtx() plugincontract.BizCtxService { return nil }

// Config returns the configured fake config service.
func (s fakeHostServices) Config() plugincontract.ConfigService { return s.configSvc }

// I18n returns no i18n service.
func (s fakeHostServices) I18n() plugincontract.I18nService { return nil }

// Notify returns no notify service.
func (s fakeHostServices) Notify() plugincontract.NotifyService { return nil }

// PluginState returns no plugin-state service.
func (s fakeHostServices) PluginState() plugincontract.PluginStateService { return nil }

// Route returns no route service.
func (s fakeHostServices) Route() plugincontract.RouteService { return nil }

// Session returns no session service.
func (s fakeHostServices) Session() plugincontract.SessionService { return nil }

// TenantFilter returns no tenant-filter service.
func (s fakeHostServices) TenantFilter() plugincontract.TenantFilterService { return nil }

// fakeMonitorService records callback usage without touching the database or host metrics.
type fakeMonitorService struct {
	// collected reports whether CollectAndStore was called.
	collected bool
	// cleanupCalled reports whether CleanupStale was called.
	cleanupCalled bool
	// cleanupThreshold records the threshold passed to CleanupStale.
	cleanupThreshold time.Duration
}

// CollectAndStore records one collection callback.
func (s *fakeMonitorService) CollectAndStore(ctx context.Context) {
	s.collected = true
}

// Collect satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) Collect(ctx context.Context) *monitorsvc.MonitorData {
	return nil
}

// GetDBInfo satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) GetDBInfo(ctx context.Context) *monitorsvc.DBInfo {
	return nil
}

// GetLatest satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) GetLatest(ctx context.Context, nodeName string) ([]*monitorsvc.NodeMonitorData, error) {
	return nil, nil
}

// CleanupStale records one cleanup callback.
func (s *fakeMonitorService) CleanupStale(ctx context.Context, threshold time.Duration) (int64, error) {
	s.cleanupCalled = true
	s.cleanupThreshold = threshold
	return 0, nil
}

// TestCollectSnapshotUsesInjectedService verifies cron collection reuses the provided service instance.
func TestCollectSnapshotUsesInjectedService(t *testing.T) {
	monitorSvc := &fakeMonitorService{}

	if err := collectSnapshot(context.Background(), monitorSvc); err != nil {
		t.Fatalf("collect snapshot: %v", err)
	}

	if !monitorSvc.collected {
		t.Fatal("expected injected monitor service to collect")
	}
}

// TestCleanupSnapshotsSkipsNonPrimaryNode verifies cleanup is skipped outside the primary node.
func TestCleanupSnapshotsSkipsNonPrimaryNode(t *testing.T) {
	monitorSvc := &fakeMonitorService{}
	registrar := &fakeCronRegistrar{primary: false}

	if err := cleanupSnapshots(context.Background(), registrar, monitorSvc); err != nil {
		t.Fatalf("cleanup snapshots: %v", err)
	}

	if monitorSvc.cleanupCalled {
		t.Fatal("expected non-primary node to skip cleanup")
	}
}

// TestCleanupSnapshotsUsesInjectedServiceOnPrimaryNode verifies cleanup uses the shared service instance.
func TestCleanupSnapshotsUsesInjectedServiceOnPrimaryNode(t *testing.T) {
	setPluginTestConfigAdapter(t, `
monitor:
  interval: 30s
  retentionMultiplier: 4
`)

	monitorSvc := &fakeMonitorService{}
	registrar := &fakeCronRegistrar{
		primary: true,
		hostServices: fakeHostServices{
			configSvc: pluginconfig.New(),
		},
	}

	if err := cleanupSnapshots(context.Background(), registrar, monitorSvc); err != nil {
		t.Fatalf("cleanup snapshots: %v", err)
	}

	if !monitorSvc.cleanupCalled {
		t.Fatal("expected injected monitor service to clean up")
	}
	if monitorSvc.cleanupThreshold != 2*time.Minute {
		t.Fatalf("expected cleanup threshold 2m, got %s", monitorSvc.cleanupThreshold)
	}
}

// setPluginTestConfigAdapter swaps the process config adapter for one test case.
func setPluginTestConfigAdapter(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create content adapter: %v", err)
	}

	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)

	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})
}
