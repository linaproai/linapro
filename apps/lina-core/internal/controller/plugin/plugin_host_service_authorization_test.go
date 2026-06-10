// This file tests plugin host-service authorization response projections.

package plugin

import (
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

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
			Service: protocol.HostServiceNotifications,
			Methods: []string{protocol.HostServiceMethodNotificationsSend},
			Resources: []*protocol.HostServiceResourceSpec{
				{Ref: "inbox", Attributes: map[string]string{"channel": "inbox"}},
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

	notificationItem := items[1]
	if notificationItem.Service != protocol.HostServiceNotifications {
		t.Fatalf("expected second service to be notifications, got %s", notificationItem.Service)
	}
	if len(notificationItem.Resources) != 1 || notificationItem.Resources[0].Ref != "inbox" {
		t.Fatalf("expected notification resource ref to be projected, got %#v", notificationItem.Resources)
	}
	if notificationItem.Resources[0].Attributes["channel"] != "inbox" {
		t.Fatalf("expected notification resource attributes to be cloned, got %#v", notificationItem.Resources[0].Attributes)
	}
	specs[1].Resources[0].Attributes["channel"] = "mutated"
	if notificationItem.Resources[0].Attributes["channel"] != "inbox" {
		t.Fatalf("expected projected resource attributes to be independent, got %#v", notificationItem.Resources[0].Attributes)
	}
}
