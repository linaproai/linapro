// This file tests dynamic-plugin organization and tenant host-service
// dispatch through the unified capability services.

package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability"
	capabilityai "lina-core/pkg/plugin/capability/aicap"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityconfigcap "lina-core/pkg/plugin/capability/configcap"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// capabilityHostServiceTestServices is a narrow capability service set used by
// org and tenant host-service tests.
type capabilityHostServiceTestServices struct {
	org        orgcap.Service
	aiText     aitext.Service
	tenant     tenantcap.Service
	lastPlugin string
}

// Ensure capabilityHostServiceTestServices implements the contracts needed by
// org and tenant host-service configuration.
var _ capability.Services = (*capabilityHostServiceTestServices)(nil)
var _ capability.ScopedServicesFactory = (*capabilityHostServiceTestServices)(nil)

// APIDoc returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) APIDoc() apidoccap.Service { return nil }

// Auth returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) Auth() authcap.Service { return nil }

// AI returns the configured AI capability namespace.
func (s *capabilityHostServiceTestServices) AI() capabilityai.Service {
	text := s.aiText
	if recorder, ok := text.(*capabilityHostServiceAITextService); ok && s.lastPlugin != "" {
		text = &capabilityHostServiceScopedAITextService{
			base:           recorder,
			sourcePluginID: s.lastPlugin,
		}
	}
	return capabilityai.New(text)
}

// Users returns no user-domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Users() capabilityusercap.Service { return nil }

// BizCtx returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) BizCtx() bizctxcap.Service { return nil }

// Cache returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) Cache() cachecap.Service { return nil }

// PluginConfig returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) PluginConfig() plugincap.ConfigService { return nil }

// Config returns no runtime-config domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Config() capabilityconfigcap.Service { return nil }

// Dict returns no dictionary-domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Dict() capabilitydictcap.Service { return nil }

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

// Manifest returns no adapter for capability host-service tests.
func (*capabilityHostServiceTestServices) Manifest() manifestcap.Service { return nil }

// Notifications returns no notification-domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Notifications() capabilitynotifycap.Service { return nil }

// Org returns the configured organization capability service.
func (s *capabilityHostServiceTestServices) Org() orgcap.Service { return s.org }

// Plugins returns no plugin-governance domain service for capability host-service tests.
func (*capabilityHostServiceTestServices) Plugins() capabilityplugincap.Service { return nil }

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

// Tenant returns the configured tenant capability service.
func (s *capabilityHostServiceTestServices) Tenant() tenantcap.Service { return s.tenant }

// ForPlugin records the requested plugin scope and returns the same directory.
func (s *capabilityHostServiceTestServices) ForPlugin(pluginID string) capability.Services {
	s.lastPlugin = pluginID
	return s
}

// TestHandleHostServiceInvokeOrgMethods verifies organization host-service
// calls are routed through capability.Services.Org.
func TestHandleHostServiceInvokeOrgMethods(t *testing.T) {
	providerPluginID := fmt.Sprintf("plugin-test-org-provider-%d", time.Now().UnixNano())
	if err := orgcap.Provide(providerPluginID, func(context.Context, orgcap.ProviderEnv) (orgcap.Provider, error) {
		return capabilityHostServiceOrgProvider{}, nil
	}); err != nil {
		t.Fatalf("register org provider failed: %v", err)
	}

	services := &capabilityHostServiceTestServices{
		org:    orgcap.New(capabilityHostServiceOrgRuntime{pluginID: providerPluginID}),
		aiText: aitext.New(nil),
		tenant: tenantcap.New(nil, nil),
	}
	previous := orgHostServices
	if err := ConfigureOrgHostService(services); err != nil {
		t.Fatalf("configure org host service failed: %v", err)
	}
	t.Cleanup(func() { orgHostServices = previous })

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
		protocol.MarshalHostServiceCapabilityUsersRequest(
			&protocol.HostServiceCapabilityUsersRequest{UserIDs: []int{7, 8}},
		),
	)
	if assignmentsResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected org assignment success, got status=%d payload=%s", assignmentsResponse.Status, string(assignmentsResponse.Payload))
	}
	var assignments map[int]*orgcap.UserDeptAssignment
	decodeCapabilityJSONResponse(t, assignmentsResponse.Payload, &assignments)
	if assignments[7] == nil || assignments[7].DeptID != 17 || assignments[8] == nil || assignments[8].DeptName != "Dept-8" {
		t.Fatalf("unexpected assignment payload: %#v", assignments)
	}
	if services.lastPlugin != "test-capability-plugin" {
		t.Fatalf("expected plugin-scoped services, got %q", services.lastPlugin)
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
		org:    orgcap.New(nil),
		aiText: aitext.New(nil),
		tenant: tenantSvc,
	}
	previous := tenantHostServices
	if err := ConfigureTenantHostService(services); err != nil {
		t.Fatalf("configure tenant host service failed: %v", err)
	}
	t.Cleanup(func() { tenantHostServices = previous })

	response := invokeCapabilityHostService(
		t,
		orgTenantHostCallContext(),
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantListUserTenants,
		protocol.MarshalHostServiceCapabilityUserRequest(&protocol.HostServiceCapabilityUserRequest{UserID: 42}),
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
		protocol.MarshalHostServiceCapabilityUserTenantSwitchRequest(
			&protocol.HostServiceCapabilityUserTenantSwitchRequest{UserID: 42, TargetTenantID: 3},
		),
	)
	if switchResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected tenant switch success, got status=%d payload=%s", switchResponse.Status, string(switchResponse.Payload))
	}
	if tenantSvc.lastSwitchUserID != 42 || tenantSvc.lastSwitchTarget != 3 {
		t.Fatalf("unexpected switch call user=%d target=%d", tenantSvc.lastSwitchUserID, tenantSvc.lastSwitchTarget)
	}
}

// TestHandleHostServiceInvokeAITextGenerate verifies AI host-service calls are
// purpose-authorized and routed through capability.Services.AIText.
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
		org:    orgcap.New(nil),
		aiText: aiSvc,
		tenant: tenantcap.New(nil, nil),
	}
	previous := aiTextHostServices
	if err := ConfigureAITextHostService(services); err != nil {
		t.Fatalf("configure ai text host service failed: %v", err)
	}
	t.Cleanup(func() { aiTextHostServices = previous })

	response := invokeCapabilityHostServiceWithResource(
		t,
		aiTextHostCallContext("purpose:content.summary", "1024"),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		"purpose:content.summary",
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
}

// TestHandleHostServiceInvokeAITextRejectsUnauthorizedPurpose verifies the
// handler rejects calls before touching the text AI service when purpose is not authorized.
func TestHandleHostServiceInvokeAITextRejectsUnauthorizedPurpose(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgcap.New(nil),
		aiText: aiSvc,
		tenant: tenantcap.New(nil, nil),
	}
	previous := aiTextHostServices
	if err := ConfigureAITextHostService(services); err != nil {
		t.Fatalf("configure ai text host service failed: %v", err)
	}
	t.Cleanup(func() { aiTextHostServices = previous })

	response := invokeCapabilityHostServiceWithResource(
		t,
		aiTextHostCallContext("purpose:content.summary", "256"),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		"purpose:content.summary",
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
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if aiSvc.called {
		t.Fatal("expected unauthorized purpose to be rejected before AI service call")
	}
}

// TestHandleHostServiceInvokeAITextRejectsOutputLimit verifies purpose policy
// attributes constrain requests before provider dispatch.
func TestHandleHostServiceInvokeAITextRejectsOutputLimit(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgcap.New(nil),
		aiText: aiSvc,
		tenant: tenantcap.New(nil, nil),
	}
	previous := aiTextHostServices
	if err := ConfigureAITextHostService(services); err != nil {
		t.Fatalf("configure ai text host service failed: %v", err)
	}
	t.Cleanup(func() { aiTextHostServices = previous })

	response := invokeCapabilityHostServiceWithResource(
		t,
		aiTextHostCallContext("purpose:content.summary", "64"),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		"purpose:content.summary",
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
	if response.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if aiSvc.called {
		t.Fatal("expected maxOutputTokens policy to reject before AI service call")
	}
}

// TestHandleHostServiceInvokeAITextAppliesDefaultOutputLimit verifies omitted
// maxOutputTokens cannot bypass the authorized purpose policy.
func TestHandleHostServiceInvokeAITextAppliesDefaultOutputLimit(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{}
	services := &capabilityHostServiceTestServices{
		org:    orgcap.New(nil),
		aiText: aiSvc,
		tenant: tenantcap.New(nil, nil),
	}
	previous := aiTextHostServices
	if err := ConfigureAITextHostService(services); err != nil {
		t.Fatalf("configure ai text host service failed: %v", err)
	}
	t.Cleanup(func() { aiTextHostServices = previous })

	response := invokeCapabilityHostServiceWithResource(
		t,
		aiTextHostCallContext("purpose:content.summary", "96"),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		"purpose:content.summary",
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
	if aiSvc.lastRequest.MaxOutputTokens != 96 {
		t.Fatalf("expected omitted maxOutputTokens to use authorized cap, got %d", aiSvc.lastRequest.MaxOutputTokens)
	}
}

// TestHandleHostServiceInvokeAITextRedactsProviderErrors verifies provider
// failures do not leak authorization markers through the host-call response.
func TestHandleHostServiceInvokeAITextRedactsProviderErrors(t *testing.T) {
	aiSvc := &capabilityHostServiceAITextService{
		err: errors.New("provider failed authorization bearer sk-secret with full prompt body"),
	}
	services := &capabilityHostServiceTestServices{
		org:    orgcap.New(nil),
		aiText: aiSvc,
		tenant: tenantcap.New(nil, nil),
	}
	previous := aiTextHostServices
	if err := ConfigureAITextHostService(services); err != nil {
		t.Fatalf("configure ai text host service failed: %v", err)
	}
	t.Cleanup(func() { aiTextHostServices = previous })

	response := invokeCapabilityHostServiceWithResource(
		t,
		aiTextHostCallContext("purpose:content.summary", "256"),
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		"purpose:content.summary",
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

// TestConfigureCapabilityHostServicesRejectNil verifies nil service directories
// fail during startup wiring.
func TestConfigureCapabilityHostServicesRejectNil(t *testing.T) {
	if err := ConfigureOrgHostService(nil); err == nil {
		t.Fatal("expected nil org host service directory to return an error")
	}
	if err := ConfigureTenantHostService(nil); err == nil {
		t.Fatal("expected nil tenant host service directory to return an error")
	}
	if err := ConfigureAITextHostService(nil); err == nil {
		t.Fatal("expected nil ai text host service directory to return an error")
	}
}

// invokeCapabilityHostService dispatches one organization or tenant host-service request.
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
	return handleHostServiceInvoke(context.Background(), hcc, protocol.MarshalHostServiceRequestEnvelope(request))
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
	return handleHostServiceInvoke(context.Background(), hcc, protocol.MarshalHostServiceRequestEnvelope(request))
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
	}
}

// aiTextHostCallContext builds an authorized AI text host service context.
func aiTextHostCallContext(resourceRef string, maxOutputTokens string) *hostCallContext {
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
				Resources: []*protocol.HostServiceResourceSpec{
					{
						Ref: resourceRef,
						Attributes: map[string]string{
							"defaultTier":     "basic",
							"maxOutputTokens": maxOutputTokens,
						},
					},
				},
			},
		},
	}
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
func (capabilityHostServiceOrgRuntime) OrgProviderEnv(string) orgcap.ProviderEnv {
	return orgcap.ProviderEnv{}
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
