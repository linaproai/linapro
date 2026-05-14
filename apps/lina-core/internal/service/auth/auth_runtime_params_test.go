// This file verifies runtime authentication behaviors driven by managed
// sys_config parameters.

package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/orgcap"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/session"
)

// TestLoginRejectsBlacklistedIP verifies managed login IP blacklist settings
// are enforced before user lookup.
func TestLoginRejectsBlacklistedIP(t *testing.T) {
	withRuntimeParamValue(t, hostconfig.RuntimeParamKeyLoginBlackIPList, "127.0.0.1")

	username := fmt.Sprintf("blacklist-test-%s", t.Name())
	ctx := newRequestContext(t, "127.0.0.1:18080")

	_, err := newRuntimeParamAuthTestService().Login(ctx, LoginInput{
		Username: username,
		Password: "ignored",
	})
	if err == nil {
		t.Fatal("expected blacklisted login attempt to fail")
	}
	if localized := i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil)).LocalizeError(context.Background(), err); localized != "登录IP已被禁止" {
		t.Fatalf("expected blacklisted login error %q, got %q", "登录IP已被禁止", localized)
	}
}

// newRuntimeParamAuthTestService constructs auth with explicit test
// dependencies while still reading runtime params from the real config service.
func newRuntimeParamAuthTestService() Service {
	configSvc := hostconfig.New()
	bizCtxSvc := bizctx.New()
	cacheCoordSvc := cachecoord.Default(nil)
	i18nSvc := i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
	sessionStore := session.NewDBStore()
	pluginSvc := pluginsvc.New(nil, configSvc, bizCtxSvc, cacheCoordSvc, i18nSvc, sessionStore)
	cacheSvc := kvcache.New()
	return New(configSvc, pluginSvc, orgcap.New(pluginSvc), roleTestService{}, disabledTenantAuthTestService{}, sessionStore, cacheSvc)
}

// newRequestContext builds one request-backed context carrying the supplied
// remote address for auth service tests.
func newRequestContext(t *testing.T, remoteAddr string) context.Context {
	t.Helper()

	httpReq, err := http.NewRequest(http.MethodPost, "http://localhost/api/v1/auth/login", nil)
	if err != nil {
		t.Fatalf("build http request: %v", err)
	}
	httpReq.RemoteAddr = remoteAddr
	httpReq.Header.Set("User-Agent", "runtime-param-test")

	req := &ghttp.Request{Request: httpReq}
	return req.Context()
}

// withRuntimeParamValue temporarily overrides one protected runtime parameter
// and restores the original sys_config record during cleanup.
func withRuntimeParamValue(t *testing.T, key string, value string) {
	t.Helper()

	ctx := context.Background()
	original, err := queryRuntimeParam(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s: %v", key, err)
	}

	if original == nil {
		_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:   key,
			Key:    key,
			Value:  value,
			Remark: "test override",
		}).Insert()
		if err != nil {
			t.Fatalf("insert runtime param %s: %v", key, err)
		}
		markRuntimeParamChanged(t, ctx)
		t.Cleanup(func() {
			if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete(); cleanupErr != nil {
				t.Fatalf("cleanup runtime param %s: %v", key, cleanupErr)
			}
			markRuntimeParamChanged(t, ctx)
		})
		return
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: value}).
		Update()
	if err != nil {
		t.Fatalf("update runtime param %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{
				Name:   original.Name,
				Key:    original.Key,
				Value:  original.Value,
				Remark: original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

// markRuntimeParamChanged bumps the runtime-parameter revision for tests after
// direct sys_config mutations.
func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

// queryRuntimeParam loads one sys_config record by protected runtime-parameter key.
func queryRuntimeParam(ctx context.Context, key string) (*entity.SysConfig, error) {
	var runtimeParam *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Scan(&runtimeParam)
	if err != nil {
		return nil, err
	}
	return runtimeParam, nil
}
