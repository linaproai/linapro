// This file provides registration-only capability stubs for plugin service
// tests that load real source plugins without executing their business paths.

package testutil

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

type testNoopAdminCapabilities struct{}

var _ capability.AdminServices = (*testNoopAdminCapabilities)(nil)

func (testNoopAdminCapabilities) Users() capabilityusercap.AdminService {
	return testNoopUsers{}
}

func (testNoopAdminCapabilities) Auth() authcap.AdminService {
	return authcap.NewAdmin(testNoopAuthz{})
}

func (testNoopAdminCapabilities) Dict() capabilitydictcap.AdminService {
	return testNoopDict{}
}

func (testNoopAdminCapabilities) Files() capabilityfilecap.AdminService {
	return testNoopFiles{}
}

func (testNoopAdminCapabilities) Sessions() capabilitysessioncap.AdminService {
	return testNoopSessions{}
}

func (testNoopAdminCapabilities) HostConfig() hostconfigcap.AdminService {
	return testNoopRuntimeConfig{}
}

func (testNoopAdminCapabilities) Notifications() capabilitynotifycap.AdminService {
	return testNoopNotifications{}
}

func (testNoopAdminCapabilities) Plugins() capabilityplugincap.AdminService {
	return testNoopPlugins{}
}

func (testNoopAdminCapabilities) Jobs() capabilityjobcap.AdminService {
	return testNoopJobs{}
}

func (testNoopAdminCapabilities) Infra() capabilityinfracap.AdminService {
	return testNoopInfra{}
}

type testNoopAPIDoc struct{}

func (testNoopAPIDoc) ResolveRouteText(_ context.Context, input apidoccap.RouteTextInput) apidoccap.RouteTextOutput {
	return apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
}

func (testNoopAPIDoc) ResolveRouteTexts(_ context.Context, inputs []apidoccap.RouteTextInput) []apidoccap.RouteTextOutput {
	outputs := make([]apidoccap.RouteTextOutput, 0, len(inputs))
	for _, input := range inputs {
		outputs = append(outputs, apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
	}
	return outputs
}

func (testNoopAPIDoc) FindRouteTitleOperationKeys(context.Context, string) []string {
	return nil
}

type testNoopAuth struct{}

func (testNoopAuth) SelectTenant(context.Context, token.SelectTenantInput) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

func (testNoopAuth) SwitchTenant(context.Context, token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

func (testNoopAuth) IssueImpersonationToken(context.Context, token.ImpersonationTokenIssueInput) (*token.ImpersonationTokenOutput, error) {
	return &token.ImpersonationTokenOutput{}, nil
}

func (testNoopAuth) RevokeImpersonationToken(context.Context, token.ImpersonationTokenRevokeInput) error {
	return nil
}

type testNoopBizCtx struct{}

func (testNoopBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{PlatformBypass: true}
}

type testNoopCache struct{}

func (testNoopCache) Get(context.Context, string, string) (*cachecap.CacheItem, bool, error) {
	return nil, false, nil
}

func (testNoopCache) GetMany(_ context.Context, in cachecap.GetManyInput) (*cachecap.GetManyOutput, error) {
	return &cachecap.GetManyOutput{
		Items:       map[string]*cachecap.CacheItem{},
		MissingKeys: append([]string(nil), in.Keys...),
	}, nil
}

func (testNoopCache) Set(_ context.Context, namespace string, key string, value string, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindString, Value: value}, nil
}

func (testNoopCache) SetMany(_ context.Context, in cachecap.SetManyInput) (*cachecap.SetManyOutput, error) {
	output := &cachecap.SetManyOutput{Items: map[string]*cachecap.CacheItem{}}
	for _, item := range in.Items {
		output.Items[item.Key] = &cachecap.CacheItem{Key: in.Namespace + ":" + item.Key, ValueKind: cachecap.CacheValueKindString, Value: item.Value}
	}
	return output, nil
}

func (testNoopCache) Delete(context.Context, string, string) error {
	return nil
}

func (testNoopCache) DeleteMany(context.Context, cachecap.DeleteManyInput) error {
	return nil
}

func (testNoopCache) Incr(_ context.Context, namespace string, key string, delta int64, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindInt, IntValue: delta}, nil
}

func (testNoopCache) Expire(context.Context, string, string, time.Duration) (bool, *time.Time, error) {
	return false, nil, nil
}

type testNoopConfig struct{}

func (testNoopConfig) Get(context.Context, string) (*gvar.Var, error) {
	return nil, nil
}

func (testNoopConfig) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func (testNoopConfig) Scan(context.Context, string, any) error {
	return nil
}

func (testNoopConfig) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

func (testNoopConfig) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

func (testNoopConfig) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

func (testNoopConfig) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

type testNoopI18n struct{}

func (testNoopI18n) GetLocale(context.Context) string {
	return "zh-CN"
}

func (testNoopI18n) Translate(_ context.Context, _ string, fallback string) string {
	return fallback
}

func (testNoopI18n) FindMessageKeys(context.Context, string, string) []string {
	return nil
}

type testNoopPluginLifecycle struct{}

func (testNoopPluginLifecycle) EnsureTenantPluginDisableAllowed(context.Context, string, int) error {
	return nil
}

func (testNoopPluginLifecycle) NotifyTenantPluginDisabled(context.Context, string, int) {}

func (testNoopPluginLifecycle) EnsureTenantDeleteAllowed(context.Context, int) error {
	return nil
}

func (testNoopPluginLifecycle) NotifyTenantDeleted(context.Context, int) {}

type testNoopPluginState struct{}

func (testNoopPluginState) IsEnabled(context.Context, string) bool {
	return false
}

func (testNoopPluginState) IsProviderEnabled(context.Context, string) bool {
	return false
}

func (testNoopPluginState) IsEnabledAuthoritative(context.Context, string) bool {
	return false
}

type testNoopRoute struct{}

func (testNoopRoute) DynamicRouteMetadata(context.Context) *routecap.DynamicRouteMetadata {
	return nil
}

type testNoopTenantFilter struct{}

func (testNoopTenantFilter) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{PlatformBypass: true}
}

func (testNoopTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

type testNoopUsers struct{}

func (testNoopUsers) Current(context.Context, capmodel.CapabilityContext) (*capabilityusercap.UserProjection, error) {
	return nil, nil
}

func (testNoopUsers) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

func (testNoopUsers) BatchResolve(_ context.Context, _ capmodel.CapabilityContext, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey]{
		Items:      map[capabilityusercap.ResolveKey]*capabilityusercap.UserProjection{},
		MissingIDs: []capabilityusercap.ResolveKey{},
	}
	for _, id := range input.IDs {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("id:"+string(id)))
	}
	for _, username := range input.Usernames {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("username:"+username))
	}
	for _, contact := range input.Contacts {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("contact:"+contact))
	}
	return result, nil
}

func (testNoopUsers) Search(context.Context, capmodel.CapabilityContext, capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
}

func (testNoopUsers) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilityusercap.UserID) error {
	return nil
}

func (testNoopUsers) SetStatus(context.Context, capmodel.CapabilityContext, capabilityusercap.UserID, string) error {
	return nil
}

type testNoopAuthz struct{}

func (testNoopAuthz) BatchGetPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityauthz.PermissionKey) (*capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey], error) {
	return &capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey]{
		Items:      map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionProjection{},
		MissingIDs: append([]capabilityauthz.PermissionKey(nil), keys...),
	}, nil
}

func (testNoopAuthz) BatchHasPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityauthz.PermissionKey) (map[capabilityauthz.PermissionKey]bool, error) {
	result := make(map[capabilityauthz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		result[key] = false
	}
	return result, nil
}

func (testNoopAuthz) HasPermission(context.Context, capmodel.CapabilityContext, capabilityauthz.PermissionKey) (bool, error) {
	return false, nil
}

func (testNoopAuthz) IsPlatformAdmin(context.Context, capmodel.CapabilityContext, capabilityauthz.UserID) (bool, error) {
	return false, nil
}

func (testNoopAuthz) ReplaceRolePermissions(context.Context, capmodel.CapabilityContext, capabilityauthz.RoleID, []capabilityauthz.PermissionKey) error {
	return nil
}

type testNoopDict struct{}

func (testNoopDict) ResolveLabels(_ context.Context, _ capmodel.CapabilityContext, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value], error) {
	return &capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value]{
		Items:      map[capabilitydictcap.Value]*capabilitydictcap.LabelProjection{},
		MissingIDs: append([]capabilitydictcap.Value(nil), input.Values...),
	}, nil
}

func (testNoopDict) ListValues(context.Context, capmodel.CapabilityContext, capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.LabelProjection], error) {
	return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{Items: []*capabilitydictcap.LabelProjection{}}, nil
}

func (testNoopDict) EnsureValuesVisible(context.Context, capmodel.CapabilityContext, capabilitydictcap.ResolveInput) error {
	return nil
}

func (testNoopDict) Refresh(context.Context, capmodel.CapabilityContext, capabilitydictcap.Type) error {
	return nil
}

type testNoopFiles struct{}

func (testNoopFiles) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityfilecap.FileID) (*capmodel.BatchResult[*capabilityfilecap.FileProjection, capabilityfilecap.FileID], error) {
	return &capmodel.BatchResult[*capabilityfilecap.FileProjection, capabilityfilecap.FileID]{
		Items:      map[capabilityfilecap.FileID]*capabilityfilecap.FileProjection{},
		MissingIDs: append([]capabilityfilecap.FileID(nil), ids...),
	}, nil
}

func (testNoopFiles) Search(context.Context, capmodel.CapabilityContext, capabilityfilecap.SearchInput) (*capmodel.PageResult[*capabilityfilecap.FileProjection], error) {
	return &capmodel.PageResult[*capabilityfilecap.FileProjection]{Items: []*capabilityfilecap.FileProjection{}}, nil
}

func (testNoopFiles) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilityfilecap.FileID) error {
	return nil
}

func (testNoopFiles) Delete(context.Context, capmodel.CapabilityContext, []capabilityfilecap.FileID) error {
	return nil
}

type testNoopRuntimeConfig struct{}

func (testNoopRuntimeConfig) BatchGetRuntimeConfig(_ context.Context, _ capmodel.CapabilityContext, keys []hostconfigcap.RuntimeConfigKey) (*capmodel.BatchResult[*hostconfigcap.RuntimeConfigProjection, hostconfigcap.RuntimeConfigKey], error) {
	return &capmodel.BatchResult[*hostconfigcap.RuntimeConfigProjection, hostconfigcap.RuntimeConfigKey]{
		Items:      map[hostconfigcap.RuntimeConfigKey]*hostconfigcap.RuntimeConfigProjection{},
		MissingIDs: append([]hostconfigcap.RuntimeConfigKey(nil), keys...),
	}, nil
}

func (testNoopRuntimeConfig) SetRuntimeConfigJSON(context.Context, capmodel.CapabilityContext, hostconfigcap.RuntimeConfigKey, []byte) error {
	return nil
}

type testNoopNotifications struct{}

func (testNoopNotifications) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[*capabilitynotifycap.MessageProjection, capabilitynotifycap.MessageID], error) {
	return &capmodel.BatchResult[*capabilitynotifycap.MessageProjection, capabilitynotifycap.MessageID]{
		Items:      map[capabilitynotifycap.MessageID]*capabilitynotifycap.MessageProjection{},
		MissingIDs: append([]capabilitynotifycap.MessageID(nil), ids...),
	}, nil
}

func (testNoopNotifications) BatchGetBySource(_ context.Context, _ capmodel.CapabilityContext, input capabilitynotifycap.BatchGetBySourceInput) (*capabilitynotifycap.BatchGetBySourceResult, error) {
	return &capabilitynotifycap.BatchGetBySourceResult{
		Items:      map[string][]*capabilitynotifycap.MessageProjection{},
		MissingIDs: append([]string(nil), input.SourceIDs...),
	}, nil
}

func (testNoopNotifications) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilitynotifycap.MessageID) error {
	return nil
}

func (testNoopNotifications) Send(context.Context, capmodel.CapabilityContext, capabilitynotifycap.SendInput) (*capabilitynotifycap.SendResult, error) {
	return &capabilitynotifycap.SendResult{}, nil
}

func (testNoopNotifications) Delete(context.Context, capmodel.CapabilityContext, []capabilitynotifycap.MessageID) error {
	return nil
}

func (testNoopNotifications) DeleteBySource(context.Context, capmodel.CapabilityContext, capabilitynotifycap.SourceType, []string) error {
	return nil
}

type testNoopPlugins struct{}

func (testNoopPlugins) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

func (testNoopPlugins) Current(context.Context, capmodel.CapabilityContext) (*capabilityplugincap.Projection, error) {
	return nil, nil
}

func (testNoopPlugins) Search(context.Context, capmodel.CapabilityContext, capabilityplugincap.SearchInput) (*capmodel.PageResult[*capabilityplugincap.Projection], error) {
	return &capmodel.PageResult[*capabilityplugincap.Projection]{Items: []*capabilityplugincap.Projection{}}, nil
}

func (testNoopPlugins) ListTenantPlugins(context.Context, capmodel.CapabilityContext, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

func (testNoopPlugins) BatchGetCapabilityStatus(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityplugincap.CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey], error) {
	result := &capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey]{
		Items:      make(map[capabilityplugincap.CapabilityKey]*capmodel.CapabilityStatus, len(keys)),
		MissingIDs: []capabilityplugincap.CapabilityKey{},
	}
	for _, key := range keys {
		result.Items[key] = &capmodel.CapabilityStatus{Available: false, Reason: "test_no_provider"}
	}
	return result, nil
}

func (testNoopPlugins) SetEnabled(context.Context, capmodel.CapabilityContext, capabilityplugincap.PluginID, bool) error {
	return nil
}

func (testNoopPlugins) ProvisionTenantDefaults(context.Context, capmodel.CapabilityContext, capmodel.DomainID) error {
	return nil
}

type testNoopSessions struct{}

func (testNoopSessions) Current(context.Context, capmodel.CapabilityContext) (*capabilitysessioncap.Projection, error) {
	return nil, nil
}

func (testNoopSessions) Search(context.Context, capmodel.CapabilityContext, capabilitysessioncap.SearchInput) (*capmodel.PageResult[*capabilitysessioncap.Projection], error) {
	return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: []*capabilitysessioncap.Projection{}}, nil
}

func (testNoopSessions) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilitysessioncap.SessionID) (*capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID]{
		Items:      map[capabilitysessioncap.SessionID]*capabilitysessioncap.Projection{},
		MissingIDs: append([]capabilitysessioncap.SessionID(nil), ids...),
	}, nil
}

func (testNoopSessions) BatchGetUserOnlineStatus(_ context.Context, _ capmodel.CapabilityContext, userIDs []string) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string]{
		Items:      map[string]*capabilitysessioncap.UserOnlineStatusProjection{},
		MissingIDs: append([]string(nil), userIDs...),
	}, nil
}

func (testNoopSessions) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilitysessioncap.SessionID) error {
	return nil
}

func (testNoopSessions) Revoke(context.Context, capmodel.CapabilityContext, capabilitysessioncap.SessionID) error {
	return nil
}

type testNoopJobs struct{}

func (testNoopJobs) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityjobcap.JobID) (*capmodel.BatchResult[*capabilityjobcap.Projection, capabilityjobcap.JobID], error) {
	return &capmodel.BatchResult[*capabilityjobcap.Projection, capabilityjobcap.JobID]{
		Items:      map[capabilityjobcap.JobID]*capabilityjobcap.Projection{},
		MissingIDs: append([]capabilityjobcap.JobID(nil), ids...),
	}, nil
}

func (testNoopJobs) Search(context.Context, capmodel.CapabilityContext, capabilityjobcap.SearchInput) (*capmodel.PageResult[*capabilityjobcap.Projection], error) {
	return &capmodel.PageResult[*capabilityjobcap.Projection]{Items: []*capabilityjobcap.Projection{}}, nil
}

func (testNoopJobs) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilityjobcap.JobID) error {
	return nil
}

func (testNoopJobs) Run(context.Context, capmodel.CapabilityContext, capabilityjobcap.JobID) error {
	return nil
}

func (testNoopJobs) SetStatus(context.Context, capmodel.CapabilityContext, capabilityjobcap.JobID, string) error {
	return nil
}

type testNoopInfra struct{}

func (testNoopInfra) BatchGetStatus(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityinfracap.ComponentID) (*capmodel.BatchResult[*capabilityinfracap.StatusProjection, capabilityinfracap.ComponentID], error) {
	return &capmodel.BatchResult[*capabilityinfracap.StatusProjection, capabilityinfracap.ComponentID]{
		Items:      map[capabilityinfracap.ComponentID]*capabilityinfracap.StatusProjection{},
		MissingIDs: append([]capabilityinfracap.ComponentID(nil), ids...),
	}, nil
}

func (testNoopInfra) RefreshStatus(context.Context, capmodel.CapabilityContext, capabilityinfracap.ComponentID) error {
	return nil
}
