package menu

import (
	"testing"

	"lina-core/internal/model/entity"
	menusvc "lina-core/internal/service/menu"
)

// TestConvertToRouteItemsBuildsIframeRouteForHostedPluginAssets verifies hosted
// asset menus default to iframe routes when they are not marked as new-window
// or embedded-mount entries.
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

// TestConvertToRouteItemsBuildsNewWindowRouteForHostedPluginAssets verifies
// hosted asset menus marked as frames become new-window link routes.
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

// TestConvertToRouteItemsBuildsEmbeddedMountRouteForHostedPluginAssets verifies
// embedded-mount runtime menus keep the hosted shell component and forward the
// target URL through query parameters.
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

// TestConvertToRouteItemsKeepsRegularViewRouteUnchanged verifies normal
// workspace views are not rewritten by hosted-link conversion logic.
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

// TestConvertToRouteItemsKeepsAbsoluteChildPath verifies grouped directory
// menus can keep child routes on their original absolute URLs.
func TestConvertToRouteItemsKeepsAbsoluteChildPath(t *testing.T) {
	routes := convertToRouteItems([]*menusvc.MenuItem{
		{
			Id:      201,
			Name:    "定时任务",
			Path:    "scheduled-job",
			Type:    "D",
			Visible: 1,
			Status:  1,
			Children: []*menusvc.MenuItem{
				{
					Id:        202,
					ParentId:  201,
					Name:      "任务管理",
					Path:      "/system/job",
					Component: "system/job/index",
					Type:      "M",
					Visible:   1,
					Status:    1,
				},
			},
		},
	})

	if len(routes) != 1 {
		t.Fatalf("expected 1 directory route, got %d", len(routes))
	}
	if len(routes[0].Children) != 1 {
		t.Fatalf("expected 1 child route, got %#v", routes[0].Children)
	}
	if routes[0].Children[0].Path != "/system/job" {
		t.Fatalf("expected absolute child path to be preserved, got %q", routes[0].Children[0].Path)
	}
}

// TestConvertToRouteItemsSkipsDirectoryWithoutVisibleChildren verifies host
// directory menus disappear once all child nodes are filtered out.
func TestConvertToRouteItemsSkipsDirectoryWithoutVisibleChildren(t *testing.T) {
	routes := convertToRouteItems([]*menusvc.MenuItem{
		{
			Id:      301,
			Name:    "系统监控",
			Path:    "monitor",
			Type:    "D",
			Visible: 1,
			Status:  1,
			Children: []*menusvc.MenuItem{
				{
					Id:       302,
					ParentId: 301,
					Name:     "操作日志查看",
					Path:     "monitor-operlog-view",
					Type:     "B",
					Visible:  1,
					Status:   1,
				},
			},
		},
	})

	if len(routes) != 0 {
		t.Fatalf("expected empty directory route to be hidden, got %#v", routes)
	}
}

// TestBuildFilteredTreeKeepsAncestors verifies selected leaf menus project the
// full ancestor chain required by the stable host catalog tree.
func TestBuildFilteredTreeKeepsAncestors(t *testing.T) {
	menuTree := buildFilteredTree([]*entity.SysMenu{
		{Id: 1, Name: "权限管理", Path: "iam", Type: "D", Visible: 1, Status: 1},
		{Id: 2, ParentId: 1, Name: "用户治理", Path: "iam-user", Type: "D", Visible: 1, Status: 1},
		{Id: 3, ParentId: 2, Name: "用户管理", Path: "/system/user", Component: "system/user/index", Type: "M", Visible: 1, Status: 1},
	}, []int{3})

	if len(menuTree) != 1 {
		t.Fatalf("expected one root ancestor, got %#v", menuTree)
	}
	if len(menuTree[0].Children) != 1 {
		t.Fatalf("expected one middle ancestor, got %#v", menuTree[0].Children)
	}
	if len(menuTree[0].Children[0].Children) != 1 {
		t.Fatalf("expected selected leaf to remain attached, got %#v", menuTree[0].Children[0].Children)
	}
	if menuTree[0].Children[0].Children[0].Id != 3 {
		t.Fatalf("expected selected leaf id=3, got %#v", menuTree[0].Children[0].Children[0])
	}
}
