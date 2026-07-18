// This file covers home-path resolution so login redirects prefer stable host
// pages before runtime-mounted plugin asset entries.

package user

import (
	"testing"

	"lina-core/internal/service/menu"
	"lina-core/pkg/menutype"
	"lina-core/pkg/plugin/pluginhost"
)

// TestResolveHomePathPrefersStableHostRoutes verifies stable workspace routes
// are chosen before hosted plugin-asset entries.
func TestResolveHomePathPrefersStableHostRoutes(t *testing.T) {
	items := []*menu.MenuItem{
		{
			Name:      "源码插件示例",
			Path:      "linapro-demo-source-sidebar-entry",
			Component: pluginhost.DynamicPageComponentPath,
			Type:      menutype.Menu.String(),
		},
		{
			Name: "工作台",
			Path: "dashboard",
			Type: menutype.Directory.String(),
			Children: []*menu.MenuItem{
				{
					Name: "分析页",
					Path: "analytics",
					Type: menutype.Menu.String(),
				},
				{
					Name: "工作台",
					Path: "workspace",
					Type: menutype.Menu.String(),
				},
			},
		},
		{
			Name: "动态插件示例",
			Path: "/x-assets/linapro-demo-dynamic/v0.1.0/mount.js",
			Type: menutype.Menu.String(),
		},
	}

	if got := resolveHomePath(items); got != "/dashboard/analytics" {
		t.Fatalf("expected stable host route /dashboard/analytics, got %s", got)
	}
}

// TestResolveHomePathFallsBackToHostedPluginAssetWhenNeeded verifies hosted
// plugin assets are still used when no stable host route exists.
func TestResolveHomePathFallsBackToHostedPluginAssetWhenNeeded(t *testing.T) {
	items := []*menu.MenuItem{
		{
			Name: "动态插件示例",
			Path: "/x-assets/linapro-demo-dynamic/v0.1.0/mount.js",
			Type: menutype.Menu.String(),
		},
	}

	if got := resolveHomePath(items); got != "/x-assets/linapro-demo-dynamic/v0.1.0/mount.js" {
		t.Fatalf("expected hosted plugin asset fallback, got %s", got)
	}
}

// TestResolveHomePathSkipsMissingWorkbench verifies homePath lands on the next
// navigable menu when the workbench/dashboard tree is absent (disabled).
func TestResolveHomePathSkipsMissingWorkbench(t *testing.T) {
	items := []*menu.MenuItem{
		{
			Name: "系统管理",
			Path: "system",
			Type: menutype.Directory.String(),
			Children: []*menu.MenuItem{
				{
					Name: "用户管理",
					Path: "user",
					Type: menutype.Menu.String(),
				},
				{
					Name: "角色管理",
					Path: "role",
					Type: menutype.Menu.String(),
				},
			},
		},
	}

	if got := resolveHomePath(items); got != "/system/user" {
		t.Fatalf("expected first navigable menu /system/user, got %s", got)
	}
}

// TestResolveHomePathEmptyTreeUsesProfile verifies the safe fallback when the
// user has no navigable internal menus.
func TestResolveHomePathEmptyTreeUsesProfile(t *testing.T) {
	if got := resolveHomePath(nil); got != "/profile" {
		t.Fatalf("expected /profile fallback, got %s", got)
	}
}
