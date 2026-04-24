// This file verifies the pure source-plugin upgrade status helpers and
// operator-facing pending-upgrade error rendering.

package sourceupgrade

import (
	"strings"
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

// TestBuildSourcePluginUpgradePendingErrorIncludesBulkCommand verifies the
// startup guard message includes per-plugin commands and the bulk command hint.
func TestBuildSourcePluginUpgradePendingErrorIncludesBulkCommand(t *testing.T) {
	err := buildSourcePluginUpgradePendingError([]*SourceUpgradeStatus{
		{
			PluginID:          "plugin-alpha",
			EffectiveVersion:  "v0.1.0",
			DiscoveredVersion: "v0.5.0",
		},
		{
			PluginID:          "plugin-beta",
			EffectiveVersion:  "v0.2.0",
			DiscoveredVersion: "v0.6.0",
		},
	})
	if err == nil {
		t.Fatal("expected pending upgrade error")
	}

	message := err.Error()
	expectedFragments := []string{
		"plugin=plugin-alpha current=v0.1.0 discovered=v0.5.0 command=make upgrade confirm=upgrade scope=source-plugin plugin=plugin-alpha",
		"plugin=plugin-beta current=v0.2.0 discovered=v0.6.0 command=make upgrade confirm=upgrade scope=source-plugin plugin=plugin-beta",
		"make upgrade confirm=upgrade scope=source-plugin plugin=all",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(message, fragment) {
			t.Fatalf("expected pending upgrade error to include %q, got %q", fragment, message)
		}
	}
}
