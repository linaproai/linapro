// This file verifies config-management metadata localization boundaries.

package sysconfig

import (
	"bytes"
	"context"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/xuri/excelize/v2"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
)

// TestListLocalizesConfigMetadata verifies config list metadata and built-in
// public frontend text values are localized for display.
func TestListLocalizesConfigMetadata(t *testing.T) {
	ctx := newEnglishBizCtx()
	ensureConfigRecordState(
		t,
		context.Background(),
		hostconfig.PublicFrontendSettingKeyAuthPageTitle,
		"登录展示-页面标题",
		"面向可持续交付的 AI 原生全栈框架",
		"控制登录页顶部主标题文案。",
	)

	out, err := New(nil, i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))).List(ctx, ListInput{
		PageNum:  1,
		PageSize: 10,
		Key:      hostconfig.PublicFrontendSettingKeyAuthPageTitle,
	})
	if err != nil {
		t.Fatalf("list localized configs: %v", err)
	}
	if len(out.List) != 1 {
		t.Fatalf("expected one config row, got %d", len(out.List))
	}

	item := out.List[0]
	if item.Name != "Login - Page Title" {
		t.Fatalf("expected localized config name %q, got %q", "Login - Page Title", item.Name)
	}
	if item.Remark != "Controls the headline shown at the top of the login page." {
		t.Fatalf("expected localized config remark %q, got %q", "Controls the headline shown at the top of the login page.", item.Remark)
	}
	if item.Value != "An AI-native full-stack framework engineered for sustainable delivery" {
		t.Fatalf("expected localized config value %q, got %q", "An AI-native full-stack framework engineered for sustainable delivery", item.Value)
	}
}

// TestListKeepsCustomConfigValueRaw verifies customized public frontend values
// remain raw in list rows because sys_config does not store per-locale copies.
func TestListKeepsCustomConfigValueRaw(t *testing.T) {
	ctx := newEnglishBizCtx()
	record := ensureConfigRecordState(
		t,
		context.Background(),
		hostconfig.PublicFrontendSettingKeyAuthPageTitle,
		"登录展示-页面标题",
		"Custom Login Title",
		"控制登录页顶部主标题文案。",
	)

	out, err := New(nil, i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))).List(ctx, ListInput{
		PageNum:  1,
		PageSize: 10,
		Key:      hostconfig.PublicFrontendSettingKeyAuthPageTitle,
	})
	if err != nil {
		t.Fatalf("list customized configs: %v", err)
	}
	if len(out.List) != 1 {
		t.Fatalf("expected one config row, got %d", len(out.List))
	}
	if out.List[0].Value != record.Value {
		t.Fatalf("expected custom config value to remain raw %q, got %q", record.Value, out.List[0].Value)
	}
}

// TestGetByIdKeepsRawConfigMetadata verifies edit/detail backfill keeps the raw
// config metadata so localized projections are not written back into sys_config.
func TestGetByIdKeepsRawConfigMetadata(t *testing.T) {
	ctx := newEnglishBizCtx()
	record := ensureConfigRecordState(
		t,
		context.Background(),
		hostconfig.PublicFrontendSettingKeyAuthPageTitle,
		"登录展示-页面标题",
		"面向可持续交付的 AI 原生全栈框架",
		"控制登录页顶部主标题文案。",
	)

	item, err := New(nil, i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))).GetById(ctx, int(record.Id))
	if err != nil {
		t.Fatalf("get raw config detail: %v", err)
	}
	if item.Name != record.Name {
		t.Fatalf("expected raw config name %q, got %q", record.Name, item.Name)
	}
	if item.Remark != record.Remark {
		t.Fatalf("expected raw config remark %q, got %q", record.Remark, item.Remark)
	}
	if item.Value != record.Value {
		t.Fatalf("expected raw config value %q, got %q", record.Value, item.Value)
	}
}

// TestGenerateImportTemplateLocalizesHeaders verifies the import template uses
// localized metadata headers and example copy for the requested locale.
func TestGenerateImportTemplateLocalizesHeaders(t *testing.T) {
	ctx := newEnglishBizCtx()

	data, err := New(nil, i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))).GenerateImportTemplate(ctx)
	if err != nil {
		t.Fatalf("generate localized import template: %v", err)
	}

	workbook, closeWorkbook := openExcelFromBytes(t, data)
	defer closeWorkbook()

	row1, err := workbook.GetRows("Sheet1")
	if err != nil {
		t.Fatalf("read template rows: %v", err)
	}
	if len(row1) < 2 {
		t.Fatalf("expected template rows, got %d", len(row1))
	}

	headers := row1[0]
	expectedHeaders := []string{"Parameter Name", "Parameter Key", "Parameter Value", "Remark"}
	for index, expected := range expectedHeaders {
		if headers[index] != expected {
			t.Fatalf("expected header[%d] to be %q, got %q", index, expected, headers[index])
		}
	}

	example := row1[1]
	if example[0] != "Authentication - JWT Expiration" {
		t.Fatalf("expected localized example name %q, got %q", "Authentication - JWT Expiration", example[0])
	}
	if example[1] != hostconfig.RuntimeParamKeyJWTExpire {
		t.Fatalf("expected example key %q, got %q", hostconfig.RuntimeParamKeyJWTExpire, example[1])
	}
	if example[2] != "24h" {
		t.Fatalf("expected example value %q, got %q", "24h", example[2])
	}
	if example[3] != "Controls the lifetime of newly issued JWT tokens using Go duration format such as 12h or 24h." {
		t.Fatalf("expected localized example remark, got %q", example[3])
	}
}

// TestExportLocalizesHeadersButKeepsRawRows verifies export headers are
// localized while exported config records retain the raw stored metadata.
func TestExportLocalizesHeadersButKeepsRawRows(t *testing.T) {
	ctx := newEnglishBizCtx()
	record := ensureConfigRecordState(
		t,
		context.Background(),
		hostconfig.PublicFrontendSettingKeyAuthPageTitle,
		"登录展示-页面标题",
		"面向可持续交付的 AI 原生全栈框架",
		"控制登录页顶部主标题文案。",
	)

	data, err := New(nil, i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))).Export(ctx, ExportInput{Ids: []int{int(record.Id)}})
	if err != nil {
		t.Fatalf("export localized config headers: %v", err)
	}

	workbook, closeWorkbook := openExcelFromBytes(t, data)
	defer closeWorkbook()

	rows, err := workbook.GetRows("Sheet1")
	if err != nil {
		t.Fatalf("read export rows: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected exported data rows, got %d", len(rows))
	}

	headers := rows[0]
	expectedHeaders := []string{
		"Parameter Name",
		"Parameter Key",
		"Parameter Value",
		"Remark",
		"Created At",
		"Updated At",
	}
	for index, expected := range expectedHeaders {
		if headers[index] != expected {
			t.Fatalf("expected export header[%d] to be %q, got %q", index, expected, headers[index])
		}
	}

	row := rows[1]
	if row[0] != record.Name {
		t.Fatalf("expected exported raw config name %q, got %q", record.Name, row[0])
	}
	if row[1] != record.Key {
		t.Fatalf("expected exported raw config key %q, got %q", record.Key, row[1])
	}
	if row[2] != record.Value {
		t.Fatalf("expected exported raw config value %q, got %q", record.Value, row[2])
	}
	if row[3] != record.Remark {
		t.Fatalf("expected exported raw config remark %q, got %q", record.Remark, row[3])
	}
}

// newEnglishBizCtx builds one request context pinned to the English runtime locale.
func newEnglishBizCtx() context.Context {
	return context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.EnglishLocale},
	)
}

// openExcelFromBytes loads one workbook from bytes and returns a cleanup hook.
func openExcelFromBytes(t *testing.T, data []byte) (*excelize.File, func()) {
	t.Helper()

	workbook, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open workbook: %v", err)
	}
	return workbook, func() {
		if closeErr := workbook.Close(); closeErr != nil {
			t.Fatalf("close workbook: %v", closeErr)
		}
	}
}

// ensureConfigRecordState forces one config row into the requested raw state
// and restores the original row after the test completes.
func ensureConfigRecordState(
	t *testing.T,
	ctx context.Context,
	key string,
	name string,
	value string,
	remark string,
) *entity.SysConfig {
	t.Helper()

	existing, err := queryRuntimeParamRecord(ctx, key)
	if err != nil {
		t.Fatalf("query config record %s: %v", key, err)
	}
	if existing == nil {
		_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:      name,
			Key:       key,
			Value:     value,
			IsBuiltin: builtInConfigFlag(key),
			Remark:    remark,
		}).Insert()
		if err != nil {
			t.Fatalf("insert config record %s: %v", key, err)
		}
		markRuntimeParamChanged(t, ctx)
		t.Cleanup(func() {
			_, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete()
			if cleanupErr != nil {
				t.Fatalf("delete config record %s: %v", key, cleanupErr)
			}
			markRuntimeParamChanged(t, ctx)
		})

		existing, err = queryRuntimeParamRecord(ctx, key)
		if err != nil {
			t.Fatalf("re-query inserted config record %s: %v", key, err)
		}
	}

	original := *existing
	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: existing.Id}).
		Data(do.SysConfig{
			Name:   name,
			Value:  value,
			Remark: remark,
		}).
		Update()
	if err != nil {
		t.Fatalf("update config record %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{
				Name:      original.Name,
				Key:       original.Key,
				Value:     original.Value,
				IsBuiltin: original.IsBuiltin,
				Remark:    original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore config record %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})

	existing.Name = name
	existing.Value = value
	existing.Remark = remark
	return existing
}
