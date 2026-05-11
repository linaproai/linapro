// This file verifies the multi-tenant source plugin can load its optional mock
// data through the real plugin install lifecycle.

package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"gopkg.in/yaml.v3"

	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/dialect"
)

const (
	multiTenantMockPostgresChildEnv    = "LINA_MULTI_TENANT_MOCK_POSTGRES_CHILD"
	multiTenantMockPostgresBaseLinkEnv = "LINA_MULTI_TENANT_MOCK_POSTGRES_BASE_LINK"
	multiTenantMockPostgresTestLinkEnv = "LINA_TEST_PGSQL_LINK"
	multiTenantPluginID                = "multi-tenant"
	multiTenantMockExpectedTenantCount = 17
	multiTenantMockExpectedUserCount   = 21
	multiTenantMockExpectedRoleCount   = 21
	multiTenantMockExpectedConfigCount = 3
	multiTenantMockExpectedLedgerCount = 1
	multiTenantMockDemoPassword        = "admin123"
	multiTenantMockMinSharedUserCount  = 4
	multiTenantMockMinMenuPermissions  = 8
)

// multiTenantTenantManagementButtonPermissions is the exact button projection
// expected under the tenant-management page in Menu Management.
var multiTenantTenantManagementButtonPermissions = []string{
	"system:tenant:query",
	"system:tenant:add",
	"system:tenant:edit",
	"system:tenant:remove",
	"system:tenant:impersonate",
}

// multiTenantMockExpectedTenantNames lists the required tenant display names
// shipped by the multi-tenant optional mock-data asset.
var multiTenantMockExpectedTenantNames = []string{
	"摸鱼科技有限公司",
	"精神股东科技有限公司",
	"打工人企业服务有限公司",
	"一键三连文化传媒有限公司",
	"赛博养生健康科技有限公司",
	"踩点到岗人力资源有限公司",
	"疯狂星期四餐饮管理有限公司",
	"薅羊毛优选商贸有限公司",
	"稳住别浪汽车服务有限公司",
	"不想内卷企业管理有限公司",
	"多喝热水健康管理有限公司",
	"绝绝子家政服务有限公司",
	"显眼包品牌策划有限公司",
	"泼天富贵贸易有限公司",
	"破防维修服务有限公司",
	"啊对对对客服外包有限公司",
	"已读乱回客服外包有限公司",
}

// multiTenantMockHostAssetName is forbidden in host mock-data because optional
// tenant demo data must stay owned by the source plugin that provides it.
const multiTenantMockHostAssetName = "015-mock-multi-tenant-platform.sql"

// multiTenantMockExpectedNicknames maps mock usernames to display names that
// identify their tenant or platform scope and scenario purpose.
var multiTenantMockExpectedNicknames = map[string]string{
	"platform_ops":                            "平台 租户生命周期运营员",
	"platform_auditor":                        "平台 跨租户审计员",
	"tenant_alpha_admin":                      "摸鱼科技 租户管理员",
	"tenant_alpha_ops":                        "摸鱼科技 运营用户",
	"tenant_beta_admin":                       "精神股东 租户管理员",
	"tenant_beta_auditor":                     "精神股东 审计用户",
	"tenant_gamma_admin":                      "打工人 暂停租户管理员",
	"tenant_one_click_triple_media_user":      "一键三连 演示用户",
	"tenant_cyber_wellness_health_user":       "赛博养生 演示用户",
	"tenant_clock_in_on_time_hr_user":         "踩点到岗 演示用户",
	"tenant_crazy_thursday_catering_user":     "疯狂星期四 演示用户",
	"tenant_deal_hunter_trading_user":         "薅羊毛优选 演示用户",
	"tenant_stay_calm_auto_service_user":      "稳住别浪 演示用户",
	"tenant_anti_involution_management_user":  "不想内卷 演示用户",
	"tenant_drink_hot_water_health_user":      "多喝热水 演示用户",
	"tenant_juejuezi_housekeeping_user":       "绝绝子 演示用户",
	"tenant_eye_catching_brand_planning_user": "显眼包 演示用户",
	"tenant_sudden_fortune_trading_user":      "泼天富贵 演示用户",
	"tenant_breakdown_repair_service_user":    "破防维修 演示用户",
	"tenant_yep_yep_customer_service_user":    "啊对对对 演示用户",
	"tenant_read_random_reply_service_user":   "已读乱回 演示用户",
}

// multiTenantMockExpectedActiveTenantUsers maps every active demo tenant to at
// least one active user that should appear in user-list tenant filtering.
var multiTenantMockExpectedActiveTenantUsers = map[string]string{
	"alpha-retail":                "tenant_alpha_ops",
	"beta-manufacturing":          "tenant_beta_admin",
	"one-click-triple-media":      "tenant_one_click_triple_media_user",
	"cyber-wellness-health":       "tenant_cyber_wellness_health_user",
	"clock-in-on-time-hr":         "tenant_clock_in_on_time_hr_user",
	"crazy-thursday-catering":     "tenant_crazy_thursday_catering_user",
	"deal-hunter-trading":         "tenant_deal_hunter_trading_user",
	"stay-calm-auto-service":      "tenant_stay_calm_auto_service_user",
	"anti-involution-management":  "tenant_anti_involution_management_user",
	"drink-hot-water-health":      "tenant_drink_hot_water_health_user",
	"juejuezi-housekeeping":       "tenant_juejuezi_housekeeping_user",
	"eye-catching-brand-planning": "tenant_eye_catching_brand_planning_user",
	"sudden-fortune-trading":      "tenant_sudden_fortune_trading_user",
	"breakdown-repair-service":    "tenant_breakdown_repair_service_user",
	"yep-yep-customer-service":    "tenant_yep_yep_customer_service_user",
	"read-random-reply-service":   "tenant_read_random_reply_service_user",
}

// multiTenantMockSharedTenantCodesByUsername lists users intentionally bound to
// multiple tenants for switching, cross-tenant list, and permission demos.
var multiTenantMockSharedTenantCodesByUsername = map[string][]string{
	"tenant_alpha_ops":                   {"alpha-retail", "beta-manufacturing", "one-click-triple-media"},
	"tenant_beta_auditor":                {"beta-manufacturing", "alpha-retail"},
	"tenant_one_click_triple_media_user": {"one-click-triple-media", "alpha-retail", "cyber-wellness-health"},
	"tenant_cyber_wellness_health_user":  {"cyber-wellness-health", "beta-manufacturing"},
}

// multiTenantMockSharedRoleKeysByTenantCode lists the tenant-local roles that
// each shared mock user needs in each tenant it can switch to.
var multiTenantMockSharedRoleKeysByUsername = map[string]map[string]string{
	"tenant_alpha_ops": {
		"alpha-retail":           "tenant-alpha-ops",
		"beta-manufacturing":     "tenant-beta-auditor",
		"one-click-triple-media": "tenant-one-click-triple-media-user",
	},
	"tenant_beta_auditor": {
		"beta-manufacturing": "tenant-beta-auditor",
		"alpha-retail":       "tenant-alpha-ops",
	},
	"tenant_one_click_triple_media_user": {
		"one-click-triple-media": "tenant-one-click-triple-media-user",
		"alpha-retail":           "tenant-alpha-ops",
		"cyber-wellness-health":  "tenant-cyber-wellness-health-user",
	},
	"tenant_cyber_wellness_health_user": {
		"cyber-wellness-health": "tenant-cyber-wellness-health-user",
		"beta-manufacturing":    "tenant-beta-auditor",
	},
}

// multiTenantMockSharedRolePermissionCodes lists menu permissions that shared
// mock-user roles need for permission-management demos.
var multiTenantMockSharedRolePermissionCodes = []string{
	"system:tenant:member:list",
	"system:user:list",
	"system:user:query",
	"system:role:list",
	"system:role:query",
	"system:dict:list",
	"system:dict:query",
	"system:config:list",
	"system:config:query",
	"system:file:list",
	"system:file:query",
}

// TestInstallMultiTenantWithMockDataOnPostgreSQL verifies the exact source
// plugin install path that operators use when selecting "install mock data" in
// the management UI. The test runs in an isolated child process because GoFrame
// database and config adapters are process-global.
func TestInstallMultiTenantWithMockDataOnPostgreSQL(t *testing.T) {
	if os.Getenv(multiTenantMockPostgresChildEnv) == "1" {
		t.Skip("parent test only launches the isolated PostgreSQL child process")
	}

	baseLink := strings.TrimSpace(os.Getenv(multiTenantMockPostgresTestLinkEnv))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run the PostgreSQL multi-tenant mock install regression test")
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestInstallMultiTenantWithMockDataOnPostgreSQLChild$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(),
		multiTenantMockPostgresChildEnv+"=1",
		multiTenantMockPostgresBaseLinkEnv+"="+baseLink,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multi-tenant mock install child test failed: %v\n%s", err, string(output))
	}
}

// TestMultiTenantMockDataContainsExpectedTenantNames keeps the optional
// multi-tenant mock-data asset aligned with the required display-name list.
func TestMultiTenantMockDataContainsExpectedTenantNames(t *testing.T) {
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	pluginSQL := readMultiTenantMockSQLAsset(t, repoRoot, filepath.Join(
		"apps",
		"lina-plugins",
		"multi-tenant",
		"manifest",
		"sql",
		"mock-data",
		"001-multi-tenant-demo-data.sql",
	))
	for _, name := range multiTenantMockExpectedTenantNames {
		if got := strings.Count(pluginSQL, "'"+name+"'"); got != 1 {
			t.Fatalf("expected tenant mock name %q to appear once as a SQL value, got %d", name, got)
		}
	}
}

// TestMultiTenantManifestTenantManagementButtonsMatchWorkbench keeps Menu
// Management button permissions aligned with the actual tenant page buttons.
func TestMultiTenantManifestTenantManagementButtonsMatchWorkbench(t *testing.T) {
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	manifestPath := filepath.Join(repoRoot, "apps", "lina-plugins", "multi-tenant", "plugin.yaml")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read multi-tenant manifest: %v", err)
	}

	manifest := &catalog.Manifest{}
	if err = yaml.Unmarshal(manifestBytes, manifest); err != nil {
		t.Fatalf("parse multi-tenant manifest: %v", err)
	}
	buttons := make(map[string]string)
	for _, item := range manifest.Menus {
		if item == nil || item.ParentKey != "plugin:multi-tenant:platform:tenants" || item.Type != catalog.MenuTypeButton.String() {
			continue
		}
		buttons[item.Key] = item.Perms
	}

	if len(buttons) != len(multiTenantTenantManagementButtonPermissions) {
		t.Fatalf("expected %d tenant-management button permissions, got %d: %#v", len(multiTenantTenantManagementButtonPermissions), len(buttons), buttons)
	}
	for _, permission := range multiTenantTenantManagementButtonPermissions {
		if !mapContainsValue(buttons, permission) {
			t.Fatalf("expected tenant-management button permission %q, got %#v", permission, buttons)
		}
	}
	for key, permission := range buttons {
		if strings.HasPrefix(permission, "system:tenant:resolver:") ||
			strings.HasPrefix(permission, "system:tenant:plugin:") ||
			strings.HasPrefix(permission, "system:tenant:member:") {
			t.Fatalf("tenant-management button %s should not expose non-page permission %q", key, permission)
		}
	}
}

// TestMultiTenantMockDataDocumentsBlocksAndNicknames keeps the tenant mock SQL
// assets readable for operators who inspect demo data directly in management
// tables.
func TestMultiTenantMockDataDocumentsBlocksAndNicknames(t *testing.T) {
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	pluginSQL := readMultiTenantMockSQLAsset(t, repoRoot, filepath.Join(
		"apps",
		"lina-plugins",
		"multi-tenant",
		"manifest",
		"sql",
		"mock-data",
		"001-multi-tenant-demo-data.sql",
	))
	assertBilingualMockSQLComments(t, "multi-tenant plugin mock SQL", pluginSQL)
	assertMultiTenantMockNotInHostSQL(t, repoRoot)
	assertMultiTenantMockUserPasswordComments(t, pluginSQL)
	assertMultiTenantMockSharedMembership(t, pluginSQL, multiTenantMockSharedTenantCodesByUsername)
	assertMultiTenantMockActiveTenantUsers(t, pluginSQL, multiTenantMockExpectedActiveTenantUsers)
	assertMultiTenantMockSharedMembershipRoles(t, pluginSQL, multiTenantMockSharedRoleKeysByUsername)
	assertMultiTenantMockSharedRolePermissions(t, pluginSQL, multiTenantMockSharedRolePermissionCodes)

	for username, nickname := range multiTenantMockExpectedNicknames {
		assertMockSQLUserNickname(t, pluginSQL, username, nickname)
		if !containsHan(nickname) {
			t.Fatalf("expected mock user %s nickname %q to use Chinese text", username, nickname)
		}
	}
	assertAllMockSQLUserNicknamesAreChinese(t, pluginSQL)
	for _, staleNickname := range []string{
		"PLATFORM Tenant Lifecycle Operator",
		"PLATFORM Cross-Tenant Auditor",
		"Alpha Admin",
		"Alpha Ops",
		"Beta Admin",
		"Beta Auditor",
		"Gamma Admin",
	} {
		if mockSQLContainsUserNickname(pluginSQL, staleNickname) {
			t.Fatalf("multi-tenant mock SQL still contains stale nickname %q", staleNickname)
		}
	}
}

// TestInstallMultiTenantWithMockDataOnPostgreSQLChild performs the actual
// isolated PostgreSQL install and mock-data assertions.
func TestInstallMultiTenantWithMockDataOnPostgreSQLChild(t *testing.T) {
	if os.Getenv(multiTenantMockPostgresChildEnv) != "1" {
		t.Skip("PostgreSQL multi-tenant mock install child test is executed by its parent")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	baseLink := strings.TrimSpace(os.Getenv(multiTenantMockPostgresBaseLinkEnv))
	if baseLink == "" {
		t.Fatal("PostgreSQL base link must be passed to the child test")
	}

	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	dbLink := multiTenantMockPostgresDatabaseLink(t, baseLink)
	setupMultiTenantMockPostgresDatabase(t, ctx, dbLink, repoRoot)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if cleanupErr := dropMultiTenantMockPostgresDatabase(cleanupCtx, dbLink); cleanupErr != nil {
			t.Errorf("drop PostgreSQL regression database: %v", cleanupErr)
		}
	})

	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	catalog.SetPluginRootDirOverride(pluginRoot)
	t.Cleanup(func() {
		catalog.SetPluginRootDirOverride("")
	})
	configsvc.SetPluginDynamicStoragePathOverride(filepath.Join(t.TempDir(), "dynamic"))
	t.Cleanup(func() {
		configsvc.SetPluginDynamicStoragePathOverride("")
	})

	service := newTestService()
	if err = service.Install(ctx, multiTenantPluginID, InstallOptions{
		InstallMockData: true,
		InstallMode:     catalog.InstallModeGlobal.String(),
	}); err != nil {
		t.Fatalf("install multi-tenant plugin with mock data: %v", err)
	}

	assertMultiTenantMockSQLCount(t, ctx, `SELECT COUNT(1) FROM plugin_multi_tenant_tenant;`, multiTenantMockExpectedTenantCount)
	assertMultiTenantMockTenantNames(t, ctx)
	assertMultiTenantMockSQLCount(t, ctx,
		`SELECT COUNT(1) FROM sys_user WHERE username LIKE 'tenant\_%' ESCAPE '\' OR username IN ('platform_ops', 'platform_auditor');`,
		multiTenantMockExpectedUserCount,
	)
	assertMultiTenantMockSQLCount(t, ctx,
		`SELECT COUNT(1) FROM sys_role WHERE key LIKE 'tenant-%' OR key IN ('platform-ops', 'platform-tenant-auditor');`,
		multiTenantMockExpectedRoleCount,
	)
	assertMultiTenantMockSQLCount(t, ctx, `SELECT COUNT(1) FROM plugin_multi_tenant_config_override;`, multiTenantMockExpectedConfigCount)
	assertMultiTenantMockSQLCount(t, ctx,
		`SELECT COUNT(1) FROM sys_plugin_migration WHERE plugin_id = 'multi-tenant' AND phase = 'mock';`,
		multiTenantMockExpectedLedgerCount,
	)
	assertMultiTenantMockSQLCountAtLeast(t, ctx,
		`SELECT COUNT(1) FROM (SELECT u."username", COUNT(DISTINCT m."tenant_id") AS tenant_count FROM plugin_multi_tenant_user_membership m JOIN sys_user u ON u."id" = m."user_id" WHERE m."status" = 1 AND u."username" LIKE 'tenant\_%' ESCAPE '\' GROUP BY u."username" HAVING COUNT(DISTINCT m."tenant_id") > 1) shared_users;`,
		multiTenantMockMinSharedUserCount,
	)
	assertMultiTenantMockSQLCount(t, ctx,
		`SELECT COUNT(1) FROM plugin_multi_tenant_tenant t WHERE t."status" = 'active' AND NOT EXISTS (SELECT 1 FROM plugin_multi_tenant_user_membership m JOIN sys_user u ON u."id" = m."user_id" WHERE m."tenant_id" = t."id" AND m."status" = 1 AND u."status" = 1);`,
		0,
	)
	assertMultiTenantMockSQLCount(t, ctx,
		`SELECT COUNT(1) FROM plugin_multi_tenant_user_membership m JOIN sys_user u ON u."id" = m."user_id" JOIN plugin_multi_tenant_tenant t ON t."id" = m."tenant_id" WHERE m."status" = 1 AND u."username" LIKE 'tenant\_%' ESCAPE '\' AND t."status" = 'active' AND NOT EXISTS (SELECT 1 FROM sys_user_role ur JOIN sys_role r ON r."id" = ur."role_id" WHERE ur."user_id" = u."id" AND ur."tenant_id" = m."tenant_id" AND r."tenant_id" = m."tenant_id" AND r."status" = 1);`,
		0,
	)
	assertMultiTenantMockSQLCount(t, ctx,
		fmt.Sprintf(`SELECT COUNT(1) FROM (SELECT u."username", m."tenant_id", COUNT(DISTINCT rm."menu_id") AS permission_count FROM plugin_multi_tenant_user_membership m JOIN sys_user u ON u."id" = m."user_id" JOIN sys_user_role ur ON ur."user_id" = u."id" AND ur."tenant_id" = m."tenant_id" JOIN sys_role r ON r."id" = ur."role_id" AND r."tenant_id" = m."tenant_id" AND r."status" = 1 JOIN sys_role_menu rm ON rm."role_id" = r."id" AND rm."tenant_id" = r."tenant_id" WHERE m."status" = 1 AND u."username" IN ('tenant_alpha_ops', 'tenant_beta_auditor', 'tenant_one_click_triple_media_user', 'tenant_cyber_wellness_health_user') GROUP BY u."username", m."tenant_id" HAVING COUNT(DISTINCT rm."menu_id") < %d) under_permissioned_shared_users;`, len(multiTenantMockSharedRolePermissionCodes)),
		0,
	)
}

// setupMultiTenantMockPostgresDatabase creates an isolated PostgreSQL database,
// points GoFrame at it, and executes host install SQL assets before the plugin
// install lifecycle runs.
func setupMultiTenantMockPostgresDatabase(t *testing.T, ctx context.Context, link string, repoRoot string) {
	t.Helper()

	dbDialect, err := dialect.From(link)
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, link, true); err != nil {
		t.Fatalf("prepare PostgreSQL regression database: %v", err)
	}
	if err = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure GoFrame PostgreSQL database: %v", err)
	}
	adapter, err := gcfg.NewAdapterContent("database:\n  default:\n    link: \"" + link + "\"\n")
	if err != nil {
		t.Fatalf("create PostgreSQL config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)

	assets, err := filepath.Glob(filepath.Join(repoRoot, "apps", "lina-core", "manifest", "sql", "*.sql"))
	if err != nil {
		t.Fatalf("glob host SQL assets: %v", err)
	}
	sort.Strings(assets)
	for _, asset := range assets {
		content, readErr := os.ReadFile(asset)
		if readErr != nil {
			t.Fatalf("read host SQL asset %s: %v", asset, readErr)
		}
		translated, translateErr := dbDialect.TranslateDDL(ctx, asset, string(content))
		if translateErr != nil {
			t.Fatalf("translate host SQL asset %s: %v", asset, translateErr)
		}
		for index, statement := range dialect.SplitSQLStatements(translated) {
			if _, err = g.DB().Exec(ctx, statement); err != nil {
				t.Fatalf("execute host SQL asset %s statement %d: %v\n%s", asset, index+1, err, statement)
			}
		}
	}
}

// multiTenantMockPostgresDatabaseLink derives a unique temporary database link
// from the operator-provided PostgreSQL base link.
func multiTenantMockPostgresDatabaseLink(t *testing.T, baseLink string) string {
	t.Helper()

	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		t.Fatalf("parse PostgreSQL base link: %v", err)
	}
	config := db.GetConfig()
	if config == nil {
		t.Fatal("PostgreSQL base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		t.Fatalf("close PostgreSQL base link parser: %v", closeErr)
	}

	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_multi_tenant_mock_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		extra,
	)
}

// dropMultiTenantMockPostgresDatabase removes the isolated regression database.
func dropMultiTenantMockPostgresDatabase(ctx context.Context, targetLink string) error {
	targetDB, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	targetConfig := targetDB.GetConfig()
	if targetConfig == nil {
		return targetDB.Close(ctx)
	}
	targetName := strings.TrimSpace(targetConfig.Name)
	if closeErr := targetDB.Close(ctx); closeErr != nil {
		return closeErr
	}
	if targetName == "" {
		return nil
	}
	extra := strings.TrimSpace(targetConfig.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	adminLink := fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		targetConfig.User,
		targetConfig.Pass,
		targetConfig.Protocol,
		targetConfig.Host,
		targetConfig.Port,
		extra,
	)
	adminDB, err := gdb.New(gdb.ConfigNode{Link: adminLink})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := adminDB.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	_, err = adminDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotePostgresIdentifier(targetName)+" WITH (FORCE)")
	if err != nil {
		_, err = adminDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotePostgresIdentifier(targetName))
	}
	return err
}

// quotePostgresIdentifier safely quotes one generated PostgreSQL identifier.
func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

// assertMultiTenantMockSQLCount checks one SQL count query against an expected value.
func assertMultiTenantMockSQLCount(t *testing.T, ctx context.Context, sql string, expected int) {
	t.Helper()

	value, err := g.DB().GetValue(ctx, sql)
	if err != nil {
		t.Fatalf("query multi-tenant mock count failed: %v\n%s", err, sql)
	}
	if value.Int() != expected {
		t.Fatalf("expected count %d for SQL %s, got %d", expected, sql, value.Int())
	}
}

// assertMultiTenantMockSQLCountAtLeast checks one SQL count query against a
// minimum value.
func assertMultiTenantMockSQLCountAtLeast(t *testing.T, ctx context.Context, sql string, minimum int) {
	t.Helper()

	value, err := g.DB().GetValue(ctx, sql)
	if err != nil {
		t.Fatalf("query multi-tenant mock count failed: %v\n%s", err, sql)
	}
	if value.Int() < minimum {
		t.Fatalf("expected count at least %d for SQL %s, got %d", minimum, sql, value.Int())
	}
}

// assertMultiTenantMockTenantNames verifies every required tenant display name
// is installed by the optional multi-tenant mock data asset.
func assertMultiTenantMockTenantNames(t *testing.T, ctx context.Context) {
	t.Helper()

	for _, name := range multiTenantMockExpectedTenantNames {
		value, err := g.DB().GetValue(ctx, `SELECT COUNT(1) FROM plugin_multi_tenant_tenant WHERE "name" = ?;`, name)
		if err != nil {
			t.Fatalf("query multi-tenant mock tenant name %q failed: %v", name, err)
		}
		if value.Int() != 1 {
			t.Fatalf("expected one multi-tenant mock tenant named %q, got %d", name, value.Int())
		}
	}
}

// readMultiTenantMockSQLAsset reads one SQL asset relative to the repository
// root for mock-data static assertions.
func readMultiTenantMockSQLAsset(t *testing.T, repoRoot string, relativePath string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
	if err != nil {
		t.Fatalf("read multi-tenant mock SQL asset %s: %v", relativePath, err)
	}
	return string(content)
}

// assertMultiTenantMockNotInHostSQL guards the plugin boundary: multi-tenant
// optional mock data belongs to the multi-tenant source plugin, not host mock SQL.
func assertMultiTenantMockNotInHostSQL(t *testing.T, repoRoot string) {
	t.Helper()

	hostAsset := filepath.Join(
		repoRoot,
		"apps",
		"lina-core",
		"manifest",
		"sql",
		"mock-data",
		multiTenantMockHostAssetName,
	)
	if _, err := os.Stat(hostAsset); err == nil {
		t.Fatalf("multi-tenant mock SQL must not be stored under host mock-data: %s", hostAsset)
	} else if !os.IsNotExist(err) {
		t.Fatalf("check host multi-tenant mock SQL asset: %v", err)
	}
}

// assertBilingualMockSQLComments verifies every mock data block starts with an
// adjacent English and Chinese comment describing the data and its purpose.
func assertBilingualMockSQLComments(t *testing.T, assetName string, sql string) {
	t.Helper()

	lines := strings.Split(sql, "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !isMockSQLBlockStart(lines, index, trimmed) {
			continue
		}
		commentBlock := precedingCommentBlock(lines, index)
		if len(commentBlock) == 0 {
			t.Fatalf("%s line %d starts a mock SQL block without comments: %s", assetName, index+1, trimmed)
		}
		var (
			hasEnglish bool
			hasChinese bool
		)
		for _, comment := range commentBlock {
			if strings.Contains(comment, "Mock data:") {
				hasEnglish = true
			}
			if strings.Contains(comment, "模拟数据：") || containsHan(comment) {
				hasChinese = true
			}
		}
		if !hasEnglish || !hasChinese {
			t.Fatalf("%s line %d must have English and Chinese mock-data comments before %s", assetName, index+1, trimmed)
		}
	}
}

// isMockSQLBlockStart reports whether a line begins a standalone mock data
// write block that should carry an explanatory comment.
func isMockSQLBlockStart(lines []string, index int, trimmed string) bool {
	if strings.HasPrefix(trimmed, "WITH v(") {
		return true
	}
	if !strings.HasPrefix(trimmed, "INSERT INTO ") {
		return false
	}
	if index == 0 {
		return true
	}
	previous := strings.TrimSpace(lines[index-1])
	return previous != ")"
}

// precedingCommentBlock returns the contiguous SQL comment block immediately
// above a DML block.
func precedingCommentBlock(lines []string, index int) []string {
	var comments []string
	for cursor := index - 1; cursor >= 0; cursor-- {
		trimmed := strings.TrimSpace(lines[cursor])
		if !strings.HasPrefix(trimmed, "--") {
			break
		}
		comments = append(comments, trimmed)
	}
	return comments
}

// assertMockSQLUserNickname verifies the username and nickname appear in the
// same VALUES row so the display name is tied to the intended demo account.
func assertMockSQLUserNickname(t *testing.T, sql string, username string, nickname string) {
	t.Helper()

	for _, line := range strings.Split(sql, "\n") {
		if strings.Contains(line, "'"+username+"'") && strings.Contains(line, "'"+nickname+"'") {
			return
		}
	}
	t.Fatalf("expected mock user %s to use nickname %q", username, nickname)
}

// assertMultiTenantMockUserPasswordComments verifies that operators can read
// the demo login password before each mock sys_user insertion block.
func assertMultiTenantMockUserPasswordComments(t *testing.T, sql string) {
	t.Helper()

	for _, expected := range []string{
		"Demo login password for all platform mock users below: " + multiTenantMockDemoPassword + ".",
		"以下所有平台 mock 用户的演示登录密码：" + multiTenantMockDemoPassword + "。",
		"Demo login password for all tenant-scoped mock users below: " + multiTenantMockDemoPassword + ".",
		"以下所有租户范围 mock 用户的演示登录密码：" + multiTenantMockDemoPassword + "。",
	} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("multi-tenant mock SQL must document demo password with comment %q", expected)
		}
	}
}

// assertMultiTenantMockSharedMembership verifies that the mock asset includes
// several users with memberships in multiple tenants for switching demos.
func assertMultiTenantMockSharedMembership(t *testing.T, sql string, tenantCodesByUsername map[string][]string) {
	t.Helper()

	if len(tenantCodesByUsername) < multiTenantMockMinSharedUserCount {
		t.Fatalf("expected at least %d shared mock users, got %d", multiTenantMockMinSharedUserCount, len(tenantCodesByUsername))
	}
	for username, tenantCodes := range tenantCodesByUsername {
		if len(tenantCodes) < 2 {
			t.Fatalf("expected mock user %s to have at least two tenant memberships, got %d", username, len(tenantCodes))
		}
		for _, tenantCode := range tenantCodes {
			expected := "('" + username + "', '" + tenantCode + "'"
			if !strings.Contains(sql, expected) {
				t.Fatalf("expected mock user %s to have membership row for tenant %s", username, tenantCode)
			}
		}
	}
}

// assertMultiTenantMockActiveTenantUsers verifies active mock tenants have
// consistent user, membership, role, and user-role rows.
func assertMultiTenantMockActiveTenantUsers(t *testing.T, sql string, tenantUsers map[string]string) {
	t.Helper()

	for tenantCode, username := range tenantUsers {
		roleKey := strings.ReplaceAll(username, "_", "-")
		assertMockSQLLineContainsValues(t, sql, tenantCode, username)
		assertMockSQLLineContainsValues(t, sql, username, tenantCode)
		assertMockSQLLineContainsValues(t, sql, tenantCode, roleKey)
		assertMockSQLLineContainsValues(t, sql, username, roleKey)
	}
}

// assertMultiTenantMockSharedMembershipRoles verifies each shared-user tenant
// membership has a matching tenant-local role binding.
func assertMultiTenantMockSharedMembershipRoles(
	t *testing.T,
	sql string,
	roleKeysByUsername map[string]map[string]string,
) {
	t.Helper()

	for username, roleKeysByTenantCode := range roleKeysByUsername {
		for tenantCode, roleKey := range roleKeysByTenantCode {
			assertMockSQLLineContainsValues(t, sql, username, tenantCode)
			assertMockSQLLineContainsValues(t, sql, tenantCode, roleKey)
			assertMockSQLLineContainsValues(t, sql, username, roleKey)
		}
	}
}

// assertMultiTenantMockSharedRolePermissions verifies shared roles carry enough
// menu permissions for tenant permission-management demos.
func assertMultiTenantMockSharedRolePermissions(t *testing.T, sql string, permissions []string) {
	t.Helper()

	if len(permissions) < multiTenantMockMinMenuPermissions {
		t.Fatalf("expected at least %d shared role permissions, got %d", multiTenantMockMinMenuPermissions, len(permissions))
	}
	block := extractMockSQLBlock(t, sql, "-- Mock data: grant operational and auditor roles read-oriented member and user")
	for _, permission := range permissions {
		if !strings.Contains(block, "'"+permission+"'") {
			t.Fatalf("expected shared mock roles to include permission %q", permission)
		}
	}
}

// extractMockSQLBlock returns one mock-data DML block starting at a stable
// comment marker.
func extractMockSQLBlock(t *testing.T, sql string, marker string) string {
	t.Helper()

	start := strings.Index(sql, marker)
	if start < 0 {
		t.Fatalf("expected mock SQL block marker %q", marker)
	}
	remaining := sql[start:]
	end := strings.Index(remaining, "ON CONFLICT DO NOTHING;")
	if end < 0 {
		t.Fatalf("expected mock SQL block %q to end with ON CONFLICT DO NOTHING", marker)
	}
	return remaining[:end]
}

// assertMockSQLLineContainsValues verifies one SQL line contains all expected
// single-quoted values.
func assertMockSQLLineContainsValues(t *testing.T, sql string, values ...string) {
	t.Helper()

	for _, line := range strings.Split(sql, "\n") {
		matches := true
		for _, value := range values {
			if !strings.Contains(line, "'"+value+"'") {
				matches = false
				break
			}
		}
		if matches {
			return
		}
	}
	t.Fatalf("expected one mock SQL row to contain values %v", values)
}

// assertAllMockSQLUserNicknamesAreChinese verifies every mock user row uses a
// Chinese nickname, including future demo accounts added to the same asset.
func assertAllMockSQLUserNicknamesAreChinese(t *testing.T, sql string) {
	t.Helper()

	for _, line := range strings.Split(sql, "\n") {
		if !strings.Contains(line, "'tenant_") && !strings.Contains(line, "'platform_") {
			continue
		}
		if !strings.Contains(line, "$2a$10$") {
			continue
		}
		values := strings.Split(line, "'")
		nickname := ""
		for index := 1; index < len(values)-2; index += 2 {
			if strings.HasPrefix(values[index], "$2a$") {
				nickname = values[index+2]
				break
			}
		}
		if nickname == "" {
			t.Fatalf("mock user row has no detectable nickname: %s", strings.TrimSpace(line))
		}
		if !containsHan(nickname) {
			t.Fatalf("mock user row nickname %q must use Chinese text: %s", nickname, strings.TrimSpace(line))
		}
	}
}

// mockSQLContainsUserNickname reports whether a stale nickname is still present
// on one of the mock user rows.
func mockSQLContainsUserNickname(sql string, nickname string) bool {
	for _, line := range strings.Split(sql, "\n") {
		if (strings.Contains(line, "'tenant_") || strings.Contains(line, "'platform_")) &&
			strings.Contains(line, "'"+nickname+"'") {
			return true
		}
	}
	return false
}

// mapContainsValue reports whether a string map contains the expected value.
func mapContainsValue(items map[string]string, expected string) bool {
	for _, value := range items {
		if value == expected {
			return true
		}
	}
	return false
}

// containsHan reports whether text contains a CJK Han character.
func containsHan(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
