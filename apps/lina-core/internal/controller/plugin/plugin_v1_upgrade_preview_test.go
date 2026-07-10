// This file tests plugin runtime-upgrade preview API projections.

package plugin

import (
	"testing"

	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
			RequestedHostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceData,
					Methods: []string{protocol.HostServiceMethodDataList},
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
			RequestedHostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceData,
					Methods: []string{protocol.HostServiceMethodDataList},
					Tables:  []string{"sys_plugin", "sys_plugin_release"},
				},
				{
					Owner:   "linapro-ai-core",
					Service: "ai",
					Version: "v1",
					Methods: []string{
						"text.method_status.get",
					},
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
					Owner:   "linapro-ai-core",
					Service: "ai",
					Version: "v1",
					FromMethods: []string{
						"text.generate",
					},
					ToMethods: []string{
						"text.generate",
						"text.method_status.get",
					},
					FromResourceCount: 1,
					ToResourceCount:   2,
					FromResourceRefs:  []string{"purpose:summary"},
					ToResourceRefs:    []string{"purpose:rewrite", "purpose:summary"},
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
	if len(out.ToManifest.RequestedHostServices) != 2 {
		t.Fatalf("expected requested hostServices projection, got %#v", out.ToManifest.RequestedHostServices)
	}
	dataHostService := out.ToManifest.RequestedHostServices[0]
	if len(dataHostService.TableItems) != 2 ||
		dataHostService.TableItems[1].Comment != "Plugin release" {
		t.Fatalf("expected requested hostServices with table comments, got %#v", out.ToManifest.RequestedHostServices)
	}
	ownerHostService := out.ToManifest.RequestedHostServices[1]
	if ownerHostService.Owner != "linapro-ai-core" ||
		ownerHostService.Service != "ai" ||
		ownerHostService.Version != "v1" ||
		len(ownerHostService.Methods) != 1 ||
		ownerHostService.Methods[0] != "text.method_status.get" {
		t.Fatalf("expected owner-aware requested hostServices projection, got %#v", ownerHostService)
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
	if out.HostServicesDiff.Changed[0].Owner != "linapro-ai-core" || out.HostServicesDiff.Changed[0].Version != "v1" {
		t.Fatalf("expected owner-aware hostServices diff projection, got %#v", out.HostServicesDiff.Changed[0])
	}
	if out.HostServicesDiff.Changed[0].FromResourceCount != 1 ||
		out.HostServicesDiff.Changed[0].ToResourceCount != 2 ||
		len(out.HostServicesDiff.Changed[0].ToResourceRefs) != 2 ||
		out.HostServicesDiff.Changed[0].ToResourceRefs[1] != "purpose:summary" {
		t.Fatalf("expected resource refs projection, got %#v", out.HostServicesDiff.Changed[0])
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
