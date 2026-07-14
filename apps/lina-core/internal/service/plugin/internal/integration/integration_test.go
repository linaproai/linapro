// This file verifies source-plugin callback flows receive plugin-scoped
// capability services through the capabilityowner scoped-service helper.

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
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// capabilityScopeRecorder records plugin-scope binding performed by
// capabilityowner.ServicesForPlugin.
type capabilityScopeRecorder struct {
	emptySourceServicesDirectory

	mu     sync.Mutex
	scopes []string
}

var _ capability.Services = (*capabilityScopeRecorder)(nil)
var _ capabilityowner.ScopedServicesFactory = (*capabilityScopeRecorder)(nil)

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

// scopeCount returns how many times pluginID was bound through ServicesForPlugin.
func (r *capabilityScopeRecorder) scopeCount(pluginID string) int {
	normalizedPluginID := strings.TrimSpace(pluginID)
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, scope := range r.scopes {
		if scope == normalizedPluginID {
			count++
		}
	}
	return count
}

// scopedSourceServicesDirectory is the plugin-bound Services view
// returned to source-plugin callback code.
type scopedSourceServicesDirectory struct {
	emptySourceServicesDirectory

	pluginID string
}

var _ capability.Services = (*scopedSourceServicesDirectory)(nil)

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

// Auth returns a no-op auth namespace required by tenant-core and external-login
// route registration (LDAP/OIDC plugins require ExternalLogin at register time).
func (d *scopedSourceServicesDirectory) Auth() authcap.Service {
	return authcap.New(scopedCapabilityAuth{}, scopedCapabilityAuthz{}, scopedCapabilityExternalLogin{})
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

// Dict returns an empty dictionary service required by monitor route registration.
func (d *scopedSourceServicesDirectory) Dict() capabilitydictcap.Service {
	return scopedCapabilityDict{}
}

// Notifications returns a no-op notification-domain service required by content-notice route registration.
func (d *scopedSourceServicesDirectory) Notifications() capabilitynotifycap.Service {
	return scopedNotificationsFixture{}
}

// Plugins returns an empty plugin-governance service required by tenant route registration.
func (d *scopedSourceServicesDirectory) Plugins() capabilityplugincap.Service {
	return scopedCapabilityPlugins{}
}

// I18n returns a fallback translator required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) I18n() i18ncap.Service {
	return scopedCapabilityI18n{}
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
	return scopedCapabilityHostConfig{}
}

// Storage returns a no-op storage service required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) Storage() storagecap.Service {
	return scopedCapabilityStorage{}
}

// Tenant returns tenant-domain fixtures required by source-plugin route registration.
func (d *scopedSourceServicesDirectory) Tenant() tenantcapsvc.Service {
	return scopedCapabilityTenantService{
		Service: tenantspi.New(nil, nil, nil, nil),
	}
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

// scopedCapabilityExternalLogin is a no-op external-login fixture for registration-only
// tests. OIDC/LDAP plugins only require a non-nil sub-capability while wiring routes.
type scopedCapabilityExternalLogin struct{}

// LoginByVerifiedIdentity returns an empty outcome because registration-only tests
// never complete an external-identity login exchange.
func (scopedCapabilityExternalLogin) LoginByVerifiedIdentity(
	context.Context,
	extlogin.LoginInput,
) (*extlogin.LoginOutput, error) {
	return &extlogin.LoginOutput{}, nil
}

// scopedCapabilityAuthz is an empty authorization fixture for registration-only tests.
type scopedCapabilityAuthz struct{}

// BatchGetPermissions returns label projections for non-empty permission keys.
func (scopedCapabilityAuthz) BatchGetPermissions(_ context.Context, keys []capabilityauthz.PermissionKey) (*capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey], error) {
	result := &capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey]{
		Items:      make(map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionInfo, len(keys)),
		MissingIDs: []capabilityauthz.PermissionKey{},
	}
	for _, key := range keys {
		if key == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &capabilityauthz.PermissionInfo{Key: key}
	}
	return result, nil
}

// BatchHasPermissions reports false because registration-only tests never authorize requests.
func (scopedCapabilityAuthz) BatchHasPermissions(_ context.Context, keys []capabilityauthz.PermissionKey) (map[capabilityauthz.PermissionKey]bool, error) {
	result := make(map[capabilityauthz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		result[key] = false
	}
	return result, nil
}

// HasPermission reports false because registration-only tests never authorize requests.
func (scopedCapabilityAuthz) HasPermission(context.Context, capabilityauthz.PermissionKey) (bool, error) {
	return false, nil
}

// IsPlatformAdmin reports false because registration-only tests never check admin status.
func (scopedCapabilityAuthz) IsPlatformAdmin(context.Context, capabilityauthz.UserID) (bool, error) {
	return false, nil
}

// ReplaceRolePermissions accepts role permission replacement without mutating state.
func (scopedCapabilityAuthz) ReplaceRolePermissions(context.Context, capabilityauthz.RoleID, []capabilityauthz.PermissionKey) error {
	return nil
}

// scopedCapabilityBizCtx is a minimal plugin-visible business-context fixture.
type scopedCapabilityBizCtx struct{}

// Current returns a platform-scoped context for registration-only tests.
func (scopedCapabilityBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{PlatformBypass: true}
}

// scopedCapabilityHostConfig is a defaulting host-config fixture for registration-only tests.
type scopedCapabilityHostConfig struct{}

// Get returns no configured value.
func (scopedCapabilityHostConfig) Get(context.Context, string, any) (*gvar.Var, error) {
	return nil, nil
}

// Exists reports that no config key exists.
func (scopedCapabilityHostConfig) Exists(context.Context, string) (bool, error) {
	return false, nil
}

// String returns the supplied default value.
func (scopedCapabilityHostConfig) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default value.
func (scopedCapabilityHostConfig) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default value.
func (scopedCapabilityHostConfig) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default value.
func (scopedCapabilityHostConfig) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// SysConfig returns registration-only sys_config methods.
func (scopedCapabilityHostConfig) SysConfig() hostconfigcap.SysConfigService {
	return scopedCapabilitySysConfig{}
}

// scopedCapabilitySysConfig is an empty sys_config fixture for registration-only tests.
type scopedCapabilitySysConfig struct{}

// Get returns no sys_config projection in registration-only tests.
func (s scopedCapabilitySysConfig) Get(ctx context.Context, key hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	result, err := s.BatchGet(ctx, []hostconfigcap.SysConfigKey{key})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[key], nil
}

// BatchGet returns all requested config keys as opaque missing entries.
func (scopedCapabilitySysConfig) BatchGet(_ context.Context, keys []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	return &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{
		Items:      map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{},
		MissingIDs: append([]hostconfigcap.SysConfigKey(nil), keys...),
	}, nil
}

// List returns an empty sys_config page.
func (scopedCapabilitySysConfig) List(context.Context, hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return &capmodel.PageResult[*hostconfigcap.SysConfigInfo]{Items: []*hostconfigcap.SysConfigInfo{}}, nil
}

// SetValue accepts sys_config writes without mutating state.
func (scopedCapabilitySysConfig) SetValue(context.Context, hostconfigcap.SysConfigKey, string, *hostconfigcap.SetSysConfigValueOptions) error {
	return nil
}

func (scopedCapabilitySysConfig) BatchSetValue(context.Context, []hostconfigcap.SetSysConfigValueItem, *hostconfigcap.SetSysConfigValueOptions) error {
	return nil
}

// Reset accepts sys_config reset without mutating state.
func (scopedCapabilitySysConfig) Reset(context.Context, hostconfigcap.SysConfigKey) error {
	return nil
}

// EnsureVisible accepts key checks in registration-only tests.
func (scopedCapabilitySysConfig) EnsureVisible(context.Context, []hostconfigcap.SysConfigKey) error {
	return nil
}

// scopedCapabilityDict is an empty dictionary fixture for registration-only tests.
type scopedCapabilityDict struct{}

// Type returns dictionary type subresource methods.
func (scopedCapabilityDict) Type() capabilitydictcap.TypeService {
	return scopedCapabilityDictType{}
}

// Value returns dictionary value subresource methods.
func (scopedCapabilityDict) Value() capabilitydictcap.ValueService {
	return scopedCapabilityDictValue{}
}

// Refresh accepts dictionary refresh without mutating state.
func (scopedCapabilityDict) Refresh(context.Context, capabilitydictcap.Type) error {
	return nil
}

// scopedCapabilityDictType is an empty dictionary-type fixture.
type scopedCapabilityDictType struct{}

// Get returns no dictionary type projection.
func (scopedCapabilityDictType) Get(context.Context, int) (*capabilitydictcap.TypeInfo, error) {
	return nil, nil
}

// BatchGet returns all dictionary type IDs as opaque missing entries.
func (scopedCapabilityDictType) BatchGet(_ context.Context, ids []int) (*capmodel.BatchResult[*capabilitydictcap.TypeInfo, int], error) {
	return &capmodel.BatchResult[*capabilitydictcap.TypeInfo, int]{
		Items:      map[int]*capabilitydictcap.TypeInfo{},
		MissingIDs: append([]int(nil), ids...),
	}, nil
}

// List returns an empty dictionary type page.
func (scopedCapabilityDictType) List(context.Context, capabilitydictcap.ListTypesInput) (*capmodel.PageResult[*capabilitydictcap.TypeInfo], error) {
	return &capmodel.PageResult[*capabilitydictcap.TypeInfo]{Items: []*capabilitydictcap.TypeInfo{}}, nil
}

// EnsureVisible accepts dictionary type visibility checks.
func (scopedCapabilityDictType) EnsureVisible(context.Context, []int) error {
	return nil
}

// EnsureKeysVisible accepts dictionary type key checks.
func (scopedCapabilityDictType) EnsureKeysVisible(context.Context, []capabilitydictcap.Type) error {
	return nil
}

// Create accepts dictionary type creation without mutating state.
func (scopedCapabilityDictType) Create(context.Context, capabilitydictcap.CreateTypeInput) (int, error) {
	return 0, nil
}

// Update accepts dictionary type updates without mutating state.
func (scopedCapabilityDictType) Update(context.Context, capabilitydictcap.UpdateTypeInput) error {
	return nil
}

// Delete accepts dictionary type deletes without mutating state.
func (scopedCapabilityDictType) Delete(context.Context, int) error {
	return nil
}

// scopedCapabilityDictValue is an empty dictionary-value fixture.
type scopedCapabilityDictValue struct{}

// Get returns no dictionary value projection.
func (scopedCapabilityDictValue) Get(context.Context, int) (*capabilitydictcap.ValueInfo, error) {
	return nil, nil
}

// BatchGet returns all dictionary values as opaque missing entries.
func (scopedCapabilityDictValue) BatchGet(_ context.Context, input capabilitydictcap.BatchGetValuesInput) (*capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value], error) {
	return &capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value]{
		Items:      map[capabilitydictcap.Value]*capabilitydictcap.ValueInfo{},
		MissingIDs: append([]capabilitydictcap.Value(nil), input.Values...),
	}, nil
}

// ResolveLabels returns deterministic label projections for requested values.
func (scopedCapabilityDictValue) ResolveLabels(_ context.Context, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.LabelInfo, len(input.Values)),
		MissingIDs: []capabilitydictcap.Value{},
	}
	for _, value := range input.Values {
		if value == "" {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		result.Items[value] = &capabilitydictcap.LabelInfo{
			Type:  input.Type,
			Value: value,
			Label: string(value),
		}
	}
	return result, nil
}

// ListValues returns an empty dictionary page for registration-only tests.
func (scopedCapabilityDictValue) List(context.Context, capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.ValueInfo], error) {
	return &capmodel.PageResult[*capabilitydictcap.ValueInfo]{Items: []*capabilitydictcap.ValueInfo{}}, nil
}

// EnsureVisible accepts dictionary value row checks.
func (scopedCapabilityDictValue) EnsureVisible(context.Context, []int) error {
	return nil
}

// EnsureValuesVisible accepts values in registration-only tests.
func (scopedCapabilityDictValue) EnsureValuesVisible(context.Context, capabilitydictcap.ResolveInput) error {
	return nil
}

// Create accepts dictionary value creation without mutating state.
func (scopedCapabilityDictValue) Create(context.Context, capabilitydictcap.CreateValueInput) (int, error) {
	return 0, nil
}

// Update accepts dictionary value updates without mutating state.
func (scopedCapabilityDictValue) Update(context.Context, capabilitydictcap.UpdateValueInput) error {
	return nil
}

// Delete accepts dictionary value deletes without mutating state.
func (scopedCapabilityDictValue) Delete(context.Context, int) error {
	return nil
}

// DeleteByType accepts dictionary value deletes by type without mutating state.
func (scopedCapabilityDictValue) DeleteByType(context.Context, capabilitydictcap.Type) error {
	return nil
}

// scopedNotificationsFixture is a no-op notification fixture for registration-only tests.
type scopedNotificationsFixture struct{}

// Get returns no notification message projection.
func (scopedNotificationsFixture) Get(context.Context, capabilitynotifycap.MessageID) (*capabilitynotifycap.MessageInfo, error) {
	return nil, nil
}

// BatchGet returns all requested messages as opaque missing entries.
func (scopedNotificationsFixture) BatchGet(_ context.Context, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[*capabilitynotifycap.MessageInfo, capabilitynotifycap.MessageID], error) {
	return &capmodel.BatchResult[*capabilitynotifycap.MessageInfo, capabilitynotifycap.MessageID]{
		Items:      map[capabilitynotifycap.MessageID]*capabilitynotifycap.MessageInfo{},
		MissingIDs: append([]capabilitynotifycap.MessageID(nil), ids...),
	}, nil
}

// List returns an empty notification page.
func (scopedNotificationsFixture) List(context.Context, capabilitynotifycap.ListInput) (*capmodel.PageResult[*capabilitynotifycap.MessageInfo], error) {
	return &capmodel.PageResult[*capabilitynotifycap.MessageInfo]{Items: []*capabilitynotifycap.MessageInfo{}}, nil
}

// BatchGetBySource returns all requested source IDs as opaque missing entries.
func (scopedNotificationsFixture) BatchGetBySource(_ context.Context, input capabilitynotifycap.BatchGetBySourceInput) (*capabilitynotifycap.BatchGetBySourceResult, error) {
	return &capabilitynotifycap.BatchGetBySourceResult{
		Items:      map[string][]*capabilitynotifycap.MessageInfo{},
		MissingIDs: append([]string(nil), input.SourceIDs...),
	}, nil
}

// EnsureVisible accepts all message IDs in registration-only tests.
func (scopedNotificationsFixture) EnsureVisible(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

// Send records no messages in registration-only tests.
func (scopedNotificationsFixture) Send(context.Context, capabilitynotifycap.SendInput) (*capabilitynotifycap.SendResult, error) {
	return &capabilitynotifycap.SendResult{}, nil
}

// Delete removes no messages in registration-only tests.
func (scopedNotificationsFixture) Delete(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

// DeleteBySource removes no messages in registration-only tests.
func (scopedNotificationsFixture) DeleteBySource(context.Context, usermsgv1.SourceType, []string) error {
	return nil
}

// MarkRead accepts read-state changes without mutating state.
func (scopedNotificationsFixture) MarkRead(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

// MarkUnread accepts read-state changes without mutating state.
func (scopedNotificationsFixture) MarkUnread(context.Context, []capabilitynotifycap.MessageID) error {
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

// scopedCapabilityRoute is a no-op route metadata fixture for registration-only tests.
type scopedCapabilityRoute struct{}

// GetMetadata returns no dynamic-route metadata.
func (scopedCapabilityRoute) GetMetadata(context.Context) *routecap.Metadata {
	return nil
}

// scopedCapabilitySession is an empty session fixture for registration-only tests.
type scopedCapabilitySession struct{}

// Current returns no session in registration-only tests.
func (scopedCapabilitySession) Current(context.Context) (*capabilitysessioncap.SessionInfo, error) {
	return nil, nil
}

// Get returns no session in registration-only tests.
func (scopedCapabilitySession) Get(context.Context, capabilitysessioncap.SessionID) (*capabilitysessioncap.SessionInfo, error) {
	return nil, nil
}

// List returns an empty session page.
func (scopedCapabilitySession) List(context.Context, capabilitysessioncap.ListInput) (*capmodel.PageResult[*capabilitysessioncap.SessionInfo], error) {
	return &capmodel.PageResult[*capabilitysessioncap.SessionInfo]{Items: []*capabilitysessioncap.SessionInfo{}, Total: 0}, nil
}

// BatchGet returns all requested sessions as opaque missing entries.
func (scopedCapabilitySession) BatchGet(_ context.Context, ids []capabilitysessioncap.SessionID) (*capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID]{
		Items:      map[capabilitysessioncap.SessionID]*capabilitysessioncap.SessionInfo{},
		MissingIDs: append([]capabilitysessioncap.SessionID(nil), ids...),
	}, nil
}

// BatchGetUserOnlineStatus returns all requested users as opaque missing entries.
func (scopedCapabilitySession) BatchGetUserOnlineStatus(_ context.Context, userIDs []string) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string]{
		Items:      map[string]*capabilitysessioncap.UserOnlineStatus{},
		MissingIDs: append([]string(nil), userIDs...),
	}, nil
}

// EnsureVisible accepts all session IDs in registration-only tests.
func (scopedCapabilitySession) EnsureVisible(context.Context, []capabilitysessioncap.SessionID) error {
	return nil
}

// Revoke records no revocation in registration-only tests.
func (scopedCapabilitySession) Revoke(context.Context, capabilitysessioncap.SessionID) error {
	return nil
}

// RevokeMany records no revocation in registration-only tests.
func (scopedCapabilitySession) RevokeMany(context.Context, []capabilitysessioncap.SessionID) error {
	return nil
}

// scopedCapabilityTenantFilter is a no-op tenant filter fixture for registration-only tests.
type scopedCapabilityTenantFilter struct{}

// scopedCapabilityTenantService attaches tenant sub-capabilities required by source plugins.
type scopedCapabilityTenantService struct {
	tenantcapsvc.Service
}

// Plugins returns no-op tenant plugin governance for registration-only tests.
func (scopedCapabilityTenantService) Plugins() tenantcapsvc.PluginService {
	return scopedCapabilityPlugins{}
}

// Filter returns the no-op tenant filter fixture for registration-only tests.
func (scopedCapabilityTenantService) Filter() tenantcapsvc.FilterService {
	return scopedCapabilityTenantFilter{}
}

// Context returns a platform-bypass tenant context for registration-only tests.
func (scopedCapabilityTenantFilter) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{PlatformBypass: true}
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

// Get returns no plugin projection for registration-only tests.
func (scopedCapabilityPlugins) Get(context.Context, capabilityplugincap.PluginID) (*capabilityplugincap.PluginInfo, error) {
	return nil, nil
}

// BatchGet returns all requested plugin IDs as opaque missing records.
func (scopedCapabilityPlugins) BatchGet(_ context.Context, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID]{
		Items:      map[capabilityplugincap.PluginID]*capabilityplugincap.PluginInfo{},
		MissingIDs: append([]capabilityplugincap.PluginID(nil), ids...),
	}, nil
}

// Current returns no current plugin projection for registration-only tests.
func (scopedCapabilityPlugins) Current(context.Context) (*capabilityplugincap.PluginInfo, error) {
	return nil, nil
}

// List returns an empty plugin-governance page for registration-only tests.
func (scopedCapabilityPlugins) List(context.Context, capabilityplugincap.ListInput) (*capmodel.PageResult[*capabilityplugincap.PluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.PluginInfo]{Items: []*capabilityplugincap.PluginInfo{}}, nil
}

// ListTenantPlugins returns an empty page for registration-only tests.
func (scopedCapabilityPlugins) ListTenantPlugins(context.Context, capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantPluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantPluginInfo]{Items: []*capabilityplugincap.TenantPluginInfo{}}, nil
}

// Config returns a blank plugin configuration reader for registration-only tests.
func (scopedCapabilityPlugins) Config() capabilityplugincap.ConfigService {
	return scopedCapabilityPluginConfig{}
}

// Registry returns the test registry projection service.
func (s scopedCapabilityPlugins) Registry() capabilityplugincap.RegistryService {
	return s
}

// State returns the test plugin enablement lookup projection.
func (s scopedCapabilityPlugins) State() capabilityplugincap.StateService {
	return s
}

// Lifecycle returns no-op lifecycle operations for registration-only tests.
func (scopedCapabilityPlugins) Lifecycle() capabilityplugincap.LifecycleService {
	return scopedCapabilityPluginLifecycle{}
}

// IsEnabled reports false for registration-only tests.
func (scopedCapabilityPlugins) IsEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// IsProviderEnabled reports false for registration-only tests.
func (scopedCapabilityPlugins) IsProviderEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// IsEnabledAuthoritative reports false for registration-only tests.
func (scopedCapabilityPlugins) IsEnabledAuthoritative(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return false, nil
}

// SetTenantPluginEnabled accepts enablement changes without mutating test state.
func (scopedCapabilityPlugins) SetTenantPluginEnabled(context.Context, capabilityplugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantPluginDefaults accepts tenant default provisioning without mutating test state.
func (scopedCapabilityPlugins) ProvisionTenantPluginDefaults(context.Context, capmodel.DomainID) error {
	return nil
}

// scopedCapabilityUsers is an empty user-domain fixture for registration-only tests.
type scopedCapabilityUsers struct{}

// Current returns no current user in registration-only tests.
func (scopedCapabilityUsers) Current(context.Context) (*capabilityusercap.UserInfo, error) {
	return nil, nil
}

// Get returns no user in registration-only tests.
func (scopedCapabilityUsers) Get(context.Context, capabilityusercap.UserID) (*capabilityusercap.UserInfo, error) {
	return nil, nil
}

// BatchGet returns all requested user IDs as opaque missing records.
func (scopedCapabilityUsers) BatchGet(_ context.Context, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserInfo{},
		MissingIDs: append([]capabilityusercap.UserID(nil), ids...),
	}, nil
}

// BatchResolve returns all requested identifiers as opaque missing records.
func (scopedCapabilityUsers) BatchResolve(_ context.Context, input capabilityusercap.BatchResolveInput) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey]{
		Items:      map[capabilityusercap.ResolveKey]*capabilityusercap.UserInfo{},
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

// List returns an empty page because registration-only tests never query users.
func (scopedCapabilityUsers) List(context.Context, capabilityusercap.ListInput) (*capmodel.PageResult[*capabilityusercap.UserInfo], error) {
	return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: []*capabilityusercap.UserInfo{}}, nil
}

// EnsureVisible accepts all users because registration-only tests never execute route handlers.
func (scopedCapabilityUsers) EnsureVisible(context.Context, []capabilityusercap.UserID) error {
	return nil
}

// Create accepts user creation without mutating state.
func (scopedCapabilityUsers) Create(context.Context, capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	return "", nil
}

func (scopedCapabilityUsers) CreateFromExternal(context.Context, capabilityusercap.CreateFromExternalInput) (capabilityusercap.UserID, error) {
	return "", nil
}

// Update accepts user updates without mutating state.
func (scopedCapabilityUsers) Update(context.Context, capabilityusercap.UpdateInput) error {
	return nil
}

// Delete accepts user deletion without mutating state.
func (scopedCapabilityUsers) Delete(context.Context, capabilityusercap.UserID) error {
	return nil
}

// SetStatus accepts user status changes without mutating state.
func (scopedCapabilityUsers) SetStatus(context.Context, capabilityusercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword accepts password reset without mutating state.
func (scopedCapabilityUsers) ResetPassword(context.Context, capabilityusercap.UserID, string) error {
	return nil
}

// Assignment returns role assignment operations.
func (scopedCapabilityUsers) Assignment() capabilityusercap.AssignmentService {
	return scopedCapabilityUserAssignments{}
}

// scopedCapabilityUserAssignments accepts role replacement without mutating state.
type scopedCapabilityUserAssignments struct{}

// ReplaceRoles accepts role replacement without mutating state.
func (scopedCapabilityUserAssignments) ReplaceRoles(context.Context, capabilityusercap.UserID, []int) error {
	return nil
}

// scopedCapabilityPluginConfig is a blank plugin configuration reader.
type scopedCapabilityPluginConfig struct{}

// Get returns no plugin config value.
func (scopedCapabilityPluginConfig) Get(_ context.Context, _ string, defaultValue any) (*gvar.Var, error) {
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}

// Exists reports that no plugin config key exists.
func (scopedCapabilityPluginConfig) Exists(context.Context, string) (bool, error) {
	return false, nil
}

// Scan leaves target unchanged because no plugin config is present.
func (scopedCapabilityPluginConfig) Scan(context.Context, string, any) error {
	return nil
}

// String returns the supplied default value.
func (scopedCapabilityPluginConfig) String(_ context.Context, _ string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// Bool returns the supplied default value.
func (scopedCapabilityPluginConfig) Bool(_ context.Context, _ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

// Int returns the supplied default value.
func (scopedCapabilityPluginConfig) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the supplied default value.
func (scopedCapabilityPluginConfig) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// emptySourceServicesDirectory is a minimal capability services test double.
type emptySourceServicesDirectory struct{}

var _ capability.Services = (*emptySourceServicesDirectory)(nil)

// APIDoc returns no API-doc service for this capability-scope test.
func (emptySourceServicesDirectory) APIDoc() apidoccap.Service { return nil }

// Auth returns no auth namespace for this capability-scope test.
func (emptySourceServicesDirectory) Auth() authcap.Service { return nil }

// Users returns no user-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Users() capabilityusercap.Service { return nil }

// BizCtx returns no business-context service for this capability-scope test.
func (emptySourceServicesDirectory) BizCtx() bizctxcap.Service { return nil }

// Cache returns no cache service for this capability-scope test.
func (emptySourceServicesDirectory) Cache() cachecap.Service { return nil }

// Dict returns no dictionary-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Dict() capabilitydictcap.Service { return nil }

// Files returns no file-domain service for this capability-scope test.
func (emptySourceServicesDirectory) Files() capabilityfilecap.Service { return nil }

// HostConfig returns no host-config service for this capability-scope test.
func (emptySourceServicesDirectory) HostConfig() hostconfigcap.Service { return nil }

// I18n returns no i18n service for this capability-scope test.
func (emptySourceServicesDirectory) I18n() i18ncap.Service { return nil }

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

// Route returns no route service for this capability-scope test.
func (emptySourceServicesDirectory) Route() routecap.Service { return nil }

// Sessions returns no online-session domain service for this capability-scope test.
func (emptySourceServicesDirectory) Sessions() capabilitysessioncap.Service { return nil }

// Storage returns no storage service for this capability-scope test.
func (emptySourceServicesDirectory) Storage() storagecap.Service { return nil }

// Tenant returns no tenant capability for this capability-scope test.
func (emptySourceServicesDirectory) Tenant() tenantcapsvc.Service { return nil }

// TestSourcePluginCallbacksUsePluginScopedServices verifies route, jobs, hook,
// and managed-job integration flows all bind runtime services through
// capabilityowner.ServicesForPlugin before exposing them to a source plugin.
func TestSourcePluginCallbacksUsePluginScopedServices(t *testing.T) {
	const pluginID = "plugin-dev-source-capability-scope"

	recorder := &capabilityScopeRecorder{}
	services := testutil.NewServicesWithCapabilities(recorder)

	observed := make(map[string]string)
	currentPhase := ""
	recordServices := func(label string, services capability.Services) error {
		if services == nil {
			return fmt.Errorf("%s services are nil", label)
		}
		if services.Tenant() == nil || services.Tenant().Filter() == nil {
			return fmt.Errorf("%s services did not expose source tenant filter", label)
		}
		if got := recorder.scopeCount(pluginID); got < len(observed)+1 {
			return fmt.Errorf("%s services scope count = %d, want at least %d", label, got, len(observed)+1)
		}
		observed[label] = pluginID
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
		t.Fatalf("expected capabilityowner.ServicesForPlugin to bind %q", pluginID)
	}
}
