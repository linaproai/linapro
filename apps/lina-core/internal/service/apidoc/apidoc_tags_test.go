// This file verifies dynamic OpenAPI top-level tag collection and generic
// hierarchical ordering used by the Stoplight sidebar.

package apidoc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/plugin/pluginhost"
)

// testAuthLoginReq is a host auth route used for tag-order builder tests.
type testAuthLoginReq struct {
	g.Meta `path:"/auth/login" method:"post" tags:"Authentication" summary:"User login" dc:"Login with username and password."`
}

// testAuthLoginRes is the response DTO for the host auth route test handler.
type testAuthLoginRes struct{}

// testRoleListReq is a non-auth host route used to interleave unrelated tags.
type testRoleListReq struct {
	g.Meta `path:"/role" method:"get" tags:"Role Management" summary:"Query role list" dc:"Query the paginated role list."`
}

// testRoleListRes is the response DTO for the role list test handler.
type testRoleListRes struct{}

// testLDAPLoginReq is a synthetic LDAP auth-login route for ordering tests.
type testLDAPLoginReq struct {
	g.Meta `path:"/plugins/ldap/login" method:"post" tags:"Auth Login / LDAP" summary:"LDAP directory login" dc:"Verify directory credentials."`
}

// testLDAPLoginRes is the response DTO for the LDAP login test handler.
type testLDAPLoginRes struct{}

// testExtIdentityReq is a synthetic external-identity route for ordering tests.
type testExtIdentityReq struct {
	g.Meta `path:"/plugins/ext/identities" method:"get" tags:"Auth Login / External Identity" summary:"List bound external identities" dc:"Return bound identities."`
}

// testExtIdentityRes is the response DTO for the external identity test handler.
type testExtIdentityRes struct{}

// testOIDCSettingsReq is a synthetic OIDC settings route for ordering tests.
type testOIDCSettingsReq struct {
	g.Meta `path:"/plugins/oidc/settings" method:"get" tags:"Auth Login / OIDC" summary:"Query Generic OIDC settings" dc:"Return OIDC settings."`
}

// testOIDCSettingsRes is the response DTO for the OIDC settings test handler.
type testOIDCSettingsRes struct{}

// testJobTaskReq is a synthetic job-scheduling route for hierarchical ordering tests.
type testJobTaskReq struct {
	g.Meta `path:"/job" method:"get" tags:"Job Scheduling / Task Management" summary:"Get task list" dc:"Query scheduled jobs."`
}

// testJobTaskRes is the response DTO for the job task test handler.
type testJobTaskRes struct{}

// testJobGroupReq is a synthetic job-group route for hierarchical ordering tests.
type testJobGroupReq struct {
	g.Meta `path:"/job-group" method:"get" tags:"Job Scheduling / Group Management" summary:"Get group list" dc:"Query job groups."`
}

// testJobGroupRes is the response DTO for the job group test handler.
type testJobGroupRes struct{}

// testNoticeReq is an unrelated plugin route used in builder ordering tests.
type testNoticeReq struct {
	g.Meta `path:"/plugins/notice" method:"get" tags:"Notices" summary:"Query notices" dc:"Return notices."`
}

// testNoticeRes is the response DTO for the notices test handler.
type testNoticeRes struct{}

// testSlashHierarchyReq uses the compact Parent/Child separator form.
type testSlashHierarchyReq struct {
	g.Meta `path:"/plugins/smart/channel" method:"get" tags:"Smart Center/Channel Management" summary:"List channels" dc:"Return channels."`
}

// testSlashHierarchyRes is the response DTO for the slash-hierarchy test handler.
type testSlashHierarchyRes struct{}

// testSlashHierarchyModelReq is another compact hierarchical tag for clustering.
type testSlashHierarchyModelReq struct {
	g.Meta `path:"/plugins/smart/model" method:"get" tags:"Smart Center/Model Management" summary:"List models" dc:"Return models."`
}

// testSlashHierarchyModelRes is the response DTO for the model hierarchy test handler.
type testSlashHierarchyModelRes struct{}

func testAuthLoginHandler(ctx context.Context, req *testAuthLoginReq) (*testAuthLoginRes, error) {
	return &testAuthLoginRes{}, nil
}

func testRoleListHandler(ctx context.Context, req *testRoleListReq) (*testRoleListRes, error) {
	return &testRoleListRes{}, nil
}

func testLDAPLoginHandler(ctx context.Context, req *testLDAPLoginReq) (*testLDAPLoginRes, error) {
	return &testLDAPLoginRes{}, nil
}

func testExtIdentityHandler(ctx context.Context, req *testExtIdentityReq) (*testExtIdentityRes, error) {
	return &testExtIdentityRes{}, nil
}

func testOIDCSettingsHandler(ctx context.Context, req *testOIDCSettingsReq) (*testOIDCSettingsRes, error) {
	return &testOIDCSettingsRes{}, nil
}

func testJobTaskHandler(ctx context.Context, req *testJobTaskReq) (*testJobTaskRes, error) {
	return &testJobTaskRes{}, nil
}

func testJobGroupHandler(ctx context.Context, req *testJobGroupReq) (*testJobGroupRes, error) {
	return &testJobGroupRes{}, nil
}

func testNoticeHandler(ctx context.Context, req *testNoticeReq) (*testNoticeRes, error) {
	return &testNoticeRes{}, nil
}

func testSlashHierarchyHandler(ctx context.Context, req *testSlashHierarchyReq) (*testSlashHierarchyRes, error) {
	return &testSlashHierarchyRes{}, nil
}

func testSlashHierarchyModelHandler(ctx context.Context, req *testSlashHierarchyModelReq) (*testSlashHierarchyModelRes, error) {
	return &testSlashHierarchyModelRes{}, nil
}

// TestOrderOpenAPIDocumentTagsUsesGenericHierarchy verifies dynamic tags cluster
// by shared parent prefix without hardcoding module names.
func TestOrderOpenAPIDocumentTagsUsesGenericHierarchy(t *testing.T) {
	input := []string{
		"Notices",
		"Auth Login / Discord",
		"Role Management",
		"Auth Login / LDAP",
		"Job Scheduling / Task Management",
		"Authentication",
		"Auth Login / External Identity",
		"Job Scheduling / Group Management",
		"Smart Center/Model Management",
		"Smart Center/Channel Management",
		"授权登录/OIDC",
		"授权登录/LDAP",
		"身份认证",
	}
	got := orderOpenAPIDocumentTags(input)
	want := []string{
		"Auth Login / Discord",
		"Auth Login / External Identity",
		"Auth Login / LDAP",
		"Authentication",
		"Job Scheduling / Group Management",
		"Job Scheduling / Task Management",
		"Notices",
		"Role Management",
		"Smart Center/Channel Management",
		"Smart Center/Model Management",
		"授权登录/LDAP",
		"授权登录/OIDC",
		"身份认证",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d ordered tags, got %d (%v)", len(want), len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("ordered tag[%d]=%q, want %q (full=%v)", index, got[index], want[index], got)
		}
	}

	// Same-prefix families must be contiguous blocks.
	assertContiguousTagPrefix(t, got, "Auth Login / ")
	assertContiguousTagPrefix(t, got, "Job Scheduling / ")
	assertContiguousTagPrefix(t, got, "Smart Center/")
	assertContiguousTagPrefix(t, got, "授权登录/")
}

// TestAssignOpenAPIDocumentTagsCollectsAndOrdersDynamically verifies top-level
// tags are derived from operations after localization, not from a fixed catalog.
func TestAssignOpenAPIDocumentTagsCollectsAndOrdersDynamically(t *testing.T) {
	document := goai.New()
	document.Paths = goai.Paths{
		"/role": {Get: &goai.Operation{Tags: []string{"角色管理"}}},
		"/auth/login": {Post: &goai.Operation{Tags: []string{"身份认证"}}},
		"/x/ldap/login": {Post: &goai.Operation{Tags: []string{"授权登录/LDAP"}}},
		"/x/oidc/settings": {Get: &goai.Operation{Tags: []string{"授权登录/OIDC"}}},
		"/x/notice": {Get: &goai.Operation{Tags: []string{"通知公告"}}},
	}
	assignOpenAPIDocumentTags(document)
	if document.Tags == nil {
		t.Fatalf("expected top-level tags to be assigned")
	}
	got := make([]string, 0, len(*document.Tags))
	for _, item := range *document.Tags {
		got = append(got, item.Name)
	}
	want := []string{"授权登录/LDAP", "授权登录/OIDC", "角色管理", "身份认证", "通知公告"}
	if len(got) != len(want) {
		t.Fatalf("expected dynamic tags %v, got %v", want, got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("dynamic tag[%d]=%q, want %q (full=%v)", index, got[index], want[index], got)
		}
	}
	assertContiguousTagPrefix(t, got, "授权登录/")
}

// TestBuildOrdersTopLevelTagsFromProjectedGroups verifies Build publishes
// dynamically ordered top-level tags from the current route/plugin projection.
func TestBuildOrdersTopLevelTagsFromProjectedGroups(t *testing.T) {
	server := g.Server("apidoc-tag-order-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Bind(testRoleListHandler)
		group.Bind(testNoticeHandler)
		group.Bind(testAuthLoginHandler)
		group.Bind(testLDAPLoginHandler)
		group.Bind(testExtIdentityHandler)
		group.Bind(testOIDCSettingsHandler)
		group.Bind(testJobTaskHandler)
		group.Bind(testJobGroupHandler)
		group.Bind(testSlashHierarchyHandler)
		group.Bind(testSlashHierarchyModelHandler)
	})
	server.Start()
	defer server.Shutdown()
	time.Sleep(100 * time.Millisecond)

	pluginProvider := &testPluginRouteProvider{
		enabledByID: map[string]bool{
			"plugin-auth-ldap":    true,
			"plugin-extlogin":     true,
			"plugin-oidc-generic": true,
			"plugin-notice":       true,
			"plugin-smart":        true,
		},
		sourceRoutes: []pluginhost.SourceRouteBinding{
			{PluginID: "plugin-notice", Method: "GET", Path: "/api/v1/plugins/notice", Handler: testNoticeHandler, Documentable: true},
			{PluginID: "plugin-oidc-generic", Method: "GET", Path: "/api/v1/plugins/oidc/settings", Handler: testOIDCSettingsHandler, Documentable: true},
			{PluginID: "plugin-auth-ldap", Method: "POST", Path: "/api/v1/plugins/ldap/login", Handler: testLDAPLoginHandler, Documentable: true},
			{PluginID: "plugin-extlogin", Method: "GET", Path: "/api/v1/plugins/ext/identities", Handler: testExtIdentityHandler, Documentable: true},
			{PluginID: "plugin-smart", Method: "GET", Path: "/api/v1/plugins/smart/channel", Handler: testSlashHierarchyHandler, Documentable: true},
			{PluginID: "plugin-smart", Method: "GET", Path: "/api/v1/plugins/smart/model", Handler: testSlashHierarchyModelHandler, Documentable: true},
		},
	}

	service := New(&testConfigProvider{}, bizctx.New(), i18nsvc.New(bizctx.New(), configsvc.New(), cachecoord.Default(nil)), pluginProvider)
	document, err := service.Build(context.Background(), server)
	if err != nil {
		t.Fatalf("expected hosted apidoc build to succeed, got %v", err)
	}
	if document.Tags == nil {
		t.Fatalf("expected top-level OpenAPI tags to be published for sidebar order")
	}
	got := make([]string, 0, len(*document.Tags))
	for _, item := range *document.Tags {
		got = append(got, item.Name)
	}
	assertContiguousTagPrefix(t, got, "Auth Login / ")
	assertContiguousTagPrefix(t, got, "Job Scheduling / ")
	assertContiguousTagPrefix(t, got, "Smart Center/")

	// Chinese locale must re-collect and re-order from localized display names.
	restoreCatalog := registerOpenAPITestCatalog("zh-CN", map[string]string{
		"core.internal.service.apidoc.testAuthLoginReq.meta.tags":           "身份认证",
		"core.internal.service.apidoc.testExtIdentityReq.meta.tags":         "授权登录/外部身份",
		"core.internal.service.apidoc.testLDAPLoginReq.meta.tags":           "授权登录/LDAP",
		"core.internal.service.apidoc.testOIDCSettingsReq.meta.tags":        "授权登录/OIDC",
		"core.internal.service.apidoc.testRoleListReq.meta.tags":            "角色管理",
		"core.internal.service.apidoc.testNoticeReq.meta.tags":              "通知公告",
		"core.internal.service.apidoc.testJobTaskReq.meta.tags":             "任务调度/任务管理",
		"core.internal.service.apidoc.testJobGroupReq.meta.tags":            "任务调度/分组管理",
		"core.internal.service.apidoc.testSlashHierarchyReq.meta.tags":      "智能中心/通道管理",
		"core.internal.service.apidoc.testSlashHierarchyModelReq.meta.tags": "智能中心/模型管理",
	})
	defer restoreCatalog()

	zhCtx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: "zh-CN"},
	)
	zhDocument, err := service.Build(zhCtx, server)
	if err != nil {
		t.Fatalf("expected Chinese apidoc build to succeed, got %v", err)
	}
	if zhDocument.Tags == nil {
		t.Fatalf("expected localized top-level OpenAPI tags")
	}
	zhGot := make([]string, 0, len(*zhDocument.Tags))
	for _, item := range *zhDocument.Tags {
		zhGot = append(zhGot, item.Name)
	}
	assertContiguousTagPrefix(t, zhGot, "授权登录/")
	assertContiguousTagPrefix(t, zhGot, "任务调度/")
	assertContiguousTagPrefix(t, zhGot, "智能中心/")
}

// assertContiguousTagPrefix fails when tags sharing prefix are not one block.
func assertContiguousTagPrefix(t *testing.T, tags []string, prefix string) {
	t.Helper()
	first := -1
	last := -1
	count := 0
	for index, name := range tags {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		count++
		if first < 0 {
			first = index
		}
		last = index
	}
	if count == 0 {
		t.Fatalf("expected at least one tag with prefix %q in %v", prefix, tags)
	}
	if last-first+1 != count {
		t.Fatalf("expected tags with prefix %q to be contiguous in %v", prefix, tags)
	}
}
