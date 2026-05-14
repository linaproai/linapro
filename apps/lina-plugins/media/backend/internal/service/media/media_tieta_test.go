// This file tests Tieta token-based media strategy authorization.

package media

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// fakeTietaClient provides deterministic token and device-permission responses for unit tests.
type fakeTietaClient struct {
	user      *TietaUser
	hasAccess bool
}

// UserInfoByToken returns the configured test user.
func (c *fakeTietaClient) UserInfoByToken(ctx context.Context, token string) (*TietaUser, error) {
	return c.user, nil
}

// CheckTenantHasDevice returns the configured device permission result.
func (c *fakeTietaClient) CheckTenantHasDevice(
	ctx context.Context,
	token string,
	tenantID string,
	deviceID string,
) (bool, error) {
	return c.hasAccess, nil
}

// TestResolveStrategyByTokenUsesTietaTenantDevicePermission verifies token tenant and device authorization drive strategy resolution.
func TestResolveStrategyByTokenUsesTietaTenantDevicePermission(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, Username: "wj530", RealName: "王杰", Mobile: "18213268117", TenantId: "tenant-a"},
		hasAccess: true,
	})
	defer restoreTietaClient()

	strategyID := insertTestStrategy(t, ctx, "租户设备策略", int(SwitchOff), int(SwitchOn))
	if _, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).Data(do.MediaStrategyDeviceTenant{
		TenantId:   "tenant-a",
		DeviceId:   "34020000001320000001",
		StrategyId: strategyID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant-device binding: %v", err)
	}

	out, err := New().ResolveStrategyByToken(ctx, ResolveStrategyByTokenInput{
		Token:    "Bearer token-value",
		TenantId: "tenant-a",
		DeviceId: "34020000001320000001",
	})
	if err != nil {
		t.Fatalf("resolve strategy by token: %v", err)
	}
	if !out.HasAccess {
		t.Fatal("expected Tieta device access")
	}
	if !out.Matched || out.Source != string(StrategySourceTenantDevice) {
		t.Fatalf("expected tenant-device strategy match, got matched=%v source=%s", out.Matched, out.Source)
	}
	if out.UserId != 13 || out.TenantId != "tenant-a" || out.StrategyId != strategyID {
		t.Fatalf("unexpected output: %+v", out)
	}
}

// TestResolveStrategyByTokenRejectsTenantMismatch verifies callers cannot override the token tenant.
func TestResolveStrategyByTokenRejectsTenantMismatch(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, TenantId: "tenant-a"},
		hasAccess: true,
	})
	defer restoreTietaClient()

	_, err := New().ResolveStrategyByToken(ctx, ResolveStrategyByTokenInput{
		Token:    "token-value",
		TenantId: "tenant-b",
		DeviceId: "34020000001320000001",
	})
	if err == nil {
		t.Fatal("expected tenant mismatch error")
	}
	structured, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected bizerr, got %T", err)
	}
	if structured.RuntimeCode() != "MEDIA_TIETA_TENANT_MISMATCH" {
		t.Fatalf("expected tenant mismatch code, got %s", structured.RuntimeCode())
	}
}

// TestResolveStrategyByTokenDeniesWithoutDevicePermission verifies denied Tieta permission does not return a strategy.
func TestResolveStrategyByTokenDeniesWithoutDevicePermission(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, TenantId: "tenant-a"},
		hasAccess: false,
	})
	defer restoreTietaClient()

	insertTestStrategy(t, ctx, "全局策略", int(SwitchOn), int(SwitchOn))
	out, err := New().ResolveStrategyByToken(ctx, ResolveStrategyByTokenInput{
		Authorization: "Bearer token-value",
		DeviceId:      "34020000001320000001",
	})
	if err != nil {
		t.Fatalf("resolve strategy by denied token: %v", err)
	}
	if out.HasAccess || out.Matched || out.Source != string(StrategySourceNone) {
		t.Fatalf("expected no access and no strategy, got %+v", out)
	}
}

// setupMediaStrategySQLite creates the minimal media tables required by strategy resolution.
func setupMediaStrategySQLite(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	dbPath := t.TempDir() + "/media-tieta.db"
	if err := gdb.SetConfig(gdb.Config{
		"default": {
			{Link: "sqlite::@file(" + dbPath + ")"},
		},
	}); err != nil {
		t.Fatalf("set sqlite config: %v", err)
	}
	t.Cleanup(func() {
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Fatalf("restore db config: %v", err)
		}
	})

	statements := []string{
		`CREATE TABLE media_strategy (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			strategy TEXT NOT NULL,
			global INTEGER NOT NULL,
			enable INTEGER NOT NULL,
			creator_id INTEGER,
			updater_id INTEGER,
			create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time TEXT
		)`,
		`CREATE TABLE media_strategy_device (
			device_id TEXT PRIMARY KEY,
			strategy_id INTEGER NOT NULL
		)`,
		`CREATE TABLE media_strategy_tenant (
			tenant_id TEXT PRIMARY KEY,
			strategy_id INTEGER NOT NULL
		)`,
		`CREATE TABLE media_strategy_device_tenant (
			tenant_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			strategy_id INTEGER NOT NULL,
			PRIMARY KEY (tenant_id, device_id)
		)`,
		`CREATE TABLE media_device_node (device_id TEXT NOT NULL, node_num INTEGER NOT NULL)`,
		`CREATE TABLE media_node (id INTEGER PRIMARY KEY AUTOINCREMENT, node_num INTEGER NOT NULL, name TEXT NOT NULL, qn_url TEXT NOT NULL, basic_url TEXT NOT NULL, dn_url TEXT NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE media_tenant_stream_config (tenant_id TEXT PRIMARY KEY, max_concurrent INTEGER NOT NULL, node_num INTEGER NOT NULL, enable INTEGER NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE media_tenant_white (tenant_id TEXT NOT NULL, ip TEXT NOT NULL, enable INTEGER NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (tenant_id, ip))`,
		`CREATE TABLE media_stream_alias (id INTEGER PRIMARY KEY AUTOINCREMENT, alias TEXT NOT NULL, auto_remove INTEGER NOT NULL, stream_path TEXT NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, statement := range statements {
		if _, err := dao.MediaStrategy.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("exec sqlite schema: %v", err)
		}
	}
}

// insertTestStrategy inserts one enabled test strategy and returns its generated ID.
func insertTestStrategy(t *testing.T, ctx context.Context, name string, global int, enable int) int64 {
	t.Helper()

	id, err := dao.MediaStrategy.Ctx(ctx).Data(do.MediaStrategy{
		Name:     name,
		Strategy: "record:\n  enabled: true\n",
		Global:   global,
		Enable:   enable,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert strategy: %v", err)
	}
	return id
}

// replaceMediaTietaClient swaps the process Tieta client and returns a restore function.
func replaceMediaTietaClient(t *testing.T, client tietaClient) func() {
	t.Helper()

	original := mediaTietaClient
	mediaTietaClient = client
	return func() {
		mediaTietaClient = original
	}
}
