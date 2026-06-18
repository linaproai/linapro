// This file tests structured dynamic-plugin host-service dispatch through the
// shared dispatcher and unified capability services.

package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

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
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
var _ capability.ScopedServicesFactory = (*capabilityHostServiceTestServices)(nil)

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

// PluginConfig returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) PluginConfig() plugincap.ConfigService { return nil }

// Dict returns the configured dictionary-domain service.
func (s *capabilityHostServiceTestServices) Dict() capabilitydictcap.Service { return s.dict }

// Files returns no file-domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Files() capabilityfilecap.Service { return nil }

// HostConfig returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) HostConfig() hostconfigcap.Service { return nil }

// I18n returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) I18n() i18ncap.Service { return nil }

// Infra returns no infrastructure-domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Infra() capabilityinfracap.Service { return nil }

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

// PluginLifecycle returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) PluginLifecycle() plugincap.LifecycleService {
	return nil
}

// PluginState returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) PluginState() plugincap.StateService { return nil }

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
		org:           orgspi.New(orgManager, capabilityHostServiceOrgRuntime{pluginID: providerPluginID}),
		aiText:        aitext.New(nil, nil),
		users:         &capabilityHostServiceUsersService{},
		tenant:        tenantspi.New(nil, nil, nil),
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

	assignmentsResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListUserDeptAssignments,
		marshalCapabilityJSONRequest(t, intUserIDsRequest{UserIDs: []int{7, 8}}),
	)
	if assignmentsResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org assignment success, got status=%d payload=%s", assignmentsResponse.Status, string(assignmentsResponse.Payload))
	}
	var assignments map[int]*orgcap.UserDeptAssignment
	decodeCapabilityJSONResponse(t, assignmentsResponse.Payload, &assignments)
	if assignments[7] == nil || assignments[7].DeptID != 17 || assignments[8] == nil || assignments[8].DeptName != "Dept-8" {
		t.Fatalf("unexpected assignment payload: %#v", assignments)
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
		org:    orgspi.New(orgManager, capabilityHostServiceOrgRuntime{pluginID: providerPluginID}),
		aiText: aitext.New(nil, nil),
		users:  userSvc,
		tenant: tenantspi.New(nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptName,
		marshalCapabilityJSONRequest(t, intUserIDRequest{UserID: 42}),
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
		users: map[capabilityusercap.UserID]*capabilityusercap.UserProjection{
			"12": {ID: "12", Username: "operator", Nickname: "Operator", Status: "1"},
			"42": {ID: "42", Username: "admin", Nickname: "Administrator", Status: "1"},
		},
	}
	services := &capabilityHostServiceTestServices{
		org:           orgspi.New(nil, nil),
		aiText:        aitext.New(nil, nil),
		users:         userSvc,
		tenant:        tenantspi.New(nil, nil, nil),
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
	var current *capabilityusercap.UserProjection
	decodeCapabilityJSONResponse(t, currentResponse.Payload, &current)
	if current == nil || current.ID != "12" || userSvc.lastCapCtx.Actor.UserID != 12 {
		t.Fatalf("unexpected current user payload current=%#v capCtx=%#v", current, userSvc.lastCapCtx)
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
	var batch capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]
	decodeCapabilityJSONResponse(t, response.Payload, &batch)
	if batch.Items["42"] == nil || batch.Items["42"].Username != "admin" || !reflect.DeepEqual(batch.MissingIDs, []capabilityusercap.UserID{"99"}) {
		t.Fatalf("unexpected user batch payload: %#v", batch)
	}
	if services.scopeRecorder.last() != "test-user-plugin" || userSvc.lastCapCtx.PluginID != "test-user-plugin" || userSvc.lastCapCtx.Actor.UserID != 12 {
		t.Fatalf("expected plugin-scoped user context, lastPlugin=%q capCtx=%#v", services.scopeRecorder.last(), userSvc.lastCapCtx)
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
	var resolved capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey]
	decodeCapabilityJSONResponse(t, resolveResponse.Payload, &resolved)
	if resolved.Items["id:42"] == nil || resolved.Items["username:admin"] == nil || !reflect.DeepEqual(userSvc.lastResolve.Usernames, []string{"admin"}) {
		t.Fatalf("unexpected user resolve payload result=%#v input=%#v", resolved, userSvc.lastResolve)
	}

	searchResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersSearch,
		marshalCapabilityJSONRequest(t, struct {
			Keyword  string `json:"keyword,omitempty"`
			PageNum  int    `json:"pageNum,omitempty"`
			PageSize int    `json:"pageSize,omitempty"`
		}{Keyword: "adm", PageNum: 1, PageSize: 10}),
	)
	if searchResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user search success, got status=%d payload=%s", searchResponse.Status, string(searchResponse.Payload))
	}
	var page capmodel.PageResult[*capabilityusercap.UserProjection]
	decodeCapabilityJSONResponse(t, searchResponse.Payload, &page)
	if page.Total != 2 || len(page.Items) != 2 || userSvc.lastSearch.Keyword != "adm" {
		t.Fatalf("unexpected user search payload page=%#v lastSearch=%#v", page, userSvc.lastSearch)
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
// directory and pass CapabilityContext to domain adapters.
func TestHandleHostServiceInvokeAdditionalDomainMethods(t *testing.T) {
	authzSvc := &capabilityHostServiceAuthzService{}
	dictSvc := &capabilityHostServiceDictService{}
	services := &capabilityHostServiceTestServices{
		auth:   authcap.New(nil, authzSvc),
		org:    orgspi.New(nil, nil),
		aiText: aitext.New(nil, nil),
		dict:   dictSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
	var permissions capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey]
	decodeCapabilityJSONResponse(t, authzResponse.Payload, &permissions)
	if permissions.Items["system:user:list"] == nil || permissions.Items["system:user:list"].LabelKey != "permissions.system:user:list" {
		t.Fatalf("unexpected authz payload: %#v", permissions)
	}
	if authzSvc.lastCapCtx.PluginID != "test-domain-plugin" || authzSvc.lastCapCtx.Actor.UserID != 21 {
		t.Fatalf("expected authz capability context, got %#v", authzSvc.lastCapCtx)
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
		protocol.HostServiceMethodDictResolveLabels,
		marshalCapabilityJSONRequest(t, map[string]any{
			"type":         "sys_common_status",
			"values":       []string{"enabled"},
			"includeLabel": true,
		}),
	)
	if dictResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict resolve success, got status=%d payload=%s", dictResponse.Status, string(dictResponse.Payload))
	}
	var labels capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value]
	decodeCapabilityJSONResponse(t, dictResponse.Payload, &labels)
	if labels.Items["enabled"] == nil || labels.Items["enabled"].Label != "Enabled" {
		t.Fatalf("unexpected dict payload: %#v", labels)
	}
	if dictSvc.lastCapCtx.PluginID != "test-domain-plugin" || dictSvc.lastInput.Type != "sys_common_status" {
		t.Fatalf("expected dict capability context and input, capCtx=%#v input=%#v", dictSvc.lastCapCtx, dictSvc.lastInput)
	}

	dictEnsureResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictEnsureValuesVisible,
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
	var dictList capmodel.PageResult[*capabilitydictcap.LabelProjection]
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
		org:    orgspi.New(nil, nil),
		aiText: aitext.New(nil, nil),
		dict:   dictSvc,
		tenant: tenantspi.New(nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictEnsureValuesVisible,
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

// TestHandleHostServiceInvokeSessionMethods verifies session host-service calls
// are routed through the shared online-session capability service.
func TestHandleHostServiceInvokeSessionMethods(t *testing.T) {
	sessionSvc := &capabilityHostServiceSessionsService{
		sessions: map[capabilitysessioncap.SessionID]*capabilitysessioncap.Projection{
			"token-1": {ID: "token-1", UserID: "12", Username: "operator"},
		},
	}
	services := &capabilityHostServiceTestServices{
		org:      orgspi.New(nil, nil),
		aiText:   aitext.New(nil, nil),
		sessions: sessionSvc,
		tenant:   tenantspi.New(nil, nil, nil),
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
	var current *capabilitysessioncap.Projection
	decodeCapabilityJSONResponse(t, currentResponse.Payload, &current)
	if current == nil || current.ID != "token-1" || sessionSvc.lastCapCtx.PluginID != "test-session-plugin" {
		t.Fatalf("unexpected session current payload current=%#v capCtx=%#v", current, sessionSvc.lastCapCtx)
	}

	searchResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsSearch,
		marshalCapabilityJSONRequest(t, map[string]any{"username": "operator", "pageNum": 1, "pageSize": 10}),
	)
	if searchResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session search success, got status=%d payload=%s", searchResponse.Status, string(searchResponse.Payload))
	}
	if sessionSvc.lastSearch.Username != "operator" {
		t.Fatalf("unexpected session search input: %#v", sessionSvc.lastSearch)
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
	var batch capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID]
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
	var online capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string]
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
		org:    orgspi.New(nil, nil),
		aiText: aitext.New(nil, nil),
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

	switchResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantValidateSwitch,
		marshalCapabilityJSONRequest(t, tenantSwitchRequest{UserID: 42, TargetTenantID: 3}),
	)
	if switchResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant switch success, got status=%d payload=%s", switchResponse.Status, string(switchResponse.Payload))
	}
	if tenantSvc.lastSwitchUserID != 42 || tenantSvc.lastSwitchTarget != 3 {
		t.Fatalf("unexpected switch call user=%d target=%d", tenantSvc.lastSwitchUserID, tenantSvc.lastSwitchTarget)
	}
}

// TestHandleHostServiceInvokeTenantRejectsInvisibleTargetUser verifies tenant
// host services do not reveal memberships for users outside actor data scope.
func TestHandleHostServiceInvokeTenantRejectsInvisibleTargetUser(t *testing.T) {
	userSvc := &capabilityHostServiceUsersService{ensureErr: errors.New("target user is outside data scope")}
	tenantSvc := &capabilityHostServiceTenantService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil),
		aiText: aitext.New(nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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
		org:    orgspi.New(nil, nil),
		aiText: aiSvc,
		tenant: tenantspi.New(nil, nil, nil),
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

// TestHandleHostServiceInvokePluginLifecycleMethods verifies plugin lifecycle
// governance methods are exposed through the plugins domain when explicitly
// authorized by service and method.
func TestHandleHostServiceInvokePluginLifecycleMethods(t *testing.T) {
	lifecycle := &capabilityHostServicePluginLifecycle{}
	services := &capabilityHostServiceTestServices{
		org:           orgspi.New(nil, nil),
		aiText:        aitext.New(nil, nil),
		plugins:       &capabilityHostServicePluginsService{lifecycle: lifecycle},
		tenant:        tenantspi.New(nil, nil, nil),
		scopeRecorder: &capabilityHostServiceScopeRecorder{},
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	disableResponse := invokeCapabilityHostService(
		t,
		pluginLifecycleHostCallContext(true),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable,
		marshalCapabilityJSONRequest(t, map[string]any{"pluginId": "target-plugin", "tenantId": 41}),
	)
	if disableResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant-plugin disable lifecycle success, got status=%d payload=%s", disableResponse.Status, string(disableResponse.Payload))
	}
	if lifecycle.ensurePluginID != "target-plugin" || lifecycle.ensurePluginTenantID != 41 {
		t.Fatalf("expected tenant-plugin disable lifecycle call, got plugin=%q tenant=%d", lifecycle.ensurePluginID, lifecycle.ensurePluginTenantID)
	}

	notifyDisableResponse := invokeCapabilityHostService(
		t,
		pluginLifecycleHostCallContext(true),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
		marshalCapabilityJSONRequest(t, map[string]any{"pluginId": "target-plugin", "tenantId": 42}),
	)
	if notifyDisableResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant-plugin disabled notification success, got status=%d payload=%s", notifyDisableResponse.Status, string(notifyDisableResponse.Payload))
	}
	if lifecycle.notifyPluginID != "target-plugin" || lifecycle.notifyPluginTenantID != 42 {
		t.Fatalf("expected tenant-plugin disabled notification call, got plugin=%q tenant=%d", lifecycle.notifyPluginID, lifecycle.notifyPluginTenantID)
	}

	deleteResponse := invokeCapabilityHostService(
		t,
		pluginLifecycleHostCallContext(true),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantDelete,
		marshalCapabilityJSONRequest(t, map[string]any{"tenantId": 43}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant delete lifecycle success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
	if lifecycle.ensureTenantDeleteID != 43 {
		t.Fatalf("expected tenant delete lifecycle call, got tenant=%d", lifecycle.ensureTenantDeleteID)
	}

	notifyDeleteResponse := invokeCapabilityHostService(
		t,
		pluginLifecycleHostCallContext(true),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
		marshalCapabilityJSONRequest(t, map[string]any{"tenantId": 44}),
	)
	if notifyDeleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant deleted notification success, got status=%d payload=%s", notifyDeleteResponse.Status, string(notifyDeleteResponse.Payload))
	}
	if lifecycle.notifyTenantDeleteID != 44 {
		t.Fatalf("expected tenant deleted notification call, got tenant=%d", lifecycle.notifyTenantDeleteID)
	}
	if services.scopeRecorder.last() != "test-plugin-lifecycle" {
		t.Fatalf("expected plugin-scoped services, got %q", services.scopeRecorder.last())
	}
}

// TestHandleHostServiceInvokePluginLifecycleRequiresMethodAuthorization verifies
// plugins lifecycle calls are rejected before dispatch when the manifest snapshot
// did not grant the exact lifecycle method.
func TestHandleHostServiceInvokePluginLifecycleRequiresMethodAuthorization(t *testing.T) {
	lifecycle := &capabilityHostServicePluginLifecycle{}
	services := &capabilityHostServiceTestServices{
		org:     orgspi.New(nil, nil),
		aiText:  aitext.New(nil, nil),
		plugins: &capabilityHostServicePluginsService{lifecycle: lifecycle},
		tenant:  tenantspi.New(nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		pluginLifecycleHostCallContext(false),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantDelete,
		marshalCapabilityJSONRequest(t, map[string]any{"tenantId": 45}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected lifecycle method authorization to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if lifecycle.ensureTenantDeleteID != 0 {
		t.Fatalf("expected unauthorized lifecycle call to skip business dispatch, got tenant=%d", lifecycle.ensureTenantDeleteID)
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

// invokeCapabilityHostServiceWithResource dispatches one resource-scoped host-service request.
func invokeCapabilityHostServiceWithResource(
	t *testing.T,
	hcc *hostCallContext,
	service string,
	method string,
	resourceRef string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()
	request := &protocol.HostServiceRequestEnvelope{
		Service:     service,
		Method:      method,
		ResourceRef: resourceRef,
		Payload:     payload,
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
					protocol.HostServiceMethodOrgListUserDeptAssignments,
					protocol.HostServiceMethodOrgGetUserDeptInfo,
					protocol.HostServiceMethodOrgGetUserDeptName,
					protocol.HostServiceMethodOrgGetUserDeptIDs,
					protocol.HostServiceMethodOrgGetUserPostIDs,
				},
			},
			{
				Service: protocol.HostServiceTenant,
				Methods: []string{
					protocol.HostServiceMethodTenantAvailable,
					protocol.HostServiceMethodTenantStatus,
					protocol.HostServiceMethodTenantCurrent,
					protocol.HostServiceMethodTenantPlatformBypass,
					protocol.HostServiceMethodTenantEnsureVisible,
					protocol.HostServiceMethodTenantValidateUserInTenant,
					protocol.HostServiceMethodTenantListUserTenants,
					protocol.HostServiceMethodTenantValidateSwitch,
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
					protocol.HostServiceMethodUsersSearch,
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
					protocol.HostServiceMethodDictResolveLabels,
					protocol.HostServiceMethodDictListValues,
					protocol.HostServiceMethodDictEnsureValuesVisible,
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
					protocol.HostServiceMethodSessionsSearch,
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

// pluginLifecycleHostCallContext builds an authorized plugins host service context.
func pluginLifecycleHostCallContext(includeLifecycleMethods bool) *hostCallContext {
	methods := []string{
		protocol.HostServiceMethodPluginsIsEnabled,
	}
	if includeLifecycleMethods {
		methods = append(methods,
			protocol.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable,
			protocol.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
			protocol.HostServiceMethodPluginsLifecycleEnsureTenantDelete,
			protocol.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
		)
	}
	return &hostCallContext{
		pluginID: "test-plugin-lifecycle",
		capabilities: map[string]struct{}{
			protocol.CapabilityPlugins: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServicePlugins,
				Methods: methods,
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

// capabilityHostServicePluginsService exposes a fake plugin lifecycle service
// for plugins host-service dispatcher tests.
type capabilityHostServicePluginsService struct {
	lifecycle capabilityplugincap.LifecycleService
}

// Config returns no plugin config service for lifecycle dispatcher tests.
func (*capabilityHostServicePluginsService) Config() capabilityplugincap.ConfigService { return nil }

// State returns no plugin state service for lifecycle dispatcher tests.
func (*capabilityHostServicePluginsService) State() capabilityplugincap.StateService { return nil }

// Lifecycle returns the configured lifecycle service.
func (s *capabilityHostServicePluginsService) Lifecycle() capabilityplugincap.LifecycleService {
	return s.lifecycle
}

// Registry returns the fake plugin registry service.
func (s *capabilityHostServicePluginsService) Registry() capabilityplugincap.RegistryService {
	return s
}

// BatchGet returns an empty fake plugin projection batch.
func (*capabilityHostServicePluginsService) BatchGet(
	context.Context,
	capmodel.CapabilityContext,
	[]capabilityplugincap.PluginID,
) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	return &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items: map[capabilityplugincap.PluginID]*capabilityplugincap.Projection{},
	}, nil
}

// Current returns a deterministic current plugin projection.
func (*capabilityHostServicePluginsService) Current(
	context.Context,
	capmodel.CapabilityContext,
) (*capabilityplugincap.Projection, error) {
	return &capabilityplugincap.Projection{ID: "test-plugin", Installed: true, Enabled: true}, nil
}

// Search returns an empty fake plugin projection page.
func (*capabilityHostServicePluginsService) Search(
	context.Context,
	capmodel.CapabilityContext,
	capabilityplugincap.SearchInput,
) (*capmodel.PageResult[*capabilityplugincap.Projection], error) {
	return &capmodel.PageResult[*capabilityplugincap.Projection]{Items: []*capabilityplugincap.Projection{}}, nil
}

// ListTenantPlugins returns an empty fake tenant plugin page.
func (*capabilityHostServicePluginsService) ListTenantPlugins(
	context.Context,
	capmodel.CapabilityContext,
	capabilityplugincap.TenantListInput,
) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
}

// BatchGetCapabilityStatus returns an empty fake framework capability status batch.
func (*capabilityHostServicePluginsService) BatchGetCapabilityStatus(
	context.Context,
	capmodel.CapabilityContext,
	[]capabilityplugincap.CapabilityKey,
) (*capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey], error) {
	return &capmodel.BatchResult[*capmodel.CapabilityStatus, capabilityplugincap.CapabilityKey]{
		Items: map[capabilityplugincap.CapabilityKey]*capmodel.CapabilityStatus{},
	}, nil
}

// capabilityHostServicePluginLifecycle records plugin lifecycle governance calls.
type capabilityHostServicePluginLifecycle struct {
	ensurePluginID       string
	ensurePluginTenantID int
	notifyPluginID       string
	notifyPluginTenantID int
	ensureTenantDeleteID int
	notifyTenantDeleteID int
}

// EnsureTenantPluginDisableAllowed records tenant-plugin disable preconditions.
func (s *capabilityHostServicePluginLifecycle) EnsureTenantPluginDisableAllowed(_ context.Context, pluginID string, tenantID int) error {
	s.ensurePluginID = pluginID
	s.ensurePluginTenantID = tenantID
	return nil
}

// NotifyTenantPluginDisabled records tenant-plugin disable notifications.
func (s *capabilityHostServicePluginLifecycle) NotifyTenantPluginDisabled(_ context.Context, pluginID string, tenantID int) {
	s.notifyPluginID = pluginID
	s.notifyPluginTenantID = tenantID
}

// EnsureTenantDeleteAllowed records tenant delete preconditions.
func (s *capabilityHostServicePluginLifecycle) EnsureTenantDeleteAllowed(_ context.Context, tenantID int) error {
	s.ensureTenantDeleteID = tenantID
	return nil
}

// NotifyTenantDeleted records tenant deleted notifications.
func (s *capabilityHostServicePluginLifecycle) NotifyTenantDeleted(_ context.Context, tenantID int) {
	s.notifyTenantDeleteID = tenantID
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
	status := s.Status(ctx)
	available := method == aicommon.CapabilityMethodTextGenerate
	reason := ""
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
	users         map[capabilityusercap.UserID]*capabilityusercap.UserProjection
	ensureErr     error
	lastCapCtx    capmodel.CapabilityContext
	lastSearch    capabilityusercap.SearchInput
	lastEnsureIDs []capabilityusercap.UserID
	lastResolve   capabilityusercap.BatchResolveInput
}

// Current returns the projection for the current actor user.
func (s *capabilityHostServiceUsersService) Current(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
) (*capabilityusercap.UserProjection, error) {
	s.lastCapCtx = capCtx
	return s.users[capabilityusercap.UserID(fmt.Sprint(capCtx.Actor.UserID))], nil
}

// BatchGet returns configured user projections and opaque missing IDs.
func (s *capabilityHostServiceUsersService) BatchGet(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	ids []capabilityusercap.UserID,
) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	s.lastCapCtx = capCtx
	result := &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserProjection{},
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

// BatchResolve records user resolve input and returns deterministic projections.
func (s *capabilityHostServiceUsersService) BatchResolve(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilityusercap.BatchResolveInput,
) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey], error) {
	s.lastCapCtx = capCtx
	s.lastResolve = input
	result := &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey]{
		Items:      map[capabilityusercap.ResolveKey]*capabilityusercap.UserProjection{},
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

// Search returns configured users as a deterministic bounded page.
func (s *capabilityHostServiceUsersService) Search(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilityusercap.SearchInput,
) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	s.lastCapCtx = capCtx
	s.lastSearch = input
	items := make([]*capabilityusercap.UserProjection, 0, len(s.users))
	for _, user := range s.users {
		items = append(items, user)
	}
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: items, Total: len(items)}, nil
}

// EnsureVisible records visibility-check user IDs.
func (s *capabilityHostServiceUsersService) EnsureVisible(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	ids []capabilityusercap.UserID,
) error {
	s.lastCapCtx = capCtx
	s.lastEnsureIDs = append([]capabilityusercap.UserID(nil), ids...)
	return s.ensureErr
}

// capabilityHostServiceSessionsService records online-session requests in tests.
type capabilityHostServiceSessionsService struct {
	sessions          map[capabilitysessioncap.SessionID]*capabilitysessioncap.Projection
	lastCapCtx        capmodel.CapabilityContext
	lastSearch        capabilitysessioncap.SearchInput
	lastOnlineUserIDs []string
	lastEnsureIDs     []capabilitysessioncap.SessionID
}

// Current returns the session projection matching the current identity token.
func (s *capabilityHostServiceSessionsService) Current(
	ctx context.Context,
	capCtx capmodel.CapabilityContext,
) (*capabilitysessioncap.Projection, error) {
	s.lastCapCtx = capCtx
	tokenID := capabilitysessioncap.SessionID(bizctxcap.CurrentFromContext(ctx).TokenID)
	return s.sessions[tokenID], nil
}

// Search returns configured sessions as a deterministic bounded page.
func (s *capabilityHostServiceSessionsService) Search(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilitysessioncap.SearchInput,
) (*capmodel.PageResult[*capabilitysessioncap.Projection], error) {
	s.lastCapCtx = capCtx
	s.lastSearch = input
	items := make([]*capabilitysessioncap.Projection, 0, len(s.sessions))
	for _, sessionItem := range s.sessions {
		items = append(items, sessionItem)
	}
	return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: items, Total: len(items)}, nil
}

// BatchGet returns configured session projections and opaque missing IDs.
func (s *capabilityHostServiceSessionsService) BatchGet(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	ids []capabilitysessioncap.SessionID,
) (*capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID], error) {
	s.lastCapCtx = capCtx
	result := &capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID]{
		Items:      map[capabilitysessioncap.SessionID]*capabilitysessioncap.Projection{},
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
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	userIDs []string,
) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string], error) {
	s.lastCapCtx = capCtx
	s.lastOnlineUserIDs = append([]string(nil), userIDs...)
	result := &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string]{
		Items:      map[string]*capabilitysessioncap.UserOnlineStatusProjection{},
		MissingIDs: []string{},
	}
	for _, userID := range userIDs {
		count := 0
		for _, sessionItem := range s.sessions {
			if sessionItem != nil && sessionItem.UserID == userID {
				count++
			}
		}
		result.Items[userID] = &capabilitysessioncap.UserOnlineStatusProjection{
			UserID:       userID,
			Online:       count > 0,
			SessionCount: count,
		}
	}
	return result, nil
}

// EnsureVisible records requested session IDs.
func (s *capabilityHostServiceSessionsService) EnsureVisible(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	ids []capabilitysessioncap.SessionID,
) error {
	s.lastCapCtx = capCtx
	s.lastEnsureIDs = append([]capabilitysessioncap.SessionID(nil), ids...)
	return nil
}

// capabilityHostServiceAuthzService records authz-domain requests in tests.
type capabilityHostServiceAuthzService struct {
	lastCapCtx capmodel.CapabilityContext
}

// BatchGetPermissions returns deterministic permission projections.
func (s *capabilityHostServiceAuthzService) BatchGetPermissions(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	keys []capabilityauthz.PermissionKey,
) (*capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey], error) {
	s.lastCapCtx = capCtx
	result := &capmodel.BatchResult[*capabilityauthz.PermissionProjection, capabilityauthz.PermissionKey]{
		Items:      map[capabilityauthz.PermissionKey]*capabilityauthz.PermissionProjection{},
		MissingIDs: []capabilityauthz.PermissionKey{},
	}
	for _, key := range keys {
		if key == "missing" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &capabilityauthz.PermissionProjection{
			Key:      key,
			LabelKey: "permissions." + string(key),
			Label:    "Permission " + string(key),
		}
	}
	return result, nil
}

// BatchHasPermissions reports true for permissions present in the capability context.
func (s *capabilityHostServiceAuthzService) BatchHasPermissions(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	keys []capabilityauthz.PermissionKey,
) (map[capabilityauthz.PermissionKey]bool, error) {
	s.lastCapCtx = capCtx
	granted := map[string]struct{}{}
	for _, permission := range capCtx.Authorization.Permissions {
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
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	_ capabilityauthz.PermissionKey,
) (bool, error) {
	s.lastCapCtx = capCtx
	return true, nil
}

// IsPlatformAdmin reports false for deterministic tests.
func (s *capabilityHostServiceAuthzService) IsPlatformAdmin(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	_ capabilityauthz.UserID,
) (bool, error) {
	s.lastCapCtx = capCtx
	return false, nil
}

// capabilityHostServiceDictService records dictionary-domain requests in tests.
type capabilityHostServiceDictService struct {
	lastCapCtx      capmodel.CapabilityContext
	lastInput       capabilitydictcap.ResolveInput
	lastListInput   capabilitydictcap.ListValuesInput
	denyBlankValues bool
}

// ResolveLabels returns deterministic dictionary label projections.
func (s *capabilityHostServiceDictService) ResolveLabels(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilitydictcap.ResolveInput,
) (*capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value], error) {
	s.lastCapCtx = capCtx
	s.lastInput = input
	result := &capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value]{
		Items:      map[capabilitydictcap.Value]*capabilitydictcap.LabelProjection{},
		MissingIDs: []capabilitydictcap.Value{},
	}
	for _, value := range input.Values {
		result.Items[value] = &capabilitydictcap.LabelProjection{
			Type:     input.Type,
			Value:    value,
			LabelKey: "dict." + string(input.Type) + "." + string(value),
			Label:    "Enabled",
		}
	}
	return result, nil
}

// ListValues returns deterministic dictionary value candidates.
func (s *capabilityHostServiceDictService) ListValues(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilitydictcap.ListValuesInput,
) (*capmodel.PageResult[*capabilitydictcap.LabelProjection], error) {
	s.lastCapCtx = capCtx
	s.lastListInput = input
	return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{
		Items: []*capabilitydictcap.LabelProjection{{
			Type:     input.Type,
			Value:    "enabled",
			LabelKey: "dict." + string(input.Type) + ".enabled",
			Label:    "Enabled",
		}},
		Total: 1,
	}, nil
}

// EnsureValuesVisible records dictionary visibility-check input.
func (s *capabilityHostServiceDictService) EnsureValuesVisible(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilitydictcap.ResolveInput,
) error {
	s.lastCapCtx = capCtx
	s.lastInput = input
	if s.denyBlankValues {
		for _, value := range input.Values {
			if strings.TrimSpace(string(value)) == "" {
				return bizerr.NewCode(capmodel.CodeCapabilityDenied)
			}
		}
	}
	return nil
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
func (capabilityHostServiceOrgProvider) SearchDepartments(context.Context, orgcap.DeptSearchInput) (*capmodel.PageResult[*orgcap.DeptProjection], error) {
	return &capmodel.PageResult[*orgcap.DeptProjection]{Items: []*orgcap.DeptProjection{}}, nil
}

// ListPostOptionsPage returns an empty post candidate page.
func (capabilityHostServiceOrgProvider) ListPostOptionsPage(context.Context, orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
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

// OrgProviderEnv returns an empty typed provider environment in host-service tests.
func (capabilityHostServiceOrgRuntime) OrgProviderEnv(string) orgspi.ProviderEnv {
	return orgspi.ProviderEnv{}
}

// capabilityHostServiceTenantService records tenant method calls for tests.
type capabilityHostServiceTenantService struct {
	tenants          []tenantcap.TenantInfo
	lastUserID       int
	lastSwitchUserID int
	lastSwitchTarget tenantcap.TenantID
}

// Available reports that the tenant service is active.
func (*capabilityHostServiceTenantService) Available(context.Context) bool { return true }

// Status returns an active tenant capability status.
func (*capabilityHostServiceTenantService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{Available: true, ActiveProvider: "test-tenant-provider"}
}

// Current returns a deterministic current tenant.
func (*capabilityHostServiceTenantService) Current(context.Context) tenantcap.TenantID { return 3 }

// CurrentTenantInfo returns the deterministic current tenant projection.
func (*capabilityHostServiceTenantService) CurrentTenantInfo(context.Context) (*tenantcap.TenantInfo, error) {
	return &tenantcap.TenantInfo{ID: 3, Code: "tenant-a", Name: "Tenant A", Status: "active"}, nil
}

// PlatformBypass reports no bypass.
func (*capabilityHostServiceTenantService) PlatformBypass(context.Context) bool { return false }

// EnsureTenantVisible accepts all tenant identifiers.
func (*capabilityHostServiceTenantService) EnsureTenantVisible(context.Context, tenantcap.TenantID) error {
	return nil
}

// ValidateUserInTenant accepts all user and tenant pairs.
func (*capabilityHostServiceTenantService) ValidateUserInTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// BatchGetTenants returns deterministic visible tenant projections.
func (s *capabilityHostServiceTenantService) BatchGetTenants(_ context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
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

// SearchTenants returns the configured tenant page.
func (s *capabilityHostServiceTenantService) SearchTenants(context.Context, tenantcap.SearchInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	items := make([]*tenantcap.TenantInfo, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		value := tenant
		items = append(items, &value)
	}
	return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: items, Total: len(items)}, nil
}

// ListUserTenants returns configured tenants and records the requested user.
func (s *capabilityHostServiceTenantService) ListUserTenants(
	_ context.Context,
	userID int,
) ([]tenantcap.TenantInfo, error) {
	s.lastUserID = userID
	return s.tenants, nil
}

// BatchListUserTenants returns configured tenants for each requested user.
func (s *capabilityHostServiceTenantService) BatchListUserTenants(_ context.Context, userIDs []int) (map[int][]tenantcap.TenantInfo, error) {
	result := make(map[int][]tenantcap.TenantInfo, len(userIDs))
	for _, userID := range userIDs {
		result[userID] = append([]tenantcap.TenantInfo(nil), s.tenants...)
	}
	return result, nil
}

// EnsureTenantsVisible accepts all tenant identifiers.
func (*capabilityHostServiceTenantService) EnsureTenantsVisible(context.Context, []tenantcap.TenantID) error {
	return nil
}

// SwitchTenant records the requested tenant switch.
func (s *capabilityHostServiceTenantService) SwitchTenant(
	_ context.Context,
	userID int,
	target tenantcap.TenantID,
) error {
	s.lastSwitchUserID = userID
	s.lastSwitchTarget = target
	return nil
}
