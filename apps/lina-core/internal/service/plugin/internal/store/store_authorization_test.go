// This file verifies dynamic plugin host-service authorization snapshots keep
// data-service table grants inside the current plugin namespace.

package store

import (
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestBuildAuthorizedHostServiceSpecsForPluginRejectsCoreTables verifies
// install-time authorization cannot preserve a host core table grant.
func TestBuildAuthorizedHostServiceSpecsForPluginRejectsCoreTables(t *testing.T) {
	_, err := BuildAuthorizedHostServiceSpecsForPlugin(
		"linapro-demo-dynamic",
		[]*protocol.HostServiceSpec{{
			Service: protocol.HostServiceData,
			Methods: []string{protocol.HostServiceMethodDataList},
			Tables:  []string{"sys_plugin_node_state"},
		}},
		&HostServiceAuthorizationInput{Services: []*HostServiceAuthorizationDecision{{
			Service: protocol.HostServiceData,
			Methods: []string{protocol.HostServiceMethodDataList},
			Tables:  []string{"sys_plugin_node_state"},
		}}},
	)
	if err == nil {
		t.Fatal("expected host core table authorization to be rejected")
	}
}

// TestBuildAuthorizedHostServiceSpecsForPluginKeepsOwnedTables verifies
// install-time authorization preserves only explicitly confirmed plugin-owned tables.
func TestBuildAuthorizedHostServiceSpecsForPluginKeepsOwnedTables(t *testing.T) {
	authorized, err := BuildAuthorizedHostServiceSpecsForPlugin(
		"linapro-demo-dynamic",
		[]*protocol.HostServiceSpec{{
			Service: protocol.HostServiceData,
			Methods: []string{
				protocol.HostServiceMethodDataList,
				protocol.HostServiceMethodDataUpdate,
			},
			Tables: []string{
				"plugin_linapro_demo_dynamic_record",
				"plugin_linapro_demo_dynamic_archive",
			},
		}},
		&HostServiceAuthorizationInput{Services: []*HostServiceAuthorizationDecision{{
			Service: protocol.HostServiceData,
			Methods: []string{protocol.HostServiceMethodDataList},
			Tables:  []string{"plugin_linapro_demo_dynamic_record"},
		}}},
	)
	if err != nil {
		t.Fatalf("expected plugin-owned data authorization to validate, got %v", err)
	}
	if len(authorized) != 1 {
		t.Fatalf("expected one authorized host service, got %#v", authorized)
	}
	if len(authorized[0].Methods) != 1 || authorized[0].Methods[0] != protocol.HostServiceMethodDataList {
		t.Fatalf("expected method narrowing to keep list only, got %#v", authorized[0].Methods)
	}
	if len(authorized[0].Tables) != 1 || authorized[0].Tables[0] != "plugin_linapro_demo_dynamic_record" {
		t.Fatalf("expected confirmed plugin-owned table only, got %#v", authorized[0].Tables)
	}
}

// TestBuildAuthorizedHostServiceSpecsForPluginUsesOwnerAwareIdentity verifies
// authorization decisions do not collapse plugin-owned declarations with the
// same service name.
func TestBuildAuthorizedHostServiceSpecsForPluginUsesOwnerAwareIdentity(t *testing.T) {
	authorized, err := BuildAuthorizedHostServiceSpecsForPlugin(
		"linapro-demo-dynamic",
		[]*protocol.HostServiceSpec{
			{
				Owner:   "linapro-ai-core",
				Service: "ai",
				Version: "v1",
				Methods: []string{"text.generate"},
				Resources: []*protocol.HostServiceResourceSpec{{
					Ref: "purpose:summary",
				}},
			},
			{
				Owner:   "other-ai-core",
				Service: "ai",
				Version: "v1",
				Methods: []string{"text.generate"},
				Resources: []*protocol.HostServiceResourceSpec{{
					Ref: "purpose:other",
				}},
			},
		},
		&HostServiceAuthorizationInput{Services: []*HostServiceAuthorizationDecision{{
			Owner:        " LINAPRO-AI-CORE ",
			Service:      " AI ",
			Version:      " V1 ",
			Methods:      []string{"text.generate"},
			ResourceRefs: []string{"purpose:summary"},
		}}},
	)
	if err != nil {
		t.Fatalf("expected owner-aware authorization to validate, got %v", err)
	}
	if len(authorized) != 1 {
		t.Fatalf("expected one authorized owner-aware host service, got %#v", authorized)
	}
	if authorized[0].Owner != "linapro-ai-core" || authorized[0].Service != "ai" || authorized[0].Version != "v1" {
		t.Fatalf("expected owner-aware identity to be preserved, got %#v", authorized[0])
	}
	if len(authorized[0].Resources) != 1 || authorized[0].Resources[0].Ref != "purpose:summary" {
		t.Fatalf("expected confirmed owner resource only, got %#v", authorized[0].Resources)
	}
}
