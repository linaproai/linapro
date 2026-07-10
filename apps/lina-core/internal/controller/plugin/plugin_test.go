// This file consolidates controller-local plugin projection unit tests.

package plugin

import (
	"net/http"
	"testing"

	v1 "lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestBuildPluginDependencyCheckResultProjectsAllSections verifies the
// dependency projection preserves hard dependency checks, blockers, cycles,
// reverse dependents, and reverse blockers.
func TestBuildPluginDependencyCheckResultProjectsAllSections(t *testing.T) {
	out := buildPluginDependencyCheckResult(&pluginsvc.DependencyCheckResult{
		TargetID: "plugin-consumer",
		Framework: pluginsvc.DependencyFrameworkCheck{
			RequiredVersion: ">=0.1.0",
			CurrentVersion:  "v0.1.0",
			Status:          "satisfied",
		},
		Dependencies: []*pluginsvc.DependencyPluginCheck{
			{
				OwnerID:         "plugin-consumer",
				DependencyID:    "plugin-base",
				DependencyName:  "Plugin Base",
				RequiredVersion: ">=0.1.0",
				CurrentVersion:  "v0.1.0",
				Installed:       false,
				Discovered:      true,
				Status:          "missing",
				Chain:           []string{"plugin-consumer", "plugin-base"},
			},
		},
		Blockers: []*pluginsvc.DependencyBlocker{
			{
				Code:            "dependency_missing",
				PluginID:        "plugin-consumer",
				DependencyID:    "plugin-missing",
				RequiredVersion: ">=0.1.0",
				CurrentVersion:  "",
				Chain:           []string{"plugin-consumer", "plugin-missing"},
				Detail:          "missing",
			},
		},
		Cycle: []string{"a", "b", "a"},
		ReverseDependents: []*pluginsvc.DependencyReverseDependent{
			{
				PluginID:        "plugin-downstream",
				Name:            "Plugin Downstream",
				Version:         "v0.2.0",
				RequiredVersion: "<0.3.0",
				OwnerHostServices: []*pluginsvc.DependencyOwnerHostServiceSummary{
					{
						Owner:   "linapro-ai-core",
						Service: "ai",
						Version: "v1",
						Methods: []string{"text.generate"},
					},
				},
			},
		},
		ReverseBlockers: []*pluginsvc.DependencyBlocker{
			{
				Code:         "dependency_snapshot_unknown",
				PluginID:     "plugin-unknown",
				DependencyID: "",
				Detail:       "unknown snapshot",
			},
		},
	})

	if out == nil {
		t.Fatal("expected projected dependency result")
	}
	if out.TargetId != "plugin-consumer" || out.Framework.Status != "satisfied" {
		t.Fatalf("unexpected top-level projection: %#v", out)
	}
	if len(out.Dependencies) != 1 || out.Dependencies[0].DependencyId != "plugin-base" {
		t.Fatalf("expected dependency edge projection, got %#v", out.Dependencies)
	}
	if len(out.Blockers) != 1 || out.Blockers[0].Code != "dependency_missing" {
		t.Fatalf("expected blocker projection, got %#v", out.Blockers)
	}
	if len(out.Cycle) != 3 || out.Cycle[2] != "a" {
		t.Fatalf("expected cycle projection, got %#v", out.Cycle)
	}
	if len(out.ReverseDependents) != 1 || out.ReverseDependents[0].PluginId != "plugin-downstream" {
		t.Fatalf("expected reverse dependent projection, got %#v", out.ReverseDependents)
	}
	if len(out.ReverseDependents[0].OwnerHostServices) != 1 ||
		out.ReverseDependents[0].OwnerHostServices[0].Methods[0] != "text.generate" {
		t.Fatalf("expected owner host service projection, got %#v", out.ReverseDependents[0].OwnerHostServices)
	}
	if len(out.ReverseBlockers) != 1 || out.ReverseBlockers[0].Code != "dependency_snapshot_unknown" {
		t.Fatalf("expected reverse blocker projection, got %#v", out.ReverseBlockers)
	}
}

// TestBuildHostServicePermissionItemsProjectsTablesAndResources verifies the
// authorization view enriches data tables and preserves governed resources.
func TestBuildHostServicePermissionItemsProjectsTablesAndResources(t *testing.T) {
	specs := []*protocol.HostServiceSpec{
		{
			Service: protocol.HostServiceData,
			Methods: []string{protocol.HostServiceMethodDataList},
			Tables:  []string{"plugin_linapro_demo_dynamic_record"},
		},
		{
			Owner:   "linapro-ai-core",
			Service: "ai",
			Version: "v1",
			Methods: []string{"text.generate"},
			Resources: []*protocol.HostServiceResourceSpec{
				{Ref: "purpose:summary", Attributes: map[string]string{"purpose": "summary"}},
			},
		},
	}

	items := buildHostServicePermissionItems(
		specs,
		map[string]string{"plugin_linapro_demo_dynamic_record": "Dynamic plugin record table"},
	)
	if len(items) != 2 {
		t.Fatalf("expected 2 host service items, got %d", len(items))
	}

	dataItem := items[0]
	if dataItem.Service != protocol.HostServiceData {
		t.Fatalf("expected first service to be data, got %s", dataItem.Service)
	}
	if len(dataItem.TableItems) != 1 {
		t.Fatalf("expected 1 table item, got %d", len(dataItem.TableItems))
	}
	if dataItem.TableItems[0].Comment != "Dynamic plugin record table" {
		t.Fatalf("expected table comment to be preserved, got %s", dataItem.TableItems[0].Comment)
	}

	ownerItem := items[1]
	if ownerItem.Owner != "linapro-ai-core" || ownerItem.Service != "ai" || ownerItem.Version != "v1" {
		t.Fatalf("expected second service identity to be preserved, got %#v", ownerItem)
	}
	if len(ownerItem.Resources) != 1 || ownerItem.Resources[0].Ref != "purpose:summary" {
		t.Fatalf("expected owner resource ref to be projected, got %#v", ownerItem.Resources)
	}
	if ownerItem.Resources[0].Attributes["purpose"] != "summary" {
		t.Fatalf("expected owner resource attributes to be cloned, got %#v", ownerItem.Resources[0].Attributes)
	}
	specs[1].Resources[0].Attributes["purpose"] = "mutated"
	if ownerItem.Resources[0].Attributes["purpose"] != "summary" {
		t.Fatalf("expected projected resource attributes to be independent, got %#v", ownerItem.Resources[0].Attributes)
	}
}

// TestBuildAuthorizationInputPreservesOwnerAwareIdentity verifies API
// authorization payloads retain plugin-owned identity fields.
func TestBuildAuthorizationInputPreservesOwnerAwareIdentity(t *testing.T) {
	input := buildAuthorizationInput(&v1.HostServiceAuthorizationReq{
		Services: []*v1.HostServiceAuthorizationServiceReq{{
			Owner:        " linapro-ai-core ",
			Service:      " ai ",
			Version:      " v1 ",
			Methods:      []string{"text.generate"},
			ResourceRefs: []string{"purpose:summary"},
		}},
	})

	if input == nil || len(input.Services) != 1 {
		t.Fatalf("expected one authorization service input, got %#v", input)
	}
	decision := input.Services[0]
	if decision.Owner != "linapro-ai-core" || decision.Service != "ai" || decision.Version != "v1" {
		t.Fatalf("expected owner-aware authorization identity to be trimmed, got %#v", decision)
	}
	if len(decision.ResourceRefs) != 1 || decision.ResourceRefs[0] != "purpose:summary" {
		t.Fatalf("expected resource refs to be copied, got %#v", decision.ResourceRefs)
	}
}

// TestBuildPluginRouteReviewItemsBuildsPublicRouteMetadata verifies current
// release route contracts are projected with host-visible public paths and
// review-friendly access metadata.
func TestBuildPluginRouteReviewItemsBuildsPublicRouteMetadata(t *testing.T) {
	items := buildPluginRouteReviewItems(
		"plugin-dev-route-review",
		[]*protocol.RouteContract{
			{
				Method:      http.MethodGet,
				Path:        "/api/v1/review-summary",
				Access:      protocol.AccessLogin,
				Permission:  "plugin-dev-route-review:review:query",
				Summary:     "查询评审摘要",
				Description: "返回插件当前评审摘要。",
			},
			nil,
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/public-ping",
				Access:  protocol.AccessPublic,
				Summary: "公开探活",
			},
		},
	)

	if len(items) != 2 {
		t.Fatalf("expected 2 projected route items, got %d", len(items))
	}

	if items[0].Method != http.MethodGet {
		t.Fatalf("expected first route method GET, got %s", items[0].Method)
	}
	if items[0].PublicPath != "/x/plugin-dev-route-review/api/v1/review-summary" {
		t.Fatalf("unexpected first route public path: %s", items[0].PublicPath)
	}
	if items[0].Access != protocol.AccessLogin {
		t.Fatalf("expected first route access login, got %s", items[0].Access)
	}
	if items[0].Permission != "plugin-dev-route-review:review:query" {
		t.Fatalf("unexpected first route permission: %s", items[0].Permission)
	}
	if items[0].Summary != "查询评审摘要" {
		t.Fatalf("unexpected first route summary: %s", items[0].Summary)
	}
	if items[0].Description != "返回插件当前评审摘要。" {
		t.Fatalf("unexpected first route description: %s", items[0].Description)
	}

	if items[1].Method != http.MethodPost {
		t.Fatalf("expected second route method POST, got %s", items[1].Method)
	}
	if items[1].PublicPath != "/x/plugin-dev-route-review/api/v1/public-ping" {
		t.Fatalf("unexpected second route public path: %s", items[1].PublicPath)
	}
	if items[1].Access != protocol.AccessPublic {
		t.Fatalf("expected second route access public, got %s", items[1].Access)
	}
	if items[1].Permission != "" {
		t.Fatalf("expected public route permission to stay empty, got %s", items[1].Permission)
	}
}

// TestBuildPluginRouteReviewItemsPreservesPluginOwnedPathContent verifies
// route review public paths only force `/x/{pluginId}` and preserve plugin-local
// path content such as `/api/v2`, `/interface/m1`, and `/graphql`.
func TestBuildPluginRouteReviewItemsPreservesPluginOwnedPathContent(t *testing.T) {
	items := buildPluginRouteReviewItems(
		"plugin-dev-route-review",
		[]*protocol.RouteContract{
			{
				Method: http.MethodGet,
				Path:   "/api/v2/review-summary",
				Access: protocol.AccessLogin,
			},
			{
				Method: http.MethodPost,
				Path:   "/interface/m1/review-summary",
				Access: protocol.AccessLogin,
			},
			{
				Method: http.MethodPost,
				Path:   "/graphql",
				Access: protocol.AccessPublic,
			},
		},
	)

	expected := []string{
		"/x/plugin-dev-route-review/api/v2/review-summary",
		"/x/plugin-dev-route-review/interface/m1/review-summary",
		"/x/plugin-dev-route-review/graphql",
	}
	if len(items) != len(expected) {
		t.Fatalf("expected %d projected route items, got %d", len(expected), len(items))
	}
	for index, expectedPath := range expected {
		if items[index].PublicPath != expectedPath {
			t.Fatalf("expected item %d public path %s, got %s", index, expectedPath, items[index].PublicPath)
		}
	}
}

// TestBuildPluginRouteReviewItemsSkipsBlankPluginID verifies blank plugin IDs
// do not produce invalid review items.
func TestBuildPluginRouteReviewItemsSkipsBlankPluginID(t *testing.T) {
	items := buildPluginRouteReviewItems(
		" ",
		[]*protocol.RouteContract{
			{
				Method: http.MethodGet,
				Path:   "/api/v1/review-summary",
			},
		},
	)

	if items != nil {
		t.Fatalf("expected nil items for blank plugin ID, got %#v", items)
	}
}
