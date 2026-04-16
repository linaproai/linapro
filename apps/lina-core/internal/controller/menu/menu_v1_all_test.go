package menu

import (
	"testing"

	menusvc "lina-core/internal/service/menu"
)

func TestConvertToRouteItemsBuildsIframeRouteForHostedPluginAssets(t *testing.T) {
	routes := convertToRouteItems([]*menusvc.MenuItem{
		{
			Id:      101,
			Name:    "Runtime Iframe Entry",
			Path:    "/plugin-assets/plugin-runtime-demo/v0.1.0/index.html",
			Type:    "M",
			IsFrame: 0,
			Visible: 1,
			Status:  1,
		},
	})

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Component != "IFrameView" {
		t.Fatalf("expected iframe route component, got %s", route.Component)
	}
	if route.Meta == nil || route.Meta.IframeSrc != "/plugin-assets/plugin-runtime-demo/v0.1.0/index.html" {
		t.Fatalf("expected iframe src to be preserved, got %#v", route.Meta)
	}
	if route.Path == "/plugin-assets/plugin-runtime-demo/v0.1.0/index.html" {
		t.Fatalf("expected virtual router path instead of raw asset path, got %s", route.Path)
	}
}

func TestConvertToRouteItemsBuildsNewWindowRouteForHostedPluginAssets(t *testing.T) {
	routes := convertToRouteItems([]*menusvc.MenuItem{
		{
			Id:      102,
			Name:    "Runtime New Window Entry",
			Path:    "/plugin-assets/plugin-runtime-demo/v0.1.0/index.html",
			Type:    "M",
			IsFrame: 1,
			Visible: 1,
			Status:  1,
		},
	})

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Component != "BasicLayout" {
		t.Fatalf("expected new-window route to keep basic layout component, got %s", route.Component)
	}
	if route.Meta == nil || route.Meta.Link != "/plugin-assets/plugin-runtime-demo/v0.1.0/index.html" {
		t.Fatalf("expected link target to be preserved, got %#v", route.Meta)
	}
	if !route.Meta.OpenInNewWindow {
		t.Fatalf("expected route to open in new window")
	}
}

func TestConvertToRouteItemsBuildsEmbeddedMountRouteForHostedPluginAssets(t *testing.T) {
	routes := convertToRouteItems([]*menusvc.MenuItem{
		{
			Id:         103,
			Name:       "Runtime Embedded Entry",
			Path:       "/plugin-assets/plugin-runtime-demo/v0.1.0/mount.js",
			Component:  "system/plugin/dynamic-page",
			Type:       "M",
			IsFrame:    0,
			Visible:    1,
			Status:     1,
			QueryParam: `{"pluginAccessMode":"embedded-mount"}`,
		},
	})

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Component != "#/views/system/plugin/dynamic-page" {
		t.Fatalf("expected embedded route to keep runtime host component, got %s", route.Component)
	}
	if route.Meta == nil || route.Meta.Query == nil {
		t.Fatalf("expected embedded route query to be present, got %#v", route.Meta)
	}
	if route.Meta.Query["pluginAccessMode"] != "embedded-mount" {
		t.Fatalf("expected embedded access mode query, got %#v", route.Meta.Query)
	}
	if route.Meta.Query["embeddedSrc"] != "/plugin-assets/plugin-runtime-demo/v0.1.0/mount.js" {
		t.Fatalf("expected embedded asset url to be preserved, got %#v", route.Meta.Query)
	}
	if route.Path == "/plugin-assets/plugin-runtime-demo/v0.1.0/mount.js" {
		t.Fatalf("expected virtual host route instead of raw asset path, got %s", route.Path)
	}
}

func TestConvertToRouteItemsKeepsRegularViewRouteUnchanged(t *testing.T) {
	routes := convertToRouteItems([]*menusvc.MenuItem{
		{
			Id:        104,
			Name:      "Plugin Demo Source",
			Path:      "plugin-demo-source-sidebar-entry",
			Component: "system/plugin/dynamic-page",
			Type:      "M",
			IsFrame:   0,
			Visible:   1,
			Status:    1,
		},
	})

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Component != "#/views/system/plugin/dynamic-page" {
		t.Fatalf("expected host view component, got %s", route.Component)
	}
	if route.Meta == nil || route.Meta.IframeSrc != "" || route.Meta.Link != "" {
		t.Fatalf("expected regular route meta to stay without link semantics, got %#v", route.Meta)
	}
	if route.Path != "/plugin-demo-source-sidebar-entry" {
		t.Fatalf("expected normal menu path, got %s", route.Path)
	}
}
