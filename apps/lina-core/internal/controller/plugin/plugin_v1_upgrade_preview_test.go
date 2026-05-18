// This file tests plugin runtime-upgrade preview API projections.

package plugin

import (
	"testing"

	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginbridge"
)

// TestBuildUpgradePreviewResponseProjectsAllSections verifies preview DTO
// projection preserves manifest snapshots, dependency checks, SQL summary,
// hostServices diff, and risk hints.
func TestBuildUpgradePreviewResponseProjectsAllSections(t *testing.T) {
	preview := &pluginsvc.RuntimeUpgradePreview{
		PluginID:          "plugin-preview",
		RuntimeState:      pluginsvc.RuntimeUpgradeStatePendingUpgrade,
		EffectiveVersion:  "v0.1.0",
		DiscoveredVersion: "v0.2.0",
		FromManifest: &pluginsvc.RuntimeUpgradeManifestSnapshot{
			ID:                      "plugin-preview",
			Name:                    "Preview Plugin",
			Version:                 "v0.1.0",
			Type:                    "dynamic",
			ManifestDeclared:        true,
			InstallSQLCount:         1,
			HostServiceAuthRequired: true,
			RequestedHostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceData,
					Methods: []string{pluginbridge.HostServiceMethodDataList},
					Tables:  []string{"sys_plugin"},
				},
			},
		},
		ToManifest: &pluginsvc.RuntimeUpgradeManifestSnapshot{
			ID:                        "plugin-preview",
			Name:                      "Preview Plugin",
			Version:                   "v0.2.0",
			Type:                      "dynamic",
			ManifestDeclared:          true,
			InstallSQLCount:           2,
			MockSQLCount:              1,
			RuntimeFrontendAssetCount: 3,
			RuntimeSQLAssetCount:      2,
			HostServiceAuthRequired:   true,
			RequestedHostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceData,
					Methods: []string{pluginbridge.HostServiceMethodDataList},
					Tables:  []string{"sys_plugin", "sys_plugin_release"},
				},
			},
		},
		DependencyCheck: &pluginsvc.DependencyCheckResult{
			TargetID: "plugin-preview",
			Framework: pluginsvc.DependencyFrameworkCheck{
				Status: "satisfied",
			},
		},
		SQLSummary: pluginsvc.RuntimeUpgradeSQLSummary{
			InstallSQLCount:      2,
			MockSQLCount:         1,
			RuntimeSQLAssetCount: 2,
		},
		HostServicesDiff: pluginsvc.RuntimeUpgradeHostServicesDiff{
			Changed: []*pluginsvc.RuntimeUpgradeHostServiceChange{
				{
					Service:    pluginbridge.HostServiceData,
					FromTables: []string{"sys_plugin"},
					ToTables:   []string{"sys_plugin", "sys_plugin_release"},
				},
			},
			AuthorizationRequired: true,
			AuthorizationChanged:  true,
		},
		RiskHints: []string{
			pluginsvc.RuntimeUpgradeRiskHintUpgradeSQLRequiresReview,
			pluginsvc.RuntimeUpgradeRiskHintMockSQLExcluded,
		},
	}

	out := buildUpgradePreviewResponse(
		map[string]string{
			"sys_plugin":         "Plugin registry",
			"sys_plugin_release": "Plugin release",
		},
		preview,
	)
	if out == nil {
		t.Fatal("expected upgrade preview response")
	}
	if out.PluginId != "plugin-preview" || out.RuntimeState != "pending_upgrade" {
		t.Fatalf("unexpected top-level projection: %#v", out)
	}
	if out.FromManifest == nil || out.FromManifest.Version != "v0.1.0" {
		t.Fatalf("expected from manifest projection, got %#v", out.FromManifest)
	}
	if out.ToManifest == nil || out.ToManifest.Version != "v0.2.0" {
		t.Fatalf("expected to manifest projection, got %#v", out.ToManifest)
	}
	if out.ToManifest.RuntimeFrontendAssetCount != 3 {
		t.Fatalf("expected runtime frontend asset count projection, got %#v", out.ToManifest)
	}
	if len(out.ToManifest.RequestedHostServices) != 1 ||
		len(out.ToManifest.RequestedHostServices[0].TableItems) != 2 ||
		out.ToManifest.RequestedHostServices[0].TableItems[1].Comment != "Plugin release" {
		t.Fatalf("expected requested hostServices with table comments, got %#v", out.ToManifest.RequestedHostServices)
	}
	if out.DependencyCheck == nil || out.DependencyCheck.TargetId != "plugin-preview" {
		t.Fatalf("expected dependency projection, got %#v", out.DependencyCheck)
	}
	if out.SQLSummary.InstallSQLCount != 2 || out.SQLSummary.MockSQLCount != 1 {
		t.Fatalf("expected SQL summary projection, got %#v", out.SQLSummary)
	}
	if !out.HostServicesDiff.AuthorizationRequired || len(out.HostServicesDiff.Changed) != 1 {
		t.Fatalf("expected hostServices diff projection, got %#v", out.HostServicesDiff)
	}
	if len(out.RiskHints) != 2 || out.RiskHints[1] != pluginsvc.RuntimeUpgradeRiskHintMockSQLExcluded {
		t.Fatalf("expected risk hint projection, got %#v", out.RiskHints)
	}
}

// TestBuildUpgradeResponseProjectsRuntimeState verifies execution result DTO
// projection preserves version and post-upgrade runtime state metadata.
func TestBuildUpgradeResponseProjectsRuntimeState(t *testing.T) {
	result := &pluginsvc.RuntimeUpgradeResult{
		PluginID:          "plugin-upgrade",
		RuntimeState:      pluginsvc.RuntimeUpgradeStateNormal,
		EffectiveVersion:  "v0.2.0",
		DiscoveredVersion: "v0.2.0",
		FromVersion:       "v0.1.0",
		ToVersion:         "v0.2.0",
		Executed:          true,
	}

	out := buildUpgradeResponse(result)
	if out == nil {
		t.Fatal("expected upgrade response")
	}
	if out.PluginId != "plugin-upgrade" || out.RuntimeState != "normal" {
		t.Fatalf("unexpected upgrade response top-level fields: %#v", out)
	}
	if out.EffectiveVersion != "v0.2.0" || out.DiscoveredVersion != "v0.2.0" {
		t.Fatalf("expected post-upgrade versions in response, got %#v", out)
	}
	if out.FromVersion != "v0.1.0" || out.ToVersion != "v0.2.0" || !out.Executed {
		t.Fatalf("expected execution metadata in response, got %#v", out)
	}
}
