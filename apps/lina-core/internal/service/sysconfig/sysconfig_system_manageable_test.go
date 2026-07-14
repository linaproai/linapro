// This file verifies system_manageable management-surface filtering and
// mutation locks for sys_config rows.

package sysconfig

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
)

// TestListExcludesNonSystemManageableRows verifies plugin closed-loop configs
// do not appear on the system settings list.
func TestListExcludesNonSystemManageableRows(t *testing.T) {
	ctx := context.Background()
	hidden := insertConfigWithSystemManageable(t, ctx, 0)
	visible := insertConfigWithSystemManageable(t, ctx, 1)

	out, err := New(nil, nil).List(ctx, ListInput{Key: "test.sysmanage."})
	if err != nil {
		t.Fatalf("list configs: %v", err)
	}
	if findConfigProjection(out.List, hidden.Key) != nil {
		t.Fatalf("expected non system-manageable key %q to be hidden from list", hidden.Key)
	}
	if findConfigProjection(out.List, visible.Key) == nil {
		t.Fatalf("expected system-manageable key %q to appear in list", visible.Key)
	}
}

// TestGetByIdAndGetByKeyHideNonSystemManageable verifies management reads
// treat non system-manageable rows as not found.
func TestGetByIdAndGetByKeyHideNonSystemManageable(t *testing.T) {
	ctx := context.Background()
	hidden := insertConfigWithSystemManageable(t, ctx, 0)
	svc := New(nil, nil)

	_, err := svc.GetById(ctx, hidden.Id)
	if !bizerr.Is(err, CodeSysConfigNotFound) {
		t.Fatalf("GetById: expected %s, got %v", CodeSysConfigNotFound.RuntimeCode(), err)
	}

	_, err = svc.GetByKey(ctx, hidden.Key)
	if !bizerr.Is(err, CodeSysConfigKeyNotFound) {
		t.Fatalf("GetByKey: expected %s, got %v", CodeSysConfigKeyNotFound.RuntimeCode(), err)
	}
}

// TestUpdateAndDeleteRejectNonSystemManageable verifies management mutations
// cannot change plugin closed-loop configs.
func TestUpdateAndDeleteRejectNonSystemManageable(t *testing.T) {
	ctx := context.Background()
	hidden := insertConfigWithSystemManageable(t, ctx, 0)
	svc := New(hostconfig.New(), nil)

	updated := "should-not-apply"
	err := svc.Update(ctx, UpdateInput{Id: hidden.Id, Value: &updated})
	if !bizerr.Is(err, CodeSysConfigNotFound) {
		t.Fatalf("Update: expected %s, got %v", CodeSysConfigNotFound.RuntimeCode(), err)
	}

	err = svc.Delete(ctx, hidden.Id)
	if !bizerr.Is(err, CodeSysConfigNotFound) {
		t.Fatalf("Delete: expected %s, got %v", CodeSysConfigNotFound.RuntimeCode(), err)
	}

	assertConfigRecordExists(t, ctx, hidden.Id)
	var row *entity.SysConfig
	if err = dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: hidden.Id}).Scan(&row); err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if row == nil || row.Value != "initial" {
		t.Fatalf("expected value unchanged, got %+v", row)
	}
}

// TestCreateDefaultsSystemManageable verifies management Create persists
// system_manageable=1.
func TestCreateDefaultsSystemManageable(t *testing.T) {
	ctx := context.Background()
	key := fmt.Sprintf("test.sysmanage.create.%d", time.Now().UnixNano())
	id, err := New(hostconfig.New(), nil).Create(ctx, CreateInput{
		Name:  "Create manageable",
		Key:   key,
		Value: "v1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Id: id}).Delete()
	})

	var row *entity.SysConfig
	if err = dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: id}).Scan(&row); err != nil {
		t.Fatalf("query created: %v", err)
	}
	if row == nil || row.SystemManageable != 1 {
		t.Fatalf("expected system_manageable=1, got %+v", row)
	}
}

// TestImportUpdateRejectsNonSystemManageable verifies import cannot overwrite
// plugin closed-loop rows.
func TestImportUpdateRejectsNonSystemManageable(t *testing.T) {
	ctx := context.Background()
	hidden := insertConfigWithSystemManageable(t, ctx, 0)
	data := buildConfigImportFile(t, []string{
		"Imported name",
		hidden.Key,
		"imported-value",
		"text",
		"",
		"imported remark",
	})

	result, err := New(hostconfig.New(), nil).Import(ctx, bytes.NewReader(data), true)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.Success != 0 || result.Fail != 1 {
		t.Fatalf("expected one failed import, got success=%d fail=%d list=%+v", result.Success, result.Fail, result.FailList)
	}

	var row *entity.SysConfig
	if err = dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: hidden.Id}).Scan(&row); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if row == nil || row.Value != "initial" {
		t.Fatalf("expected value unchanged after import reject, got %+v", row)
	}
}

// insertConfigWithSystemManageable inserts one isolated sys_config row with the
// given system_manageable flag for management-surface tests.
func insertConfigWithSystemManageable(t *testing.T, ctx context.Context, systemManageable int) *entity.SysConfig {
	t.Helper()

	key := fmt.Sprintf("test.sysmanage.%d.%d", systemManageable, time.Now().UnixNano())
	id, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		Name:             "System manageable test",
		Key:              key,
		Value:            "initial",
		IsBuiltin:        0,
		SystemManageable: systemManageable,
		Remark:           "system manageable test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert system_manageable=%d config: %v", systemManageable, err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Id: id}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup config %s: %v", key, cleanupErr)
		}
	})
	return &entity.SysConfig{
		Id:               id,
		Key:              key,
		Value:            "initial",
		SystemManageable: systemManageable,
	}
}
