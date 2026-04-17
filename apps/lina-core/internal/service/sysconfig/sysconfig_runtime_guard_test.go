// This file verifies built-in runtime parameter validation and guardrails in
// the sysconfig management service.

package sysconfig

import (
	"bytes"
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
)

func TestDeleteRejectsProtectedRuntimeParam(t *testing.T) {
	ctx := context.Background()
	runtimeParam := ensureRuntimeParamRecord(t, ctx, hostconfig.RuntimeParamKeyJWTExpire, "24h")

	err := New().Delete(ctx, int(runtimeParam.Id))
	if err == nil {
		t.Fatal("expected deleting protected runtime param to fail")
	}
}

func TestUpdateRejectsProtectedRuntimeParamRename(t *testing.T) {
	ctx := context.Background()
	runtimeParam := ensureRuntimeParamRecord(t, ctx, hostconfig.RuntimeParamKeyJWTExpire, "24h")
	newKey := "sys.jwt.expire.renamed"

	err := New().Update(ctx, UpdateInput{
		Id:  int(runtimeParam.Id),
		Key: &newKey,
	})
	if err == nil {
		t.Fatal("expected renaming protected runtime param to fail")
	}
}

func TestValidateManagedConfigValueRejectsInvalidValues(t *testing.T) {
	testCases := []struct {
		key   string
		value string
	}{
		{key: hostconfig.RuntimeParamKeyJWTExpire, value: "bad"},
		{key: hostconfig.RuntimeParamKeySessionTimeout, value: "0s"},
		{key: hostconfig.RuntimeParamKeyUploadMaxSize, value: "-1"},
		{key: hostconfig.RuntimeParamKeyLoginBlackIPList, value: "invalid-ip"},
	}

	for _, testCase := range testCases {
		if err := validateManagedConfigValue(testCase.key, testCase.value); err == nil {
			t.Fatalf("expected invalid runtime value to fail validation: %s=%q", testCase.key, testCase.value)
		}
	}
}

func TestUpdateProtectedRuntimeParamRefreshesConfigSnapshot(t *testing.T) {
	ctx := context.Background()
	runtimeParam := ensureRuntimeParamRecord(t, ctx, hostconfig.RuntimeParamKeyJWTExpire, "24h")

	cfgSvc := hostconfig.New()
	if cfg := cfgSvc.GetJwt(ctx); cfg.Expire != 24*time.Hour {
		t.Fatalf("expected initial jwt expire to be 24h, got %s", cfg.Expire)
	}

	updatedValue := "8h"
	err := New().Update(ctx, UpdateInput{
		Id:    int(runtimeParam.Id),
		Value: &updatedValue,
	})
	if err != nil {
		t.Fatalf("update protected runtime param: %v", err)
	}

	if cfg := cfgSvc.GetJwt(ctx); cfg.Expire != 8*time.Hour {
		t.Fatalf("expected jwt expire to refresh to 8h after update, got %s", cfg.Expire)
	}
}

func TestCreateProtectedRuntimeParamRefreshesConfigSnapshot(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamRemoved(t, ctx, hostconfig.RuntimeParamKeyUploadMaxSize)

	cfgSvc := hostconfig.New()
	_ = cfgSvc.GetUpload(ctx)

	createdID, err := New().Create(ctx, CreateInput{
		Name:   "文件管理-上传大小上限",
		Key:    hostconfig.RuntimeParamKeyUploadMaxSize,
		Value:  "3",
		Remark: "test create",
	})
	if err != nil {
		t.Fatalf("create protected runtime param: %v", err)
	}
	if createdID <= 0 {
		t.Fatalf("expected created runtime param id to be positive, got %d", createdID)
	}

	if cfg := cfgSvc.GetUpload(ctx); cfg.MaxSize != 3 {
		t.Fatalf("expected upload max size to refresh to 3 after create, got %d", cfg.MaxSize)
	}
}

func TestImportProtectedRuntimeParamRefreshesConfigSnapshot(t *testing.T) {
	ctx := context.Background()
	ensureRuntimeParamRecord(t, ctx, hostconfig.RuntimeParamKeyJWTExpire, "24h")

	cfgSvc := hostconfig.New()
	if cfg := cfgSvc.GetJwt(ctx); cfg.Expire != 24*time.Hour {
		t.Fatalf("expected initial jwt expire to be 24h, got %s", cfg.Expire)
	}

	importData := buildConfigImportFile(t, []string{
		"认证管理-JWT Token 有效期",
		hostconfig.RuntimeParamKeyJWTExpire,
		"6h",
		"test import update",
	})

	result, err := New().Import(ctx, bytes.NewReader(importData), true)
	if err != nil {
		t.Fatalf("import protected runtime param: %v", err)
	}
	if result.Success != 1 || result.Fail != 0 {
		t.Fatalf("expected one successful import, got success=%d fail=%d", result.Success, result.Fail)
	}

	if cfg := cfgSvc.GetJwt(ctx); cfg.Expire != 6*time.Hour {
		t.Fatalf("expected jwt expire to refresh to 6h after import, got %s", cfg.Expire)
	}
}

func ensureRuntimeParamRecord(
	t *testing.T,
	ctx context.Context,
	key string,
	value string,
) *entity.SysConfig {
	t.Helper()

	existing, err := queryRuntimeParamRecord(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s: %v", key, err)
	}
	if existing != nil {
		original := *existing
		_, err = dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: existing.Id}).
			Data(do.SysConfig{Value: value}).
			Update()
		if err != nil {
			t.Fatalf("update runtime param %s: %v", key, err)
		}
		markRuntimeParamChanged(t, ctx)
		t.Cleanup(func() {
			_, cleanupErr := dao.SysConfig.Ctx(ctx).
				Unscoped().
				Where(do.SysConfig{Id: original.Id}).
				Data(do.SysConfig{
					Name:   original.Name,
					Key:    original.Key,
					Value:  original.Value,
					Remark: original.Remark,
				}).
				Update()
			if cleanupErr != nil {
				t.Fatalf("restore runtime param %s: %v", key, cleanupErr)
			}
			markRuntimeParamChanged(t, ctx)
		})
		return existing
	}

	_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		Name:   key,
		Key:    key,
		Value:  value,
		Remark: "test runtime param",
	}).Insert()
	if err != nil {
		t.Fatalf("insert runtime param %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)

	inserted, err := queryRuntimeParamRecord(ctx, key)
	if err != nil {
		t.Fatalf("query inserted runtime param %s: %v", key, err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
	return inserted
}

func withRuntimeParamRemoved(t *testing.T, ctx context.Context, key string) {
	t.Helper()

	existing, err := queryRuntimeParamRecord(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s for removal: %v", key, err)
	}
	if existing == nil {
		return
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: existing.Id}).
		Delete()
	if err != nil {
		t.Fatalf("delete runtime param %s for removal: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)

	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete(); cleanupErr != nil {
			t.Fatalf("delete recreated runtime param %s before restore: %v", key, cleanupErr)
		}
		_, cleanupErr := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Id:     existing.Id,
			Name:   existing.Name,
			Key:    existing.Key,
			Value:  existing.Value,
			Remark: existing.Remark,
		}).Insert()
		if cleanupErr != nil {
			t.Fatalf("restore removed runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

func buildConfigImportFile(t *testing.T, row []string) []byte {
	t.Helper()

	f := excelize.NewFile()
	sheet := "Sheet1"

	headers := []string{"参数名称", "参数键名", "参数键值", "备注"}
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			t.Fatalf("build import header cell name: %v", err)
		}
		if err = f.SetCellValue(sheet, cell, header); err != nil {
			t.Fatalf("set import header %s: %v", header, err)
		}
	}
	for i, value := range row {
		cell, err := excelize.CoordinatesToCellName(i+1, 2)
		if err != nil {
			t.Fatalf("build import row cell name: %v", err)
		}
		if err = f.SetCellValue(sheet, cell, value); err != nil {
			t.Fatalf("set import row value %s: %v", value, err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write import workbook: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close import workbook: %v", err)
	}
	return buf.Bytes()
}

func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

func queryRuntimeParamRecord(ctx context.Context, key string) (*entity.SysConfig, error) {
	var runtimeParam *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Scan(&runtimeParam)
	if err != nil {
		return nil, err
	}
	return runtimeParam, nil
}
