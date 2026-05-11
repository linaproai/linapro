// This file verifies generic plugin resource permission checks and request
// data-scope propagation used by plugin-owned resource queries.

package plugin

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/service/datascope"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
)

// pluginResourceFakeBizCtx stores one mutable business context for controller tests.
type pluginResourceFakeBizCtx struct {
	ctx *model.Context
}

// Init assigns the supplied context to the fake service.
func (f *pluginResourceFakeBizCtx) Init(_ *ghttp.Request, ctx *model.Context) {
	f.ctx = ctx
}

// Get returns the fake request business context.
func (f *pluginResourceFakeBizCtx) Get(_ context.Context) *model.Context {
	return f.ctx
}

// SetLocale records the current request locale.
func (f *pluginResourceFakeBizCtx) SetLocale(_ context.Context, locale string) {
	if f.ctx != nil {
		f.ctx.Locale = locale
	}
}

// SetUser records the current authenticated user snapshot.
func (f *pluginResourceFakeBizCtx) SetUser(_ context.Context, tokenId string, userId int, username string, status int) {
	if f.ctx != nil {
		f.ctx.TokenId = tokenId
		f.ctx.UserId = userId
		f.ctx.Username = username
		f.ctx.Status = status
	}
}

// SetTenant records the current tenant snapshot.
func (f *pluginResourceFakeBizCtx) SetTenant(_ context.Context, tenantId int) {
	if f.ctx != nil {
		f.ctx.TenantId = tenantId
	}
}

// SetImpersonation records the current impersonation snapshot.
func (f *pluginResourceFakeBizCtx) SetImpersonation(_ context.Context, actingUserId int, tenantId int, actingAsTenant bool, isImpersonation bool) {
	if f.ctx != nil {
		f.ctx.ActingUserId = actingUserId
		f.ctx.TenantId = tenantId
		f.ctx.ActingAsTenant = actingAsTenant
		f.ctx.IsImpersonation = isImpersonation
	}
}


// SetUserAccess records the role-derived data-scope snapshot.
func (f *pluginResourceFakeBizCtx) SetUserAccess(_ context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int) {
	if f.ctx != nil {
		f.ctx.DataScope = dataScope
		f.ctx.DataScopeUnsupported = dataScopeUnsupported
		f.ctx.UnsupportedDataScope = unsupportedDataScope
	}
}

// pluginResourceFakeRoleService returns one fixed access context.
type pluginResourceFakeRoleService struct {
	role.Service

	accessContext *role.UserAccessContext
}

// GetUserAccessContext returns the configured role access snapshot.
func (f pluginResourceFakeRoleService) GetUserAccessContext(_ context.Context, _ int) (*role.UserAccessContext, error) {
	return f.accessContext, nil
}

// pluginResourceFakeService resolves one fixed resource permission.
type pluginResourceFakeService struct {
	pluginsvc.Service

	permission string
}

// ResolveResourcePermission returns the configured resource permission string.
func (f pluginResourceFakeService) ResolveResourcePermission(_ context.Context, _ string, _ string) (string, error) {
	return f.permission, nil
}

// TestEnsurePluginResourcePermissionPropagatesDataScope verifies the generic
// plugin resource route carries the role data-scope snapshot into downstream
// plugin resource queries after controller-level permission checks.
func TestEnsurePluginResourcePermissionPropagatesDataScope(t *testing.T) {
	const requiredPermission = "plugin-dynamic-governance:records:list"

	bizCtx := &pluginResourceFakeBizCtx{
		ctx: &model.Context{
			UserId:   1001,
			Username: "runtime_governance_user",
		},
	}
	controller := &ControllerV1{
		pluginSvc: pluginResourceFakeService{
			permission: requiredPermission,
		},
		bizCtxSvc: bizCtx,
		i18nSvc:   fakePluginI18nTranslator{},
		roleSvc: pluginResourceFakeRoleService{
			accessContext: &role.UserAccessContext{
				Permissions:          []string{requiredPermission},
				DataScope:            datascope.ScopeSelf,
				DataScopeUnsupported: true,
				UnsupportedDataScope: 99,
			},
		},
	}

	allowed, err := controller.ensurePluginResourcePermission(
		context.Background(),
		"plugin-dynamic-governance",
		"records",
	)
	if err != nil {
		t.Fatalf("expected permission check to succeed, got error: %v", err)
	}
	if !allowed {
		t.Fatal("expected plugin resource permission to be allowed")
	}
	if bizCtx.ctx.DataScope != int(datascope.ScopeSelf) {
		t.Fatalf("expected data-scope self to be propagated, got %d", bizCtx.ctx.DataScope)
	}
	if !bizCtx.ctx.DataScopeUnsupported {
		t.Fatal("expected unsupported data-scope flag to be propagated")
	}
	if bizCtx.ctx.UnsupportedDataScope != 99 {
		t.Fatalf("expected unsupported data-scope value 99, got %d", bizCtx.ctx.UnsupportedDataScope)
	}
}
