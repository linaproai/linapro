// This file verifies sys_config capability write guards and first-insert
// system_manageable defaults.

package hostconfig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/cachecoord"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityhostconfigcap "lina-core/pkg/plugin/capability/hostconfigcap"
)

// TestSetValueRequiresCacheCoord verifies sys_config writes fail before data
// mutation when no unified cache coordinator is injected.
func TestSetValueRequiresCacheCoord(t *testing.T) {
	err := NewSysConfigCapabilityAdapter(nil, nil).SetValue(
		context.Background(),
		"site.title",
		"Lina",
		nil,
	)
	if !bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
		t.Fatalf("expected cachecoord unavailable error, got %v", err)
	}
}

// TestSetValueFirstInsertIsNotSystemManageable verifies plugin SetValue creates
// rows with system_manageable=0 and subsequent writes keep that flag when
// SystemManageable is omitted or explicit false.
func TestSetValueFirstInsertIsNotSystemManageable(t *testing.T) {
	ctx := context.Background()
	key := capabilityhostconfigcap.SysConfigKey(
		fmt.Sprintf("plugin.test-sysmanage.setting.%d", time.Now().UnixNano()),
	)
	adapter := NewSysConfigCapabilityAdapter(nil, cachecoord.Default(nil))

	if err := adapter.SetValue(ctx, key, "first", &capabilityhostconfigcap.SetSysConfigValueOptions{
		SystemManageable: gconv.PtrBool(false),
	}); err != nil {
		t.Fatalf("first SetValue: %v", err)
	}
	t.Cleanup(func() {
		_, _ = dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: string(key)}).Delete()
	})

	row := loadSysConfigByKey(t, ctx, string(key))
	if row.SystemManageable != 0 {
		t.Fatalf("first insert: expected system_manageable=0, got %d", row.SystemManageable)
	}
	if row.Value != "first" {
		t.Fatalf("first insert: expected value first, got %q", row.Value)
	}

	if err := adapter.SetValue(ctx, key, "second", nil); err != nil {
		t.Fatalf("second SetValue: %v", err)
	}
	row = loadSysConfigByKey(t, ctx, string(key))
	if row.SystemManageable != 0 {
		t.Fatalf("second write: expected system_manageable stay 0, got %d", row.SystemManageable)
	}
	if row.Value != "second" {
		t.Fatalf("second write: expected value second, got %q", row.Value)
	}
}

// TestBatchSetValueWritesAllKeysInOneRound verifies multi-key upsert shares
// one revision path and honors shared SystemManageable options.
func TestBatchSetValueWritesAllKeysInOneRound(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	keyA := capabilityhostconfigcap.SysConfigKey(fmt.Sprintf("plugin.test-batch.a.%d", suffix))
	keyB := capabilityhostconfigcap.SysConfigKey(fmt.Sprintf("plugin.test-batch.b.%d", suffix))
	adapter := NewSysConfigCapabilityAdapter(nil, cachecoord.Default(nil))

	if err := adapter.BatchSetValue(ctx, []capabilityhostconfigcap.SetSysConfigValueItem{
		{Key: keyA, Value: "va"},
		{Key: keyB, Value: "vb"},
	}, &capabilityhostconfigcap.SetSysConfigValueOptions{
		SystemManageable: gconv.PtrBool(false),
	}); err != nil {
		t.Fatalf("BatchSetValue: %v", err)
	}
	t.Cleanup(func() {
		_, _ = dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: string(keyA)}).Delete()
		_, _ = dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: string(keyB)}).Delete()
	})

	rowA := loadSysConfigByKey(t, ctx, string(keyA))
	rowB := loadSysConfigByKey(t, ctx, string(keyB))
	if rowA.Value != "va" || rowB.Value != "vb" {
		t.Fatalf("unexpected values a=%q b=%q", rowA.Value, rowB.Value)
	}
	if rowA.SystemManageable != 0 || rowB.SystemManageable != 0 {
		t.Fatalf("expected both system_manageable=0, got a=%d b=%d", rowA.SystemManageable, rowB.SystemManageable)
	}
}

// TestSetValueCanMarkSystemManageable verifies plugins may opt a key into the
// system config management page via SystemManageable=true.
func TestSetValueCanMarkSystemManageable(t *testing.T) {
	ctx := context.Background()
	key := capabilityhostconfigcap.SysConfigKey(
		fmt.Sprintf("plugin.test-sysmanage.public.%d", time.Now().UnixNano()),
	)
	adapter := NewSysConfigCapabilityAdapter(nil, cachecoord.Default(nil))

	if err := adapter.SetValue(ctx, key, "public-v1", &capabilityhostconfigcap.SetSysConfigValueOptions{
		SystemManageable: gconv.PtrBool(true),
	}); err != nil {
		t.Fatalf("SetValue system manageable: %v", err)
	}
	t.Cleanup(func() {
		_, _ = dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: string(key)}).Delete()
	})

	row := loadSysConfigByKey(t, ctx, string(key))
	if row.SystemManageable != 1 {
		t.Fatalf("expected system_manageable=1, got %d", row.SystemManageable)
	}
}

func loadSysConfigByKey(t *testing.T, ctx context.Context, key string) *entity.SysConfig {
	t.Helper()
	var row *entity.SysConfig
	if err := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Key: key}).Scan(&row); err != nil {
		t.Fatalf("query sys_config %s: %v", key, err)
	}
	if row == nil {
		t.Fatalf("expected sys_config row for key %s", key)
	}
	return row
}
