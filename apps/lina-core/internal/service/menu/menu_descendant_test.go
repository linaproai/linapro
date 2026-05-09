// This file verifies menu hierarchy descendant detection.

package menu

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

// TestIsDescendantUsesInMemoryHierarchy verifies the optimized descendant
// lookup handles direct, cross-level, self, sibling, and missing-target cases.
func TestIsDescendantUsesInMemoryHierarchy(t *testing.T) {
	ctx := context.Background()
	svc := &serviceImpl{}
	prefix := fmt.Sprintf("menu-descendant-%d", time.Now().UnixNano())

	rootID := insertTestMenu(t, ctx, prefix+"-root", 0)
	childID := insertTestMenu(t, ctx, prefix+"-child", rootID)
	grandchildID := insertTestMenu(t, ctx, prefix+"-grandchild", childID)
	siblingID := insertTestMenu(t, ctx, prefix+"-sibling", rootID)
	t.Cleanup(func() {
		if _, err := dao.SysMenu.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysMenu.Columns().Id, []int{rootID, childID, grandchildID, siblingID}).
			Delete(); err != nil {
			t.Fatalf("cleanup menu hierarchy: %v", err)
		}
	})

	if !svc.isDescendant(ctx, rootID, childID) {
		t.Fatal("expected child to be descendant of root")
	}
	if !svc.isDescendant(ctx, rootID, grandchildID) {
		t.Fatal("expected grandchild to be descendant of root")
	}
	if svc.isDescendant(ctx, rootID, rootID) {
		t.Fatal("expected menu to not be descendant of itself")
	}
	if svc.isDescendant(ctx, childID, siblingID) {
		t.Fatal("expected sibling to not be descendant of child")
	}
	if svc.isDescendant(ctx, rootID, grandchildID+1000000) {
		t.Fatal("expected missing target to not be reported as descendant")
	}
}

// insertTestMenu inserts one minimal menu row for hierarchy tests.
func insertTestMenu(t *testing.T, ctx context.Context, key string, parentID int) int {
	t.Helper()

	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId: parentID,
		MenuKey:  key,
		Name:     key,
		Type:     "M",
		Sort:     1,
		Visible:  1,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test menu %s: %v", key, err)
	}
	return int(id)
}
