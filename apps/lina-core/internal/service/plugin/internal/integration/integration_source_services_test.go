// This file verifies source-plugin callback flows receive plugin-scoped
// capability services through the public scoping mechanism.

package integration_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
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
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
)

// capabilityScopeRecorder records plugin-scope binding performed by
// capability.ServicesForPlugin.
type capabilityScopeRecorder struct {
	emptySourceServicesDirectory

	mu     sync.Mutex
	scopes []string
}

var _ capability.Services = (*capabilityScopeRecorder)(nil)
var _ capability.ScopedServicesFactory = (*capabilityScopeRecorder)(nil)

// ForPlugin returns a source-plugin capability view bound to pluginID.
func (r *capabilityScopeRecorder) ForPlugin(pluginID string) capability.Services {
	normalizedPluginID := strings.TrimSpace(pluginID)
	r.mu.Lock()
	r.scopes = append(r.scopes, normalizedPluginID)
	r.mu.Unlock()
	return &scopedSourceServicesDirectory{pluginID: normalizedPluginID}
}

// seenScope reports whether ServicesForPlugin bound pluginID at least once.
func (r *capabilityScopeRecorder) seenScope(pluginID string) bool {
	normalizedPluginID := strings.TrimSpace(pluginID)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, scope := range r.scopes {
		if scope == normalizedPluginID {
			return true
		}
	}
	return false
}

// scopedSourceServicesDirectory is the plugin-bound Services view
// returned to source-plugin callback code.
type scopedSourceServicesDirectory struct {
	emptySourceServicesDirectory

	pluginID string
}

var _ pluginhost.Services = (*scopedSourceServicesDirectory)(nil)

// scopedPluginID exposes the test-only scope marker for assertions.
func (d *scopedSourceServicesDirectory) scopedPluginID() string {
	if d == nil {
		return ""
	}
	return d.pluginID
}

// APIDoc returns a fallback API-doc service required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) APIDoc() apidoccap.Service {
	return scopedCapabilityAPIDoc{}
}

// Auth returns a no-op auth namespace required by tenant-core route registration.
func (d *scopedSourceServicesDirectory) Auth() authcap.Service {
	return authcap.New(scopedCapabilityAuth{}, scopedCapabilityAuthz{})
}

// BizCtx returns a minimal non-nil business context service required by source
// plugin route registration callbacks in this test.
func (d *scopedSourceServicesDirectory) BizCtx() bizctxcap.Service {
	return scopedCapabilityBizCtx{}
}

// Cache returns a no-op cache service required by Smart Center route registration.
func (d *scopedSourceServicesDirectory) Cache() cachecap.Service {
	return scopedCapabilityCache{}
}

// PluginConfig returns a defaulting plugin configuration service required by plugin job registration.
func (d *scopedSourceServicesDirectory) PluginConfig() plugincap.ConfigService {
	return scopedCapabilityConfig{}
}

// Dict returns an empty dictionary service required by monitor route registration.
func (d *scopedSourceServicesDirectory) Dict() capabilitydictcap.Service {
	return scopedCapabilityDict{}
}

// Notifications returns a no-op notification-domain service required by content-notice route registration.
func (d *scopedSourceServicesDirectory) Notifications() capabilitynotifycap.Service {
	return scopedNotificationsFixture{}
}

// Admin returns a minimal management directory required by tenant route registration.
func (d *scopedSourceServicesDirectory) Admin() capability.AdminServices {
	return scopedCapabilityAdminServices{}
}

// Plugins returns an empty plugin-governance service required by tenant route registration.
func (d *scopedSourceServicesDirectory) Plugins() capabilityplugincap.Service {
	return scopedCapabilityPlugins{}
}

// I18n returns a fallback translator required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) I18n() i18ncap.Service {
	return scopedCapabilityI18n{}
}

// PluginLifecycle returns no-op lifecycle operations required by tenant-core route registration.
func (d *scopedSourceServicesDirectory) PluginLifecycle() plugincap.LifecycleService {
	return scopedCapabilityPluginLifecycle{}
}

// PluginState returns a disabled-state reader required by global middleware registration.
func (d *scopedSourceServicesDirectory) PluginState() plugincap.StateService {
	return scopedCapabilityPluginState{}
}

// Route returns a no-op dynamic-route metadata reader required by audit middleware registration.
func (d *scopedSourceServicesDirectory) Route() routecap.Service {
	return scopedCapabilityRoute{}
}

// Sessions returns an empty session domain service required by monitor-online route registration.
func (d *scopedSourceServicesDirectory) Sessions() capabilitysessioncap.Service {
	return scopedCapabilitySession{}
}

// Users returns an empty user-domain service required by notice route registration.
func (d *scopedSourceServicesDirectory) Users() capabilityusercap.Service {
	return scopedCapabilityUsers{}
}

// HostConfig returns a defaulting host-config service required by cleanup job registration.
func (d *scopedSourceServicesDirectory) HostConfig() hostconfigcap.Service {
	return scopedCapabilityConfig{}
}

// Storage returns a no-op storage service required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) Storage() storagecap.Service {
	return scopedCapabilityStorage{}
}

// TenantFilter returns a no-op tenant filter service required by source-plugin
// registrations that construct tenant-aware services.
func (d *scopedSourceServicesDirectory) TenantFilter() tenantspi.PluginTableFilterService {
	return scopedCapabilityTenantFilter{}
}

// scopedCapabilityAPIDoc is a fallback API-doc fixture for registration-only tests.
type scopedCapabilityAPIDoc struct{}

// ResolveRouteText returns the supplied fallback route text.
func (scopedCapabilityAPIDoc) ResolveRouteText(_ context.Context, input apidoccap.RouteTextInput) apidoccap.RouteTextOutput {
	return apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
}

// ResolveRouteTexts returns fallback route text for each input.
func (scopedCapabilityAPIDoc) ResolveRouteTexts(_ context.Context, inputs []apidoccap.RouteTextInput) []apidoccap.RouteTextOutput {
	outputs := make([]apidoccap.RouteTextOutput, 0, len(inputs))
	for _, input := range inputs {
		outputs = append(outputs, apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
	}
	return outputs
}

// FindRouteTitleOperationKeys returns no matches in registration-only tests.
func (scopedCapabilityAPIDoc) FindRouteTitleOperationKeys(context.Context, string) []string {
	return nil
}

// scopedCapabilityAuth is a no-op auth fixture for registration-only tests.
type scopedCapabilityAuth struct{}

// SelectTenant returns an empty token output because registration-only tests never authenticate.
func (scopedCapabilityAuth) SelectTenant(context.Context, token.SelectTenantInput) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

// SwitchTenant returns an empty token output because registration-only tests never authenticate.
func (scopedCapabilityAuth) SwitchTenant(context.Context, token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

// IssueImpersonationToken returns an empty token output for registration-only tests.
func (scopedCapabilityAuth) IssueImpersonationToken(context.Context, token.ImpersonationTokenIssueInput) (*token.ImpersonationTokenOutput, error) {
	return &token.ImpersonationTokenOutput{}, nil
}

// RevokeImpersonationToken performs no revocation in registration-only tests.
func (scopedCapabilityAuth) RevokeImpersonationToken(context.Context, token.ImpersonationTokenRevokeInput) error {
	return nil
}

// scopedCapabilityAuthz is an empty authorization fixture for registration-only tests.
type scopedCapabilityAuthz struct{}

// BatchGetPermissions returns label projections for non-empty permission keys.
func (scopedCapabilityAuthz) BatchGetPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityauthz.PermissionKey) (*capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey], error) {
	result := &capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey]{
		Items:      make(map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionProjection, len(keys)),
		MissingIDs: []capabilityauthz.PermissionKey{},
	}
	for _, key := range keys {
		if key == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &capabilityauthz.PermissionProjection{Key: key}
	}
	return result, nil
}

// BatchHasPermissions reports false because registration-only tests never authorize requests.
func (scopedCapabilityAuthz) BatchHasPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityauthz.PermissionKey) (map[capabilityauthz.PermissionKey]bool, error) {
	result := make(map[capabilityauthz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		result[key] = false
	}
	return result, nil
}

// HasPermission reports false because registration-only tests never authorize requests.
func (scopedCapabilityAuthz) HasPermission(context.Context, capmodel.CapabilityContext, capabilityauthz.PermissionKey) (bool, error) {
	return false, nil
}

// IsPlatformAdmin reports false because registration-only tests never check admin status.
func (scopedCapabilityAuthz) IsPlatformAdmin(context.Context, capmodel.CapabilityContext, capabilityauthz.UserID) (bool, error) {
	return false, nil
}

// scopedCapabilityBizCtx is a minimal plugin-visible business-context fixture.
type scopedCapabilityBizCtx struct{}

// Current returns a platform-scoped context for registration-only tests.
func (scopedCapabilityBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{PlatformBypass: true}
}

// scopedCapabilityConfig is a defaulting config fixture for registration-only tests.
type scopedCapabilityConfig struct{}

// Get returns no configured value.
func (scopedCapabilityConfig) Get(context.Context, string) (*gvar.Var, error) {
	return nil, nil
}

// Exists reports that no config key exists.
func (scopedCapabilityConfig) Exists(context.Context, string) (bool, error) {
	return false, nil
}

// Scan leaves target unchanged because no test config is present.
func (scopedCapabilityConfig) Scan(context.Context, string, any) error {
	return nil
}

// String returns the supplied default value.
func (scopedCapabilityConfig) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default value.
func (scopedCapabilityConfig) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default value.
func (scopedCapabilityConfig) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default value.
func (scopedCapabilityConfig) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// BatchGetRuntimeConfig returns all requested config keys as opaque missing entries.
func (scopedCapabilityConfig) BatchGetRuntimeConfig(_ context.Context, _ capmodel.CapabilityContext, keys []hostconfigcap.RuntimeConfigKey) (*capmodel.BatchResult[*hostconfigcap.RuntimeConfigProjection, hostconfigcap.RuntimeConfigKey], error) {
	return &capmodel.BatchResult[*hostconfigcap.RuntimeConfigProjection, hostconfigcap.RuntimeConfigKey]{
		Items:      map[hostconfigcap.RuntimeConfigKey]*hostconfigcap.RuntimeConfigProjection{},
		MissingIDs: append([]hostconfigcap.RuntimeConfigKey(nil), keys...),
	}, nil
}

// scopedCapabilityDict is an empty dictionary fixture for registration-only tests.
type scopedCapabilityDict struct{}

// ResolveLabels returns deterministic label projections for requested values.
func (scopedCapabilityDict) ResolveLabels(_ context.Context, _ capmodel.CapabilityContext, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.LabelProjection, len(input.Values)),
		MissingIDs: []capabilitydictcap.Value{},
	}
	for _, value := range input.Values {
		if value == "" {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		result.Items[value] = &capabilitydictcap.LabelProjection{
			Type:  input.Type,
			Value: value,
			Label: string(value),
		}
	}
	return result, nil
}

// ListValues returns an empty dictionary page for registration-only tests.
func (scopedCapabilityDict) ListValues(context.Context, capmodel.CapabilityContext, capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.LabelProjection], error) {
	return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{Items: []*capabilitydictcap.LabelProjection{}}, nil
}

// EnsureValuesVisible accepts values in registration-only tests.
func (scopedCapabilityDict) EnsureValuesVisible(context.Context, capmodel.CapabilityContext, capabilitydictcap.ResolveInput) error {
	return nil
}

// scopedNotificationsFixture is a no-op notification fixture for registration-only tests.
type scopedNotificationsFixture struct{}

// BatchGet returns all requested messages as opaque missing entries.
func (scopedNotificationsFixture) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[*capabilitynotifycap.MessageProjection, capabilitynotifycap.MessageID], error) {
	return &capmodel.BatchResult[*capabilitynotifycap.MessageProjection, capabilitynotifycap.MessageID]{
		Items:      map[capabilitynotifycap.MessageID]*capabilitynotifycap.MessageProjection{},
		MissingIDs: append([]capabilitynotifycap.MessageID(nil), ids...),
	}, nil
}

// BatchGetBySource returns all requested source IDs as opaque missing entries.
func (scopedNotificationsFixture) BatchGetBySource(_ context.Context, _ capmodel.CapabilityContext, input capabilitynotifycap.BatchGetBySourceInput) (*capabilitynotifycap.BatchGetBySourceResult, error) {
	return &capabilitynotifycap.BatchGetBySourceResult{
		Items:      map[string][]*capabilitynotifycap.MessageProjection{},
		MissingIDs: append([]string(nil), input.SourceIDs...),
	}, nil
}

// EnsureVisible accepts all message IDs in registration-only tests.
func (scopedNotificationsFixture) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilitynotifycap.MessageID) error {
	return nil
}

// Send records no messages in registration-only tests.
func (scopedNotificationsFixture) Send(context.Context, capmodel.CapabilityContext, capabilitynotifycap.SendInput) (*capabilitynotifycap.SendResult, error) {
	return &capabilitynotifycap.SendResult{}, nil
}

// Delete removes no messages in registration-only tests.
func (scopedNotificationsFixture) Delete(context.Context, capmodel.CapabilityContext, []capabilitynotifycap.MessageID) error {
	return nil
}

// DeleteBySource removes no messages in registration-only tests.
func (scopedNotificationsFixture) DeleteBySource(context.Context, capmodel.CapabilityContext, capabilitynotifycap.SourceType, []string) error {
	return nil
}

// scopedCapabilityStorage is a no-op object storage fixture for registration-only tests.
type scopedCapabilityStorage struct{}

// Put drains the supplied body and returns metadata without persisting content.
func (scopedCapabilityStorage) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	if in.Body != nil {
		if _, err := io.Copy(io.Discard, in.Body); err != nil {
			return nil, err
		}
	}
	return &storagecap.PutOutput{
		Object: &storagecap.Object{
			Path:        in.Path,
			Size:        in.Size,
			ContentType: in.ContentType,
			Visibility:  storagecap.VisibilityPrivate,
		},
	}, nil
}

// Get reports a miss because registration-only tests do not persist objects.
func (scopedCapabilityStorage) Get(context.Context, storagecap.GetInput) (*storagecap.GetOutput, error) {
	return &storagecap.GetOutput{Found: false}, nil
}

// Delete accepts deletion without touching persistent state.
func (scopedCapabilityStorage) Delete(context.Context, storagecap.DeleteInput) error {
	return nil
}

// DeleteMany accepts batch deletion without touching persistent state.
func (scopedCapabilityStorage) DeleteMany(context.Context, storagecap.DeleteManyInput) error {
	return nil
}

// List returns an empty bounded object list.
func (scopedCapabilityStorage) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	return &storagecap.ListOutput{Objects: []*storagecap.Object{}, Limit: in.Limit}, nil
}

// ListCursor returns an empty bounded cursor object list.
func (scopedCapabilityStorage) ListCursor(_ context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	return &storagecap.ListCursorOutput{Objects: []*storagecap.Object{}, Limit: in.Limit}, nil
}

// Stat reports a miss because registration-only tests do not persist objects.
func (scopedCapabilityStorage) Stat(context.Context, storagecap.StatInput) (*storagecap.StatOutput, error) {
	return &storagecap.StatOutput{Found: false}, nil
}

// BatchStat reports all registration-only objects as missing.
func (scopedCapabilityStorage) BatchStat(_ context.Context, in storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	return &storagecap.BatchStatOutput{MissingPaths: append([]string(nil), in.Paths...)}, nil
}

// ProviderStatuses returns one available local provider for registration-only tests.
func (scopedCapabilityStorage) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return []*storagecap.ProviderStatus{{
		ProviderID: storagecap.LocalProviderID,
		Active:     true,
		Available:  true,
	}}, nil
}

// scopedCapabilityI18n is a fallback translator fixture for registration-only tests.
type scopedCapabilityI18n struct{}

// GetLocale returns the default locale used by registration-only tests.
func (scopedCapabilityI18n) GetLocale(context.Context) string {
	return "zh-CN"
}

// Translate returns fallback text because registration-only tests do not render messages.
func (scopedCapabilityI18n) Translate(_ context.Context, _ string, fallback string) string {
	return fallback
}

// FindMessageKeys returns no keys because registration-only tests do not search messages.
func (scopedCapabilityI18n) FindMessageKeys(context.Context, string, string) []string {
	return nil
}

// scopedCapabilityPluginLifecycle is a no-op lifecycle fixture for registration-only tests.
type scopedCapabilityPluginLifecycle struct{}

// EnsureTenantPluginDisableAllowed always allows tenant plugin disable in registration-only tests.
func (scopedCapabilityPluginLifecycle) EnsureTenantPluginDisableAllowed(context.Context, string, int) error {
	return nil
}

// NotifyTenantPluginDisabled records no notification in registration-only tests.
func (scopedCapabilityPluginLifecycle) NotifyTenantPluginDisabled(context.Context, string, int) {}

// EnsureTenantDeleteAllowed always allows tenant delete in registration-only tests.
func (scopedCapabilityPluginLifecycle) EnsureTenantDeleteAllowed(context.Context, int) error {
	return nil
}

// NotifyTenantDeleted records no notification in registration-only tests.
func (scopedCapabilityPluginLifecycle) NotifyTenantDeleted(context.Context, int) {}

// scopedCapabilityPluginState is a disabled-state fixture for registration-only tests.
type scopedCapabilityPluginState struct{}

// IsEnabled reports false because registration-only tests never execute plugin branches.
func (scopedCapabilityPluginState) IsEnabled(context.Context, string) bool {
	return false
}

// IsProviderEnabled reports false because registration-only tests never activate providers.
func (scopedCapabilityPluginState) IsProviderEnabled(context.Context, string) bool {
	return false
}

// IsEnabledAuthoritative reports false for registration-only global middleware fixtures.
func (scopedCapabilityPluginState) IsEnabledAuthoritative(context.Context, string) bool {
	return false
}

// scopedCapabilityRoute is a no-op route metadata fixture for registration-only tests.
type scopedCapabilityRoute struct{}

// DynamicRouteMetadata returns no dynamic-route metadata.
func (scopedCapabilityRoute) DynamicRouteMetadata(context.Context) *routecap.DynamicRouteMetadata {
	return nil
}

// scopedCapabilitySession is an empty session fixture for registration-only tests.
type scopedCapabilitySession struct{}

// Current returns no session in registration-only tests.
func (scopedCapabilitySession) Current(context.Context, capmodel.CapabilityContext) (*capabilitysessioncap.Projection, error) {
	return nil, nil
}

// Search returns an empty session page.
func (scopedCapabilitySession) Search(context.Context, capmodel.CapabilityContext, capabilitysessioncap.SearchInput) (*capmodel.PageResult[*capabilitysessioncap.Projection], error) {
	return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: []*capabilitysessioncap.Projection{}, Total: 0}, nil
}

// BatchGet returns all requested sessions as opaque missing entries.
func (scopedCapabilitySession) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilitysessioncap.SessionID) (*capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID]{
		Items:      map[capabilitysessioncap.SessionID]*capabilitysessioncap.Projection{},
		MissingIDs: append([]capabilitysessioncap.SessionID(nil), ids...),
	}, nil
}

// BatchGetUserOnlineStatus returns all requested users as opaque missing entries.
func (scopedCapabilitySession) BatchGetUserOnlineStatus(_ context.Context, _ capmodel.CapabilityContext, userIDs []string) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string]{
		Items:      map[string]*capabilitysessioncap.UserOnlineStatusProjection{},
		MissingIDs: append([]string(nil), userIDs...),
	}, nil
}

// EnsureVisible accepts all session IDs in registration-only tests.
func (scopedCapabilitySession) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilitysessioncap.SessionID) error {
	return nil
}

// Revoke records no revocation in registration-only tests.
func (scopedCapabilitySession) Revoke(context.Context, capmodel.CapabilityContext, capabilitysessioncap.SessionID) error {
	return nil
}

// scopedCapabilityTenantFilter is a no-op tenant filter fixture for registration-only tests.
type scopedCapabilityTenantFilter struct{}

// Context returns a platform-bypass tenant context for registration-only tests.
func (scopedCapabilityTenantFilter) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{PlatformBypass: true}
}

// Apply returns the model unchanged because registration-only tests never query plugin tables.
func (scopedCapabilityTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

// scopedCapabilityCache is a no-op cache fixture for registration-only tests.
type scopedCapabilityCache struct{}

// Get reports a cache miss because registration-only tests never persist cache data.
func (scopedCapabilityCache) Get(context.Context, string, string) (*cachecap.CacheItem, bool, error) {
	return nil, false, nil
}

// GetMany reports all cache keys as missing.
func (scopedCapabilityCache) GetMany(_ context.Context, in cachecap.GetManyInput) (*cachecap.GetManyOutput, error) {
	return &cachecap.GetManyOutput{
		Items:       map[string]*cachecap.CacheItem{},
		MissingKeys: append([]string(nil), in.Keys...),
	}, nil
}

// Set returns the stored projection without touching shared cache state.
func (scopedCapabilityCache) Set(_ context.Context, namespace string, key string, value string, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindString, Value: value}, nil
}

// SetMany returns stored projections without touching shared cache state.
func (scopedCapabilityCache) SetMany(_ context.Context, in cachecap.SetManyInput) (*cachecap.SetManyOutput, error) {
	output := &cachecap.SetManyOutput{Items: map[string]*cachecap.CacheItem{}}
	for _, item := range in.Items {
		output.Items[item.Key] = &cachecap.CacheItem{Key: in.Namespace + ":" + item.Key, ValueKind: cachecap.CacheValueKindString, Value: item.Value}
	}
	return output, nil
}

// Delete is a successful no-op for registration-only tests.
func (scopedCapabilityCache) Delete(context.Context, string, string) error {
	return nil
}

// DeleteMany is a successful no-op for registration-only tests.
func (scopedCapabilityCache) DeleteMany(context.Context, cachecap.DeleteManyInput) error {
	return nil
}

// Incr returns the requested delta as an isolated integer cache item.
func (scopedCapabilityCache) Incr(_ context.Context, namespace string, key string, delta int64, _ time.Duration) (*cachecap.CacheItem, error) {
	return &cachecap.CacheItem{Key: namespace + ":" + key, ValueKind: cachecap.CacheValueKindInt, IntValue: delta}, nil
}

// Expire reports that no cache item existed to expire.
func (scopedCapabilityCache) Expire(context.Context, string, string, time.Duration) (bool, *time.Time, error) {
	return false, nil, nil
}

// scopedCapabilityPlugins is an empty plugin-governance fixture for registration-only tests.
type scopedCapabilityPlugins struct{}

// BatchGet returns all requested plugin IDs as opaque missing records.
func (scopedCapabilityPlugins) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// Current returns no current plugin projection for registration-only tests.
func (scopedCapabilityPlugins) Current(context.Context, capmodel.CapabilityContext) (*capabilityplugincap.Projection, error) {
	return nil, nil
}

// Search returns an empty plugin-governance page for registration-only tests.
func (scopedCapabilityPlugins) Search(context.Context, capmodel.CapabilityContext, capabilityplugincap.SearchInput) (*capmodel.PageResult[*capabilityplugincap.Projection], error) {
	return &capmodel.PageResult[*capabilityplugincap.Projection]{Items: []*capabilityplugincap.Projection{}}, nil
}

// ListTenantPlugins returns an empty page for registration-only tests.
func (scopedCapabilityPlugins) ListTenantPlugins(context.Context, capmodel.CapabilityContext, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

// BatchGetCapabilityStatus returns all requested capability keys as unavailable.
func (scopedCapabilityPlugins) BatchGetCapabilityStatus(_ context.Context, _ capmodel.CapabilityContext, keys []capabilityplugincap.CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey], error) {
	result := &capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey]{
		Items:      make(map[capabilityplugincap.CapabilityKey]*capmodel.CapabilityStatus, len(keys)),
		MissingIDs: []capabilityplugincap.CapabilityKey{},
	}
	for _, key := range keys {
		result.Items[key] = &capmodel.CapabilityStatus{Available: false, Reason: "test_no_provider"}
	}
	return result, nil
}

// Config returns a blank plugin configuration reader for registration-only tests.
func (scopedCapabilityPlugins) Config() capabilityplugincap.ConfigService {
	return scopedCapabilityConfig{}
}

// State returns a nil-backed plugin state reader for registration-only tests.
func (scopedCapabilityPlugins) State() capabilityplugincap.StateService {
	return plugincap.NewState(nil)
}

// Lifecycle returns a nil-backed lifecycle service for registration-only tests.
func (scopedCapabilityPlugins) Lifecycle() capabilityplugincap.LifecycleService {
	return plugincap.NewLifecycle(nil)
}

// Registry returns the test registry projection service.
func (s scopedCapabilityPlugins) Registry() capabilityplugincap.RegistryService {
	return s
}

// scopedCapabilityPluginAdmin is a no-op plugin-governance admin fixture.
type scopedCapabilityPluginAdmin struct{}

// SetEnabled accepts enablement changes without mutating test state.
func (scopedCapabilityPluginAdmin) SetEnabled(context.Context, capmodel.CapabilityContext, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantDefaults accepts tenant default provisioning without mutating test state.
func (scopedCapabilityPluginAdmin) ProvisionTenantDefaults(context.Context, capmodel.CapabilityContext, capmodel.DomainID) error {
	return nil
}

// scopedCapabilityAdminServices exposes only the admin slices needed by registration tests.
type scopedCapabilityAdminServices struct{}

// Users returns no user management commands for registration-only tests.
func (scopedCapabilityAdminServices) Users() capabilityusercap.AdminService { return nil }

// Auth returns no authentication or authorization management commands for registration-only tests.
func (scopedCapabilityAdminServices) Auth() authcap.AdminService { return nil }

// Dict returns no dictionary management commands for registration-only tests.
func (scopedCapabilityAdminServices) Dict() capabilitydictcap.AdminService { return nil }

// Files returns no file management commands for registration-only tests.
func (scopedCapabilityAdminServices) Files() capabilityfilecap.AdminService { return nil }

// Sessions returns no-op session management commands for registration-only tests.
func (scopedCapabilityAdminServices) Sessions() capabilitysessioncap.AdminService {
	return scopedCapabilitySession{}
}

// HostConfig returns no runtime host-configuration management commands for registration-only tests.
func (scopedCapabilityAdminServices) HostConfig() hostconfigcap.AdminService { return nil }

// Notifications returns no-op notification management commands for registration-only tests.
func (scopedCapabilityAdminServices) Notifications() capabilitynotifycap.AdminService {
	return scopedNotificationsFixture{}
}

// Plugins returns no-op plugin management commands for tenant route construction.
func (scopedCapabilityAdminServices) Plugins() capabilityplugincap.AdminService {
	return scopedCapabilityPluginAdmin{}
}

// Jobs returns no scheduled-job management commands for registration-only tests.
func (scopedCapabilityAdminServices) Jobs() capabilityjobcap.AdminService { return nil }

// Infra returns no infrastructure management commands for registration-only tests.
func (scopedCapabilityAdminServices) Infra() capabilityinfracap.AdminService { return nil }

// scopedCapabilityUsers is an empty user-domain fixture for registration-only tests.
type scopedCapabilityUsers struct{}

// Current returns no current user in registration-only tests.
func (scopedCapabilityUsers) Current(context.Context, capmodel.CapabilityContext) (*capabilityusercap.UserProjection, error) {
	return nil, nil
}

// BatchGet returns all requested user IDs as opaque missing records.
func (scopedCapabilityUsers) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// BatchResolve returns all requested identifiers as opaque missing records.
func (scopedCapabilityUsers) BatchResolve(_ context.Context, _ capmodel.CapabilityContext, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey], error) {
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

// Search returns an empty page because registration-only tests never query users.
func (scopedCapabilityUsers) Search(context.Context, capmodel.CapabilityContext, capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
}

// EnsureVisible accepts all users because registration-only tests never execute route handlers.
func (scopedCapabilityUsers) EnsureVisible(context.Context, capmodel.CapabilityContext, []capabilityusercap.UserID) error {
	return nil
}

// emptySourceServicesDirectory is a minimal Services test double.
type emptySourceServicesDirectory struct{}

var _ pluginhost.Services = (*emptySourceServicesDirectory)(nil)

// APIDoc returns no API-doc service for this capability-scope test.
func (emptySourceServicesDirectory) APIDoc() apidoccap.Service { return nil }

// Auth returns no auth namespace for this capability-scope test.
func (emptySourceServicesDirectory) Auth() authcap.Service { return nil }

// Admin returns no management directory for this capability-scope test.
func (emptySourceServicesDirectory) Admin() capability.AdminServices { return nil }

// AI returns the default AI fallback namespace for this capability-scope test.
func (emptySourceServicesDirectory) AI() capabilityai.Service { return capabilityai.New(nil) }

// Users returns no user-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Users() capabilityusercap.Service { return nil }

// BizCtx returns no business-context service for this capability-scope test.
func (emptySourceServicesDirectory) BizCtx() bizctxcap.Service { return nil }

// Cache returns no cache service for this capability-scope test.
func (emptySourceServicesDirectory) Cache() cachecap.Service { return nil }

// PluginConfig returns no plugin config service for this capability-scope test.
func (emptySourceServicesDirectory) PluginConfig() plugincap.ConfigService { return nil }

// Dict returns no dictionary-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Dict() capabilitydictcap.Service { return nil }

// Files returns no file-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Files() capabilityfilecap.Service { return nil }

// HostConfig returns no host-config service for this capability-scope test.
func (emptySourceServicesDirectory) HostConfig() hostconfigcap.Service { return nil }

// I18n returns no i18n service for this capability-scope test.
func (emptySourceServicesDirectory) I18n() i18ncap.Service { return nil }

// Infra returns no infrastructure-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Infra() capabilityinfracap.Service { return nil }

// Jobs returns no scheduled-job domain service for this capability-scope test.
func (emptySourceServicesDirectory) Jobs() capabilityjobcap.Service { return nil }

// Lock returns no lock service for this capability-scope test.
func (emptySourceServicesDirectory) Lock() lockcap.Service { return nil }

// Manifest returns no manifest service for this capability-scope test.
func (emptySourceServicesDirectory) Manifest() manifestcap.Service { return nil }

// Notifications returns no notification-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Notifications() capabilitynotifycap.Service { return nil }

// Org returns no organization capability for this capability-scope test.
func (emptySourceServicesDirectory) Org() capabilityorgcap.Service { return nil }

// Plugins returns no plugin-governance domain service for this capability-scope test.
func (emptySourceServicesDirectory) Plugins() capabilityplugincap.Service { return nil }

// PluginLifecycle returns no plugin lifecycle service for this capability-scope test.
func (emptySourceServicesDirectory) PluginLifecycle() plugincap.LifecycleService { return nil }

// PluginState returns no plugin-state service for this capability-scope test.
func (emptySourceServicesDirectory) PluginState() plugincap.StateService { return nil }

// Route returns no route service for this capability-scope test.
func (emptySourceServicesDirectory) Route() routecap.Service { return nil }

// Sessions returns no online-session domain service for this capability-scope test.
func (emptySourceServicesDirectory) Sessions() capabilitysessioncap.Service { return nil }

// Storage returns no storage service for this capability-scope test.
func (emptySourceServicesDirectory) Storage() storagecap.Service { return nil }

// Tenant returns no tenant capability for this capability-scope test.
func (emptySourceServicesDirectory) Tenant() tenantcapsvc.Service { return nil }

// TenantFilter returns no tenant-filter service for this capability-scope test.
func (emptySourceServicesDirectory) TenantFilter() tenantspi.PluginTableFilterService { return nil }

// scopedCapabilityView is implemented by test doubles returned from ForPlugin.
type scopedCapabilityView interface {
	scopedPluginID() string
}

// TestSourcePluginCallbacksUsePluginScopedServices verifies route, jobs, hook,
// and managed-job integration flows all bind runtime services through
// capability.ServicesForPlugin before exposing them to a source plugin.
func TestSourcePluginCallbacksUsePluginScopedServices(t *testing.T) {
	const pluginID = "plugin-dev-source-capability-scope"

	recorder := &capabilityScopeRecorder{}
	services := testutil.NewServicesWithCapabilities(recorder)

	observed := make(map[string]string)
	currentPhase := ""
	recordServices := func(label string, services pluginhost.Services) error {
		if services == nil {
			return fmt.Errorf("%s services are nil", label)
		}
		scoped, ok := services.(scopedCapabilityView)
		if !ok {
			return fmt.Errorf("%s services were not plugin-scoped: %T", label, services)
		}
		if got := scoped.scopedPluginID(); got != pluginID {
			return fmt.Errorf("%s services scope = %q, want %q", label, got, pluginID)
		}
		observed[label] = scoped.scopedPluginID()
		return nil
	}

	sourcePlugin := pluginhost.NewDeclarations(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(
			"id: " + pluginID + "\n" +
				"name: Source Capability Scope Plugin\n" +
				"version: 0.1.0\n" +
				"type: source\n" +
				"scope_nature: tenant_aware\n" +
				"supports_multi_tenant: true\n" +
				"default_install_mode: tenant_scoped\n",
		)},
		"backend/plugin.go":             &fstest.MapFile{Data: []byte("package backend\n")},
		"frontend/pages/main-entry.vue": &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
	})

	if err := sourcePlugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		func(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
			return recordServices("route", registrar.Services())
		},
	); err != nil {
		t.Fatalf("failed to register source route handler: %v", err)
	}
	if err := sourcePlugin.Jobs().RegisterJobs(
		pluginhost.ExtensionPointJobsRegister,
		pluginhost.CallbackExecutionModeBlocking,
		func(_ context.Context, registrar pluginhost.JobsRegistrar) error {
			if currentPhase == "" {
				return fmt.Errorf("job registration phase was not set")
			}
			return recordServices(currentPhase, registrar.Services())
		},
	); err != nil {
		t.Fatalf("failed to register source job handler: %v", err)
	}
	if err := sourcePlugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.CallbackExecutionModeBlocking,
		func(_ context.Context, payload pluginhost.HookPayload) error {
			return recordServices("hook", payload.Services())
		},
	); err != nil {
		t.Fatalf("failed to register source hook handler: %v", err)
	}

	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("failed to register source plugin fixture: %v", err)
	}
	t.Cleanup(cleanup)

	ctx := context.Background()
	server := g.Server("integration-source-capability-scope")
	server.SetDumpRouterMap(false)
	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	noopMiddleware := func(req *ghttp.Request) {
		req.Middleware.Next()
	}
	middlewares := pluginhost.NewRouteMiddlewares(
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
		noopMiddleware,
	)

	if err = services.Integration.RegisterHTTPRoutes(ctx, server, rootGroup, middlewares); err != nil {
		t.Fatalf("expected route registration to receive scoped services, got error: %v", err)
	}

	currentPhase = "jobs"
	if err = services.Integration.RegisterJobs(ctx); err != nil {
		t.Fatalf("expected job registration to receive scoped services, got error: %v", err)
	}

	if err = services.Integration.DispatchPluginHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: pluginID,
			Name:     "Source Capability Scope Plugin",
			Version:  "0.1.0",
		}),
	); err != nil {
		t.Fatalf("expected hook dispatch to receive scoped services, got error: %v", err)
	}

	currentPhase = "managed-job"
	if _, err = services.Integration.ListJobDeclarationsByPlugin(ctx, pluginID); err != nil {
		t.Fatalf("expected managed job collection to receive scoped services, got error: %v", err)
	}

	for _, label := range []string{"route", "jobs", "hook", "managed-job"} {
		if got := observed[label]; got != pluginID {
			t.Fatalf("expected %s callback to receive plugin-scoped services for %q, got %q", label, pluginID, got)
		}
	}
	if !recorder.seenScope(pluginID) {
		t.Fatalf("expected capability.ServicesForPlugin to bind %q", pluginID)
	}
}
