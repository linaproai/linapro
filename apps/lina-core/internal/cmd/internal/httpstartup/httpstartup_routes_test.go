// This file verifies host HTTP route registration and protected API routing.

package httpstartup

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/dict"
	filesvc "lina-core/internal/service/file"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/menu"
	"lina-core/internal/service/middleware"
	"lina-core/internal/service/notify"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/internal/service/sysconfig"
	sysinfosvc "lina-core/internal/service/sysinfo"
	"lina-core/internal/service/user"
	"lina-core/internal/service/usermsg"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestUploadedFileAccessRouteIsPublic verifies direct upload URLs remain
// browser-loadable without making the whole file controller public.
func TestUploadedFileAccessRouteIsPublic(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-upload-public-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	bindHostAPIRoutes(ctx, server, runtime)

	if err := server.Start(); err != nil {
		t.Fatalf("start route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown route test server: %v", err)
		}
	})

	uploadAccess := mustFindRoute(t, server, "GET", "/api/v1/uploads/*path")
	if strings.Contains(uploadAccess.Middleware, "Service.Auth") {
		t.Fatalf("expected upload URL access route to be public, middleware=%s", uploadAccess.Middleware)
	}
	if strings.Contains(uploadAccess.Middleware, "Service.Permission") {
		t.Fatalf("expected upload URL access route to skip permission middleware, middleware=%s", uploadAccess.Middleware)
	}

	fileUpload := mustFindRoute(t, server, "POST", "/api/v1/file/upload")
	if !strings.Contains(fileUpload.Middleware, "Service.Auth") {
		t.Fatalf("expected file upload route to remain authenticated, middleware=%s", fileUpload.Middleware)
	}
	if !strings.Contains(fileUpload.Middleware, "Service.Permission") {
		t.Fatalf("expected file upload route to keep permission middleware, middleware=%s", fileUpload.Middleware)
	}

	userBatchUpdate := mustFindRoute(t, server, "PUT", "/api/v1/user")
	if !strings.Contains(userBatchUpdate.Middleware, "Service.Auth") {
		t.Fatalf("expected user batch-update route to remain authenticated, middleware=%s", userBatchUpdate.Middleware)
	}
	if !strings.Contains(userBatchUpdate.Middleware, "Service.Permission") {
		t.Fatalf("expected user batch-update route to keep permission middleware, middleware=%s", userBatchUpdate.Middleware)
	}
}

// TestPluginManagementRuntimeRoutesAreBound verifies plugin runtime-management
// endpoints added by the upgrade workflow are reachable through the protected API tree.
func TestPluginManagementRuntimeRoutesAreBound(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-plugin-runtime-routes-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	bindHostAPIRoutes(ctx, server, runtime)

	if err := server.Start(); err != nil {
		t.Fatalf("start plugin route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown plugin route test server: %v", err)
		}
	})

	routes := []struct {
		method string
		route  string
	}{
		{method: "GET", route: "/api/v1/plugins/{id}"},
		{method: "GET", route: "/api/v1/plugins/{id}/dependencies"},
		{method: "GET", route: "/api/v1/plugins/{id}/upgrade/preview"},
		{method: "POST", route: "/api/v1/plugins/{id}/upgrade"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.route, func(t *testing.T) {
			item := mustFindRoute(t, server, route.method, route.route)
			if !strings.Contains(item.Middleware, "Service.Auth") {
				t.Fatalf("expected plugin runtime route to remain authenticated, middleware=%s", item.Middleware)
			}
			if !strings.Contains(item.Middleware, "Service.Permission") {
				t.Fatalf("expected plugin runtime route to keep permission middleware, middleware=%s", item.Middleware)
			}
		})
	}
}

// TestDynamicPluginRootRoutesPrecedeSPAFallback verifies root-level dynamic
// plugin paths are claimed before the catch-all frontend fallback can redirect
// API-style requests to the host SPA entry.
func TestDynamicPluginRootRoutesPrecedeSPAFallback(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-dynamic-plugin-root-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	bindHostAPIRoutes(ctx, server, runtime)
	if err := bindFrontendAssetRoutes(ctx, server, runtime.pluginSvc, "/admin"); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("start dynamic route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown dynamic route test server: %v", err)
		}
	})

	response, err := http.Get(fmt.Sprintf(
		"http://127.0.0.1:%d/x/plugin-dev-route-missing/api/v1/backend-summary",
		server.GetListenedPort(),
	))
	if err != nil {
		t.Fatalf("request root dynamic plugin route: %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatalf("close dynamic route response body: %v", closeErr)
		}
	}()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dynamic route response body: %v", err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected dynamic route 404, got status=%d body=%q", response.StatusCode, string(body))
	}
	if strings.TrimSpace(string(body)) != "Dynamic route not found" {
		t.Fatalf("expected dynamic route not found body, got %q", string(body))
	}
	if response.Header.Get("Location") != "" {
		t.Fatalf("expected no SPA redirect for dynamic route, got location=%q", response.Header.Get("Location"))
	}
}

// newRouteBindingTestRuntime creates the shared service graph required by
// route-binding tests without starting cluster, plugin, or cron lifecycles.
func newRouteBindingTestRuntime(ctx context.Context) *httpRuntime {
	var (
		configSvc     = config.New()
		clusterSvc    = cluster.New(configSvc.GetCluster(ctx))
		bizCtxSvc     = bizctx.New()
		sessionStore  = session.NewDBStore()
		cacheCoordSvc = cachecoord.Default(clusterSvc)
		i18nService   = i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
		lockerSvc     = locker.New()
		pluginRuntime = pluginsvc.NewRuntimeDelegate()
		orgCapSvc     = orgspi.New(orgspi.NewManager(), pluginRuntime, pluginRuntime.OrgProviderEnv)
		tenantSvc     = tenantspi.New(tenantspi.NewManager(), pluginRuntime, pluginRuntime.TenantProviderEnv, bizCtxSvc)
		roleSvc       = role.New(pluginRuntime, bizCtxSvc, configSvc, i18nService, orgCapSvc, tenantSvc)
		kvCacheSvc    = kvcache.New()
		dictSvc       = dict.New(i18nService)
		scopeSvc      = datascope.New(bizCtxSvc, roleSvc, orgCapSvc.Scope())
	)
	roleSvc.SetDataScopeService(scopeSvc)
	var (
		menuSvc   = menu.New(pluginRuntime, i18nService, roleSvc, tenantSvc)
		notifySvc = notify.New(tenantSvc)
		authSvc   = auth.New(configSvc, pluginRuntime, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
	)
	var (
		objectStorage = storagesvc.New(storagesvc.Config{NamespaceRoots: map[string]string{
			storagesvc.NamespaceFiles:   configSvc.GetUploadPath(ctx),
			storagesvc.NamespacePlugins: configSvc.GetPluginDynamicStoragePath(ctx),
		}})
		storageRuntime       = pluginsvc.NewStorageProviderRuntime(pluginRuntime)
		localStorageProvider = pluginsvc.NewLocalStorageProvider(objectStorage)
		fileStorage          = storagesvc.NewResolvingService(objectStorage, storageRuntime, localStorageProvider)
		fileSvc              = filesvc.New(configSvc, fileStorage, bizCtxSvc, dictSvc, scopeSvc)
		sysConfigSvc         = sysconfig.New(configSvc, i18nService)
	)
	sysInfoSvc, err := sysinfosvc.New(configSvc, clusterSvc, nil, cacheCoordSvc)
	if err != nil {
		panic(err)
	}
	var (
		userSvc             = user.New(authSvc, bizCtxSvc, i18nService, orgCapSvc, roleSvc, scopeSvc, tenantSvc)
		userMsgSvc          = usermsg.New(bizCtxSvc, notifySvc, i18nService)
		jobRegistry         = jobhandlersvc.New()
		jobMgmtSvc          = jobmgmtsvc.New(bizCtxSvc, configSvc, i18nService, jobRegistry, nil, scopeSvc)
		hostConfigSvc       = pluginsvc.NewHostConfigService(configSvc)
		pluginConfigFactory = pluginsvc.NewPluginConfigFactoryWithHostStaticConfig("", "", configSvc)
	)
	hostLockSvc, err := hostlock.New(lockerSvc)
	if err != nil {
		panic(err)
	}
	capabilities, err := pluginsvc.NewHostServices(
		apidoc.New(configSvc, bizCtxSvc, i18nService, pluginRuntime),
		authSvc,
		bizCtxSvc,
		roleSvc,
		hostConfigSvc,
		scopeSvc,
		cacheCoordSvc,
		i18nService,
		pluginRuntime,
		pluginRuntime,
		userSvc,
		fileSvc,
		jobMgmtSvc,
		orgCapSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
		hostLockSvc,
		pluginConfigFactory,
		storageRuntime,
		localStorageProvider,
	)
	if err != nil {
		panic(err)
	}
	pluginSvc, err := pluginsvc.New(
		clusterSvc,
		configSvc,
		bizCtxSvc,
		cacheCoordSvc,
		i18nService,
		sessionStore,
		roleSvc,
		lockerSvc,
		nil,
		capabilities,
		orgCapSvc,
		tenantSvc,
		pluginConfigFactory,
		hostConfigSvc,
	)
	if err != nil {
		panic(err)
	}
	if err = pluginRuntime.BindService(pluginSvc); err != nil {
		panic(err)
	}
	return &httpRuntime{
		configSvc:     configSvc,
		clusterSvc:    clusterSvc,
		pluginSvc:     pluginSvc,
		authSvc:       authSvc,
		bizCtxSvc:     bizCtxSvc,
		i18nSvc:       i18nService,
		orgSvc:        orgCapSvc,
		roleSvc:       roleSvc,
		dictSvc:       dictSvc,
		fileSvc:       fileSvc,
		menuSvc:       menuSvc,
		sysConfigSvc:  sysConfigSvc,
		sysInfoSvc:    sysInfoSvc,
		userSvc:       userSvc,
		userMsgSvc:    userMsgSvc,
		jobRegistry:   jobRegistry,
		jobMgmtSvc:    jobMgmtSvc,
		middlewareSvc: middleware.New(authSvc, bizCtxSvc, configSvc, i18nService, roleSvc, tenantSvc),
	}
}

// mustFindRoute returns one route item by method and path.
func mustFindRoute(t *testing.T, server *ghttp.Server, method string, route string) ghttp.RouterItem {
	t.Helper()

	for _, item := range server.GetRoutes() {
		if item.Method == method && item.Route == route {
			return item
		}
	}
	t.Fatalf("expected route %s %s to be registered", method, route)
	return ghttp.RouterItem{}
}
