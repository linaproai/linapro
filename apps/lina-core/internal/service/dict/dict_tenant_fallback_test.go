// This file verifies tenant fallback and override guardrails for dictionary
// records.

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

// findDictTypeName returns the name for one dictionary type in a result set.
func findDictTypeName(rows []*entity.SysDictType, dictType string) string {
	for _, row := range rows {
		if row != nil && row.Type == dictType {
			return row.Name
		}
	}
	return ""
}

// findDictDataLabel returns the label for one dictionary data value.
func findDictDataLabel(rows []*entity.SysDictData, value string) string {
	for _, row := range rows {
		if row != nil && row.Value == value {
			return row.Label
		}
	}
	return ""
}
