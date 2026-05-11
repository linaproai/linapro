// This file verifies tenant fallback behavior for system configuration records.

package sysconfig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
)

// TestTenantSysconfigReadPrefersTenantOverrideWithPlatformFallback verifies
// key lookups and lists prefer tenant records while falling back to platform
// defaults.
func TestTenantSysconfigReadPrefersTenantOverrideWithPlatformFallback(t *testing.T) {
	ctx := context.Background()
	tenantCtx := datascope.WithTenantForTest(ctx, 93)
	key := fmt.Sprintf("tenant.fallback.%d", time.Now().UnixNano())
	insertTenantFallbackConfig(t, ctx, datascope.PlatformTenantID, key, "platform-value")
	insertTenantFallbackConfig(t, ctx, 93, key, "tenant-value")
	insertTenantFallbackConfig(t, ctx, datascope.PlatformTenantID, key+".platform", "platform-only")

	cfg, err := New().GetByKey(tenantCtx, key)
	if err != nil {
		t.Fatalf("get tenant fallback config by key: %v", err)
	}
	if cfg.Value != "tenant-value" || cfg.TenantId != 93 {
		t.Fatalf("expected tenant config override, got value=%q tenant=%d", cfg.Value, cfg.TenantId)
	}

	out, err := New().List(tenantCtx, ListInput{Key: key})
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

	platformCfg, err := New().GetByKey(ctx, key)
	if err != nil {
		t.Fatalf("get platform config by key: %v", err)
	}
	if platformCfg.Value != "platform-value" || platformCfg.TenantId != datascope.PlatformTenantID {
		t.Fatalf("expected platform config, got value=%q tenant=%d", platformCfg.Value, platformCfg.TenantId)
	}
}

// TestTenantSysconfigCreatePersistsCurrentTenant verifies tenant writes use the
// current tenant id instead of platform.
func TestTenantSysconfigCreatePersistsCurrentTenant(t *testing.T) {
	ctx := context.Background()
	tenantCtx := datascope.WithTenantForTest(ctx, 94)
	key := fmt.Sprintf("tenant.create.%d", time.Now().UnixNano())

	createdID, err := New().Create(tenantCtx, CreateInput{
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

	cfg, err := New().GetByKey(tenantCtx, key)
	if err != nil {
		t.Fatalf("get created tenant config: %v", err)
	}
	if cfg.TenantId != 94 {
		t.Fatalf("expected created config tenant 94, got %d", cfg.TenantId)
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

// findConfigValue returns the value for one config key in a result set.
func findConfigValue(rows []*entity.SysConfig, key string) string {
	for _, row := range rows {
		if row != nil && row.Key == key {
			return row.Value
		}
	}
	return ""
}
