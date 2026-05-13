// This file verifies deletion guardrails for built-in dictionary records.

package dict

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
)

// TestDeleteRejectsBuiltInDictType verifies built-in dictionary types cannot
// be removed through the management service.
func TestDeleteRejectsBuiltInDictType(t *testing.T) {
	ctx := context.Background()
	record := insertDictTypeForDeleteGuard(t, ctx, true)

	err := New(nil).Delete(ctx, record.Id)
	if !bizerr.Is(err, CodeDictTypeBuiltinDeleteDenied) {
		t.Fatalf("expected %s, got %v", CodeDictTypeBuiltinDeleteDenied.RuntimeCode(), err)
	}

	assertDictTypeExists(t, ctx, record.Id)
}

// TestDataDeleteRejectsBuiltInDictData verifies built-in dictionary data
// cannot be removed through the management service.
func TestDataDeleteRejectsBuiltInDictData(t *testing.T) {
	ctx := context.Background()
	dictType := insertDictTypeForDeleteGuard(t, ctx, false)
	record := insertDictDataForDeleteGuard(t, ctx, dictType.Type, true)

	err := New(nil).DataDelete(ctx, record.Id)
	if !bizerr.Is(err, CodeDictDataBuiltinDeleteDenied) {
		t.Fatalf("expected %s, got %v", CodeDictDataBuiltinDeleteDenied.RuntimeCode(), err)
	}

	assertDictDataExists(t, ctx, record.Id)
}

// TestDeleteRejectsDictTypeContainingBuiltInData verifies cascading dictionary
// type deletion cannot remove built-in dictionary data.
func TestDeleteRejectsDictTypeContainingBuiltInData(t *testing.T) {
	ctx := context.Background()
	dictType := insertDictTypeForDeleteGuard(t, ctx, false)
	record := insertDictDataForDeleteGuard(t, ctx, dictType.Type, true)

	err := New(nil).Delete(ctx, dictType.Id)
	if !bizerr.Is(err, CodeDictDataBuiltinDeleteDenied) {
		t.Fatalf("expected %s, got %v", CodeDictDataBuiltinDeleteDenied.RuntimeCode(), err)
	}

	assertDictTypeExists(t, ctx, dictType.Id)
	assertDictDataExists(t, ctx, record.Id)
}

// insertDictTypeForDeleteGuard creates one isolated dictionary type row and
// registers cleanup for the current test.
func insertDictTypeForDeleteGuard(t *testing.T, ctx context.Context, builtin bool) *entity.SysDictType {
	t.Helper()

	suffix := time.Now().UnixNano()
	dictType := fmt.Sprintf("delete_guard_%d", suffix)
	builtinFlag := 0
	if builtin {
		builtinFlag = 1
	}

	insertedID, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
		Name:      "Delete guard dictionary",
		Type:      dictType,
		Status:    1,
		IsBuiltin: builtinFlag,
		Remark:    "delete guard test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dictionary type: %v", err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := dao.SysDictData.Ctx(ctx).
			Unscoped().
			Where(do.SysDictData{DictType: dictType}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup dictionary data for %s: %v", dictType, cleanupErr)
		}
		if _, cleanupErr := dao.SysDictType.Ctx(ctx).
			Unscoped().
			Where(do.SysDictType{Id: int(insertedID)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup dictionary type %s: %v", dictType, cleanupErr)
		}
	})

	return &entity.SysDictType{
		Id:        int(insertedID),
		Type:      dictType,
		IsBuiltin: builtinFlag,
	}
}

// insertDictDataForDeleteGuard creates one isolated dictionary data row and
// registers cleanup for the current test.
func insertDictDataForDeleteGuard(
	t *testing.T,
	ctx context.Context,
	dictType string,
	builtin bool,
) *entity.SysDictData {
	t.Helper()

	value := fmt.Sprintf("value_%d", time.Now().UnixNano())
	builtinFlag := 0
	if builtin {
		builtinFlag = 1
	}

	insertedID, err := dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
		DictType:  dictType,
		Label:     "Delete guard data",
		Value:     value,
		Sort:      1,
		TagStyle:  "primary",
		Status:    1,
		IsBuiltin: builtinFlag,
		Remark:    "delete guard test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dictionary data: %v", err)
	}

	return &entity.SysDictData{
		Id:        int(insertedID),
		DictType:  dictType,
		Value:     value,
		IsBuiltin: builtinFlag,
	}
}

// assertDictTypeExists verifies a dictionary type row remains queryable.
func assertDictTypeExists(t *testing.T, ctx context.Context, id int) {
	t.Helper()

	count, err := dao.SysDictType.Ctx(ctx).Where(do.SysDictType{Id: id}).Count()
	if err != nil {
		t.Fatalf("query dictionary type %d: %v", id, err)
	}
	if count != 1 {
		t.Fatalf("expected dictionary type %d to remain, got count %d", id, count)
	}
}

// assertDictDataExists verifies a dictionary data row remains queryable.
func assertDictDataExists(t *testing.T, ctx context.Context, id int) {
	t.Helper()

	count, err := dao.SysDictData.Ctx(ctx).Where(do.SysDictData{Id: id}).Count()
	if err != nil {
		t.Fatalf("query dictionary data %d: %v", id, err)
	}
	if count != 1 {
		t.Fatalf("expected dictionary data %d to remain, got count %d", id, count)
	}
}
