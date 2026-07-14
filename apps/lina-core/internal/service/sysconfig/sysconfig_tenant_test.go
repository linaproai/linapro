// This file verifies tenant fallback behavior for system configuration records.

package sysconfig

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
)

// TestTenantSysconfigReadPrefersTenantOverrideWithPlatformFallback verifies
// key lookups and lists prefer tenant records while falling back to platform
// defaults.
func TestTenantSysconfigReadPrefersTenantOverrideWithPlatformFallback(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantCtx = datascope.WithTenantScope(ctx, 93)
		key       = fmt.Sprintf("tenant.fallback.%d", time.Now().UnixNano())
	)
	insertTenantFallbackConfig(t, ctx, datascope.PlatformTenantID, key, "platform-value")
	insertTenantFallbackConfig(t, ctx, 93, key, "tenant-value")
	insertTenantFallbackConfig(t, ctx, datascope.PlatformTenantID, key+".platform", "platform-only")

	cfg, err := New(nil, nil).GetByKey(tenantCtx, key)
	if err != nil {
		t.Fatalf("get tenant fallback config by key: %v", err)
	}
	if cfg.Value != "tenant-value" || cfg.TenantId != 93 {
		t.Fatalf("expected tenant config override, got value=%q tenant=%d", cfg.Value, cfg.TenantId)
	}

	out, err := New(nil, nil).List(tenantCtx, ListInput{Key: key})
	if err != nil {
		t.Fatalf("list tenant fallback configs: %v", err)
	}
	if out.Total != 2 {
		t.Fatalf("expected two effective configs, got %d", out.Total)
	}
	if value := findConfigValue(out.List, key); value != "tenant-value" {
		t.Fatalf("expected tenant list override, got %q", value)
	}
	if value := findConfigValue(out.List, key+".platform"); value != "platform-only" {
		t.Fatalf("expected platform fallback config, got %q", value)
	}
	tenantItem := findConfigProjection(out.List, key)
	assertConfigMetadata(t, tenantItem, 93, false, true, false, FallbackOverrideModeNone)
	fallbackItem := findConfigProjection(out.List, key+".platform")
	assertConfigMetadata(t, fallbackItem, datascope.PlatformTenantID, true, false, true, FallbackOverrideModeCreateTenantOverride)

	platformCfg, err := New(nil, nil).GetByKey(ctx, key)
	if err != nil {
		t.Fatalf("get platform config by key: %v", err)
	}
	if platformCfg.Value != "platform-value" || platformCfg.TenantId != datascope.PlatformTenantID {
		t.Fatalf("expected platform config, got value=%q tenant=%d", platformCfg.Value, platformCfg.TenantId)
	}
	assertConfigMetadata(t, platformCfg, datascope.PlatformTenantID, false, true, false, FallbackOverrideModeNone)
}

// TestTenantSysconfigCreatePersistsCurrentTenant verifies tenant writes use the
// current tenant id instead of platform.
func TestTenantSysconfigCreatePersistsCurrentTenant(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantCtx = datascope.WithTenantScope(ctx, 94)
		key       = fmt.Sprintf("tenant.create.%d", time.Now().UnixNano())
	)

	createdID, err := New(nil, nil).Create(tenantCtx, CreateInput{
		Name:  "Tenant create",
		Key:   key,
		Value: "tenant-created",
	})
	if err != nil {
		t.Fatalf("create tenant config: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: createdID}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup tenant config %s: %v", key, cleanupErr)
		}
	})

	cfg, err := New(nil, nil).GetByKey(tenantCtx, key)
	if err != nil {
		t.Fatalf("get created tenant config: %v", err)
	}
	if cfg.TenantId != 94 {
		t.Fatalf("expected created config tenant 94, got %d", cfg.TenantId)
	}
	assertConfigMetadata(t, cfg, 94, false, true, false, FallbackOverrideModeNone)
}

// TestTenantSysconfigImportPersistsCurrentTenant verifies imported config rows
// are tenant-owned instead of being written to the platform fallback scope.
func TestTenantSysconfigImportPersistsCurrentTenant(t *testing.T) {
	var (
		ctx        = context.Background()
		tenantID   = 95
		tenantCtx  = datascope.WithTenantScope(ctx, tenantID)
		key        = fmt.Sprintf("tenant.import.%d", time.Now().UnixNano())
		importData = buildConfigImportFile(t, []string{"Tenant import", key, "tenant-imported", "text", "", "tenant import test"})
	)

	result, err := New(nil, nil).Import(tenantCtx, bytes.NewReader(importData), false)
	if err != nil {
		t.Fatalf("import tenant config: %v", err)
	}
	if result.Success != 1 || result.Fail != 0 {
		t.Fatalf("expected one successful tenant config import, got success=%d fail=%d failures=%#v", result.Success, result.Fail, result.FailList)
	}
	t.Cleanup(func() { cleanupTenantFallbackConfigsByKey(t, ctx, key) })

	cfg, err := New(nil, nil).GetByKey(tenantCtx, key)
	if err != nil {
		t.Fatalf("get imported tenant config: %v", err)
	}
	if cfg.TenantId != tenantID {
		t.Fatalf("expected imported config tenant_id=%d, got %d", tenantID, cfg.TenantId)
	}
	assertConfigMetadata(t, cfg, tenantID, false, true, false, FallbackOverrideModeNone)
}

// TestTenantSysconfigImportCreatesOverrideInsteadOfUpdatingPlatformFallback
// verifies tenant imports do not mutate inherited platform defaults.
func TestTenantSysconfigImportCreatesOverrideInsteadOfUpdatingPlatformFallback(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantID  = 96
		tenantCtx = datascope.WithTenantScope(ctx, tenantID)
		key       = fmt.Sprintf("tenant.import.override.%d", time.Now().UnixNano())
	)
	insertTenantFallbackConfig(t, ctx, datascope.PlatformTenantID, key, "platform-value")
	importData := buildConfigImportFile(t, []string{"Tenant override", key, "tenant-value", "text", "", "tenant import override"})

	result, err := New(nil, nil).Import(tenantCtx, bytes.NewReader(importData), true)
	if err != nil {
		t.Fatalf("import tenant config override: %v", err)
	}
	if result.Success != 1 || result.Fail != 0 {
		t.Fatalf("expected one successful tenant config override import, got success=%d fail=%d failures=%#v", result.Success, result.Fail, result.FailList)
	}
	t.Cleanup(func() { cleanupTenantFallbackConfigsByKey(t, ctx, key) })

	tenantCfg, err := New(nil, nil).GetByKey(tenantCtx, key)
	if err != nil {
		t.Fatalf("get imported tenant override config: %v", err)
	}
	if tenantCfg.Value != "tenant-value" || tenantCfg.TenantId != tenantID {
		t.Fatalf("expected tenant override value and tenant, got value=%q tenant=%d", tenantCfg.Value, tenantCfg.TenantId)
	}

	platformCfg, err := New(nil, nil).GetByKey(ctx, key)
	if err != nil {
		t.Fatalf("get platform config after tenant import: %v", err)
	}
	if platformCfg.Value != "platform-value" || platformCfg.TenantId != datascope.PlatformTenantID {
		t.Fatalf("expected platform config to remain unchanged, got value=%q tenant=%d", platformCfg.Value, platformCfg.TenantId)
	}
}

// insertTenantFallbackConfig creates one isolated sys_config row.
func insertTenantFallbackConfig(t *testing.T, ctx context.Context, tenantID int, key string, value string) {
	t.Helper()

	insertedID, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		TenantId: tenantID,
		Name:     key,
		Key:      key,
		Value:    value,
		Remark:   "tenant fallback test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert config %s tenant %d: %v", key, tenantID, err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: int64(insertedID)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup config %s tenant %d: %v", key, tenantID, cleanupErr)
		}
	})
}

// cleanupTenantFallbackConfigsByKey removes all test config rows for one key.
func cleanupTenantFallbackConfigsByKey(t *testing.T, ctx context.Context, key string) {
	t.Helper()
	if _, err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Delete(); err != nil {
		t.Fatalf("cleanup config key %s: %v", key, err)
	}
}

// findConfigValue returns the value for one config key in a result set.
func findConfigValue(rows []*ConfigProjection, key string) string {
	row := findConfigProjection(rows, key)
	if row == nil {
		return ""
	}
	return row.Value
}

// findConfigProjection returns one config projection by key.
func findConfigProjection(rows []*ConfigProjection, key string) *ConfigProjection {
	for _, row := range rows {
		if row != nil && row.Key == key {
			return row
		}
	}
	return nil
}

// assertConfigMetadata verifies fallback action metadata for one config row.
func assertConfigMetadata(
	t *testing.T,
	row *ConfigProjection,
	sourceTenantID int,
	isFallback bool,
	canEdit bool,
	canOverride bool,
	overrideMode FallbackOverrideMode,
) {
	t.Helper()

	if row == nil {
		t.Fatal("expected config projection, got nil")
	}
	if row.SourceTenantId != sourceTenantID {
		t.Fatalf("expected source tenant %d, got %d", sourceTenantID, row.SourceTenantId)
	}
	if row.IsFallback != isFallback {
		t.Fatalf("expected isFallback=%v, got %v", isFallback, row.IsFallback)
	}
	if row.CanEdit != canEdit {
		t.Fatalf("expected canEdit=%v, got %v", canEdit, row.CanEdit)
	}
	if row.CanOverride != canOverride {
		t.Fatalf("expected canOverride=%v, got %v", canOverride, row.CanOverride)
	}
	if row.OverrideMode != overrideMode {
		t.Fatalf("expected overrideMode=%q, got %q", overrideMode, row.OverrideMode)
	}
}
