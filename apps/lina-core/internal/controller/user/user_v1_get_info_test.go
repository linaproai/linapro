// This file covers home-path resolution so login redirects prefer stable host
// pages before runtime-mounted plugin asset entries.

package user

import (
	"testing"

	"lina-core/internal/service/menu"
	"lina-core/pkg/menutype"
)

// TestResolveHomePathPrefersStableHostRoutes verifies stable workspace routes
// are chosen before hosted plugin-asset entries.
func TestResolveHomePathPrefersStableHostRoutes(t *testing.T) {
	items := []*menu.MenuItem{
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
			Path: "/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js",
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
			Path: "/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js",
			Type: menutype.Menu.String(),
		},
	}

	if got := resolveHomePath(items); got != "/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js" {
		t.Fatalf("expected hosted plugin asset fallback, got %s", got)
	}
}
