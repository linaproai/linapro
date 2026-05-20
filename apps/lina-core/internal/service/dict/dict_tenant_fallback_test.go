// This file verifies tenant fallback and override guardrails for dictionary
// records.

package dict

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// TestTenantDictReadPrefersTenantOverrideWithPlatformFallback verifies
// dictionary type and data reads use tenant rows before platform defaults.
func TestTenantDictReadPrefersTenantOverrideWithPlatformFallback(t *testing.T) {
	ctx := context.Background()
	tenantCtx := datascope.WithTenantForTest(ctx, 91)
	dictType := fmt.Sprintf("tenant_fallback_%d", time.Now().UnixNano())
	insertTenantFallbackDictType(t, ctx, datascope.PlatformTenantID, dictType, "Platform", true)
	insertTenantFallbackDictType(t, ctx, 91, dictType, "Tenant", true)
	insertTenantFallbackDictType(t, ctx, datascope.PlatformTenantID, dictType+"_platform_only", "Platform Only", false)
	insertTenantFallbackDictData(t, ctx, datascope.PlatformTenantID, dictType, "shared", "Platform Shared")
	insertTenantFallbackDictData(t, ctx, 91, dictType, "shared", "Tenant Shared")
	insertTenantFallbackDictData(t, ctx, datascope.PlatformTenantID, dictType, "platform-only", "Platform Only")

	typeOut, err := New(nil).List(tenantCtx, ListInput{Type: dictType})
	if err != nil {
		t.Fatalf("list tenant dict types: %v", err)
	}
	if typeOut.Total != 2 {
		t.Fatalf("expected two effective dict types, got %d", typeOut.Total)
	}
	tenantType := findDictTypeName(typeOut.List, dictType)
	if tenantType != "Tenant" {
		t.Fatalf("expected tenant type override, got %q", tenantType)
	}
	assertDictTypeMetadata(t, findDictTypeProjection(typeOut.List, dictType), 91, false, true, false, FallbackOverrideModeNone)
	assertDictTypeMetadata(
		t,
		findDictTypeProjection(typeOut.List, dictType+"_platform_only"),
		datascope.PlatformTenantID,
		true,
		false,
		false,
		FallbackOverrideModeNone,
	)

	dataList, err := New(nil).DataByType(tenantCtx, dictType)
	if err != nil {
		t.Fatalf("list tenant dict data: %v", err)
	}
	if len(dataList) != 2 {
		t.Fatalf("expected two effective dict data rows, got %d", len(dataList))
	}
	if label := findDictDataLabel(dataList, "shared"); label != "Tenant Shared" {
		t.Fatalf("expected tenant data override, got %q", label)
	}
	if label := findDictDataLabel(dataList, "platform-only"); label != "Platform Only" {
		t.Fatalf("expected platform fallback data, got %q", label)
	}
	assertDictDataMetadata(t, findDictDataProjection(dataList, "shared"), 91, false, true, false, FallbackOverrideModeNone)
	assertDictDataMetadata(
		t,
		findDictDataProjection(dataList, "platform-only"),
		datascope.PlatformTenantID,
		true,
		false,
		true,
		FallbackOverrideModeCreateTenantOverride,
	)

	platformList, err := New(nil).DataByType(ctx, dictType)
	if err != nil {
		t.Fatalf("list platform dict data: %v", err)
	}
	if len(platformList) != 2 {
		t.Fatalf("expected platform context to see platform rows only, got %d", len(platformList))
	}
	if label := findDictDataLabel(platformList, "shared"); label != "Platform Shared" {
		t.Fatalf("expected platform shared data, got %q", label)
	}
	assertDictDataMetadata(t, findDictDataProjection(platformList, "shared"), datascope.PlatformTenantID, false, true, false, FallbackOverrideModeNone)
}

// TestTenantDictOverrideRequiresPlatformPermission verifies tenant dictionary
// overrides are rejected unless the platform type allows them.
func TestTenantDictOverrideRequiresPlatformPermission(t *testing.T) {
	ctx := context.Background()
	tenantCtx := datascope.WithTenantForTest(ctx, 92)
	dictType := fmt.Sprintf("tenant_override_denied_%d", time.Now().UnixNano())
	insertTenantFallbackDictType(t, ctx, datascope.PlatformTenantID, dictType, "Platform", false)

	_, err := New(nil).Create(tenantCtx, CreateInput{
		Name:   "Tenant denied",
		Type:   dictType,
		Status: 1,
	})
	if !bizerr.Is(err, CodeDictTypeTenantOverrideDenied) {
		t.Fatalf("expected %s from type override, got %v", CodeDictTypeTenantOverrideDenied.RuntimeCode(), err)
	}

	_, err = New(nil).DataCreate(tenantCtx, DataCreateInput{
		DictType: dictType,
		Label:    "Tenant denied",
		Value:    "tenant-denied",
		Status:   1,
	})
	if !bizerr.Is(err, CodeDictTypeTenantOverrideDenied) {
		t.Fatalf("expected %s from data override, got %v", CodeDictTypeTenantOverrideDenied.RuntimeCode(), err)
	}
}

// TestTenantDictTypeImportPersistsCurrentTenant verifies imported dictionary
// types are owned by the request tenant.
func TestTenantDictTypeImportPersistsCurrentTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := 97
	tenantCtx := datascope.WithTenantForTest(ctx, tenantID)
	dictType := fmt.Sprintf("tenant_import_type_%d", time.Now().UnixNano())
	importData := buildDictTypeImportFile(t, []string{"Tenant Import Type", dictType, "1", "tenant import test"})

	result, err := New(nil).TypeImport(tenantCtx, bytes.NewReader(importData), false)
	if err != nil {
		t.Fatalf("import tenant dictionary type: %v", err)
	}
	if result.Success != 1 || result.Fail != 0 {
		t.Fatalf("expected one successful type import, got success=%d fail=%d failures=%#v", result.Success, result.Fail, result.FailList)
	}
	t.Cleanup(func() { cleanupTenantFallbackDictByType(t, ctx, dictType) })

	var row *struct {
		TenantId int `orm:"tenant_id"`
	}
	if err = dao.SysDictType.Ctx(ctx).
		Fields(dao.SysDictType.Columns().TenantId).
		Where(do.SysDictType{Type: dictType}).
		Scan(&row); err != nil {
		t.Fatalf("query imported dictionary type: %v", err)
	}
	if row == nil || row.TenantId != tenantID {
		t.Fatalf("expected imported dictionary type tenant_id=%d, got %#v", tenantID, row)
	}
}

// TestTenantDictDataImportCreatesOverrideInsteadOfUpdatingPlatformFallback
// verifies tenant data imports create tenant rows while preserving platform
// fallback dictionary data.
func TestTenantDictDataImportCreatesOverrideInsteadOfUpdatingPlatformFallback(t *testing.T) {
	ctx := context.Background()
	tenantID := 98
	tenantCtx := datascope.WithTenantForTest(ctx, tenantID)
	dictType := fmt.Sprintf("tenant_import_data_%d", time.Now().UnixNano())
	insertTenantFallbackDictType(t, ctx, datascope.PlatformTenantID, dictType, "Platform Type", true)
	insertTenantFallbackDictData(t, ctx, datascope.PlatformTenantID, dictType, "shared", "Platform Label")
	importData := buildDictDataImportFile(t, []string{dictType, "Tenant Label", "shared", "1", "primary", "", "1", "tenant import override"})

	result, err := New(nil).DataImport(tenantCtx, bytes.NewReader(importData), true)
	if err != nil {
		t.Fatalf("import tenant dictionary data: %v", err)
	}
	if result.Success != 1 || result.Fail != 0 {
		t.Fatalf("expected one successful data import, got success=%d fail=%d failures=%#v", result.Success, result.Fail, result.FailList)
	}
	t.Cleanup(func() { cleanupTenantFallbackDictByType(t, ctx, dictType) })

	tenantRows, err := New(nil).DataByType(tenantCtx, dictType)
	if err != nil {
		t.Fatalf("list imported tenant dictionary data: %v", err)
	}
	tenantRow := findDictDataProjection(tenantRows, "shared")
	if tenantRow == nil || tenantRow.Label != "Tenant Label" || tenantRow.TenantId != tenantID {
		t.Fatalf("expected tenant override data row, got %#v", tenantRow)
	}

	platformRows, err := New(nil).DataByType(ctx, dictType)
	if err != nil {
		t.Fatalf("list platform dictionary data after tenant import: %v", err)
	}
	platformRow := findDictDataProjection(platformRows, "shared")
	if platformRow == nil || platformRow.Label != "Platform Label" || platformRow.TenantId != datascope.PlatformTenantID {
		t.Fatalf("expected platform dictionary data to remain unchanged, got %#v", platformRow)
	}
}

// insertTenantFallbackDictType creates one isolated dictionary type row.
func insertTenantFallbackDictType(
	t *testing.T,
	ctx context.Context,
	tenantID int,
	dictType string,
	name string,
	allowOverride bool,
) {
	t.Helper()

	insertedID, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
		TenantId:            tenantID,
		Name:                name,
		Type:                dictType,
		Status:              1,
		IsBuiltin:           0,
		AllowTenantOverride: allowOverride,
		Remark:              "tenant fallback test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dict type %s tenant %d: %v", dictType, tenantID, err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := dao.SysDictType.Ctx(ctx).
			Unscoped().
			Where(do.SysDictType{Id: int(insertedID)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup dict type %s tenant %d: %v", dictType, tenantID, cleanupErr)
		}
	})
}

// insertTenantFallbackDictData creates one isolated dictionary data row.
func insertTenantFallbackDictData(
	t *testing.T,
	ctx context.Context,
	tenantID int,
	dictType string,
	value string,
	label string,
) {
	t.Helper()

	insertedID, err := dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
		TenantId: tenantID,
		DictType: dictType,
		Label:    label,
		Value:    value,
		Sort:     1,
		Status:   1,
		Remark:   "tenant fallback test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dict data %s/%s tenant %d: %v", dictType, value, tenantID, err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := dao.SysDictData.Ctx(ctx).
			Unscoped().
			Where(do.SysDictData{Id: int(insertedID)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup dict data %s/%s tenant %d: %v", dictType, value, tenantID, cleanupErr)
		}
	})
}

// cleanupTenantFallbackDictByType removes all test dictionary rows for one
// dictionary type.
func cleanupTenantFallbackDictByType(t *testing.T, ctx context.Context, dictType string) {
	t.Helper()
	if _, err := dao.SysDictData.Ctx(ctx).
		Unscoped().
		Where(do.SysDictData{DictType: dictType}).
		Delete(); err != nil {
		t.Fatalf("cleanup dict data type %s: %v", dictType, err)
	}
	if _, err := dao.SysDictType.Ctx(ctx).
		Unscoped().
		Where(do.SysDictType{Type: dictType}).
		Delete(); err != nil {
		t.Fatalf("cleanup dict type %s: %v", dictType, err)
	}
}

// buildDictTypeImportFile builds a dictionary-type import workbook with one
// data row.
func buildDictTypeImportFile(t *testing.T, row []string) []byte {
	t.Helper()
	return buildDictImportWorkbook(t, "Sheet1", []string{"Dictionary Name", "Dictionary Type", "Status", "Remark"}, row)
}

// buildDictDataImportFile builds a dictionary-data import workbook with one
// data row.
func buildDictDataImportFile(t *testing.T, row []string) []byte {
	t.Helper()
	return buildDictImportWorkbook(
		t,
		"Sheet1",
		[]string{"Dictionary Type", "Dictionary Label", "Dictionary Value", "Sort", "Tag Style", "CSS Class", "Status", "Remark"},
		row,
	)
}

// buildDictImportWorkbook writes one in-memory Excel workbook for import tests.
func buildDictImportWorkbook(t *testing.T, sheet string, headers []string, row []string) []byte {
	t.Helper()

	f := excelize.NewFile()
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			t.Fatalf("build dict import header cell name: %v", err)
		}
		if err = f.SetCellValue(sheet, cell, header); err != nil {
			t.Fatalf("set dict import header %s: %v", header, err)
		}
	}
	for i, value := range row {
		cell, err := excelize.CoordinatesToCellName(i+1, 2)
		if err != nil {
			t.Fatalf("build dict import row cell name: %v", err)
		}
		if err = f.SetCellValue(sheet, cell, value); err != nil {
			t.Fatalf("set dict import row value %s: %v", value, err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write dict import workbook: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close dict import workbook: %v", err)
	}
	return buf.Bytes()
}

// findDictTypeName returns the name for one dictionary type in a result set.
func findDictTypeName(rows []*DictTypeProjection, dictType string) string {
	row := findDictTypeProjection(rows, dictType)
	if row == nil {
		return ""
	}
	return row.Name
}

// findDictTypeProjection returns one dictionary type projection by type.
func findDictTypeProjection(rows []*DictTypeProjection, dictType string) *DictTypeProjection {
	for _, row := range rows {
		if row != nil && row.Type == dictType {
			return row
		}
	}
	return nil
}

// findDictDataLabel returns the label for one dictionary data value.
func findDictDataLabel(rows []*DictDataProjection, value string) string {
	row := findDictDataProjection(rows, value)
	if row == nil {
		return ""
	}
	return row.Label
}

// findDictDataProjection returns one dictionary data projection by value.
func findDictDataProjection(rows []*DictDataProjection, value string) *DictDataProjection {
	for _, row := range rows {
		if row != nil && row.Value == value {
			return row
		}
	}
	return nil
}

// assertDictTypeMetadata verifies fallback action metadata for one dictionary
// type row.
func assertDictTypeMetadata(
	t *testing.T,
	row *DictTypeProjection,
	sourceTenantID int,
	isFallback bool,
	canEdit bool,
	canOverride bool,
	overrideMode FallbackOverrideMode,
) {
	t.Helper()

	if row == nil {
		t.Fatal("expected dictionary type projection, got nil")
	}
	assertDictMetadata(t, row.FallbackMetadata, sourceTenantID, isFallback, canEdit, canOverride, overrideMode)
}

// assertDictDataMetadata verifies fallback action metadata for one dictionary
// data row.
func assertDictDataMetadata(
	t *testing.T,
	row *DictDataProjection,
	sourceTenantID int,
	isFallback bool,
	canEdit bool,
	canOverride bool,
	overrideMode FallbackOverrideMode,
) {
	t.Helper()

	if row == nil {
		t.Fatal("expected dictionary data projection, got nil")
	}
	assertDictMetadata(t, row.FallbackMetadata, sourceTenantID, isFallback, canEdit, canOverride, overrideMode)
}

// assertDictMetadata verifies one fallback metadata value set.
func assertDictMetadata(
	t *testing.T,
	metadata FallbackMetadata,
	sourceTenantID int,
	isFallback bool,
	canEdit bool,
	canOverride bool,
	overrideMode FallbackOverrideMode,
) {
	t.Helper()

	if metadata.SourceTenantId != sourceTenantID {
		t.Fatalf("expected source tenant %d, got %d", sourceTenantID, metadata.SourceTenantId)
	}
	if metadata.IsFallback != isFallback {
		t.Fatalf("expected isFallback=%v, got %v", isFallback, metadata.IsFallback)
	}
	if metadata.CanEdit != canEdit {
		t.Fatalf("expected canEdit=%v, got %v", canEdit, metadata.CanEdit)
	}
	if metadata.CanOverride != canOverride {
		t.Fatalf("expected canOverride=%v, got %v", canOverride, metadata.CanOverride)
	}
	if metadata.OverrideMode != overrideMode {
		t.Fatalf("expected overrideMode=%q, got %q", overrideMode, metadata.OverrideMode)
	}
}
