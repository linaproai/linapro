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

	jobv1 "lina-core/api/job/v1"
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
	hostConfig    hostconfigcap.Service
	org           orgcap.Service
	aiText        aitext.Service
	jobs          capabilityjobcap.Service
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

// HostConfig returns the configured host-config capability namespace.
func (s *capabilityHostServiceTestServices) HostConfig() hostconfigcap.Service { return s.hostConfig }

// I18n returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) I18n() i18ncap.Service { return nil }

// Jobs returns the configured scheduled-job domain service.
func (s *capabilityHostServiceTestServices) Jobs() capabilityjobcap.Service { return s.jobs }

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
	provider := &capabilityHostServiceOrgProvider{}
	orgManager := orgspi.NewManager()
	if err := orgManager.RegisterFactory(providerPluginID, func(context.Context, orgspi.ProviderEnv) (orgspi.Provider, error) {
		return provider, nil
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

	deptCreateResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgDepartmentCreate,
		marshalCapabilityJSONRequest(t, orgcap.DeptCreateInput{DeptName: "Engineering", DeptCode: "ENG"}),
	)
	if deptCreateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org department create success, got status=%d payload=%s", deptCreateResponse.Status, string(deptCreateResponse.Payload))
	}
	var deptID int
	decodeCapabilityJSONResponse(t, deptCreateResponse.Payload, &deptID)
	if deptID != 301 || provider.lastDeptCreate.DeptName != "Engineering" {
		t.Fatalf("unexpected org department create result=%d input=%#v", deptID, provider.lastDeptCreate)
	}

	deptName := "Platform"
	deptUpdateResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgDepartmentUpdate,
		marshalCapabilityJSONRequest(t, orgcap.DeptUpdateInput{DeptID: 301, DeptName: &deptName}),
	)
	if deptUpdateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org department update success, got status=%d payload=%s", deptUpdateResponse.Status, string(deptUpdateResponse.Payload))
	}
	if provider.lastDeptUpdate.DeptID != 301 || provider.lastDeptUpdate.DeptName == nil || *provider.lastDeptUpdate.DeptName != deptName {
		t.Fatalf("unexpected org department update input: %#v", provider.lastDeptUpdate)
	}

	deptDeleteResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgDepartmentDelete,
		marshalCapabilityJSONRequest(t, orgDeptIDRequest{DeptID: 301}),
	)
	if deptDeleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org department delete success, got status=%d payload=%s", deptDeleteResponse.Status, string(deptDeleteResponse.Payload))
	}
	if provider.lastDeptDeleteID != 301 {
		t.Fatalf("unexpected org department delete ID: %d", provider.lastDeptDeleteID)
	}

	postCreateResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgPostCreate,
		marshalCapabilityJSONRequest(t, orgcap.PostCreateInput{DeptID: 301, PostCode: "DEV", PostName: "Developer"}),
	)
	if postCreateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org post create success, got status=%d payload=%s", postCreateResponse.Status, string(postCreateResponse.Payload))
	}
	var postID int
	decodeCapabilityJSONResponse(t, postCreateResponse.Payload, &postID)
	if postID != 401 || provider.lastPostCreate.PostName != "Developer" {
		t.Fatalf("unexpected org post create result=%d input=%#v", postID, provider.lastPostCreate)
	}

	postName := "Senior Developer"
	postUpdateResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgPostUpdate,
		marshalCapabilityJSONRequest(t, orgcap.PostUpdateInput{PostID: 401, PostName: &postName}),
	)
	if postUpdateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org post update success, got status=%d payload=%s", postUpdateResponse.Status, string(postUpdateResponse.Payload))
	}
	if provider.lastPostUpdate.PostID != 401 || provider.lastPostUpdate.PostName == nil || *provider.lastPostUpdate.PostName != postName {
		t.Fatalf("unexpected org post update input: %#v", provider.lastPostUpdate)
	}

	postDeleteResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgPostDelete,
		marshalCapabilityJSONRequest(t, orgPostIDRequest{PostID: 401}),
	)
	if postDeleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org post delete success, got status=%d payload=%s", postDeleteResponse.Status, string(postDeleteResponse.Payload))
	}
	if provider.lastPostDeleteID != 401 {
		t.Fatalf("unexpected org post delete ID: %d", provider.lastPostDeleteID)
	}

	assignDeptID := 301
	assignResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgAssignmentReplaceByUser,
		marshalCapabilityJSONRequest(t, orgAssignmentReplaceByUserRequest{
			UserID:  42,
			DeptID:  &assignDeptID,
			PostIDs: []int{401, 402},
		}),
	)
	if assignResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org assignment replace success, got status=%d payload=%s", assignResponse.Status, string(assignResponse.Payload))
	}
	if provider.lastAssignmentUserID != 42 || provider.lastAssignmentDeptID == nil || *provider.lastAssignmentDeptID != 301 ||
		!reflect.DeepEqual(provider.lastAssignmentPostIDs, []int{401, 402}) {
		t.Fatalf("unexpected org assignment replace request user=%d dept=%v posts=%#v", provider.lastAssignmentUserID, provider.lastAssignmentDeptID, provider.lastAssignmentPostIDs)
	}

	cleanupResponse := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgAssignmentCleanupByUser,
		marshalCapabilityJSONRequest(t, intUserIDRequest{UserID: 42}),
	)
	if cleanupResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org assignment cleanup success, got status=%d payload=%s", cleanupResponse.Status, string(cleanupResponse.Payload))
	}
	if provider.lastCleanupUserID != 42 {
		t.Fatalf("unexpected org assignment cleanup user: %d", provider.lastCleanupUserID)
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
		return &capabilityHostServiceOrgProvider{}, nil
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

	createResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersCreate,
		marshalCapabilityJSONRequest(t, capabilityusercap.CreateInput{
			Username: "created",
			Password: "secret",
			Nickname: "Created User",
		}),
	)
	if createResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user create success, got status=%d payload=%s", createResponse.Status, string(createResponse.Payload))
	}
	var createdID capabilityusercap.UserID
	decodeCapabilityJSONResponse(t, createResponse.Payload, &createdID)
	if createdID != "created-user" || userSvc.lastCreate.Username != "created" || userSvc.lastCurrent.UserID != 12 {
		t.Fatalf("unexpected user create result=%q input=%#v bizctx=%#v", createdID, userSvc.lastCreate, userSvc.lastCurrent)
	}

	nickname := "Updated User"
	updateResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersUpdate,
		marshalCapabilityJSONRequest(t, capabilityusercap.UpdateInput{ID: "42", Nickname: &nickname}),
	)
	if updateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user update success, got status=%d payload=%s", updateResponse.Status, string(updateResponse.Payload))
	}
	if userSvc.lastUpdate.ID != "42" || userSvc.lastUpdate.Nickname == nil || *userSvc.lastUpdate.Nickname != nickname {
		t.Fatalf("unexpected user update input: %#v", userSvc.lastUpdate)
	}

	deleteResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersDelete,
		marshalCapabilityJSONRequest(t, userIDRequest{UserID: "42"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user delete success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
	if userSvc.lastDeleteID != "42" {
		t.Fatalf("unexpected user delete ID: %q", userSvc.lastDeleteID)
	}

	statusResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersSetStatus,
		marshalCapabilityJSONRequest(t, usersSetStatusRequest{UserID: "42", Status: int(statusflag.Disabled)}),
	)
	if statusResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user status success, got status=%d payload=%s", statusResponse.Status, string(statusResponse.Payload))
	}
	if userSvc.lastStatusID != "42" || userSvc.lastStatus != statusflag.Disabled {
		t.Fatalf("unexpected user status request id=%q status=%d", userSvc.lastStatusID, userSvc.lastStatus)
	}

	resetResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersResetPassword,
		marshalCapabilityJSONRequest(t, usersResetPasswordRequest{UserID: "42", Password: "new-secret"}),
	)
	if resetResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user password reset success, got status=%d payload=%s", resetResponse.Status, string(resetResponse.Payload))
	}
	if userSvc.lastResetID != "42" || userSvc.lastPassword != "new-secret" {
		t.Fatalf("unexpected user password reset id=%q password=%q", userSvc.lastResetID, userSvc.lastPassword)
	}

	rolesResponse := invokeCapabilityHostService(
		t,
		userHostCallContext(),
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersReplaceRoles,
		marshalCapabilityJSONRequest(t, usersReplaceRolesRequest{UserID: "42", RoleIDs: []int{1, 2}}),
	)
	if rolesResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected user role replacement success, got status=%d payload=%s", rolesResponse.Status, string(rolesResponse.Payload))
	}
	if userSvc.assignments == nil || userSvc.assignments.lastUserID != "42" || !reflect.DeepEqual(userSvc.assignments.lastRoleIDs, []int{1, 2}) {
		t.Fatalf("unexpected role replacement request: %#v", userSvc.assignments)
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
		protocol.HostServiceAuth,
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
		protocol.HostServiceAuth,
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

	authzReplaceResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceAuth,
		protocol.HostServiceMethodAuthzReplaceRolePermissions,
		marshalCapabilityJSONRequest(t, authzReplaceRolePermissionsRequest{
			RoleID: "role-1",
			Keys:   []string{"system:user:list", "system:user:create"},
		}),
	)
	if authzReplaceResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected authz replace success, got status=%d payload=%s", authzReplaceResponse.Status, string(authzReplaceResponse.Payload))
	}
	if authzSvc.lastRoleID != "role-1" || !reflect.DeepEqual(authzSvc.lastPermissionKeys, []capabilityauthz.PermissionKey{"system:user:list", "system:user:create"}) {
		t.Fatalf("unexpected authz role replacement role=%q keys=%#v", authzSvc.lastRoleID, authzSvc.lastPermissionKeys)
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

	dictRefreshResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictRefresh,
		marshalCapabilityJSONRequest(t, dictTypeRequest{Type: "sys_common_status"}),
	)
	if dictRefreshResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict refresh success, got status=%d payload=%s", dictRefreshResponse.Status, string(dictRefreshResponse.Payload))
	}
	if dictSvc.lastRefreshType != "sys_common_status" {
		t.Fatalf("unexpected dict refresh type: %q", dictSvc.lastRefreshType)
	}

	dictTypeCreateResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictTypeCreate,
		marshalCapabilityJSONRequest(t, capabilitydictcap.CreateTypeInput{
			Type:   "plugin_state",
			Name:   "Plugin State",
			Status: statusflag.EnabledValue,
		}),
	)
	if dictTypeCreateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict type create success, got status=%d payload=%s", dictTypeCreateResponse.Status, string(dictTypeCreateResponse.Payload))
	}
	var dictTypeID int
	decodeCapabilityJSONResponse(t, dictTypeCreateResponse.Payload, &dictTypeID)
	if dictTypeID != 101 || dictSvc.lastTypeCreate.Type != "plugin_state" {
		t.Fatalf("unexpected dict type create result=%d input=%#v", dictTypeID, dictSvc.lastTypeCreate)
	}

	dictTypeUpdateResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictTypeUpdate,
		marshalCapabilityJSONRequest(t, capabilitydictcap.UpdateTypeInput{ID: 101}),
	)
	if dictTypeUpdateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict type update success, got status=%d payload=%s", dictTypeUpdateResponse.Status, string(dictTypeUpdateResponse.Payload))
	}
	if dictSvc.lastTypeUpdate.ID != 101 {
		t.Fatalf("unexpected dict type update input: %#v", dictSvc.lastTypeUpdate)
	}

	dictTypeDeleteResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictTypeDelete,
		marshalCapabilityJSONRequest(t, dictIDRequest{ID: 101}),
	)
	if dictTypeDeleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict type delete success, got status=%d payload=%s", dictTypeDeleteResponse.Status, string(dictTypeDeleteResponse.Payload))
	}
	if dictSvc.lastTypeDelete != 101 {
		t.Fatalf("unexpected dict type delete ID: %d", dictSvc.lastTypeDelete)
	}

	dictValueCreateResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueCreate,
		marshalCapabilityJSONRequest(t, capabilitydictcap.CreateValueInput{
			Type:   "plugin_state",
			Value:  "enabled",
			Label:  "Enabled",
			Status: statusflag.EnabledValue,
		}),
	)
	if dictValueCreateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict value create success, got status=%d payload=%s", dictValueCreateResponse.Status, string(dictValueCreateResponse.Payload))
	}
	var dictValueID int
	decodeCapabilityJSONResponse(t, dictValueCreateResponse.Payload, &dictValueID)
	if dictValueID != 202 || dictSvc.lastValueCreate.Value != "enabled" {
		t.Fatalf("unexpected dict value create result=%d input=%#v", dictValueID, dictSvc.lastValueCreate)
	}

	dictValueUpdateResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueUpdate,
		marshalCapabilityJSONRequest(t, capabilitydictcap.UpdateValueInput{ID: 202}),
	)
	if dictValueUpdateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict value update success, got status=%d payload=%s", dictValueUpdateResponse.Status, string(dictValueUpdateResponse.Payload))
	}
	if dictSvc.lastValueUpdate.ID != 202 {
		t.Fatalf("unexpected dict value update input: %#v", dictSvc.lastValueUpdate)
	}

	dictValueDeleteResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueDelete,
		marshalCapabilityJSONRequest(t, dictIDRequest{ID: 202}),
	)
	if dictValueDeleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict value delete success, got status=%d payload=%s", dictValueDeleteResponse.Status, string(dictValueDeleteResponse.Payload))
	}
	if dictSvc.lastValueDelete != 202 {
		t.Fatalf("unexpected dict value delete ID: %d", dictSvc.lastValueDelete)
	}

	dictDeleteByTypeResponse := invokeCapabilityHostService(
		t,
		additionalDomainHostCallContext(),
		protocol.HostServiceDict,
		protocol.HostServiceMethodDictValueDeleteByType,
		marshalCapabilityJSONRequest(t, dictTypeRequest{Type: "plugin_state"}),
	)
	if dictDeleteByTypeResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected dict values delete-by-type success, got status=%d payload=%s", dictDeleteByTypeResponse.Status, string(dictDeleteByTypeResponse.Payload))
	}
	if dictSvc.lastDeleteType != "plugin_state" {
		t.Fatalf("unexpected dict delete-by-type key: %q", dictSvc.lastDeleteType)
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

	renamed := "renamed.txt"
	metadataResponse := invokeCapabilityHostService(
		t,
		filesHostCallContext(),
		protocol.HostServiceFiles,
		protocol.HostServiceMethodFilesUpdateMetadata,
		marshalCapabilityJSONRequest(t, capabilityfilecap.UpdateMetadataInput{
			ID:   "file-from-storage",
			Name: &renamed,
		}),
	)
	if metadataResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected files metadata update success, got status=%d payload=%s", metadataResponse.Status, string(metadataResponse.Payload))
	}
	if filesSvc.lastMetadataInput.ID != "file-from-storage" || filesSvc.lastMetadataInput.Name == nil || *filesSvc.lastMetadataInput.Name != renamed {
		t.Fatalf("unexpected files metadata update input: %#v", filesSvc.lastMetadataInput)
	}

	deleteResponse := invokeCapabilityHostService(
		t,
		filesHostCallContext(),
		protocol.HostServiceFiles,
		protocol.HostServiceMethodFilesDelete,
		marshalCapabilityJSONRequest(t, fileIDRequest{FileID: "file-from-storage"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected files delete success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
	if filesSvc.lastDeleteID != "file-from-storage" {
		t.Fatalf("unexpected files delete ID: %q", filesSvc.lastDeleteID)
	}

	deleteManyResponse := invokeCapabilityHostService(
		t,
		filesHostCallContext(),
		protocol.HostServiceFiles,
		protocol.HostServiceMethodFilesDeleteMany,
		marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{"file-a", "file-b"}}),
	)
	if deleteManyResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected files delete_many success, got status=%d payload=%s", deleteManyResponse.Status, string(deleteManyResponse.Payload))
	}
	if !reflect.DeepEqual(filesSvc.lastDeleteManyIDs, []capabilityfilecap.FileID{"file-a", "file-b"}) {
		t.Fatalf("unexpected files delete_many IDs: %#v", filesSvc.lastDeleteManyIDs)
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

	revokeResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsRevoke,
		marshalCapabilityJSONRequest(t, sessionIDRequest{SessionID: "token-1"}),
	)
	if revokeResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session revoke success, got status=%d payload=%s", revokeResponse.Status, string(revokeResponse.Payload))
	}
	if sessionSvc.lastRevokeID != "token-1" {
		t.Fatalf("unexpected session revoke ID: %q", sessionSvc.lastRevokeID)
	}

	revokeManyResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceSessions,
		protocol.HostServiceMethodSessionsRevokeMany,
		marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{"token-1", "token-2"}}),
	)
	if revokeManyResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected session revoke_many success, got status=%d payload=%s", revokeManyResponse.Status, string(revokeManyResponse.Payload))
	}
	if !reflect.DeepEqual(sessionSvc.lastRevokeManyIDs, []capabilitysessioncap.SessionID{"token-1", "token-2"}) {
		t.Fatalf("unexpected session revoke_many IDs: %#v", sessionSvc.lastRevokeManyIDs)
	}
}

// TestHandleHostServiceInvokeJobsRuntimeMethods verifies runtime scheduled-job
// commands dispatch to the shared job capability service.
func TestHandleHostServiceInvokeJobsRuntimeMethods(t *testing.T) {
	jobsSvc := &capabilityHostServiceJobsService{}
	services := &capabilityHostServiceTestServices{
		org:    orgspi.New(nil, nil, nil),
		aiText: aitext.New(nil, nil, nil),
		jobs:   jobsSvc,
		tenant: tenantspi.New(nil, nil, nil, nil),
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	hcc := jobsHostCallContext()
	createResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsCreate,
		marshalCapabilityJSONRequest(t, capabilityjobcap.SaveInput{
			GroupID:  "default",
			Name:     "Heartbeat",
			CronExpr: "*/5 * * * *",
			Status:   jobv1.StatusEnabled,
			LogRetentionOverride: &capabilityjobcap.LogRetentionOption{
				Mode:  jobv1.RetentionModeCount,
				Value: 500,
			},
		}),
	)
	if createResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs create success, got status=%d payload=%s", createResponse.Status, string(createResponse.Payload))
	}
	var createdID capabilityjobcap.JobID
	decodeCapabilityJSONResponse(t, createResponse.Payload, &createdID)
	if createdID != "job-created" || jobsSvc.lastCreate.Name != "Heartbeat" || jobsSvc.lastCurrent.UserID != 21 {
		t.Fatalf("unexpected jobs create result=%q input=%#v bizctx=%#v", createdID, jobsSvc.lastCreate, jobsSvc.lastCurrent)
	}
	if jobsSvc.lastCreate.LogRetentionOverride == nil ||
		jobsSvc.lastCreate.LogRetentionOverride.Mode != jobv1.RetentionModeCount ||
		jobsSvc.lastCreate.LogRetentionOverride.Value != 500 {
		t.Fatalf("unexpected jobs create retention override: %#v", jobsSvc.lastCreate.LogRetentionOverride)
	}

	jobsSvc.jobItems = []*capabilityjobcap.JobInfo{{
		ID:     "job-created",
		Name:   "Heartbeat",
		Group:  "default",
		Status: jobv1.StatusEnabled,
		LogRetentionOverride: &capabilityjobcap.LogRetentionOption{
			Mode:  jobv1.RetentionModeCount,
			Value: 500,
		},
	}}
	batchGetResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsBatchGet,
		marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{"job-created"}}),
	)
	if batchGetResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs batch_get success, got status=%d payload=%s", batchGetResponse.Status, string(batchGetResponse.Payload))
	}
	var batchGetResult capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID]
	decodeCapabilityJSONResponse(t, batchGetResponse.Payload, &batchGetResult)
	if !reflect.DeepEqual(jobsSvc.lastBatchGetIDs, []capabilityjobcap.JobID{"job-created"}) {
		t.Fatalf("unexpected jobs batch_get IDs: %#v", jobsSvc.lastBatchGetIDs)
	}
	if got := batchGetResult.Items["job-created"]; got == nil ||
		got.LogRetentionOverride == nil ||
		got.LogRetentionOverride.Mode != jobv1.RetentionModeCount ||
		got.LogRetentionOverride.Value != 500 {
		t.Fatalf("unexpected jobs batch_get retention projection: %#v", got)
	}

	listResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsList,
		marshalCapabilityJSONRequest(t, jobsListRequest{Keyword: "Heartbeat", PageNum: 1, PageSize: 20}),
	)
	if listResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs list success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	var listResult capmodel.PageResult[*capabilityjobcap.JobInfo]
	decodeCapabilityJSONResponse(t, listResponse.Payload, &listResult)
	if jobsSvc.lastList.Keyword != "Heartbeat" || jobsSvc.lastList.Page.PageNum != 1 || jobsSvc.lastList.Page.PageSize != 20 {
		t.Fatalf("unexpected jobs list input: %#v", jobsSvc.lastList)
	}
	if len(listResult.Items) != 1 ||
		listResult.Items[0].LogRetentionOverride == nil ||
		listResult.Items[0].LogRetentionOverride.Mode != jobv1.RetentionModeCount ||
		listResult.Items[0].LogRetentionOverride.Value != 500 {
		t.Fatalf("unexpected jobs list retention projection: %#v", listResult.Items)
	}

	updateResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsUpdate,
		marshalCapabilityJSONRequest(t, capabilityjobcap.UpdateInput{
			ID: "job-created",
			SaveInput: capabilityjobcap.SaveInput{
				Name: "Updated Heartbeat",
				LogRetentionOverride: &capabilityjobcap.LogRetentionOption{
					Mode: jobv1.RetentionModeNone,
				},
			},
		}),
	)
	if updateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs update success, got status=%d payload=%s", updateResponse.Status, string(updateResponse.Payload))
	}
	if jobsSvc.lastUpdate.ID != "job-created" || jobsSvc.lastUpdate.Name != "Updated Heartbeat" {
		t.Fatalf("unexpected jobs update input: %#v", jobsSvc.lastUpdate)
	}
	if jobsSvc.lastUpdate.LogRetentionOverride == nil ||
		jobsSvc.lastUpdate.LogRetentionOverride.Mode != jobv1.RetentionModeNone ||
		jobsSvc.lastUpdate.LogRetentionOverride.Value != 0 {
		t.Fatalf("unexpected jobs update retention override: %#v", jobsSvc.lastUpdate.LogRetentionOverride)
	}

	deleteResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsDelete,
		marshalCapabilityJSONRequest(t, jobIDRequest{JobID: "job-created"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs delete success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
	if jobsSvc.lastDeleteID != "job-created" {
		t.Fatalf("unexpected jobs delete ID: %q", jobsSvc.lastDeleteID)
	}

	runResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsRun,
		marshalCapabilityJSONRequest(t, jobIDRequest{JobID: "job-created"}),
	)
	if runResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs run success, got status=%d payload=%s", runResponse.Status, string(runResponse.Payload))
	}
	if jobsSvc.lastRunID != "job-created" {
		t.Fatalf("unexpected jobs run ID: %q", jobsSvc.lastRunID)
	}

	statusResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsSetStatus,
		marshalCapabilityJSONRequest(t, jobsSetStatusRequest{JobID: "job-created", Status: string(jobv1.StatusDisabled)}),
	)
	if statusResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs status success, got status=%d payload=%s", statusResponse.Status, string(statusResponse.Payload))
	}
	if jobsSvc.lastStatusID != "job-created" || jobsSvc.lastStatus != jobv1.StatusDisabled {
		t.Fatalf("unexpected jobs status id=%q status=%q", jobsSvc.lastStatusID, jobsSvc.lastStatus)
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
					protocol.HostServiceMethodOrgDepartmentCreate,
					protocol.HostServiceMethodOrgDepartmentUpdate,
					protocol.HostServiceMethodOrgDepartmentDelete,
					protocol.HostServiceMethodOrgPostCreate,
					protocol.HostServiceMethodOrgPostUpdate,
					protocol.HostServiceMethodOrgPostDelete,
					protocol.HostServiceMethodOrgAssignmentReplaceByUser,
					protocol.HostServiceMethodOrgAssignmentCleanupByUser,
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
					protocol.HostServiceMethodUsersCreate,
					protocol.HostServiceMethodUsersUpdate,
					protocol.HostServiceMethodUsersDelete,
					protocol.HostServiceMethodUsersSetStatus,
					protocol.HostServiceMethodUsersResetPassword,
					protocol.HostServiceMethodUsersReplaceRoles,
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
				Service: protocol.HostServiceAuth,
				Methods: []string{
					protocol.HostServiceMethodAuthzBatchGetPermissions,
					protocol.HostServiceMethodAuthzBatchHasPermissions,
					protocol.HostServiceMethodAuthzReplaceRolePermissions,
				},
			},
			{
				Service: protocol.HostServiceDict,
				Methods: []string{
					protocol.HostServiceMethodDictRefresh,
					protocol.HostServiceMethodDictTypeCreate,
					protocol.HostServiceMethodDictTypeUpdate,
					protocol.HostServiceMethodDictTypeDelete,
					protocol.HostServiceMethodDictValueResolveLabels,
					protocol.HostServiceMethodDictListValues,
					protocol.HostServiceMethodDictValueEnsureValuesVisible,
					protocol.HostServiceMethodDictValueCreate,
					protocol.HostServiceMethodDictValueUpdate,
					protocol.HostServiceMethodDictValueDelete,
					protocol.HostServiceMethodDictValueDeleteByType,
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
					protocol.HostServiceMethodFilesUpdateMetadata,
					protocol.HostServiceMethodFilesDelete,
					protocol.HostServiceMethodFilesDeleteMany,
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
					protocol.HostServiceMethodSessionsRevoke,
					protocol.HostServiceMethodSessionsRevokeMany,
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

// jobsHostCallContext builds an authorized scheduled-job host service context.
func jobsHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-jobs-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityJobs: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceJobs,
				Methods: []string{
					protocol.HostServiceMethodJobsBatchGet,
					protocol.HostServiceMethodJobsList,
					protocol.HostServiceMethodJobsCreate,
					protocol.HostServiceMethodJobsUpdate,
					protocol.HostServiceMethodJobsDelete,
					protocol.HostServiceMethodJobsRun,
					protocol.HostServiceMethodJobsSetStatus,
				},
			},
		},
		identity: &bridgecontract.IdentitySnapshotV1{
			TenantId: 9,
			UserID:   21,
			Username: "jobs-user",
		},
		requestID: "trace-jobs",
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
	lastCreate    capabilityusercap.CreateInput
	lastUpdate    capabilityusercap.UpdateInput
	lastDeleteID  capabilityusercap.UserID
	lastStatusID  capabilityusercap.UserID
	lastStatus    statusflag.Enabled
	lastResetID   capabilityusercap.UserID
	lastPassword  string
	assignments   *capabilityHostServiceUserAssignments
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

// Create records one user create request.
func (s *capabilityHostServiceUsersService) Create(ctx context.Context, input capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastCreate = input
	return capabilityusercap.UserID("created-user"), nil
}

// Update records one user update request.
func (s *capabilityHostServiceUsersService) Update(ctx context.Context, input capabilityusercap.UpdateInput) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastUpdate = input
	return nil
}

// Delete records one user delete request.
func (s *capabilityHostServiceUsersService) Delete(ctx context.Context, id capabilityusercap.UserID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastDeleteID = id
	return nil
}

// SetStatus records one user status update request.
func (s *capabilityHostServiceUsersService) SetStatus(ctx context.Context, id capabilityusercap.UserID, status statusflag.Enabled) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastStatusID = id
	s.lastStatus = status
	return nil
}

// ResetPassword records one password reset request.
func (s *capabilityHostServiceUsersService) ResetPassword(ctx context.Context, id capabilityusercap.UserID, password string) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastResetID = id
	s.lastPassword = password
	return nil
}

// Assignment returns user-role assignment operations.
func (s *capabilityHostServiceUsersService) Assignment() capabilityusercap.AssignmentService {
	if s.assignments == nil {
		s.assignments = &capabilityHostServiceUserAssignments{}
	}
	return s.assignments
}

// capabilityHostServiceUserAssignments records role replacements.
type capabilityHostServiceUserAssignments struct {
	lastUserID  capabilityusercap.UserID
	lastRoleIDs []int
}

// ReplaceRoles records one role replacement request.
func (s *capabilityHostServiceUserAssignments) ReplaceRoles(_ context.Context, id capabilityusercap.UserID, roleIDs []int) error {
	s.lastUserID = id
	s.lastRoleIDs = append([]int(nil), roleIDs...)
	return nil
}

// capabilityHostServiceSessionsService records online-session requests in tests.
type capabilityHostServiceSessionsService struct {
	sessions          map[capabilitysessioncap.SessionID]*capabilitysessioncap.SessionInfo
	lastCurrent       bizctxcap.CurrentContext
	lastList          capabilitysessioncap.ListInput
	lastOnlineUserIDs []string
	lastEnsureIDs     []capabilitysessioncap.SessionID
	lastRevokeID      capabilitysessioncap.SessionID
	lastRevokeManyIDs []capabilitysessioncap.SessionID
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

// Revoke records one session revoke request.
func (s *capabilityHostServiceSessionsService) Revoke(ctx context.Context, id capabilitysessioncap.SessionID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastRevokeID = id
	return nil
}

// RevokeMany records one batch session revoke request.
func (s *capabilityHostServiceSessionsService) RevokeMany(ctx context.Context, ids []capabilitysessioncap.SessionID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastRevokeManyIDs = append([]capabilitysessioncap.SessionID(nil), ids...)
	return nil
}

// capabilityHostServiceAuthzService records authz-domain requests in tests.
type capabilityHostServiceAuthzService struct {
	lastCurrent        bizctxcap.CurrentContext
	lastRoleID         capabilityauthz.RoleID
	lastPermissionKeys []capabilityauthz.PermissionKey
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

// ReplaceRolePermissions records one role permission replacement.
func (s *capabilityHostServiceAuthzService) ReplaceRolePermissions(ctx context.Context, roleID capabilityauthz.RoleID, keys []capabilityauthz.PermissionKey) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastRoleID = roleID
	s.lastPermissionKeys = append([]capabilityauthz.PermissionKey(nil), keys...)
	return nil
}

// capabilityHostServiceDictService records dictionary-domain requests in tests.
type capabilityHostServiceDictService struct {
	lastCurrent     bizctxcap.CurrentContext
	lastInput       capabilitydictcap.ResolveInput
	lastListInput   capabilitydictcap.ListValuesInput
	lastRefreshType capabilitydictcap.Type
	lastTypeCreate  capabilitydictcap.CreateTypeInput
	lastTypeUpdate  capabilitydictcap.UpdateTypeInput
	lastTypeDelete  int
	lastValueCreate capabilitydictcap.CreateValueInput
	lastValueUpdate capabilitydictcap.UpdateValueInput
	lastValueDelete int
	lastDeleteType  capabilitydictcap.Type
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

// Create records one dictionary type create request.
func (s capabilityHostServiceDictTypeService) Create(ctx context.Context, input capabilitydictcap.CreateTypeInput) (int, error) {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastTypeCreate = input
	return 101, nil
}

// Update records one dictionary type update request.
func (s capabilityHostServiceDictTypeService) Update(ctx context.Context, input capabilitydictcap.UpdateTypeInput) error {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastTypeUpdate = input
	return nil
}

// Delete records one dictionary type delete request.
func (s capabilityHostServiceDictTypeService) Delete(ctx context.Context, id int) error {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastTypeDelete = id
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

// Create records one dictionary value create request.
func (s capabilityHostServiceDictValueService) Create(ctx context.Context, input capabilitydictcap.CreateValueInput) (int, error) {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastValueCreate = input
	return 202, nil
}

// Update records one dictionary value update request.
func (s capabilityHostServiceDictValueService) Update(ctx context.Context, input capabilitydictcap.UpdateValueInput) error {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastValueUpdate = input
	return nil
}

// Delete records one dictionary value delete request.
func (s capabilityHostServiceDictValueService) Delete(ctx context.Context, id int) error {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastValueDelete = id
	return nil
}

// DeleteByType records one dictionary type-wide value delete request.
func (s capabilityHostServiceDictValueService) DeleteByType(ctx context.Context, dictType capabilitydictcap.Type) error {
	s.parent.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.parent.lastDeleteType = dictType
	return nil
}

// Refresh records one dictionary refresh request.
func (s *capabilityHostServiceDictService) Refresh(ctx context.Context, dictType capabilitydictcap.Type) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastRefreshType = dictType
	return nil
}

// capabilityHostServiceFilesService records dynamic file write calls.
type capabilityHostServiceFilesService struct {
	capabilityfilecap.Service
	lastUpload        capabilityfilecap.UploadInput
	lastUploadBody    string
	lastStorageInput  capabilityfilecap.CreateFromStorageInput
	lastMetadataInput capabilityfilecap.UpdateMetadataInput
	lastDeleteID      capabilityfilecap.FileID
	lastDeleteManyIDs []capabilityfilecap.FileID
	lastCurrent       bizctxcap.CurrentContext
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

// UpdateMetadata records one visible file metadata update.
func (s *capabilityHostServiceFilesService) UpdateMetadata(
	ctx context.Context,
	input capabilityfilecap.UpdateMetadataInput,
) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastMetadataInput = input
	return nil
}

// Delete records one visible file delete.
func (s *capabilityHostServiceFilesService) Delete(ctx context.Context, id capabilityfilecap.FileID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastDeleteID = id
	return nil
}

// DeleteMany records one visible file batch delete.
func (s *capabilityHostServiceFilesService) DeleteMany(ctx context.Context, ids []capabilityfilecap.FileID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastDeleteManyIDs = append([]capabilityfilecap.FileID(nil), ids...)
	return nil
}

// capabilityHostServiceJobsService records scheduled-job requests in tests.
type capabilityHostServiceJobsService struct {
	jobItems        []*capabilityjobcap.JobInfo
	lastCurrent     bizctxcap.CurrentContext
	lastBatchGetIDs []capabilityjobcap.JobID
	lastList        capabilityjobcap.ListInput
	lastCreate      capabilityjobcap.SaveInput
	lastUpdate      capabilityjobcap.UpdateInput
	lastDeleteID    capabilityjobcap.JobID
	lastRunID       capabilityjobcap.JobID
	lastStatusID    capabilityjobcap.JobID
	lastStatus      jobv1.Status
}

// Get returns no job projection for dispatcher tests.
func (s *capabilityHostServiceJobsService) Get(ctx context.Context, id capabilityjobcap.JobID) (*capabilityjobcap.JobInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityjobcap.JobID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// BatchGet returns configured visible job projections.
func (s *capabilityHostServiceJobsService) BatchGet(ctx context.Context, ids []capabilityjobcap.JobID) (*capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastBatchGetIDs = append([]capabilityjobcap.JobID(nil), ids...)
	result := &capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID]{
		Items:      map[capabilityjobcap.JobID]*capabilityjobcap.JobInfo{},
		MissingIDs: []capabilityjobcap.JobID{},
	}
	for _, id := range ids {
		found := false
		for _, item := range s.jobItems {
			if item != nil && item.ID == id {
				result.Items[id] = item
				found = true
				break
			}
		}
		if !found {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// List returns configured visible job projections as one deterministic page.
func (s *capabilityHostServiceJobsService) List(ctx context.Context, input capabilityjobcap.ListInput) (*capmodel.PageResult[*capabilityjobcap.JobInfo], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastList = input
	items := append([]*capabilityjobcap.JobInfo(nil), s.jobItems...)
	return &capmodel.PageResult[*capabilityjobcap.JobInfo]{Items: items, Total: len(items)}, nil
}

// EnsureVisible accepts all requested jobs.
func (s *capabilityHostServiceJobsService) EnsureVisible(ctx context.Context, _ []capabilityjobcap.JobID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	return nil
}

// Create records one scheduled-job create request.
func (s *capabilityHostServiceJobsService) Create(ctx context.Context, input capabilityjobcap.SaveInput) (capabilityjobcap.JobID, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastCreate = input
	return capabilityjobcap.JobID("job-created"), nil
}

// Update records one scheduled-job update request.
func (s *capabilityHostServiceJobsService) Update(ctx context.Context, input capabilityjobcap.UpdateInput) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastUpdate = input
	return nil
}

// Delete records one scheduled-job delete request.
func (s *capabilityHostServiceJobsService) Delete(ctx context.Context, id capabilityjobcap.JobID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastDeleteID = id
	return nil
}

// Run records one scheduled-job execution request.
func (s *capabilityHostServiceJobsService) Run(ctx context.Context, id capabilityjobcap.JobID) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastRunID = id
	return nil
}

// SetStatus records one scheduled-job lifecycle status update.
func (s *capabilityHostServiceJobsService) SetStatus(ctx context.Context, id capabilityjobcap.JobID, status jobv1.Status) error {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastStatusID = id
	s.lastStatus = status
	return nil
}

// capabilityHostServiceOrgProvider is a deterministic organization provider for tests.
type capabilityHostServiceOrgProvider struct {
	lastDeptCreate        orgcap.DeptCreateInput
	lastDeptUpdate        orgcap.DeptUpdateInput
	lastDeptDeleteID      int
	lastPostCreate        orgcap.PostCreateInput
	lastPostUpdate        orgcap.PostUpdateInput
	lastPostDeleteID      int
	lastAssignmentUserID  int
	lastAssignmentDeptID  *int
	lastAssignmentPostIDs []int
	lastCleanupUserID     int
}

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

// CreateDepartment records one department create request.
func (p *capabilityHostServiceOrgProvider) CreateDepartment(_ context.Context, input orgcap.DeptCreateInput) (int, error) {
	p.lastDeptCreate = input
	return 301, nil
}

// UpdateDepartment records one department update request.
func (p *capabilityHostServiceOrgProvider) UpdateDepartment(_ context.Context, input orgcap.DeptUpdateInput) error {
	p.lastDeptUpdate = input
	return nil
}

// DeleteDepartment records one department delete request.
func (p *capabilityHostServiceOrgProvider) DeleteDepartment(_ context.Context, deptID int) error {
	p.lastDeptDeleteID = deptID
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

// CreatePost records one post create request.
func (p *capabilityHostServiceOrgProvider) CreatePost(_ context.Context, input orgcap.PostCreateInput) (int, error) {
	p.lastPostCreate = input
	return 401, nil
}

// UpdatePost records one post update request.
func (p *capabilityHostServiceOrgProvider) UpdatePost(_ context.Context, input orgcap.PostUpdateInput) error {
	p.lastPostUpdate = input
	return nil
}

// DeletePost records one post delete request.
func (p *capabilityHostServiceOrgProvider) DeletePost(_ context.Context, postID int) error {
	p.lastPostDeleteID = postID
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

// ReplaceUserAssignments records one organization assignment rewrite.
func (p *capabilityHostServiceOrgProvider) ReplaceUserAssignments(_ context.Context, userID int, deptID *int, postIDs []int) error {
	p.lastAssignmentUserID = userID
	if deptID != nil {
		value := *deptID
		p.lastAssignmentDeptID = &value
	} else {
		p.lastAssignmentDeptID = nil
	}
	p.lastAssignmentPostIDs = append([]int(nil), postIDs...)
	return nil
}

// CleanupUserAssignments records one organization assignment cleanup.
func (p *capabilityHostServiceOrgProvider) CleanupUserAssignments(_ context.Context, userID int) error {
	p.lastCleanupUserID = userID
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
