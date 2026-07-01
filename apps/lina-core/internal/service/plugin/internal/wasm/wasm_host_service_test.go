// This file tests structured dynamic-plugin host-service dispatch through the
// shared dispatcher and unified capability services.

package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	capabilityauthz "lina-core/pkg/plugin/capability/authcap/authz"
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
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// capabilityHostServiceTestServices is a narrow capability service set used by
// org and tenant host-service tests.
type capabilityHostServiceTestServices struct {
	auth          authcap.Service
	cache         cachecap.Service
	org           orgcap.Service
	aiText        aitext.Service
	users         capabilityusercap.Service
	dict          capabilitydictcap.Service
	files         capabilityfilecap.Service
	lock          lockcap.Service
	notifications capabilitynotifycap.Service
	plugins       capabilityplugincap.Service
	sessions      capabilitysessioncap.Service
	storage       storagecap.Service
	tenant        tenantcap.Service
	scopeRecorder *capabilityHostServiceScopeRecorder
	scopedPlugin  string
}

type capabilityHostServiceScopeRecorder struct {
	mu         sync.Mutex
	lastPlugin string
}

func (r *capabilityHostServiceScopeRecorder) record(pluginID string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastPlugin = pluginID
}

func (r *capabilityHostServiceScopeRecorder) last() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastPlugin
}

// Ensure capabilityHostServiceTestServices implements the contracts needed by
// org and tenant host-service configuration.
var _ capability.Services = (*capabilityHostServiceTestServices)(nil)
var _ capabilityowner.ScopedServicesFactory = (*capabilityHostServiceTestServices)(nil)

// APIDoc returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) APIDoc() apidoccap.Service { return nil }

// Auth returns the configured auth capability namespace.
func (s *capabilityHostServiceTestServices) Auth() authcap.Service { return s.auth }

// AI returns the configured AI capability namespace.
func (s *capabilityHostServiceTestServices) AI() capabilityai.Service {
	text := s.aiText
	if recorder, ok := text.(*capabilityHostServiceAITextService); ok && s.scopedPlugin != "" {
		text = &capabilityHostServiceScopedAITextService{
			base:           recorder,
			sourcePluginID: s.scopedPlugin,
		}
	}
	return capabilityai.New(text)
}

// Users returns the configured user-domain capability service.
func (s *capabilityHostServiceTestServices) Users() capabilityusercap.Service { return s.users }

// BizCtx returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) BizCtx() bizctxcap.Service { return nil }

// Cache returns the configured cache-domain service.
func (s *capabilityHostServiceTestServices) Cache() cachecap.Service { return s.cache }

// Dict returns the configured dictionary-domain service.
func (s *capabilityHostServiceTestServices) Dict() capabilitydictcap.Service { return s.dict }

// Files returns the configured file-domain service.
func (s *capabilityHostServiceTestServices) Files() capabilityfilecap.Service { return s.files }

// HostConfig returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) HostConfig() hostconfigcap.Service { return nil }

// I18n returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) I18n() i18ncap.Service { return nil }

// Jobs returns no scheduled-job domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Jobs() capabilityjobcap.Service { return nil }

// Lock returns the configured lock-domain service.
func (s *capabilityHostServiceTestServices) Lock() lockcap.Service { return s.lock }

// Manifest returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) Manifest() manifestcap.Service { return nil }

// Notifications returns the configured notification-domain service.
func (s *capabilityHostServiceTestServices) Notifications() capabilitynotifycap.Service {
	return s.notifications
}

// Org returns the configured organization capability service.
func (s *capabilityHostServiceTestServices) Org() orgcap.Service { return s.org }

// Plugins returns the configured plugin-governance domain service.
func (s *capabilityHostServiceTestServices) Plugins() capabilityplugincap.Service { return s.plugins }

// Route returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) Route() routecap.Service { return nil }

// Sessions returns the configured online-session domain service.
func (s *capabilityHostServiceTestServices) Sessions() capabilitysessioncap.Service {
	return s.sessions
}

// Storage returns the configured storage-domain service.
func (s *capabilityHostServiceTestServices) Storage() storagecap.Service { return s.storage }

// Tenant returns the configured tenant capability service.
func (s *capabilityHostServiceTestServices) Tenant() tenantcap.Service { return s.tenant }

// ForPlugin records the requested plugin scope and returns the same directory.
func (s *capabilityHostServiceTestServices) ForPlugin(pluginID string) capability.Services {
	scoped := *s
	if scoped.scopeRecorder != nil {
		scoped.scopeRecorder.record(pluginID)
	}
	scoped.scopedPlugin = pluginID
	if cacheSvc, ok := scoped.cache.(*cacheDomainTestService); ok {
		copySvc := *cacheSvc
		copySvc.pluginID = pluginID
		scoped.cache = &copySvc
	}
	if lockSvc, ok := scoped.lock.(*lockDomainTestService); ok {
		copySvc := *lockSvc
		copySvc.pluginID = pluginID
		scoped.lock = &copySvc
	}
	return &scoped
}

// configureDomainHostServicesForCapabilityTest installs one shared domain
// services directory and restores the previous package state after the test.
func configureDomainHostServicesForCapabilityTest(t *testing.T, services capability.Services) {
	t.Helper()
	if services == nil {
		t.Fatal("configure domain host services failed: services is nil")
	}
	bindTestHostServiceRuntime(t, withTestDomainServices(services))
}

// TestHandleHostServiceInvokeOrgMethods verifies organization host-service
// calls are routed through capability.Services.Org.
func TestHandleHostServiceInvokeOrgMethods(t *testing.T) {
	providerPluginID := fmt.Sprintf("plugin-test-org-provider-%d", time.Now().UnixNano())
	orgManager := orgspi.NewManager()
	if err := orgManager.RegisterFactory(providerPluginID, func(context.Context, orgspi.ProviderEnv) (orgspi.Provider, error) {
		return capabilityHostServiceOrgProvider{}, nil
	}); err != nil {
		t.Fatalf("register org provider failed: %v", err)
	}

	services := &capabilityHostServiceTestServices{
		org:           orgspi.New(orgManager, capabilityHostServiceOrgRuntime{pluginID: providerPluginID}, nil),
		aiText:        aitext.New(nil, nil, nil),
		users:         &capabilityHostServiceUsersService{},
		tenant:        tenantspi.New(nil, nil, nil, nil),
		scopeRecorder: &capabilityHostServiceScopeRecorder{},
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	statusResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgStatus,
		nil,
	)
	if statusResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org status success, got status=%d payload=%s", statusResponse.Status, string(statusResponse.Payload))
	}
	var status capmodel.CapabilityStatus
	decodeCapabilityJSONResponse(t, statusResponse.Payload, &status)
	if !status.Available || status.ActiveProvider != providerPluginID {
		t.Fatalf("expected active org provider status, got %#v", status)
	}

	profilesResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgBatchGetUserOrgProfiles,
		marshalCapabilityJSONRequest(t, intUserIDsRequest{UserIDs: []int{7, 8}}),
	)
	if profilesResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org profile success, got status=%d payload=%s", profilesResponse.Status, string(profilesResponse.Payload))
	}
	var profiles capmodel.BatchResult[*orgcap.UserOrgProfile, int]
	decodeCapabilityJSONResponse(t, profilesResponse.Payload, &profiles)
	if profiles.Items[7] == nil || profiles.Items[7].DeptID != 17 || profiles.Items[8] == nil || profiles.Items[8].DeptName != "Dept-8" {
		t.Fatalf("unexpected profile payload: %#v", profiles.Items)
	}
	deptsResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgDepartmentBatchGet,
		marshalCapabilityJSONRequest(t, intDeptIDsRequest{DeptIDs: []int{17}}),
	)
	if deptsResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org department batch success, got status=%d payload=%s", deptsResponse.Status, string(deptsResponse.Payload))
	}
	var depts capmodel.BatchResult[*orgcap.DeptInfo, int]
	decodeCapabilityJSONResponse(t, deptsResponse.Payload, &depts)
	if depts.Items[17] == nil || depts.Items[17].DeptName != "Dept-17" {
		t.Fatalf("unexpected department batch payload: %#v", depts.Items)
	}
	postsResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgPostBatchGet,
		marshalCapabilityJSONRequest(t, intPostIDsRequest{PostIDs: []int{107}}),
	)
	if postsResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org post batch success, got status=%d payload=%s", postsResponse.Status, string(postsResponse.Payload))
	}
	var posts capmodel.BatchResult[*orgcap.PostInfo, int]
	decodeCapabilityJSONResponse(t, postsResponse.Payload, &posts)
	if posts.Items[107] == nil || posts.Items[107].PostName != "Post-107" {
		t.Fatalf("unexpected post batch payload: %#v", posts.Items)
	}
	if services.scopeRecorder.last() != "test-capability-plugin" {
		t.Fatalf("expected plugin-scoped services, got %q", services.scopeRecorder.last())
	}
}

// TestHandleHostServiceInvokeOrgRejectsInvisibleTargetUser verifies user-scoped
// organization host services cannot inspect users outside the actor data scope.
func TestHandleHostServiceInvokeOrgRejectsInvisibleTargetUser(t *testing.T) {
	providerPluginID := fmt.Sprintf("plugin-test-org-visible-%d", time.Now().UnixNano())
	orgManager := orgspi.NewManager()
	if err := orgManager.RegisterFactory(providerPluginID, func(context.Context, orgspi.ProviderEnv) (orgspi.Provider, error) {
		return capabilityHostServiceOrgProvider{}, nil
	}); err != nil {
		t.Fatalf("register org provider failed: %v", err)
	}
	userSvc := &capabilityHostServiceUsersService{ensureErr: errors.New("target user is outside data scope")}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(orgManager, capabilityHostServiceOrgRuntime{pluginID: providerPluginID}, nil),
		aiText: aitext.New(nil, nil, nil),
		users:  userSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgBatchGetUserOrgProfiles,
		marshalCapabilityJSONRequest(t, intUserIDsRequest{UserIDs: []int{42}}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected invisible org target user to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if !reflect.DeepEqual(userSvc.lastEnsureIDs, []capabilityusercap.UserID{"42"}) {
		t.Fatalf("expected org host service to check target user visibility, got %#v", userSvc.lastEnsureIDs)
	}
}

// TestHandleHostServiceInvokeUserMethods verifies user host-service calls are
// routed through capability.Services.Users with plugin-scoped context.
func TestHandleHostServiceInvokeUserMethods(t *testing.T) {
	userSvc := &capabilityHostServiceUsersService{
		users: map[capabilityusercap.UserID]*capabilityusercap.UserInfo{
			"12": {ID: "12", Username: "operator", Nickname: "Operator", Status: statusflag.EnabledValue},
			"42": {ID: "42", Username: "admin", Nickname: "Administrator", Status: statusflag.EnabledValue},
		},
	}
	services := &capabilityHostServiceTestServices{
		org:           orgspi.New(nil, nil, nil),
		aiText:        aitext.New(nil, nil, nil),
		users:         userSvc,
		tenant:        tenantspi.New(nil, nil, nil, nil),
		scopeRecorder: &capabilityHostServiceScopeRecorder{},
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	currentResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersCurrent,
		nil,
	)
	if currentResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user current success, got status=%d payload=%s", currentResponse.Status, string(currentResponse.Payload))
	}
	var current *capabilityusercap.UserInfo
	decodeCapabilityJSONResponse(t, currentResponse.Payload, &current)
	if current == nil || current.ID != "12" || userSvc.lastCurrent.UserID != 12 {
		t.Fatalf("unexpected current user payload current=%#v bizctx=%#v", current, userSvc.lastCurrent)
	}

	response := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersBatchGet,
		marshalCapabilityJSONRequest(t, struct {
			UserIDs []string `json:"userIds"`
		}{UserIDs: []string{"42", "99"}}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user batch success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var batch capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]
	decodeCapabilityJSONResponse(t, response.Payload, &batch)
	if batch.Items["42"] == nil || batch.Items["42"].Username != "admin" || !reflect.DeepEqual(batch.MissingIDs, []capabilityusercap.UserID{"99"}) {
		t.Fatalf("unexpected user batch payload: %#v", batch)
	}
	if services.scopeRecorder.last() != "test-user-plugin" || userSvc.lastCurrent.UserID != 12 {
		t.Fatalf("expected plugin-scoped user directory and current user context, lastPlugin=%q bizctx=%#v", services.scopeRecorder.last(), userSvc.lastCurrent)
	}

	resolveResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersBatchResolve,
		marshalCapabilityJSONRequest(t, struct {
			UserIDs   []string `json:"userIds"`
			Usernames []string `json:"usernames"`
			Contacts  []string `json:"contacts"`
		}{UserIDs: []string{"42"}, Usernames: []string{"admin"}, Contacts: []string{"missing@example.test"}}),
	)
	if resolveResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user resolve success, got status=%d payload=%s", resolveResponse.Status, string(resolveResponse.Payload))
	}
	var resolved capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey]
	decodeCapabilityJSONResponse(t, resolveResponse.Payload, &resolved)
	if resolved.Items["id:42"] == nil || resolved.Items["username:admin"] == nil || !reflect.DeepEqual(userSvc.lastResolve.Usernames, []string{"admin"}) {
		t.Fatalf("unexpected user resolve payload result=%#v input=%#v", resolved, userSvc.lastResolve)
	}

	listResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersList,
		marshalCapabilityJSONRequest(t, struct {
			Keyword  string `json:"keyword,omitempty"`
			PageNum  int    `json:"pageNum,omitempty"`
			PageSize int    `json:"pageSize,omitempty"`
		}{Keyword: "adm", PageNum: 1, PageSize: 10}),
	)
	if listResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user list success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	var page capmodel.PageResult[*capabilityusercap.UserInfo]
	decodeCapabilityJSONResponse(t, listResponse.Payload, &page)
	if page.Total != 2 || len(page.Items) != 2 || userSvc.lastList.Keyword != "adm" {
		t.Fatalf("unexpected user list payload page=%#v lastList=%#v", page, userSvc.lastList)
	}

	ensureResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersEnsureVisible,
		marshalCapabilityJSONRequest(t, struct {
			UserIDs []string `json:"userIds"`
		}{UserIDs: []string{"42"}}),
	)
	if ensureResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user ensure success, got status=%d payload=%s", ensureResponse.Status, string(ensureResponse.Payload))
	}
	if !reflect.DeepEqual(userSvc.lastEnsureIDs, []capabilityusercap.UserID{"42"}) {
		t.Fatalf("unexpected ensure user IDs: %#v", userSvc.lastEnsureIDs)
	}
}

// TestHandleHostServiceInvokeAdditionalDomainMethods verifies newly published
// resource-less domain services route through the shared capability.Services
// directory and pass standard business context to domain adapters.
func TestHandleHostServiceInvokeAdditionalDomainMethods(t *testing.T) {
	authzSvc := &capabilityHostServiceAuthzService{}
	dictSvc := &capabilityHostServiceDictService{}
	services := &capabilityHostServiceTestServices{
		auth:   authcap.New(nil, authzSvc),
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		dict:   dictSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	authzResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceAuthz,
		protocol.HostServiceMethodAuthzBatchGetPermissions,
		marshalCapabilityJSONRequest(t, map[string]any{"ids": []string{"system:user:list", "missing"}}),
	)
	if authzResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected authz batch success, got status=%d payload=%s", authzResponse.Status, string(authzResponse.Payload))
	}
	var permissions capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey]
	decodeCapabilityJSONResponse(t, authzResponse.Payload, &permissions)
	if permissions.Items["system:user:list"] == nil || permissions.Items["system:user:list"].LabelKey != "permissions.system:user:list" {
		t.Fatalf("unexpected authz payload: %#v", permissions)
	}
	if authzSvc.lastCurrent.UserID != 21 {
		t.Fatalf("expected authz business context, got %#v", authzSvc.lastCurrent)
	}

	authzHasResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceAuthz,
		protocol.HostServiceMethodAuthzBatchHasPermissions,
		marshalCapabilityJSONRequest(t, map[string]any{"ids": []string{"system:user:list", "system:user:delete"}}),
	)
	if authzHasResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected authz batch-has success, got status=%d payload=%s", authzHasResponse.Status, string(authzHasResponse.Payload))
	}
	var permissionChecks map[capabilityauthz.PermissionKey]bool
	decodeCapabilityJSONResponse(t, authzHasResponse.Payload, &permissionChecks)
	if !permissionChecks["system:user:list"] || permissionChecks["system:user:delete"] {
		t.Fatalf("unexpected authz batch-has payload: %#v", permissionChecks)
	}

	dictResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueResolveLabels,
		marshalCapabilityJSONRequest(t, map[string]any{
			"type":         "sys_common_status",
			"values":       []string{"enabled"},
			"includeLabel": true,
		}),
	)
	if dictResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict resolve success, got status=%d payload=%s", dictResponse.Status, string(dictResponse.Payload))
	}
	var labels capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value]
	decodeCapabilityJSONResponse(t, dictResponse.Payload, &labels)
	if labels.Items["enabled"] == nil || labels.Items["enabled"].Label != "Enabled" {
		t.Fatalf("unexpected dict payload: %#v", labels)
	}
	if dictSvc.lastCurrent.UserID != 21 || dictSvc.lastInput.Type != "sys_common_status" {
		t.Fatalf("expected dict business context and input, bizctx=%#v input=%#v", dictSvc.lastCurrent, dictSvc.lastInput)
	}

	dictEnsureResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueEnsureValuesVisible,
		marshalCapabilityJSONRequest(t, map[string]any{
			"type":   "sys_common_status",
			"values": []string{"enabled"},
		}),
	)
	if dictEnsureResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict ensure success, got status=%d payload=%s", dictEnsureResponse.Status, string(dictEnsureResponse.Payload))
	}
	if dictSvc.lastInput.Type != "sys_common_status" || !reflect.DeepEqual(dictSvc.lastInput.Values, []capabilitydictcap.Value{"enabled"}) {
		t.Fatalf("unexpected dict ensure input: %#v", dictSvc.lastInput)
	}

	dictListResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictListValues,
		marshalCapabilityJSONRequest(t, map[string]any{
			"type":         "sys_common_status",
			"includeLabel": true,
			"pageNum":      1,
			"pageSize":     10,
		}),
	)
	if dictListResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict list success, got status=%d payload=%s", dictListResponse.Status, string(dictListResponse.Payload))
	}
	var dictList capmodel.PageResult[*capabilitydictcap.LabelInfo]
	decodeCapabilityJSONResponse(t, dictListResponse.Payload, &dictList)
	if dictList.Total != 1 || dictSvc.lastListInput.Type != "sys_common_status" || dictSvc.lastListInput.Page.PageSize != 10 {
		t.Fatalf("unexpected dict list payload=%#v input=%#v", dictList, dictSvc.lastListInput)
	}
}

// TestHandleHostServiceInvokeDictEnsurePreservesBlankValues verifies dynamic
// dictionary visibility checks do not drop invalid blank values before the
// domain service can fail closed.
func TestHandleHostServiceInvokeDictEnsurePreservesBlankValues(t *testing.T) {
	dictSvc := &capabilityHostServiceDictService{denyBlankValues: true}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		dict:   dictSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueEnsureValuesVisible,
		marshalCapabilityJSONRequest(t, map[string]any{
			"type":   "sys_common_status",
			"values": []string{" "},
		}),
	)
	if response.Status == protocol.HostCallStatusSuccess {
		t.Fatalf("expected blank dictionary value to be rejected")
	}
	if !reflect.DeepEqual(dictSvc.lastInput.Values, []capabilitydictcap.Value{""}) {
		t.Fatalf("expected blank dictionary value to reach domain service, got %#v", dictSvc.lastInput.Values)
	}
}

// TestHandleHostServiceInvokeFilesWriteMethods verifies dynamic file writes
// dispatch to the scoped file capability with decoded payloads.
func TestHandleHostServiceInvokeFilesWriteMethods(t *testing.T) {
	filesSvc := &capabilityHostServiceFilesService{}
	services := &capabilityHostServiceTestServices{
		files:  filesSvc,
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	uploadResponse := invokeCapabilityHostService(
		t,
		filesHostCallContext(),
		protocol.HostServiceFiles,
		protocol.HostServiceMethodFilesUpload,
		marshalCapabilityJSONRequest(t, map[string]any{
			"filename":      "dynamic-upload.txt",
			"businessScene": "dynamic",
			"body":          []byte("dynamic content"),
			"sizeBytes":     int64(len("dynamic content")),
		}),
	)
	if uploadResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected files upload success, got status=%d payload=%s", uploadResponse.Status, string(uploadResponse.Payload))
	}
	var uploaded capabilityfilecap.FileInfo
	decodeCapabilityJSONResponse(t, uploadResponse.Payload, &uploaded)
	if uploaded.ID != "file-uploaded" || filesSvc.lastUpload.Filename != "dynamic-upload.txt" || filesSvc.lastUploadBody != "dynamic content" {
		t.Fatalf("unexpected files upload result=%#v input=%#v body=%q", uploaded, filesSvc.lastUpload, filesSvc.lastUploadBody)
	}
	if filesSvc.lastCurrent.UserID != 21 {
		t.Fatalf("expected files upload business context, got %#v", filesSvc.lastCurrent)
	}

	storageResponse := invokeCapabilityHostService(
		t,
		filesHostCallContext(),
		protocol.HostServiceFiles,
		protocol.HostServiceMethodFilesCreateFromStorage,
		marshalCapabilityJSONRequest(t, map[string]any{
			"storagePath":   "exports/source.txt",
			"filename":      "source.txt",
			"businessScene": "export",
			"sizeBytes":     int64(17),
		}),
	)
	if storageResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected files create-from-storage success, got status=%d payload=%s", storageResponse.Status, string(storageResponse.Payload))
	}
	var created capabilityfilecap.FileInfo
	decodeCapabilityJSONResponse(t, storageResponse.Payload, &created)
	if created.ID != "file-from-storage" || filesSvc.lastStorageInput.StoragePath != "exports/source.txt" || filesSvc.lastStorageInput.Filename != "source.txt" {
		t.Fatalf("unexpected files create-from-storage result=%#v input=%#v", created, filesSvc.lastStorageInput)
	}
}

// TestHandleHostServiceInvokeFilesCreateFromStorageRejectsUnauthorizedStorage
// verifies storage promotion cannot bypass dynamic storage path authorization.
func TestHandleHostServiceInvokeFilesCreateFromStorageRejectsUnauthorizedStorage(t *testing.T) {
	filesSvc := &capabilityHostServiceFilesService{}
	services := &capabilityHostServiceTestServices{
		files:  filesSvc,
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)
	hcc := filesHostCallContext()
	hcc.hostServices[1].Paths = []string{"reports/"}

	response := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceFiles,
		protocol.HostServiceMethodFilesCreateFromStorage,
		marshalCapabilityJSONRequest(t, map[string]any{
			"storagePath":   "exports/source.txt",
			"filename":      "source.txt",
			"businessScene": "export",
		}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected unauthorized storage path to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if filesSvc.lastStorageInput.StoragePath != "" {
		t.Fatalf("expected file service not to be called, got %#v", filesSvc.lastStorageInput)
	}
}

// TestHandleHostServiceInvokeSessionMethods verifies session host-service calls
// are routed through the shared online-session capability service.
func TestHandleHostServiceInvokeSessionMethods(t *testing.T) {
	sessionSvc := &capabilityHostServiceSessionsService{
		sessions: map[capabilitysessioncap.SessionID]*capabilitysessioncap.SessionInfo{
			"token-1": {ID: "token-1", UserID: "12", Username: "operator"},
		},
	}
	services := &capabilityHostServiceTestServices{
		org:      orgspi.New(nil, nil, nil),
		aiText:   aitext.New(nil, nil, nil),
		sessions: sessionSvc,
		tenant:   tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	hcc := sessionHostCallContext()
	currentResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsCurrent,
		nil,
	)
	if currentResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session current success, got status=%d payload=%s", currentResponse.Status, string(currentResponse.Payload))
	}
	var current *capabilitysessioncap.SessionInfo
	decodeCapabilityJSONResponse(t, currentResponse.Payload, &current)
	if current == nil || current.ID != "token-1" || sessionSvc.lastCurrent.TokenID != "token-1" {
		t.Fatalf("unexpected session current payload current=%#v bizctx=%#v", current, sessionSvc.lastCurrent)
	}

	listResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsList,
		marshalCapabilityJSONRequest(t, map[string]any{"username": "operator", "pageNum": 1, "pageSize": 10}),
	)
	if listResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session list success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	if sessionSvc.lastList.Username != "operator" {
		t.Fatalf("unexpected session list input: %#v", sessionSvc.lastList)
	}

	batchResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsBatchGet,
		marshalCapabilityJSONRequest(t, map[string]any{"ids": []string{"token-1", "missing"}}),
	)
	if batchResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session batch success, got status=%d payload=%s", batchResponse.Status, string(batchResponse.Payload))
	}
	var batch capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID]
	decodeCapabilityJSONResponse(t, batchResponse.Payload, &batch)
	if batch.Items["token-1"] == nil || !reflect.DeepEqual(batch.MissingIDs, []capabilitysessioncap.SessionID{"missing"}) {
		t.Fatalf("unexpected session batch payload: %#v", batch)
	}

	onlineResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsBatchGetUserOnlineStatus,
		marshalCapabilityJSONRequest(t, map[string]any{"userIds": []string{"12", "99"}}),
	)
	if onlineResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session online status success, got status=%d payload=%s", onlineResponse.Status, string(onlineResponse.Payload))
	}
	var online capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string]
	decodeCapabilityJSONResponse(t, onlineResponse.Payload, &online)
	if online.Items["12"] == nil || !online.Items["12"].Online || !reflect.DeepEqual(sessionSvc.lastOnlineUserIDs, []string{"12", "99"}) {
		t.Fatalf("unexpected online status payload=%#v last=%#v", online, sessionSvc.lastOnlineUserIDs)
	}

	ensureResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsEnsureVisible,
		marshalCapabilityJSONRequest(t, map[string]any{"ids": []string{"token-1"}}),
	)
	if ensureResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session ensure success, got status=%d payload=%s", ensureResponse.Status, string(ensureResponse.Payload))
	}
	if !reflect.DeepEqual(sessionSvc.lastEnsureIDs, []capabilitysessioncap.SessionID{"token-1"}) {
		t.Fatalf("unexpected session ensure IDs: %#v", sessionSvc.lastEnsureIDs)
	}
}

// TestHandleHostServiceInvokeTenantMethods verifies tenant host-service calls
// are routed through capability.Services.Tenant.
func TestHandleHostServiceInvokeTenantMethods(t *testing.T) {
	tenantSvc := &capabilityHostServiceTenantService{
		tenants: []tenantcap.TenantInfo{
			{ID: 3, Code: "tenant-a", Name: "Tenant A", Status: "active"},
		},
	}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		users:  &capabilityHostServiceUsersService{},
		tenant: tenantSvc,
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantListUserTenants,
		marshalCapabilityJSONRequest(t, intUserIDRequest{UserID: 42}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant list success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var tenants []tenantcap.TenantInfo
	decodeCapabilityJSONResponse(t, response.Payload, &tenants)
	if !reflect.DeepEqual(tenants, tenantSvc.tenants) || tenantSvc.lastUserID != 42 {
		t.Fatalf("unexpected tenant payload tenants=%#v lastUserID=%d", tenants, tenantSvc.lastUserID)
	}

	currentResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantCurrent,
		nil,
	)
	if currentResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant current success, got status=%d payload=%s", currentResponse.Status, string(currentResponse.Payload))
	}
	var current tenantcap.TenantID
	decodeCapabilityJSONResponse(t, currentResponse.Payload, &current)
	if current != 3 {
		t.Fatalf("unexpected current tenant: %d", current)
	}
}

// TestHandleHostServiceInvokeTenantGovernanceMethods verifies dynamic tenant
// governance methods dispatch through Tenant().Plugins().
func TestHandleHostServiceInvokeTenantGovernanceMethods(t *testing.T) {
	tenantPluginSvc := &trackingTenantPluginService{}
	tenantSvc := &capabilityHostServiceTenantService{plugins: tenantPluginSvc}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		users:  &capabilityHostServiceUsersService{},
		tenant: tenantSvc,
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	hcc := tenantGovernanceHostCallContext()
	setResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantPluginSetEnabled,
		marshalCapabilityJSONRequest(t, tenantPluginSetEnabledRequest{PluginID: "linapro-demo-dynamic", Enabled: true}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant plugin set success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}
	if tenantPluginSvc.lastPluginID != "linapro-demo-dynamic" || !tenantPluginSvc.lastEnabled {
		t.Fatalf("unexpected tenant plugin set call: plugin=%q enabled=%v", tenantPluginSvc.lastPluginID, tenantPluginSvc.lastEnabled)
	}

	provisionResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantPluginProvisionDefaults,
		marshalCapabilityJSONRequest(t, tenantPluginProvisionDefaultsRequest{TenantID: "42"}),
	)
	if provisionResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant plugin default provisioning success, got status=%d payload=%s", provisionResponse.Status, string(provisionResponse.Payload))
	}
	if tenantPluginSvc.lastProvisionTenantID != "42" {
		t.Fatalf("unexpected tenant default provisioning tenant: %q", tenantPluginSvc.lastProvisionTenantID)
	}
}

// TestHandleHostServiceInvokeTenantFilterContext verifies dynamic tenant filter
// exposure is limited to the serializable Tenant().Filter().Context() method.
func TestHandleHostServiceInvokeTenantFilterContext(t *testing.T) {
	filterSvc := &trackingTenantFilterService{contextValue: tenantspi.TenantFilterContext{
		UserID:         21,
		TenantID:       9,
		ActingUserID:   20,
		PlatformBypass: true,
	}}
	tenantSvc := &capabilityHostServiceTenantService{filter: filterSvc}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		users:  &capabilityHostServiceUsersService{},
		tenant: tenantSvc,
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		tenantGovernanceHostCallContext(),
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantFilterContext,
		nil,
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant filter context success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var result tenantspi.TenantFilterContext
	decodeCapabilityJSONResponse(t, response.Payload, &result)
	if result != filterSvc.contextValue || filterSvc.contextCalls != 1 {
		t.Fatalf("unexpected tenant filter context result=%#v calls=%d", result, filterSvc.contextCalls)
	}
}

// TestHandleHostServiceInvokeTenantRejectsInvisibleTargetUser verifies tenant
// host services do not reveal memberships for users outside actor data scope.
func TestHandleHostServiceInvokeTenantRejectsInvisibleTargetUser(t *testing.T) {
	userSvc := &capabilityHostServiceUsersService{ensureErr: errors.New("target user is outside data scope")}
	tenantSvc := &capabilityHostServiceTenantService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		users:  userSvc,
		tenant: tenantSvc,
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantListUserTenants,
		marshalCapabilityJSONRequest(t, intUserIDRequest{UserID: 42}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected invisible tenant target user to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if !reflect.DeepEqual(userSvc.lastEnsureIDs, []capabilityusercap.UserID{"42"}) {
		t.Fatalf("expected tenant host service to check target user visibility, got %#v", userSvc.lastEnsureIDs)
	}
	if tenantSvc.lastUserID != 0 {
		t.Fatalf("expected tenant service not to be called after visibility denial, got user=%d", tenantSvc.lastUserID)
	}
}

// TestHandleHostServiceInvokeAITextGenerate verifies AI host-service calls are
// method-authorized and routed through capability.Services.AIText.
func TestHandleHostServiceInvokeAITextGenerate(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{
		response: &aitext.GenerateResponse{
			Text:         "summary",
			Tier:         aitext.TierBasic,
			ProviderName: "Fake AI",
			ModelName:    "fake-text",
			Protocol:     "test",
			Usage:        aitext.Usage{InputTokens: 3, OutputTokens: 4},
			GeneratedAt:  time.Now().UnixMilli(),
		},
	}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		protocol.MarshalHostServiceAITextGenerateRequest(
			&protocol.HostServiceAITextGenerateRequest{
				Purpose:         "content.summary",
				Tier:            aitext.TierBasic,
				MaxOutputTokens: 512,
				Messages: []aitext.Message{
					{Role: aitext.MessageRoleUser, Content: "sensitive prompt must not be logged"},
				},
			},
		),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected ai text success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var out aitext.GenerateResponse
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if out.Text != "summary" {
		t.Fatalf("unexpected ai text response: %#v", out)
	}
	if aiSvc.lastSourcePluginID != "test-ai-plugin" || aiSvc.lastRequest.Purpose != "content.summary" {
		t.Fatalf("unexpected routed ai request: source=%s request=%#v", aiSvc.lastSourcePluginID, aiSvc.lastRequest)
	}
	if aiSvc.lastRequest.MaxOutputTokens != 512 {
		t.Fatalf("expected dto maxOutputTokens to pass through, got %d", aiSvc.lastRequest.MaxOutputTokens)
	}
}

// TestHandleHostServiceInvokeAITextRoutesPurposeFromDTO verifies purpose stays
// a request DTO field instead of a host-service resource authorization key.
func TestHandleHostServiceInvokeAITextRoutesPurposeFromDTO(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		protocol.MarshalHostServiceAITextGenerateRequest(
			&protocol.HostServiceAITextGenerateRequest{
				Purpose:         "code.review",
				Tier:            aitext.TierBasic,
				MaxOutputTokens: 128,
				Messages: []aitext.Message{
					{Role: aitext.MessageRoleUser, Content: "hello"},
				},
			},
		),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected ai text success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if !aiSvc.called || aiSvc.lastRequest.Purpose != "code.review" {
		t.Fatalf("expected purpose to reach AI service from DTO, called=%v request=%#v", aiSvc.called, aiSvc.lastRequest)
	}
}

// TestHandleHostServiceInvokeAITextDoesNotEnforceOutputLimit verifies bridge
// dispatch no longer enforces resource maxOutputTokens policy.
func TestHandleHostServiceInvokeAITextDoesNotEnforceOutputLimit(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		protocol.MarshalHostServiceAITextGenerateRequest(
			&protocol.HostServiceAITextGenerateRequest{
				Purpose:         "content.summary",
				Tier:            aitext.TierBasic,
				MaxOutputTokens: 128,
				Messages: []aitext.Message{
					{Role: aitext.MessageRoleUser, Content: "hello"},
				},
			},
		),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected ai text success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if !aiSvc.called || aiSvc.lastRequest.MaxOutputTokens != 128 {
		t.Fatalf("expected maxOutputTokens to pass through, called=%v request=%#v", aiSvc.called, aiSvc.lastRequest)
	}
}

// TestHandleHostServiceInvokeAITextDoesNotApplyDefaultOutputLimit verifies the
// bridge does not derive default output limits from host-service resources.
func TestHandleHostServiceInvokeAITextDoesNotApplyDefaultOutputLimit(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		protocol.MarshalHostServiceAITextGenerateRequest(
			&protocol.HostServiceAITextGenerateRequest{
				Purpose: "content.summary",
				Tier:    aitext.TierBasic,
				Messages: []aitext.Message{
					{Role: aitext.MessageRoleUser, Content: "hello"},
				},
			},
		),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected ai text success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if aiSvc.lastRequest.MaxOutputTokens != 0 {
		t.Fatalf("expected omitted maxOutputTokens to stay unset, got %d", aiSvc.lastRequest.MaxOutputTokens)
	}
}

// TestHandleHostServiceInvokeAITextMethodStatus verifies text method status
// is dynamically dispatched without exposing provider configuration.
func TestHandleHostServiceInvokeAITextMethodStatus(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextMethodStatus,
		marshalCapabilityJSONRequest(t, capabilityai.MethodStatusQuery{
			CapabilityType:   capabilityai.CapabilityTypeText,
			CapabilityMethod: capabilityai.CapabilityMethod(aicommon.CapabilityMethodTextGenerate),
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected ai text method status success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var out aicommon.MethodStatus
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if !out.Available || out.CapabilityType != aicommon.CapabilityTypeText || out.CapabilityMethod != aicommon.CapabilityMethodTextGenerate {
		t.Fatalf("unexpected ai text method status: %#v", out)
	}
	if out.CapabilityStatus.ActiveProvider != aitext.ProviderPluginID {
		t.Fatalf("expected public active provider projection, got %#v", out.CapabilityStatus)
	}
}

// TestHandleHostServiceInvokeAIMethodStatuses verifies cross-sub-capability
// method statuses are dynamically dispatched through the AI namespace.
func TestHandleHostServiceInvokeAIMethodStatuses(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAIMethodStatuses,
		marshalCapabilityJSONRequest(t, capabilityai.MethodStatusesInput{
			Methods: []capabilityai.MethodStatusQuery{
				{
					CapabilityType:   capabilityai.CapabilityTypeText,
					CapabilityMethod: capabilityai.CapabilityMethod(aicommon.CapabilityMethodTextGenerate),
				},
				{
					CapabilityType:   capabilityai.CapabilityTypeImage,
					CapabilityMethod: capabilityai.CapabilityMethod(aicommon.CapabilityMethodImageGenerate),
				},
			},
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected ai method status batch success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var out capabilityai.MethodStatusesResult
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if len(out.Items) != 2 {
		t.Fatalf("expected two ai method statuses, got %#v", out)
	}
	if !out.Items[0].Available || out.Items[0].CapabilityType != aicommon.CapabilityTypeText {
		t.Fatalf("unexpected text status: %#v", out.Items[0])
	}
	if out.Items[1].Available || out.Items[1].CapabilityType != aicommon.CapabilityTypeImage {
		t.Fatalf("expected unavailable fallback image status, got %#v", out.Items[1])
	}
}

// TestHandleHostServiceInvokeAIRejectsUnauthorizedMethod verifies host-service
// method authorization rejects undeclared AI methods before dispatch.
func TestHandleHostServiceInvokeAIRejectsUnauthorizedMethod(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	hcc := aiTextHostCallContext()
	hcc.capabilities[protocol.CapabilityAIDocument] = struct{}{}
	hcc.hostServices = []*protocol.HostServiceSpec{
		{
			Service: protocol.HostServiceAI,
			Methods: []string{
				protocol.HostServiceMethodAITextGenerate,
			},
		},
	}
	response := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceAI,
		protocol.HostServiceMethodAIDocumentCite,
		nil,
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if aiSvc.called {
		t.Fatal("expected unauthorized method to be rejected before AI service call")
	}
}

// TestHandleHostServiceInvokeAITextRedactsProviderErrors verifies provider
// failures do not leak authorization markers through the host-call response.
func TestHandleHostServiceInvokeAITextRedactsProviderErrors(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{
		err: errors.New("provider failed authorization bearer sk-secret with full prompt body"),
	}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		aiTextHostCallContext(),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		protocol.MarshalHostServiceAITextGenerateRequest(
			&protocol.HostServiceAITextGenerateRequest{
				Purpose:         "content.summary",
				Tier:            aitext.TierBasic,
				MaxOutputTokens: 128,
				Messages: []aitext.Message{
					{Role: aitext.MessageRoleUser, Content: "full prompt body"},
				},
			},
		),
	)
	if response.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request on provider failure, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload := string(response.Payload)
	for _, forbidden := range []string{"sk-secret", "bearer", "full prompt body"} {
		if strings.Contains(strings.ToLower(payload), forbidden) {
			t.Fatalf("expected redacted ai error, got payload=%q", payload)
		}
	}
}

// TestHandleHostServiceInvokeRejectsRemovedPluginGovernanceMethods verifies
// legacy unscoped plugin governance method names are rejected dynamically.
func TestHandleHostServiceInvokeRejectsRemovedPluginGovernanceMethods(t *testing.T) {
	services := &capabilityHostServiceTestServices{
		org:           orgspi.New(nil, nil, nil),
		aiText:        aitext.New(nil, nil, nil),
		plugins:       &capabilityHostServicePluginsService{},
		tenant:        tenantspi.New(nil, nil, nil, nil),
		scopeRecorder: &capabilityHostServiceScopeRecorder{},
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	for _, method := range []string{
		"plugins.enabled.check",
		"plugins.provider_enabled.check",
		"plugins.enabled_authoritative.check",
		"lifecycle.tenant_plugin_disable.ensure",
		"lifecycle.tenant_plugin_disabled.notify",
		"lifecycle.tenant_delete.ensure",
		"lifecycle.tenant_deleted.notify",
	} {
		method := method
		t.Run(method, func(t *testing.T) {
			response := invokeCapabilityHostService(
				t,
				removedPluginGovernanceHostCallContext(method),
				protocol.HostServicePlugins,
				method,
				marshalCapabilityJSONRequest(t, map[string]any{"pluginId": "target-plugin", "tenantId": 41}),
			)
			if response.Status != protocol.HostCallStatusNotFound {
				t.Fatalf("expected removed plugin governance method to be unregistered, got status=%d payload=%s", response.Status, string(response.Payload))
			}
		})
	}
}

// TestConfigureDomainHostServicesRejectNil verifies nil domain service
// directories fail during startup wiring.
func TestConfigureDomainHostServicesRejectNil(t *testing.T) {
	if _, err := NewRuntime(
		nil,
		noopTestConfigFactory{},
		noopTestHostConfigService{},
		noopTestManifestFactory{},
	); err == nil {
		t.Fatal("expected nil domain host service directory to return an error")
	}
}

// invokeCapabilityHostService dispatches one method-authorized host-service request.
func invokeCapabilityHostService(
	t *testing.T,
	hcc *hostCallContext,
	service string,
	method string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()
	request := &protocol.HostServiceRequestEnvelope{
		Service: service,
		Method:  method,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// orgTenantHostCallContext builds an authorized org and tenant host service context.
func orgTenantHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-capability-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityOrg:    {},
			protocol.CapabilityTenant: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceOrg,
				Methods: []string{
					protocol.HostServiceMethodOrgAvailable,
					protocol.HostServiceMethodOrgStatus,
					protocol.HostServiceMethodOrgBatchGetUserOrgProfiles,
					protocol.HostServiceMethodOrgDepartmentBatchGet,
					protocol.HostServiceMethodOrgPostBatchGet,
				},
			},
			{
				Service: protocol.HostServiceTenant,
				Methods: []string{
					protocol.HostServiceMethodTenantAvailable,
					protocol.HostServiceMethodTenantStatus,
					protocol.HostServiceMethodTenantCurrent,
					protocol.HostServiceMethodTenantPlatformBypass,
					protocol.HostServiceMethodTenantValidateUserInTenant,
					protocol.HostServiceMethodTenantListUserTenants,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 7,
			UserID:   12,
			Username: "operator",
		},
	}
}

// tenantGovernanceHostCallContext builds an authorized tenant governance host
// service context for dynamically published tenant-domain methods.
func tenantGovernanceHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-tenant-governance-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityTenant: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceTenant,
				Methods: []string{
					protocol.HostServiceMethodTenantPluginSetEnabled,
					protocol.HostServiceMethodTenantPluginProvisionDefaults,
					protocol.HostServiceMethodTenantFilterContext,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 9,
			UserID:   21,
			Username: "tenant-governance-user",
		},
		requestID: "trace-tenant-governance",
	}
}

// userHostCallContext builds an authorized user host service context.
func userHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-user-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityUsers: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceUsers,
				Methods: []string{
					protocol.HostServiceMethodUsersCurrent,
					protocol.HostServiceMethodUsersBatchGet,
					protocol.HostServiceMethodUsersBatchResolve,
					protocol.HostServiceMethodUsersList,
					protocol.HostServiceMethodUsersEnsureVisible,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 7,
			UserID:   12,
			Username: "operator",
		},
	}
}

// aiTextHostCallContext builds an authorized AI text host service context.
func aiTextHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-ai-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityAIText: {},
			protocol.CapabilityAI:     {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceAI,
				Methods: []string{
					protocol.HostServiceMethodAITextGenerate,
					protocol.HostServiceMethodAITextMethodStatus,
					protocol.HostServiceMethodAIMethodStatuses,
				},
			},
		},
	}
}

// additionalDomainHostCallContext builds an authorized context for newly
// published ordinary domain host services.
func additionalDomainHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-domain-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityAuthz: {},
			protocol.CapabilityDict:  {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceAuthz,
				Methods: []string{
					protocol.HostServiceMethodAuthzBatchGetPermissions,
					protocol.HostServiceMethodAuthzBatchHasPermissions,
				},
			},
			{
				Service: protocol.HostServiceDict,
				Methods: []string{
					protocol.HostServiceMethodDictValueResolveLabels,
					protocol.HostServiceMethodDictListValues,
					protocol.HostServiceMethodDictValueEnsureValuesVisible,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId:    9,
			UserID:      21,
			Username:    "domain-user",
			Permissions: []string{"system:user:list"},
		},
		requestID: "trace-domain",
	}
}

// filesHostCallContext builds an authorized files host-service context.
func filesHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-files-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityFiles:   {},
			protocol.CapabilityStorage: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceFiles,
				Methods: []string{
					protocol.HostServiceMethodFilesUpload,
					protocol.HostServiceMethodFilesCreateFromStorage,
				},
			},
			{
				Service: protocol.HostServiceStorage,
				Methods: []string{
					protocol.HostServiceMethodStorageGet,
				},
				Paths: []string{"exports/"},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 9,
			UserID:   21,
			Username: "files-user",
		},
		requestID: "trace-files",
	}
}

// sessionHostCallContext builds an authorized online-session host service context.
func sessionHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-session-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilitySessions: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceSessions,
				Methods: []string{
					protocol.HostServiceMethodSessionsCurrent,
					protocol.HostServiceMethodSessionsList,
					protocol.HostServiceMethodSessionsBatchGet,
					protocol.HostServiceMethodSessionsBatchGetUserOnlineStatus,
					protocol.HostServiceMethodSessionsEnsureVisible,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TokenID:  "token-1",
			TenantId: 7,
			UserID:   12,
			Username: "operator",
		},
		requestID: "trace-session",
	}
}

// removedPluginGovernanceHostCallContext builds a plugins host-service context
// that declares a removed method by raw string to prove registry rejection.
func removedPluginGovernanceHostCallContext(method string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-lifecycle",
		capabilities: map[string]struct{}{
			protocol.CapabilityPlugins: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServicePlugins,
				Methods: []string{method},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 9,
			UserID:   21,
			Username: "plugin-lifecycle-user",
		},
		requestID: "trace-plugin-lifecycle",
	}
}

// marshalCapabilityJSONRequest encodes one JSON request for domain host-service tests.
func marshalCapabilityJSONRequest(t *testing.T, value any) []byte {
	t.Helper()
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal capability JSON request failed: %v", err)
	}
	return protocol.MarshalHostServiceCapabilityJSONRequest(&protocol.HostServiceCapabilityJSONRequest{Value: content})
}

// decodeCapabilityJSONResponse decodes one transport JSON response for tests.
func decodeCapabilityJSONResponse(t *testing.T, payload []byte, out any) {
	t.Helper()
	response, err := protocol.UnmarshalHostServiceCapabilityJSONResponse(payload)
	if err != nil {
		t.Fatalf("decode capability response failed: %v", err)
	}
	if err = json.Unmarshal(response.Value, out); err != nil {
		t.Fatalf("decode capability JSON failed: %v", err)
	}
}

// capabilityHostServicePluginsService exposes fake plugin registry, state, and
// lifecycle services for plugins host-service dispatcher tests.
type capabilityHostServicePluginsService struct {
	state     capabilityplugincap.StateService
	lifecycle capabilityplugincap.LifecycleService
}

// Config returns no plugin config service for dispatcher tests.
func (*capabilityHostServicePluginsService) Config() capabilityplugincap.ConfigService { return nil }

// Registry returns the fake plugin registry service.
func (s *capabilityHostServicePluginsService) Registry() capabilityplugincap.RegistryService {
	return s
}

// State returns the fake plugin enablement service.
func (s *capabilityHostServicePluginsService) State() capabilityplugincap.StateService {
	if s.state != nil {
		return s.state
	}
	return s
}

// Lifecycle returns no lifecycle service for dispatcher tests.
func (s *capabilityHostServicePluginsService) Lifecycle() capabilityplugincap.LifecycleService {
	return s.lifecycle
}

// BatchGet returns an empty fake plugin projection batch.
func (*capabilityHostServicePluginsService) BatchGet(
	context.Context,
	[]capabilityplugincap.PluginID,
) (*capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID]{
		Items: map[capabilityplugincap.PluginID]*capabilityplugincap.PluginInfo{},
	}, nil
}

// Current returns a deterministic current plugin projection.
func (*capabilityHostServicePluginsService) Current(
	context.Context,
) (*capabilityplugincap.PluginInfo, error) {
	return &capabilityplugincap.PluginInfo{ID: "test-plugin", Installed: true, Enabled: true}, nil
}

// Get returns a deterministic plugin projection.
func (*capabilityHostServicePluginsService) Get(
	context.Context,
	capabilityplugincap.PluginID,
) (*capabilityplugincap.PluginInfo, error) {
	return &capabilityplugincap.PluginInfo{ID: "test-plugin", Installed: true, Enabled: true}, nil
}

// List returns an empty fake plugin projection page.
func (*capabilityHostServicePluginsService) List(
	context.Context,
	capabilityplugincap.ListInput,
) (*capmodel.PageResult[*capabilityplugincap.PluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.PluginInfo]{Items: []*capabilityplugincap.PluginInfo{}}, nil
}

// ListTenantPlugins returns an empty fake tenant plugin page.
func (*capabilityHostServicePluginsService) ListTenantPlugins(
	context.Context,
	capabilityplugincap.TenantListInput,
) (*capmodel.PageResult[*capabilityplugincap.TenantPluginInfo], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantPluginInfo]{Items: []*capabilityplugincap.TenantPluginInfo{}}, nil
}

// IsEnabled returns enabled for fake plugin capability tests.
func (*capabilityHostServicePluginsService) IsEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return true, nil
}

// IsProviderEnabled returns enabled for fake plugin capability tests.
func (*capabilityHostServicePluginsService) IsProviderEnabled(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return true, nil
}

// IsEnabledAuthoritative returns enabled for fake plugin capability tests.
func (*capabilityHostServicePluginsService) IsEnabledAuthoritative(context.Context, capabilityplugincap.PluginID) (bool, error) {
	return true, nil
}

// capabilityHostServiceAITextService records text AI requests in host-service tests.
type capabilityHostServiceAITextService struct {
	response           *aitext.GenerateResponse
	err                error
	lastRequest        aitext.GenerateRequest
	lastSourcePluginID string
	called             bool
}

// Available reports that the fake service is usable.
func (s *capabilityHostServiceAITextService) Available(context.Context) bool { return true }

// Status returns an available fake text AI capability status.
func (s *capabilityHostServiceAITextService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{
		CapabilityID:   aitext.CapabilityAITextV1,
		Available:      true,
		ActiveProvider: aitext.ProviderPluginID,
	}
}

// MethodStatus returns a fake text AI method status without provider internals.
func (s *capabilityHostServiceAITextService) MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	var (
		status    = s.Status(ctx)
		available = method == aicommon.CapabilityMethodTextGenerate
		reason    = ""
	)
	if !available {
		reason = "method_unsupported"
	}
	return aicommon.MethodStatus{
		CapabilityType:   aicommon.CapabilityTypeText,
		CapabilityMethod: method,
		Available:        available,
		Reason:           reason,
		CapabilityStatus: status,
	}
}

// GenerateText records and returns a deterministic fake response.
func (s *capabilityHostServiceAITextService) GenerateText(
	_ context.Context,
	request aitext.GenerateRequest,
) (*aitext.GenerateResponse, error) {
	s.called = true
	s.lastRequest = request
	if s.err != nil {
		return nil, s.err
	}
	if s.response == nil {
		return &aitext.GenerateResponse{Text: "ok", Tier: request.Tier}, nil
	}
	return s.response, nil
}

// capabilityHostServiceScopedAITextService records source plugin identity
// injected by the scoped capability directory in host-service tests.
type capabilityHostServiceScopedAITextService struct {
	base           *capabilityHostServiceAITextService
	sourcePluginID string
}

// Available delegates to the base fake service.
func (s *capabilityHostServiceScopedAITextService) Available(ctx context.Context) bool {
	return s.base.Available(ctx)
}

// Status delegates to the base fake service.
func (s *capabilityHostServiceScopedAITextService) Status(ctx context.Context) capmodel.CapabilityStatus {
	return s.base.Status(ctx)
}

// MethodStatus delegates to the base fake service.
func (s *capabilityHostServiceScopedAITextService) MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return s.base.MethodStatus(ctx, method)
}

// GenerateText records scoped source identity before delegating.
func (s *capabilityHostServiceScopedAITextService) GenerateText(
	ctx context.Context,
	request aitext.GenerateRequest,
) (*aitext.GenerateResponse, error) {
	s.base.lastSourcePluginID = s.sourcePluginID
	return s.base.GenerateText(ctx, request)
}

// capabilityHostServiceUsersService records user-domain requests in tests.
type capabilityHostServiceUsersService struct {
	users         map[capabilityusercap.UserID]*capabilityusercap.UserInfo
	ensureErr     error
	lastCurrent   bizctxcap.CurrentContext
	lastList      capabilityusercap.ListInput
	lastEnsureIDs []capabilityusercap.UserID
	lastResolve   capabilityusercap.BatchResolveInput
}

// Current returns the projection for the current actor user.
func (s *capabilityHostServiceUsersService) Current(
	ctx context.Context,
) (*capabilityusercap.UserInfo, error) {
	current := bizctxcap.CurrentFromContext(ctx)
	s.lastCurrent = current
	return s.users[capabilityusercap.UserID(fmt.Sprint(current.UserID))], nil
}

// BatchGet returns configured user projections and opaque missing IDs.
func (s *capabilityHostServiceUsersService) BatchGet(
	ctx context.Context,
	ids []capabilityusercap.UserID,
) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserInfo{},
		MissingIDs: []capabilityusercap.UserID{},
	}
	for _, id := range ids {
		if user := s.users[id]; user != nil {
			result.Items[id] = user
			continue
		}
		result.MissingIDs = append(result.MissingIDs, id)
	}
	return result, nil
}

// Get returns one configured user projection.
func (s *capabilityHostServiceUsersService) Get(ctx context.Context, id capabilityusercap.UserID) (*capabilityusercap.UserInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityusercap.UserID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// BatchResolve records user resolve input and returns deterministic projections.
func (s *capabilityHostServiceUsersService) BatchResolve(
	ctx context.Context,
	input capabilityusercap.BatchResolveInput,
) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastResolve = input
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey]{
		Items:      map[capabilityusercap.ResolveKey]*capabilityusercap.UserInfo{},
		MissingIDs: []capabilityusercap.ResolveKey{},
	}
	for _, id := range input.IDs {
		key := capabilityusercap.ResolveKey("id:" + string(id))
		if user := s.users[id]; user != nil {
			result.Items[key] = user
			continue
		}
		result.MissingIDs = append(result.MissingIDs, key)
	}
	for _, username := range input.Usernames {
		key := capabilityusercap.ResolveKey("username:" + username)
		for _, user := range s.users {
			if user.Username == username {
				result.Items[key] = user
				break
			}
		}
		if _, ok := result.Items[key]; !ok {
			result.MissingIDs = append(result.MissingIDs, key)
		}
	}
	for _, contact := range input.Contacts {
		result.MissingIDs = append(result.MissingIDs, capabilityusercap.ResolveKey("contact:"+contact))
	}
	return result, nil
}

// List returns configured users as a deterministic bounded page.
func (s *capabilityHostServiceUsersService) List(
	ctx context.Context,
	input capabilityusercap.ListInput,
) (*capmodel.PageResult[*capabilityusercap.UserInfo], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastList = input
	items := make([]*capabilityusercap.UserInfo, 0, len(s.users))
	for _, user := range s.users {
		items = append(items, user)
	}
	return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: items, Total: len(items)}, nil
}

// EnsureVisible records visibility-check user IDs.
func (s *capabilityHostServiceUsersService) EnsureVisible(
	ctx context.Context,
	ids []capabilityusercap.UserID,
) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastEnsureIDs = append([]capabilityusercap.UserID(nil), ids...)
	return s.ensureErr
}

// Create is unused by host-service dispatch tests.
func (s *capabilityHostServiceUsersService) Create(context.Context, capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	return "", nil
}

// Update is unused by host-service dispatch tests.
func (s *capabilityHostServiceUsersService) Update(context.Context, capabilityusercap.UpdateInput) error {
	return nil
}

// Delete is unused by host-service dispatch tests.
func (s *capabilityHostServiceUsersService) Delete(context.Context, capabilityusercap.UserID) error {
	return nil
}

// SetStatus is unused by host-service dispatch tests.
func (s *capabilityHostServiceUsersService) SetStatus(context.Context, capabilityusercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword is unused by host-service dispatch tests.
func (s *capabilityHostServiceUsersService) ResetPassword(context.Context, capabilityusercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations unused by host-service dispatch tests.
func (s *capabilityHostServiceUsersService) Assignment() capabilityusercap.AssignmentService {
	return capabilityHostServiceUserAssignments{}
}

// capabilityHostServiceUserAssignments accepts unused role replacements.
type capabilityHostServiceUserAssignments struct{}

// ReplaceRoles is unused by host-service dispatch tests.
func (capabilityHostServiceUserAssignments) ReplaceRoles(context.Context, capabilityusercap.UserID, []int) error {
	return nil
}

// capabilityHostServiceSessionsService records online-session requests in tests.
type capabilityHostServiceSessionsService struct {
	sessions          map[capabilitysessioncap.SessionID]*capabilitysessioncap.SessionInfo
	lastCurrent       bizctxcap.CurrentContext
	lastList          capabilitysessioncap.ListInput
	lastOnlineUserIDs []string
	lastEnsureIDs     []capabilitysessioncap.SessionID
}

// Current returns the session projection matching the current identity token.
func (s *capabilityHostServiceSessionsService) Current(
	ctx context.Context,
) (*capabilitysessioncap.SessionInfo, error) {
	current := bizctxcap.CurrentFromContext(ctx)
	s.lastCurrent = current
	tokenID := capabilitysessioncap.SessionID(current.TokenID)
	return s.sessions[tokenID], nil
}

// Get returns one configured session projection.
func (s *capabilityHostServiceSessionsService) Get(ctx context.Context, id capabilitysessioncap.SessionID) (*capabilitysessioncap.SessionInfo, error) {
	result, err := s.BatchGet(ctx, []capabilitysessioncap.SessionID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// List returns configured sessions as a deterministic bounded page.
func (s *capabilityHostServiceSessionsService) List(
	ctx context.Context,
	input capabilitysessioncap.ListInput,
) (*capmodel.PageResult[*capabilitysessioncap.SessionInfo], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastList = input
	items := make([]*capabilitysessioncap.SessionInfo, 0, len(s.sessions))
	for _, sessionItem := range s.sessions {
		items = append(items, sessionItem)
	}
	return &capmodel.PageResult[*capabilitysessioncap.SessionInfo]{Items: items, Total: len(items)}, nil
}

// BatchGet returns configured session projections and opaque missing IDs.
func (s *capabilityHostServiceSessionsService) BatchGet(
	ctx context.Context,
	ids []capabilitysessioncap.SessionID,
) (*capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	result := &capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID]{
		Items:      map[capabilitysessioncap.SessionID]*capabilitysessioncap.SessionInfo{},
		MissingIDs: []capabilitysessioncap.SessionID{},
	}
	for _, id := range ids {
		if sessionItem := s.sessions[id]; sessionItem != nil {
			result.Items[id] = sessionItem
			continue
		}
		result.MissingIDs = append(result.MissingIDs, id)
	}
	return result, nil
}

// BatchGetUserOnlineStatus returns deterministic online states for configured sessions.
func (s *capabilityHostServiceSessionsService) BatchGetUserOnlineStatus(
	ctx context.Context,
	userIDs []string,
) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastOnlineUserIDs = append([]string(nil), userIDs...)
	result := &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string]{
		Items:      map[string]*capabilitysessioncap.UserOnlineStatus{},
		MissingIDs: []string{},
	}
	for _, userID := range userIDs {
		count := 0
		for _, sessionItem := range s.sessions {
			if sessionItem != nil && sessionItem.UserID == userID {
				count++
			}
		}
		result.Items[userID] = &capabilitysessioncap.UserOnlineStatus{
			UserID:       userID,
			Online:       count > 0,
			SessionCount: count,
		}
	}
	return result, nil
}

// EnsureVisible records requested session IDs.
func (s *capabilityHostServiceSessionsService) EnsureVisible(
	ctx context.Context,
	ids []capabilitysessioncap.SessionID,
) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastEnsureIDs = append([]capabilitysessioncap.SessionID(nil), ids...)
	return nil
}

// Revoke is unused by host-service dispatch tests.
func (s *capabilityHostServiceSessionsService) Revoke(context.Context, capabilitysessioncap.SessionID) error {
	return nil
}

// RevokeMany is unused by host-service dispatch tests.
func (s *capabilityHostServiceSessionsService) RevokeMany(context.Context, []capabilitysessioncap.SessionID) error {
	return nil
}

// capabilityHostServiceAuthzService records authz-domain requests in tests.
type capabilityHostServiceAuthzService struct {
	lastCurrent bizctxcap.CurrentContext
}

// BatchGetPermissions returns deterministic permission projections.
func (s *capabilityHostServiceAuthzService) BatchGetPermissions(
	ctx context.Context,
	keys []capabilityauthz.PermissionKey,
) (*capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	result := &capmodel.BatchResult[*capabilityauthz.PermissionInfo, capabilityauthz.PermissionKey]{
		Items:      map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionInfo{},
		MissingIDs: []capabilityauthz.PermissionKey{},
	}
	for _, key := range keys {
		if key == "missing" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &capabilityauthz.PermissionInfo{
			Key:      key,
			LabelKey: "permissions." + string(key),
			Label:    "Permission " + string(key),
		}
	}
	return result, nil
}

// BatchHasPermissions reports true for permissions present in the standard business context.
func (s *capabilityHostServiceAuthzService) BatchHasPermissions(
	ctx context.Context,
	keys []capabilityauthz.PermissionKey,
) (map[capabilityauthz.PermissionKey]bool, error) {
	current := bizctxcap.CurrentFromContext(ctx)
	s.lastCurrent = current
	granted := map[string]struct{}{}
	for _, permission := range current.Permissions {
		granted[permission] = struct{}{}
	}
	result := make(map[capabilityauthz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		_, ok := granted[string(key)]
		result[key] = ok
	}
	return result, nil
}

// HasPermission reports true for deterministic tests.
func (s *capabilityHostServiceAuthzService) HasPermission(
	ctx context.Context,
	_ capabilityauthz.PermissionKey,
) (bool, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	return true, nil
}

// IsPlatformAdmin reports false for deterministic tests.
func (s *capabilityHostServiceAuthzService) IsPlatformAdmin(
	ctx context.Context,
	_ capabilityauthz.UserID,
) (bool, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	return false, nil
}

// ReplaceRolePermissions is unused by host-service dispatch tests.
func (s *capabilityHostServiceAuthzService) ReplaceRolePermissions(context.Context, capabilityauthz.RoleID, []capabilityauthz.PermissionKey) error {
	return nil
}

// capabilityHostServiceDictService records dictionary-domain requests in tests.
type capabilityHostServiceDictService struct {
	lastCurrent     bizctxcap.CurrentContext
	lastInput       capabilitydictcap.ResolveInput
	lastListInput   capabilitydictcap.ListValuesInput
	denyBlankValues bool
}

// Type returns fake dictionary type subresource methods.
func (s *capabilityHostServiceDictService) Type() capabilitydictcap.TypeService {
	return capabilityHostServiceDictTypeService{parent: s}
}

// Value returns fake dictionary value subresource methods.
func (s *capabilityHostServiceDictService) Value() capabilitydictcap.ValueService {
	return capabilityHostServiceDictValueService{parent: s}
}

// capabilityHostServiceDictTypeService implements unused dictionary type methods.
type capabilityHostServiceDictTypeService struct {
	parent *capabilityHostServiceDictService
}

// capabilityHostServiceDictValueService records dictionary value-domain requests in tests.
type capabilityHostServiceDictValueService struct {
	parent *capabilityHostServiceDictService
}

// Get is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) Get(context.Context, int) (*capabilitydictcap.TypeInfo, error) {
	return nil, nil
}

// BatchGet is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) BatchGet(context.Context, []int) (*capmodel.BatchResult[*capabilitydictcap.TypeInfo, int], error) {
	return &capmodel.BatchResult[*capabilitydictcap.TypeInfo, int]{Items: map[int]*capabilitydictcap.TypeInfo{}}, nil
}

// List is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) List(context.Context, capabilitydictcap.ListTypesInput) (*capmodel.PageResult[*capabilitydictcap.TypeInfo], error) {
	return &capmodel.PageResult[*capabilitydictcap.TypeInfo]{Items: []*capabilitydictcap.TypeInfo{}}, nil
}

// EnsureVisible is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) EnsureVisible(context.Context, []int) error {
	return nil
}

// EnsureKeysVisible is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) EnsureKeysVisible(context.Context, []capabilitydictcap.Type) error {
	return nil
}

// Create is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) Create(context.Context, capabilitydictcap.CreateTypeInput) (int, error) {
	return 0, nil
}

// Update is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) Update(context.Context, capabilitydictcap.UpdateTypeInput) error {
	return nil
}

// Delete is unused by dictionary type dispatcher tests.
func (s capabilityHostServiceDictTypeService) Delete(context.Context, int) error {
	return nil
}

// Get is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) Get(context.Context, int) (*capabilitydictcap.ValueInfo, error) {
	return nil, nil
}

// BatchGet is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) BatchGet(context.Context, capabilitydictcap.BatchGetValuesInput) (*capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value], error) {
	return &capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value]{Items: map[capabilitydictcap.Value]*capabilitydictcap.ValueInfo{}}, nil
}

// ResolveLabels returns deterministic dictionary label projections.
func (s capabilityHostServiceDictValueService) ResolveLabels(
	ctx context.Context,
	input capabilitydictcap.ResolveInput,
) (*capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value], error) {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastInput = input
	result := &capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value]{
		Items:      map[capabilitydictcap.Value]*capabilitydictcap.LabelInfo{},
		MissingIDs: []capabilitydictcap.Value{},
	}
	for _, value := range input.Values {
		result.Items[value] = &capabilitydictcap.LabelInfo{
			Type:     input.Type,
			Value:    value,
			LabelKey: "dict." + string(input.Type) + "." + string(value),
			Label:    "Enabled",
		}
	}
	return result, nil
}

// List returns deterministic dictionary value candidates.
func (s capabilityHostServiceDictValueService) List(
	ctx context.Context,
	input capabilitydictcap.ListValuesInput,
) (*capmodel.PageResult[*capabilitydictcap.ValueInfo], error) {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastListInput = input
	return &capmodel.PageResult[*capabilitydictcap.ValueInfo]{
		Items: []*capabilitydictcap.ValueInfo{{
			Type:     input.Type,
			Value:    "enabled",
			LabelKey: "dict." + string(input.Type) + ".enabled",
			Label:    "Enabled",
		}},
		Total: 1,
	}, nil
}

// EnsureVisible is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) EnsureVisible(context.Context, []int) error {
	return nil
}

// EnsureValuesVisible records dictionary visibility-check input.
func (s capabilityHostServiceDictValueService) EnsureValuesVisible(
	ctx context.Context,
	input capabilitydictcap.ResolveInput,
) error {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastInput = input
	if s.parent.denyBlankValues {
		for _, value := range input.Values {
			if strings.TrimSpace(string(value)) == "" {
				return bizerr.NewCode(capmodel.CodeCapabilityDenied)
			}
		}
	}
	return nil
}

// Create is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) Create(context.Context, capabilitydictcap.CreateValueInput) (int, error) {
	return 0, nil
}

// Update is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) Update(context.Context, capabilitydictcap.UpdateValueInput) error {
	return nil
}

// Delete is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) Delete(context.Context, int) error {
	return nil
}

// DeleteByType is unused by dictionary value dispatcher tests.
func (s capabilityHostServiceDictValueService) DeleteByType(context.Context, capabilitydictcap.Type) error {
	return nil
}

// Refresh is unused by host-service dispatch tests.
func (s *capabilityHostServiceDictService) Refresh(context.Context, capabilitydictcap.Type) error {
	return nil
}

// capabilityHostServiceFilesService records dynamic file write calls.
type capabilityHostServiceFilesService struct {
	capabilityfilecap.Service
	lastUpload       capabilityfilecap.UploadInput
	lastUploadBody   string
	lastStorageInput capabilityfilecap.CreateFromStorageInput
	lastCurrent      bizctxcap.CurrentContext
}

// Upload records one decoded dynamic file upload request.
func (s *capabilityHostServiceFilesService) Upload(
	ctx context.Context,
	input capabilityfilecap.UploadInput,
) (*capabilityfilecap.FileInfo, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastUpload = input
	body, err := io.ReadAll(input.Reader)
	if err != nil {
		return nil, err
	}
	s.lastUploadBody = string(body)
	return &capabilityfilecap.FileInfo{
		ID:            "file-uploaded",
		Name:          input.Filename,
		SizeBytes:     int64(len(body)),
		BusinessScene: input.BusinessScene,
	}, nil
}

// CreateFromStorage records one decoded dynamic storage-promotion request.
func (s *capabilityHostServiceFilesService) CreateFromStorage(
	ctx context.Context,
	input capabilityfilecap.CreateFromStorageInput,
) (*capabilityfilecap.FileInfo, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastStorageInput = input
	return &capabilityfilecap.FileInfo{
		ID:            "file-from-storage",
		Name:          input.Filename,
		SizeBytes:     input.SizeBytes,
		BusinessScene: input.BusinessScene,
	}, nil
}

// capabilityHostServiceOrgProvider is a deterministic organization provider for tests.
type capabilityHostServiceOrgProvider struct{}

// ListUserDeptAssignments returns deterministic department assignments.
func (capabilityHostServiceOrgProvider) ListUserDeptAssignments(
	_ context.Context,
	userIDs []int,
) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment, len(userIDs))
	for _, userID := range userIDs {
		assignments[userID] = &orgcap.UserDeptAssignment{
			DeptID:   userID + 10,
			DeptName: fmt.Sprintf("Dept-%d", userID),
		}
	}
	return assignments, nil
}

// GetUserDeptInfo returns a deterministic department projection.
func (capabilityHostServiceOrgProvider) GetUserDeptInfo(_ context.Context, userID int) (int, string, error) {
	return userID + 10, fmt.Sprintf("Dept-%d", userID), nil
}

// GetUserDeptIDs returns deterministic department identifiers.
func (capabilityHostServiceOrgProvider) GetUserDeptIDs(_ context.Context, userID int) ([]int, error) {
	return []int{userID + 10}, nil
}

// ApplyUserDeptScope returns the input model unchanged.
func (capabilityHostServiceOrgProvider) ApplyUserDeptScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ int,
) (*gdb.Model, bool, error) {
	return model, true, nil
}

// BuildUserDeptScopeExists returns an empty scope marker.
func (capabilityHostServiceOrgProvider) BuildUserDeptScopeExists(context.Context, string, int) (*gdb.Model, bool, error) {
	return nil, true, nil
}

// ApplyUserDeptFilter returns the input model unchanged.
func (capabilityHostServiceOrgProvider) ApplyUserDeptFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ int,
) (*gdb.Model, bool, error) {
	return model, true, nil
}

// ApplyUserDeptUnassignedFilter returns the input model unchanged.
func (capabilityHostServiceOrgProvider) ApplyUserDeptUnassignedFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// GetUserPostIDs returns deterministic post identifiers.
func (capabilityHostServiceOrgProvider) GetUserPostIDs(_ context.Context, userID int) ([]int, error) {
	return []int{userID + 100}, nil
}

// BatchGetUserOrgProfiles returns deterministic organization profiles.
func (capabilityHostServiceOrgProvider) BatchGetUserOrgProfiles(
	_ context.Context,
	userIDs []int,
) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	result := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      map[int]*orgcap.UserOrgProfile{},
		MissingIDs: []int{},
	}
	for _, userID := range userIDs {
		result.Items[userID] = &orgcap.UserOrgProfile{
			UserID:    userID,
			DeptID:    userID + 10,
			DeptName:  fmt.Sprintf("Dept-%d", userID),
			PostIDs:   []int{userID + 100},
			PostNames: []string{fmt.Sprintf("Post-%d", userID)},
		}
	}
	return result, nil
}

// ListDeptTree returns an empty bounded department tree.
func (capabilityHostServiceOrgProvider) ListDeptTree(context.Context, orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}, nil
}

// SearchDepartments returns an empty department candidate page.
func (capabilityHostServiceOrgProvider) SearchDepartments(context.Context, orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
}

// BatchGetDepartments returns deterministic department projections.
func (capabilityHostServiceOrgProvider) BatchGetDepartments(
	_ context.Context,
	deptIDs []int,
) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.DeptInfo, int]{
		Items:      map[int]*orgcap.DeptInfo{},
		MissingIDs: []int{},
	}
	for _, deptID := range deptIDs {
		result.Items[deptID] = &orgcap.DeptInfo{DeptID: deptID, DeptName: fmt.Sprintf("Dept-%d", deptID)}
	}
	return result, nil
}

// CreateDepartment returns a deterministic department identifier.
func (capabilityHostServiceOrgProvider) CreateDepartment(context.Context, orgcap.DeptCreateInput) (int, error) {
	return 0, nil
}

// UpdateDepartment accepts department updates.
func (capabilityHostServiceOrgProvider) UpdateDepartment(context.Context, orgcap.DeptUpdateInput) error {
	return nil
}

// DeleteDepartment accepts department deletes.
func (capabilityHostServiceOrgProvider) DeleteDepartment(context.Context, int) error {
	return nil
}

// ListPostOptionsPage returns an empty post candidate page.
func (capabilityHostServiceOrgProvider) ListPostOptionsPage(context.Context, orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
}

// GetPost returns no post projection.
func (capabilityHostServiceOrgProvider) GetPost(context.Context, int) (*orgcap.PostInfo, error) {
	return nil, nil
}

// BatchGetPosts returns deterministic post projections.
func (capabilityHostServiceOrgProvider) BatchGetPosts(
	_ context.Context,
	postIDs []int,
) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.PostInfo, int]{
		Items:      map[int]*orgcap.PostInfo{},
		MissingIDs: []int{},
	}
	for _, postID := range postIDs {
		result.Items[postID] = &orgcap.PostInfo{PostID: postID, PostName: fmt.Sprintf("Post-%d", postID)}
	}
	return result, nil
}

// ListPosts returns no post projections.
func (capabilityHostServiceOrgProvider) ListPosts(context.Context, orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{}}, nil
}

// CreatePost returns a deterministic post identifier.
func (capabilityHostServiceOrgProvider) CreatePost(context.Context, orgcap.PostCreateInput) (int, error) {
	return 0, nil
}

// UpdatePost accepts post updates.
func (capabilityHostServiceOrgProvider) UpdatePost(context.Context, orgcap.PostUpdateInput) error {
	return nil
}

// DeletePost accepts post deletes.
func (capabilityHostServiceOrgProvider) DeletePost(context.Context, int) error {
	return nil
}

// EnsureDepartmentsVisible accepts all test department identifiers.
func (capabilityHostServiceOrgProvider) EnsureDepartmentsVisible(context.Context, []int) error {
	return nil
}

// EnsurePostsVisible accepts all test post identifiers.
func (capabilityHostServiceOrgProvider) EnsurePostsVisible(context.Context, []int) error {
	return nil
}

// ReplaceUserAssignments ignores assignment writes.
func (capabilityHostServiceOrgProvider) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

// CleanupUserAssignments ignores cleanup writes.
func (capabilityHostServiceOrgProvider) CleanupUserAssignments(context.Context, int) error {
	return nil
}

// UserDeptTree returns an empty department tree.
func (capabilityHostServiceOrgProvider) UserDeptTree(context.Context) ([]*orgcap.DeptTreeNode, error) {
	return []*orgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns no post options.
func (capabilityHostServiceOrgProvider) ListPostOptions(context.Context, *int) ([]*orgcap.PostOption, error) {
	return []*orgcap.PostOption{}, nil
}

// capabilityHostServiceOrgRuntime marks the test provider plugin enabled.
type capabilityHostServiceOrgRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the test organization provider plugin is enabled.
func (r capabilityHostServiceOrgRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
}

// capabilityHostServiceTenantService records tenant method calls for tests.
type capabilityHostServiceTenantService struct {
	tenants    []tenantcap.TenantInfo
	lastUserID int
	plugins    tenantcap.PluginService
	filter     tenantcap.FilterService
}

// Available reports that the tenant service is active.
func (*capabilityHostServiceTenantService) Available(context.Context) bool { return true }

// Status returns an active tenant capability status.
func (*capabilityHostServiceTenantService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{Available: true, ActiveProvider: "test-tenant-provider"}
}

// Context returns tenant context operations.
func (s *capabilityHostServiceTenantService) Context() tenantcap.ContextService { return s }

// Directory returns tenant directory operations.
func (s *capabilityHostServiceTenantService) Directory() tenantcap.DirectoryService { return s }

// Membership returns tenant membership operations.
func (s *capabilityHostServiceTenantService) Membership() tenantcap.MembershipService { return s }

// Plugins returns no tenant-plugin governance service by default.
func (s *capabilityHostServiceTenantService) Plugins() tenantcap.PluginService { return s.plugins }

// Filter returns no tenant filter context service by default.
func (s *capabilityHostServiceTenantService) Filter() tenantcap.FilterService {
	return s.filter
}

// trackingTenantPluginService records tenant-plugin governance calls.
type trackingTenantPluginService struct {
	lastPluginID          capabilityplugincap.PluginID
	lastEnabled           bool
	lastProvisionTenantID capmodel.DomainID
}

// SetTenantPluginEnabled records one tenant-plugin enablement update.
func (s *trackingTenantPluginService) SetTenantPluginEnabled(
	_ context.Context,
	pluginID capabilityplugincap.PluginID,
	enabled bool,
) error {
	s.lastPluginID = pluginID
	s.lastEnabled = enabled
	return nil
}

// ProvisionTenantPluginDefaults records one default tenant-plugin provisioning call.
func (s *trackingTenantPluginService) ProvisionTenantPluginDefaults(
	_ context.Context,
	tenantID capmodel.DomainID,
) error {
	s.lastProvisionTenantID = tenantID
	return nil
}

// trackingTenantFilterService records tenant filter context calls.
type trackingTenantFilterService struct {
	contextValue tenantspi.TenantFilterContext
	contextCalls int
}

// Context returns a configured tenant filter context.
func (s *trackingTenantFilterService) Context(context.Context) tenantspi.TenantFilterContext {
	s.contextCalls++
	return s.contextValue
}

// Current returns a deterministic current tenant.
func (*capabilityHostServiceTenantService) Current(context.Context) tenantcap.TenantID { return 3 }

// Info returns the deterministic current tenant projection.
func (*capabilityHostServiceTenantService) Info(context.Context) (*tenantcap.TenantInfo, error) {
	return &tenantcap.TenantInfo{ID: 3, Code: "tenant-a", Name: "Tenant A", Status: "active"}, nil
}

// PlatformBypass reports no bypass.
func (*capabilityHostServiceTenantService) PlatformBypass(context.Context) bool { return false }

// EnsureVisible accepts all tenant identifiers.
func (*capabilityHostServiceTenantService) EnsureVisible(context.Context, []tenantcap.TenantID) error {
	return nil
}

// BatchGet returns deterministic visible tenant projections.
func (s *capabilityHostServiceTenantService) BatchGet(_ context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	result := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{
		Items:      map[tenantcap.TenantID]*tenantcap.TenantInfo{},
		MissingIDs: []tenantcap.TenantID{},
	}
	for _, tenantID := range tenantIDs {
		for _, tenant := range s.tenants {
			if tenant.ID == tenantID {
				value := tenant
				result.Items[tenantID] = &value
				break
			}
		}
		if _, ok := result.Items[tenantID]; !ok {
			result.MissingIDs = append(result.MissingIDs, tenantID)
		}
	}
	return result, nil
}

// Get returns one visible tenant projection.
func (s *capabilityHostServiceTenantService) Get(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error) {
	result, err := s.BatchGet(ctx, []tenantcap.TenantID{tenantID})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[tenantID], nil
}

// List returns the configured tenant page.
func (s *capabilityHostServiceTenantService) List(context.Context, tenantcap.ListInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	items := make([]*tenantcap.TenantInfo, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		value := tenant
		items = append(items, &value)
	}
	return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: items, Total: len(items)}, nil
}

// Validate accepts all user and tenant pairs.
func (*capabilityHostServiceTenantService) Validate(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// ListByUser returns configured tenants and records the requested user.
func (s *capabilityHostServiceTenantService) ListByUser(
	_ context.Context,
	userID int,
) ([]tenantcap.TenantInfo, error) {
	s.lastUserID = userID
	return s.tenants, nil
}
