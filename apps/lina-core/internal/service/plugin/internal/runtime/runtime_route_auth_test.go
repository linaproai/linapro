// 本文件验证动态路由在认证后使用真实租户检查插件启用状态，覆盖公开入口和登录失败边界。

package runtime

import (
	"context"
	"net/http"
	"testing"
	"time"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
)

// routeTenantGate记录检查实际使用的租户；平台允许访问，普通租户按测试设置处理。
type routeTenantGate struct {
	IntegrationService
	enabled bool
	tenants []int
}

// CanExposeBusinessEntries模拟租户启用判定，记录每次检查的身份边界。
func (g *routeTenantGate) CanExposeBusinessEntries(ctx context.Context, _ string) bool {
	tenantID := datascope.CurrentTenantID(ctx)
	g.tenants = append(g.tenants, tenantID)
	return tenantID == 0 || g.enabled
}

// routeTenantSession隔离持久化依赖，只接受本测试签发的会话。
type routeTenantSession struct {
	session.Store
	tenantID int
}

// TouchOrValidate验证认证层传递的租户及会话标识。
func (s routeTenantSession) TouchOrValidate(_ context.Context, tenantID int, tokenID string, _ time.Duration) (bool, error) {
	return tenantID == s.tenantID && tokenID == "tenant-order-token", nil
}

// TestAuthorizeDynamicRouteChecksAuthenticatedTenant确保默认平台上下文不能替代真实租户启用状态。
func TestAuthorizeDynamicRouteChecksAuthenticatedTenant(t *testing.T) {
	for _, tc := range []struct {
		name          string
		tenantID      int
		enabled       bool
		public        bool
		invalidToken  bool
		staleScope    bool
		withoutBizCtx bool
		wantStatus    int
		wantChecks    int
	}{
		{name: "租户未启用", tenantID: 42, wantStatus: http.StatusNotFound, wantChecks: 1},
		{name: "租户已启用", tenantID: 42, enabled: true, wantChecks: 1},
		{name: "已认证平台身份", tenantID: 0, wantChecks: 1},
		{name: "无效登录先拒绝", tenantID: 42, invalidToken: true, wantStatus: http.StatusUnauthorized},
		{name: "公开入口保留平台上下文", public: true, wantChecks: 1},
		{name: "公开入口遵守已有租户禁用状态", public: true, tenantID: 42, wantStatus: http.StatusNotFound, wantChecks: 1},
		{name: "JWT覆盖预置平台作用域", tenantID: 42, staleScope: true, wantStatus: http.StatusNotFound, wantChecks: 1},
		{name: "无业务上下文仍按JWT租户检查", tenantID: 42, withoutBizCtx: true, wantStatus: http.StatusNotFound, wantChecks: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				gate   = &routeTenantGate{enabled: tc.enabled}
				ctxSvc = bizctx.New()
				svc    = &serviceImpl{
					configSvc:      routeTestJwtConfig{secret: "tenant-order-secret"},
					sessionStore:   routeTenantSession{tenantID: tc.tenantID},
					userCtx:        ctxSvc,
					integrationSvc: gate,
					roleAccess:     testRoleAccessProjector{projection: &rolesvc.DynamicRouteAccessProjection{}},
				}
				token  = signDynamicRouteImpersonationTestToken(t, svc.configSvc, "tenant-order-token", tc.tenantID, 7, 9)
				access = bridgecontract.AccessLogin
			)
			if tc.invalidToken {
				token = "invalid"
			}
			if tc.public {
				access = bridgecontract.AccessPublic
				token = ""
			}
			request := buildDynamicRouteAccessTestRequest(token)
			if tc.public {
				request.Header.Del("Authorization")
			}
			ctxSvc.Init(request, &model.Context{})
			ctx := request.Context()
			if tc.public && tc.tenantID > 0 {
				ctxSvc.SetTenant(ctx, tc.tenantID)
			}
			if tc.staleScope {
				ctx = datascope.WithTenantScope(ctx, 0)
			}
			if tc.withoutBizCtx {
				svc.userCtx = nil
				ctx = context.Background()
			}
			identity, failure, err := svc.authorizeDynamicRouteRequest(ctx, &dynamicRouteRuntimeState{
				Match: &dynamicRouteMatch{PluginID: "test-tenant-gate", Route: &bridgecontract.RouteContract{Access: access}},
			}, request)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantStatus == 0 {
				if failure != nil {
					t.Fatalf("请求应通过，实际为 %#v", failure)
				}
				if !tc.public && (identity == nil || int(identity.TenantId) != tc.tenantID) {
					t.Fatalf("身份租户错误：%#v", identity)
				}
			} else if failure == nil || int(failure.StatusCode) != tc.wantStatus || identity != nil {
				t.Fatalf("期望拒绝状态 %d 且不发布身份，实际 identity=%#v failure=%#v", tc.wantStatus, identity, failure)
			}
			if len(gate.tenants) != tc.wantChecks {
				t.Fatalf("期望检查 %d 次，实际租户为 %v", tc.wantChecks, gate.tenants)
			}
			if tc.wantChecks > 0 && gate.tenants[0] != tc.tenantID {
				t.Fatalf("期望检查租户 %d，实际 %v", tc.tenantID, gate.tenants)
			}
		})
	}
}
