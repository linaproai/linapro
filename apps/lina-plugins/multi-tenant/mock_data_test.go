// This file verifies the multi-tenant plugin mock-data and manifest assets.

package multitenant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"gopkg.in/yaml.v3"
)

const (
	multiTenantPluginID               = "multi-tenant"
	multiTenantMockDemoPassword       = "admin123"
	multiTenantMockMinSharedUserCount = 4
	multiTenantMockMinMenuPermissions = 8
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
// shipped by the optional mock-data asset.
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

// multiTenantMockSharedRoleKeysByUsername lists the tenant-local roles that each
// shared mock user needs in each tenant it can switch to.
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

// testPluginManifest stores the subset of plugin.yaml needed by asset tests.
type testPluginManifest struct {
	ID    string                `yaml:"id"`
	Menus []*testPluginMenuSpec `yaml:"menus"`
}

// testPluginMenuSpec stores the subset of one plugin menu entry used by tests.
type testPluginMenuSpec struct {
	Key       string `yaml:"key"`
	Type      string `yaml:"type"`
	ParentKey string `yaml:"parent_key"`
	Perms     string `yaml:"perms"`
}

// TestMultiTenantMockDataContainsExpectedTenantNames keeps the optional
// mock-data asset aligned with the required display-name list.
func TestMultiTenantMockDataContainsExpectedTenantNames(t *testing.T) {
	pluginSQL := readMultiTenantMockSQLAsset(t)
	for _, name := range multiTenantMockExpectedTenantNames {
		if got := strings.Count(pluginSQL, "'"+name+"'"); got != 1 {
			t.Fatalf("expected tenant mock name %q to appear once as a SQL value, got %d", name, got)
		}
	}
}

// TestMultiTenantManifestTenantManagementButtonsMatchWorkbench keeps Menu
// Management button permissions aligned with the actual tenant page buttons.
func TestMultiTenantManifestTenantManagementButtonsMatchWorkbench(t *testing.T) {
	manifestBytes, err := os.ReadFile(filepath.Join(pluginRootDir(t), "plugin.yaml"))
	if err != nil {
		t.Fatalf("read multi-tenant manifest: %v", err)
	}

	manifest := &testPluginManifest{}
	if err = yaml.Unmarshal(manifestBytes, manifest); err != nil {
		t.Fatalf("parse multi-tenant manifest: %v", err)
	}
	if manifest.ID != multiTenantPluginID {
		t.Fatalf("expected plugin id %q, got %q", multiTenantPluginID, manifest.ID)
	}

	buttons := make(map[string]string)
	for _, item := range manifest.Menus {
		if item == nil || item.ParentKey != "plugin:multi-tenant:platform:tenants" || item.Type != "B" {
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

// TestMultiTenantMockDataDocumentsBlocksAndNicknames keeps mock SQL assets
// readable for operators who inspect demo data directly in management tables.
func TestMultiTenantMockDataDocumentsBlocksAndNicknames(t *testing.T) {
	pluginSQL := readMultiTenantMockSQLAsset(t)
	assertBilingualMockSQLComments(t, "multi-tenant plugin mock SQL", pluginSQL)
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

// readMultiTenantMockSQLAsset reads the plugin-owned mock SQL asset.
func readMultiTenantMockSQLAsset(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(pluginRootDir(t), "manifest", "sql", "mock-data", "001-multi-tenant-demo-data.sql"))
	if err != nil {
		t.Fatalf("read multi-tenant mock SQL asset: %v", err)
	}
	return string(content)
}

// pluginRootDir resolves the plugin module root from the current test package.
func pluginRootDir(t *testing.T) string {
	t.Helper()

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory: %v", err)
	}
	return workingDir
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
	block := extractMockSQLBlock(t, sql, "-- Mock data: grant operational and auditor roles read-oriented user")
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
