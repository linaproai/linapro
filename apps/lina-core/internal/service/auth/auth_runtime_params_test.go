// This file verifies runtime authentication behaviors driven by managed
// sys_config parameters.

package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	_ "lina-core/pkg/dbdriver"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/session"
	pluginservicebizctx "lina-core/pkg/plugin/capability/bizctx"
	pluginserviceconfig "lina-core/pkg/plugin/capability/config"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	pluginservicetenantfilter "lina-core/pkg/plugin/capability/tenantfilter"
	"lina-core/pkg/plugin/pluginhost"
)

// TestLoginRejectsBlacklistedIP verifies managed login IP blacklist settings
// are enforced before user lookup.
func TestLoginRejectsBlacklistedIP(t *testing.T) {
	withRuntimeParamValue(t, hostconfig.RuntimeParamKeyLoginBlackIPList, "127.0.0.1")

	username := fmt.Sprintf("blacklist-test-%s", t.Name())
	ctx := newRequestContext(t, "127.0.0.1:19120")

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
	pluginSvc, err := pluginsvc.New(nil, configSvc, bizCtxSvc, cacheCoordSvc, i18nSvc, sessionStore, nil)
	if err != nil {
		panic(err)
	}
	pluginSvc.SetCapabilities(newRuntimeParamAuthTestCapabilities(i18nSvc))
	cacheSvc := kvcache.New()
	return New(configSvc, pluginSvc, orgcap.New(pluginSvc), roleTestService{}, disabledTenantAuthTestService{}, sessionStore, cacheSvc)
}

// runtimeParamAuthTestCapabilities publishes the capabilities required by
// official source-plugin auth hooks during plugin-full tests.
type runtimeParamAuthTestCapabilities struct {
	config       plugincontract.ConfigService
	i18n         plugincontract.I18nService
	tenantFilter plugincontract.TenantFilterService
}

// Ensure runtimeParamAuthTestCapabilities satisfies the source-plugin directory.
var _ pluginhost.Services = (*runtimeParamAuthTestCapabilities)(nil)

// newRuntimeParamAuthTestCapabilities creates the minimal source-plugin
// capability directory needed by auth runtime-parameter tests.
func newRuntimeParamAuthTestCapabilities(i18nSvc i18nsvc.Service) pluginhost.Services {
	bizCtxSvc := pluginservicebizctx.New(nil)
	tenantFilterSvc, err := pluginservicetenantfilter.New(bizCtxSvc, nil)
	if err != nil {
		panic(err)
	}
	return &runtimeParamAuthTestCapabilities{
		config:       pluginserviceconfig.New(),
		i18n:         runtimeParamAuthTestI18n{service: i18nSvc},
		tenantFilter: tenantFilterSvc,
	}
}

// APIDoc returns no apidoc service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) APIDoc() plugincontract.APIDocService { return nil }

// Auth returns no auth service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) Auth() plugincontract.AuthService { return nil }

// BizCtx returns no bizctx service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) BizCtx() plugincontract.BizCtxService { return nil }

// Cache returns no cache service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) Cache() plugincontract.CacheService { return nil }

// Config returns the test host configuration service.
func (s *runtimeParamAuthTestCapabilities) Config() plugincontract.ConfigService {
	if s == nil {
		return nil
	}
	return s.config
}

// HostConfig returns no host config service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) HostConfig() plugincontract.HostConfigService {
	return nil
}

// Manifest returns no manifest service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) Manifest() plugincontract.ManifestService {
	return nil
}

// I18n returns the runtime translation adapter used by auth hooks.
func (s *runtimeParamAuthTestCapabilities) I18n() plugincontract.I18nService {
	if s == nil {
		return nil
	}
	return s.i18n
}

// Notify returns no notification service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) Notify() plugincontract.NotifyService { return nil }

// PluginSettings returns no plugin settings service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) PluginSettings() plugincontract.PluginSettingsService {
	return nil
}

// Org returns the default organization capability fallback service.
func (s *runtimeParamAuthTestCapabilities) Org() orgcap.Service {
	return orgcap.New(nil)
}

// PluginLifecycle returns no plugin lifecycle service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) PluginLifecycle() plugincontract.PluginLifecycleService {
	return nil
}

// PluginState returns no plugin-state service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) PluginState() plugincontract.PluginStateService {
	return nil
}

// Route returns no route service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) Route() plugincontract.RouteService { return nil }

// Session returns no session service for auth runtime-parameter tests.
func (s *runtimeParamAuthTestCapabilities) Session() plugincontract.SessionService { return nil }

// TenantFilter returns the tenant-filter service used by auth hooks.
func (s *runtimeParamAuthTestCapabilities) TenantFilter() plugincontract.TenantFilterService {
	if s == nil {
		return nil
	}
	return s.tenantFilter
}

// Tenant returns the default tenant capability fallback service.
func (s *runtimeParamAuthTestCapabilities) Tenant() tenantcap.Service {
	return tenantcap.New(nil, nil)
}

// runtimeParamAuthTestI18n adapts internal i18n to the source-plugin contract
// for auth runtime-parameter tests.
type runtimeParamAuthTestI18n struct {
	service i18nsvc.Service
}

// GetLocale returns the effective request locale.
func (s runtimeParamAuthTestI18n) GetLocale(ctx context.Context) string {
	if s.service == nil {
		return i18nsvc.DefaultLocale
	}
	return s.service.GetLocale(ctx)
}

// Translate resolves one runtime message.
func (s runtimeParamAuthTestI18n) Translate(ctx context.Context, key string, fallback string) string {
	if s.service == nil {
		return fallback
	}
	return s.service.Translate(ctx, key, fallback)
}

// FindMessageKeys is unused by auth hooks and returns no matches.
func (runtimeParamAuthTestI18n) FindMessageKeys(context.Context, string, string) []string {
	return []string{}
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
