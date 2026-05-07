// This file verifies role authorization menu tree projection rules.

package menu

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/pkg/menutype"
)

// TestBuildPermissionTreeGroupsRootDynamicNodesAfterHostDirectories verifies
// root-level dynamic menus and buttons are grouped after stable host folders.
func TestBuildPermissionTreeGroupsRootDynamicNodesAfterHostDirectories(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: menuTestTranslator{
			"menu.dynamic-permissions.title":                   "动态权限",
			"menu.dynamic-permissions.route-permissions.title": "运行时路由权限",
		},
	}
	list := []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "dashboard",
			Name:    "工作台",
			Type:    menutype.Directory.String(),
			Sort:    1,
		},
		{
			Id:      2,
			MenuKey: "iam",
			Name:    "权限管理",
			Type:    menutype.Directory.String(),
			Sort:    2,
		},
		{
			Id:      10,
			MenuKey: "plugin:demo:main",
			Name:    "动态插件入口",
			Type:    menutype.Menu.String(),
			Sort:    -1,
		},
		{
			Id:      13,
			MenuKey: "plugin:demo:directory",
			Name:    "插件自有目录",
			Type:    menutype.Directory.String(),
			Sort:    -2,
		},
		{
			Id:       11,
			ParentId: 10,
			MenuKey:  "plugin:demo:perm:abc",
			Name:     "Dynamic Route Permission:plugin-demo:review:view",
			Type:     menutype.Button.String(),
			Sort:     1,
		},
		{
			Id:      12,
			MenuKey: "plugin:orphan:perm:def",
			Name:    "Dynamic Route Permission:plugin-orphan:review:check",
			Type:    menutype.Button.String(),
			Sort:    -10,
		},
	}

	tree := svc.buildPermissionTreeNodes(context.Background(), list)

	if len(tree) != 3 {
		t.Fatalf("expected two host roots plus dynamic group, got %d", len(tree))
	}
	if tree[0].Id != 1 || tree[1].Id != 2 {
		t.Fatalf("expected host directories to stay first, got root ids %d and %d", tree[0].Id, tree[1].Id)
	}

	dynamicRoot := tree[2]
	if dynamicRoot.Id >= 0 || dynamicRoot.Type != menutype.Directory.String() {
		t.Fatalf("expected synthetic dynamic directory as last root, got %#v", dynamicRoot)
	}
	if dynamicRoot.Label != "动态权限" {
		t.Fatalf("expected localized dynamic directory label, got %q", dynamicRoot.Label)
	}
	if len(dynamicRoot.Children) != 3 {
		t.Fatalf("expected dynamic menu, directory, and orphan-button menu under dynamic group, got %d", len(dynamicRoot.Children))
	}

	dynamicMenu := dynamicRoot.Children[0]
	if dynamicMenu.Id != 10 || dynamicMenu.Type != menutype.Menu.String() {
		t.Fatalf("expected root dynamic menu to become child menu, got %#v", dynamicMenu)
	}
	if len(dynamicMenu.Children) != 1 || dynamicMenu.Children[0].Id != 11 {
		t.Fatalf("expected dynamic route button to stay under its plugin menu, got %#v", dynamicMenu.Children)
	}

	dynamicDirectory := dynamicRoot.Children[1]
	if dynamicDirectory.Id != 13 || dynamicDirectory.Type != menutype.Directory.String() {
		t.Fatalf("expected root plugin directory to move under dynamic group, got %#v", dynamicDirectory)
	}

	orphanButtonMenu := dynamicRoot.Children[2]
	if orphanButtonMenu.Id >= 0 || orphanButtonMenu.Type != menutype.Menu.String() {
		t.Fatalf("expected synthetic menu for orphan buttons, got %#v", orphanButtonMenu)
	}
	if orphanButtonMenu.Label != "运行时路由权限" {
		t.Fatalf("expected localized synthetic button menu label, got %q", orphanButtonMenu.Label)
	}
	if len(orphanButtonMenu.Children) != 1 || orphanButtonMenu.Children[0].Id != 12 {
		t.Fatalf("expected orphan button under synthetic menu, got %#v", orphanButtonMenu.Children)
	}
}

// TestBuildPermissionTreeMovesDirectoryButtonsIntoSyntheticMenu verifies direct
// buttons under a directory are displayed under a menu row.
func TestBuildPermissionTreeMovesDirectoryButtonsIntoSyntheticMenu(t *testing.T) {
	svc := &serviceImpl{}
	list := []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "extension",
			Name:    "扩展中心",
			Type:    menutype.Directory.String(),
			Sort:    8,
		},
		{
			Id:       2,
			ParentId: 1,
			MenuKey:  "plugin:demo:perm:root",
			Name:     "Dynamic Route Permission:plugin-demo:review:view",
			Type:     menutype.Button.String(),
			Sort:     1,
		},
	}

	tree := svc.buildPermissionTreeNodes(context.Background(), list)

	if len(tree) != 1 {
		t.Fatalf("expected one root directory, got %d", len(tree))
	}
	root := tree[0]
	if root.Id != 1 || root.Type != menutype.Directory.String() {
		t.Fatalf("expected original directory root, got %#v", root)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected synthetic menu child for direct button, got %d", len(root.Children))
	}
	syntheticMenu := root.Children[0]
	if syntheticMenu.Id >= 0 || syntheticMenu.Type != menutype.Menu.String() {
		t.Fatalf("expected synthetic menu, got %#v", syntheticMenu)
	}
	if len(syntheticMenu.Children) != 1 || syntheticMenu.Children[0].Id != 2 {
		t.Fatalf("expected button to move under synthetic menu, got %#v", syntheticMenu.Children)
	}
}
