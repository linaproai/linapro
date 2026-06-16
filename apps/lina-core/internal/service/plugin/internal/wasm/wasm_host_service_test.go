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

	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
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

// Sessions returns no online-session domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Sessions() capabilitysessioncap.Service { return nil }

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
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != "42" || userSvc.lastSearch.Keyword != "adm" {
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
					protocol.HostServiceMethodUsersBatchGet,
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
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceAI,
				Methods: []string{
					protocol.HostServiceMethodAITextGenerate,
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
				},
			},
			{
				Service: protocol.HostServiceDict,
				Methods: []string{
					protocol.HostServiceMethodDictResolveLabels,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 9,
			UserID:   21,
			Username: "domain-user",
		},
		requestID: "trace-domain",
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

// ListTenantPlugins returns an empty fake tenant plugin page.
func (*capabilityHostServicePluginsService) ListTenantPlugins(
	context.Context,
	capmodel.CapabilityContext,
) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: []*capabilityplugincap.TenantProjection{}}, nil
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
	lastCapCtx capmodel.CapabilityContext
	lastInput  capabilitydictcap.ResolveInput
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

// ListUserTenants returns configured tenants and records the requested user.
func (s *capabilityHostServiceTenantService) ListUserTenants(
	_ context.Context,
	userID int,
) ([]tenantcap.TenantInfo, error) {
	s.lastUserID = userID
	return s.tenants, nil
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
