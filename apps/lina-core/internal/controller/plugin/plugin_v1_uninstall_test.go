// This file verifies plugin uninstall controller request mapping before the
// service lifecycle receives host-side uninstall policy options.

package plugin

import (
	"context"
	"testing"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
)

// pluginUninstallFakeService records uninstall calls from the controller.
type pluginUninstallFakeService struct {
	pluginsvc.Service

	calls    int
	pluginID string
	options  pluginsvc.UninstallOptions
}

// Uninstall records the supplied uninstall policy snapshot.
func (f *pluginUninstallFakeService) Uninstall(
	_ context.Context,
	pluginID string,
	options pluginsvc.UninstallOptions,
) error {
	f.calls++
	f.pluginID = pluginID
	f.options = options
	return nil
}

// pluginUninstallFakeRoleService records access-topology invalidation calls.
type pluginUninstallFakeRoleService struct {
	role.Service

	calls int
}

// NotifyAccessTopologyChanged records that plugin topology changed.
func (f *pluginUninstallFakeRoleService) NotifyAccessTopologyChanged(_ context.Context) {
	f.calls++
}

// TestUninstallMapsPurgeStorageData verifies the DELETE request flag is
// converted into the service uninstall policy before lifecycle callbacks run.
func TestUninstallMapsPurgeStorageData(t *testing.T) {
	var (
		zero = 0
		one  = 1
	)
	cases := []struct {
		name        string
		purge       *int
		force       bool
		expectPurge bool
	}{
		{name: "default purges storage data", expectPurge: true},
		{name: "query one purges storage data", purge: &one, expectPurge: true},
		{name: "query zero preserves storage data", purge: &zero, expectPurge: false},
		{name: "force preserves requested purge policy", purge: &one, force: true, expectPurge: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pluginSvc := &pluginUninstallFakeService{}
			roleSvc := &pluginUninstallFakeRoleService{}
			controller := &ControllerV1{
				pluginSvc: pluginSvc,
				roleSvc:   roleSvc,
			}

			res, err := controller.Uninstall(context.Background(), &v1.UninstallReq{
				Id:               "multi-tenant",
				PurgeStorageData: tc.purge,
				Force:            tc.force,
			})
			if err != nil {
				t.Fatalf("expected uninstall request to succeed, got error: %v", err)
			}
			if res == nil || res.Id != "multi-tenant" || res.Installed != 0 || res.Enabled != 0 {
				t.Fatalf("unexpected uninstall response: %#v", res)
			}
			if pluginSvc.calls != 1 {
				t.Fatalf("expected one uninstall service call, got %d", pluginSvc.calls)
			}
			if pluginSvc.pluginID != "multi-tenant" {
				t.Fatalf("expected plugin id multi-tenant, got %s", pluginSvc.pluginID)
			}
			if pluginSvc.options.PurgeStorageData != tc.expectPurge {
				t.Fatalf(
					"expected purgeStorageData=%v, got %v",
					tc.expectPurge,
					pluginSvc.options.PurgeStorageData,
				)
			}
			if pluginSvc.options.Force != tc.force {
				t.Fatalf("expected force=%v, got %v", tc.force, pluginSvc.options.Force)
			}
			if roleSvc.calls != 1 {
				t.Fatalf("expected one access topology notification, got %d", roleSvc.calls)
			}
		})
	}
}
