// This file verifies the pure source-plugin upgrade status helpers.

package sourceupgrade

import (
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// TestBuildSourceUpgradeStatusMarksPendingUpgrade verifies installed source
// plugins report a pending upgrade when discovery finds a higher version.
func TestBuildSourceUpgradeStatusMarksPendingUpgrade(t *testing.T) {
	status, err := buildSourceUpgradeStatus(
		&catalog.Manifest{
			ID:      "plugin-source-upgrade-status",
			Name:    "Source Upgrade Status Plugin",
			Version: "v0.5.0",
		},
		&entity.SysPlugin{
			PluginId:  "plugin-source-upgrade-status",
			Name:      "Source Upgrade Status Plugin",
			Version:   "v0.1.0",
			Installed: catalog.InstalledYes,
			Status:    catalog.StatusEnabled,
		},
	)
	if err != nil {
		t.Fatalf("expected source upgrade status build to succeed, got error: %v", err)
	}
	if status == nil {
		t.Fatal("expected source upgrade status result")
	}
	if !status.NeedsUpgrade {
		t.Fatalf("expected pending upgrade status, got %#v", status)
	}
	if status.DowngradeDetected {
		t.Fatalf("expected no downgrade flag for higher discovered version, got %#v", status)
	}
}

// TestBuildSourceUpgradeStatusKeepsUninstalledPluginsOutOfUpgradeFlow verifies
// discovery drift does not mark an uninstalled source plugin as pending upgrade.
func TestBuildSourceUpgradeStatusKeepsUninstalledPluginsOutOfUpgradeFlow(t *testing.T) {
	status, err := buildSourceUpgradeStatus(
		&catalog.Manifest{
			ID:      "plugin-source-upgrade-uninstalled",
			Name:    "Source Upgrade Uninstalled Plugin",
			Version: "v0.5.0",
		},
		&entity.SysPlugin{
			PluginId:  "plugin-source-upgrade-uninstalled",
			Name:      "Source Upgrade Uninstalled Plugin",
			Version:   "v0.1.0",
			Installed: catalog.InstalledNo,
			Status:    catalog.StatusDisabled,
		},
	)
	if err != nil {
		t.Fatalf("expected uninstalled source upgrade status build to succeed, got error: %v", err)
	}
	if status == nil {
		t.Fatal("expected source upgrade status result")
	}
	if status.NeedsUpgrade {
		t.Fatalf("expected uninstalled plugin not to require upgrade, got %#v", status)
	}
	if status.DowngradeDetected {
		t.Fatalf("expected uninstalled plugin not to report downgrade detection, got %#v", status)
	}
}

// TestBuildSourceUpgradeStatusMarksDowngrade verifies discovery of a lower
// version is surfaced as an unsupported downgrade signal.
func TestBuildSourceUpgradeStatusMarksDowngrade(t *testing.T) {
	status, err := buildSourceUpgradeStatus(
		&catalog.Manifest{
			ID:      "plugin-source-upgrade-downgrade",
			Name:    "Source Upgrade Downgrade Plugin",
			Version: "v0.1.0",
		},
		&entity.SysPlugin{
			PluginId:  "plugin-source-upgrade-downgrade",
			Name:      "Source Upgrade Downgrade Plugin",
			Version:   "v0.5.0",
			Installed: catalog.InstalledYes,
			Status:    catalog.StatusEnabled,
		},
	)
	if err != nil {
		t.Fatalf("expected downgrade status build to succeed, got error: %v", err)
	}
	if status == nil {
		t.Fatal("expected source upgrade status result")
	}
	if status.NeedsUpgrade {
		t.Fatalf("expected downgrade case not to mark upgrade-needed, got %#v", status)
	}
	if !status.DowngradeDetected {
		t.Fatalf("expected downgrade detection flag, got %#v", status)
	}
}

// TestSourceManifestSnapshotViewPublishesTypedSnapshot verifies catalog
// snapshots are projected through the pluginhost typed manifest contract.
func TestSourceManifestSnapshotViewPublishesTypedSnapshot(t *testing.T) {
	view := sourceManifestSnapshotView(&catalog.ManifestSnapshot{
		ID:                      "plugin-source-upgrade-snapshot",
		Name:                    "Source Upgrade Snapshot Plugin",
		Version:                 "v1.2.3",
		Type:                    "source",
		ScopeNature:             "tenant_aware",
		SupportsMultiTenant:     true,
		DefaultInstallMode:      "tenant_scoped",
		Description:             "Snapshot projection test",
		InstallSQLCount:         3,
		UninstallSQLCount:       2,
		MockSQLCount:            1,
		MenuCount:               4,
		BackendHookCount:        5,
		ResourceSpecCount:       6,
		HostServiceAuthRequired: true,
	})
	if view == nil {
		t.Fatal("expected projected manifest snapshot")
	}
	if view.ID() != "plugin-source-upgrade-snapshot" ||
		view.Name() != "Source Upgrade Snapshot Plugin" ||
		view.Version() != "v1.2.3" ||
		view.Type() != "source" {
		t.Fatalf("expected typed getters to expose core fields, got %#v", view)
	}

	values := view.Values()
	if values["scopeNature"] != "tenant_aware" ||
		values["supportsMultiTenant"] != true ||
		values["defaultInstallMode"] != "tenant_scoped" ||
		values["description"] != "Snapshot projection test" ||
		values["installSqlCount"] != 3 ||
		values["uninstallSqlCount"] != 2 ||
		values["mockSqlCount"] != 1 ||
		values["menuCount"] != 4 ||
		values["backendHookCount"] != 5 ||
		values["resourceSpecCount"] != 6 ||
		values["hostServiceAuthNeeded"] != true {
		t.Fatalf("expected all published snapshot values, got %#v", values)
	}
}
