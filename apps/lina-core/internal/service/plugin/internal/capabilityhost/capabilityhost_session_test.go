// This file verifies session capability conversion, visibility filtering, and
// revocation behavior inside the sessioncap component.

package capabilityhost

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/datascope"
	internalsession "lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestToInternalFilter verifies the published filter contract is converted explicitly.
func TestToInternalFilter(t *testing.T) {
	if result := toInternalFilter(capabilitysessioncap.SearchInput{}); result != nil {
		t.Fatalf("expected nil filter, got %#v", result)
	}

	filter := capabilitysessioncap.SearchInput{
		Username: "admin",
		IP:       "127.0.0.1",
	}
	result := toInternalFilter(filter)
	if result == nil {
		t.Fatal("expected converted filter, got nil")
	}
	if result.Username != "admin" || result.Ip != "127.0.0.1" {
		t.Fatalf("unexpected converted filter: %#v", result)
	}
}

// TestFromInternalSession verifies host-internal session projections are copied into plugin DTOs.
func TestFromInternalSession(t *testing.T) {
	loginTime := time.Now()
	sessionItem := &internalsession.Session{
		TokenId:        "token-1",
		UserId:         100,
		Username:       "admin",
		ClientType:     "desktop",
		DeptName:       "Engineering",
		Ip:             "127.0.0.1",
		Browser:        "Chrome",
		Os:             "macOS",
		LoginTime:      &loginTime,
		LastActiveTime: &loginTime,
	}

	result := fromInternalSession(sessionItem)
	if result == nil {
		t.Fatal("expected converted session, got nil")
	}
	if string(result.ID) != sessionItem.TokenId ||
		result.UserID != "100" ||
		result.Username != sessionItem.Username ||
		result.ClientType != sessionItem.ClientType ||
		result.DeptName != sessionItem.DeptName ||
		result.Ip != sessionItem.Ip ||
		result.Browser != sessionItem.Browser ||
		result.Os != sessionItem.Os ||
		result.LoginAt != sessionItem.LoginTime ||
		result.LastActiveAt != sessionItem.LastActiveTime {
		t.Fatalf("unexpected converted session: %#v", result)
	}
}

// TestFromInternalListResult verifies nil-safe list conversion and item projection.
func TestFromInternalListResult(t *testing.T) {
	empty := fromInternalListResult(nil)
	if empty == nil {
		t.Fatal("expected empty result, got nil")
	}
	if empty.Total != 0 || len(empty.Items) != 0 {
		t.Fatalf("unexpected empty result: %#v", empty)
	}

	loginTime := time.Now()
	result := fromInternalListResult(&internalsession.ListResult{
		Items: []*internalsession.Session{
			{
				TokenId:        "token-2",
				UserId:         101,
				Username:       "demo",
				ClientType:     "mobile",
				DeptName:       "QA",
				Ip:             "10.0.0.1",
				Browser:        "Firefox",
				Os:             "Linux",
				LoginTime:      &loginTime,
				LastActiveTime: &loginTime,
			},
		},
		Total: 1,
	})
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected converted list result: %#v", result)
	}
	if result.Items[0] == nil || result.Items[0].ID != "token-2" || result.Items[0].ClientType != "mobile" {
		t.Fatalf("unexpected converted item: %#v", result.Items[0])
	}
}

// TestSessionListPageAndRevokeApplyDataScope verifies online-user operations are scope-bound.
func TestSessionListPageAndRevokeApplyDataScope(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := &sessionDataScopeStore{
		sessions: []*internalsession.Session{
			{TokenId: "visible-token", TenantId: 22, UserId: 10, Username: "visible", ClientType: "web", LoginTime: &now, LastActiveTime: &now},
			{TokenId: "hidden-token", TenantId: 33, UserId: 20, Username: "hidden", ClientType: "web", LoginTime: &now, LastActiveTime: &now},
		},
	}
	svc := newSessionCapabilityAdapter(
		nil,
		sessionDataScopeService{visibleUserIDs: map[int]bool{10: true}},
		store,
		sessionTenantScopeService{visibleTenantIDs: map[int]bool{22: true}},
	)

	out, err := svc.Search(ctx, capmodel.CapabilityContext{}, capabilitysessioncap.SearchInput{Page: capmodel.PageRequest{PageNum: 1, PageSize: 20}})
	if err != nil {
		t.Fatalf("list scoped sessions: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].ID != "visible-token" {
		t.Fatalf("expected only visible session, got %#v", out)
	}
	if out.Items[0].TenantID != "22" {
		t.Fatalf("expected visible session tenant projection 22, got %s", out.Items[0].TenantID)
	}

	batch, err := svc.BatchGet(ctx, capmodel.CapabilityContext{}, []capabilitysessioncap.SessionID{"visible-token", "hidden-token", "missing-token"})
	if err != nil {
		t.Fatalf("batch get scoped sessions: %v", err)
	}
	if got := batch.Items["visible-token"]; got == nil || got.ID != "visible-token" {
		t.Fatalf("expected visible token in batch result, got %#v", batch)
	}
	if len(batch.MissingIDs) != 2 ||
		!containsSessionID(batch.MissingIDs, "hidden-token") ||
		!containsSessionID(batch.MissingIDs, "missing-token") {
		t.Fatalf("expected hidden and missing tokens to be opaque missing IDs, got %#v", batch.MissingIDs)
	}
	if store.batchRequested == 0 {
		t.Fatal("expected batch query path to be used")
	}

	if err = svc.Revoke(ctx, capmodel.CapabilityContext{}, "hidden-token"); err == nil {
		t.Fatal("expected hidden session revoke to be denied")
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected hidden session not to be deleted, got token %q", store.deletedTokenID)
	}

	store.sessions = append(store.sessions, &internalsession.Session{
		TokenId:        "hidden-tenant-token",
		TenantId:       33,
		UserId:         10,
		Username:       "visible-user-hidden-tenant",
		ClientType:     "web",
		LoginTime:      &now,
		LastActiveTime: &now,
	})
	if err = svc.Revoke(ctx, capmodel.CapabilityContext{}, "hidden-tenant-token"); err == nil {
		t.Fatal("expected hidden tenant session revoke to be denied")
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected hidden tenant session not to be deleted, got token %q", store.deletedTokenID)
	}

	if err = svc.Revoke(ctx, capmodel.CapabilityContext{}, "visible-token"); err != nil {
		t.Fatalf("expected visible non-platform session revoke, got %v", err)
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected adapter without auth service to only authorize visible token, got deleted token %q", store.deletedTokenID)
	}
}

// containsSessionID reports whether ids already contains id.
func containsSessionID(ids []capabilitysessioncap.SessionID, id capabilitysessioncap.SessionID) bool {
	for _, existing := range ids {
		if existing == id {
			return true
		}
	}
	return false
}

// sessionDataScopeStore is an in-memory session store for capability tests.
type sessionDataScopeStore struct {
	sessions       []*internalsession.Session
	deletedTokenID string
	batchRequested int
}

// Get returns one session by token ID.
func (s *sessionDataScopeStore) Get(_ context.Context, tokenID string) (*internalsession.Session, error) {
	for _, sessionItem := range s.sessions {
		if sessionItem != nil && sessionItem.TokenId == tokenID {
			return sessionItem, nil
		}
	}
	return nil, nil
}

// BatchGetScoped returns requested sessions whose users and tenants are visible.
func (s *sessionDataScopeStore) BatchGetScoped(
	ctx context.Context,
	tokenIDs []string,
	scopeSvc datascope.Service,
	tenantSvc tenantspi.ScopeService,
) ([]*internalsession.Session, error) {
	s.batchRequested++
	requested := make(map[string]struct{}, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		requested[tokenID] = struct{}{}
	}
	items := make([]*internalsession.Session, 0, len(tokenIDs))
	for _, sessionItem := range s.sessions {
		if sessionItem == nil {
			continue
		}
		if _, ok := requested[sessionItem.TokenId]; !ok {
			continue
		}
		if tenantVisibility, ok := tenantSvc.(interface {
			EnsureTenantVisible(context.Context, tenantcapsvc.TenantID) error
		}); ok && tenantVisibility != nil {
			if err := tenantVisibility.EnsureTenantVisible(ctx, tenantcapsvc.TenantID(sessionItem.TenantId)); err != nil {
				if bizerr.Is(err, tenantcap.CodeTenantForbidden) {
					continue
				}
				return nil, err
			}
		}
		if scopeSvc != nil {
			if err := scopeSvc.EnsureUsersVisible(ctx, []int{sessionItem.UserId}); err != nil {
				if _, ok := bizerr.As(err); ok {
					continue
				}
				return nil, err
			}
		}
		items = append(items, sessionItem)
	}
	return items, nil
}

// ListPageScoped returns only sessions whose users are visible to the supplied scope service.
func (s *sessionDataScopeStore) ListPageScoped(
	ctx context.Context,
	filter *internalsession.ListFilter,
	pageNum, pageSize int,
	scopeSvc datascope.Service,
	tenantSvc tenantspi.ScopeService,
) (*internalsession.ListResult, error) {
	items := make([]*internalsession.Session, 0, len(s.sessions))
	for _, sessionItem := range s.sessions {
		if sessionItem == nil {
			continue
		}
		if tenantVisibility, ok := tenantSvc.(interface {
			EnsureTenantVisible(context.Context, tenantcapsvc.TenantID) error
		}); ok && tenantVisibility != nil {
			if err := tenantVisibility.EnsureTenantVisible(ctx, tenantcapsvc.TenantID(sessionItem.TenantId)); err != nil {
				if bizerr.Is(err, tenantcap.CodeTenantForbidden) {
					continue
				}
				return nil, err
			}
		}
		if scopeSvc != nil {
			if err := scopeSvc.EnsureUsersVisible(ctx, []int{sessionItem.UserId}); err != nil {
				if _, ok := bizerr.As(err); ok {
					continue
				}
				return nil, err
			}
		}
		items = append(items, sessionItem)
	}
	return &internalsession.ListResult{Items: items, Total: len(items)}, nil
}

// Set persists one session in memory.
func (s *sessionDataScopeStore) Set(_ context.Context, sessionItem *internalsession.Session) error {
	s.sessions = append(s.sessions, sessionItem)
	return nil
}

// Delete records the deleted token ID.
func (s *sessionDataScopeStore) Delete(_ context.Context, tokenID string) error {
	s.deletedTokenID = tokenID
	return nil
}

// DeleteByUserId is unused by sessioncap data-scope tests.
func (s *sessionDataScopeStore) DeleteByUserId(context.Context, int, int) error { return nil }

// List returns all configured sessions.
func (s *sessionDataScopeStore) List(context.Context, *internalsession.ListFilter) ([]*internalsession.Session, error) {
	return append([]*internalsession.Session(nil), s.sessions...), nil
}

// ListPage returns all configured sessions without scope filtering.
func (s *sessionDataScopeStore) ListPage(context.Context, *internalsession.ListFilter, int, int) (*internalsession.ListResult, error) {
	items := append([]*internalsession.Session(nil), s.sessions...)
	return &internalsession.ListResult{Items: items, Total: len(items)}, nil
}

// Count returns the number of configured sessions.
func (s *sessionDataScopeStore) Count(context.Context) (int, error) { return len(s.sessions), nil }

// TouchOrValidate is unused by sessioncap data-scope tests.
func (s *sessionDataScopeStore) TouchOrValidate(context.Context, int, string, time.Duration) (bool, error) {
	return true, nil
}

// CleanupInactive is unused by sessioncap data-scope tests.
func (s *sessionDataScopeStore) CleanupInactive(context.Context, time.Duration) (int64, error) {
	return 0, nil
}

// sessionDataScopeService allows only configured user IDs.
type sessionDataScopeService struct {
	visibleUserIDs map[int]bool
}

// Current returns a minimal all-scope context.
func (s sessionDataScopeService) Current(context.Context) (*datascope.Context, error) {
	return &datascope.Context{UserID: 10, Scope: datascope.ScopeAll}, nil
}

// ApplyUserScope is unused by this in-memory fake.
func (s sessionDataScopeService) ApplyUserScope(context.Context, *gdb.Model, string) (*gdb.Model, bool, error) {
	return nil, false, nil
}

// ApplyUserScopeWithBypass is unused by this in-memory fake.
func (s sessionDataScopeService) ApplyUserScopeWithBypass(
	context.Context,
	*gdb.Model,
	string,
	string,
	any,
) (*gdb.Model, bool, error) {
	return nil, false, nil
}

// EnsureUsersVisible verifies all requested users are configured as visible.
func (s sessionDataScopeService) EnsureUsersVisible(_ context.Context, userIDs []int) error {
	for _, userID := range userIDs {
		if !s.visibleUserIDs[userID] {
			return bizerr.NewCode(datascope.CodeDataScopeDenied)
		}
	}
	return nil
}

// EnsureRowsVisible is unused by this in-memory fake.
func (s sessionDataScopeService) EnsureRowsVisible(context.Context, *gdb.Model, string, int) error {
	return nil
}

// sessionTenantScopeService allows only configured tenant IDs.
type sessionTenantScopeService struct {
	visibleTenantIDs map[int]bool
}

// Available reports an active tenant provider for tenant visibility tests.
func (s sessionTenantScopeService) Available(context.Context) bool { return true }

// Status returns an available tenant capability status.
func (s sessionTenantScopeService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{Available: true, ActiveProvider: tenantcap.ProviderPluginID}
}

// Current returns the first configured tenant ID.
func (s sessionTenantScopeService) Current(context.Context) tenantcapsvc.TenantID {
	for tenantID := range s.visibleTenantIDs {
		return tenantcapsvc.TenantID(tenantID)
	}
	return tenantcap.PLATFORM
}

// Apply is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) Apply(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, error) {
	return model, nil
}

// PlatformBypass reports no platform bypass in tenant visibility tests.
func (s sessionTenantScopeService) PlatformBypass(context.Context) bool { return false }

// EnsureTenantVisible verifies the requested tenant is configured as visible.
func (s sessionTenantScopeService) EnsureTenantVisible(_ context.Context, tenantID tenantcapsvc.TenantID) error {
	if s.visibleTenantIDs[int(tenantID)] {
		return nil
	}
	return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
}

// ValidateUserInTenant is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ValidateUserInTenant(context.Context, int, tenantcapsvc.TenantID) error {
	return nil
}

// ResolveTenant is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ResolveTenant(ctx context.Context, _ *ghttp.Request) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: s.Current(ctx), Matched: true}, nil
}

// ApplyUserTenantScope is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ApplyUserTenantScope(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenants is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{}, nil
}

// SwitchTenant is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) SwitchTenant(context.Context, int, tenantcapsvc.TenantID) error {
	return nil
}

// ApplyUserTenantFilter is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ tenantcapsvc.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*tenantcap.UserTenantProjection, error) {
	return map[int]*tenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ResolveUserTenantAssignment(
	context.Context,
	[]tenantcapsvc.TenantID,
	tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	return &tenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*tenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) EnsureUsersInTenant(context.Context, []int, tenantcapsvc.TenantID) error {
	return nil
}

// ValidateUserMembershipStartupConsistency is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ValidateUserMembershipStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// ProvisionAutoEnabledTenantPlugins is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ProvisionAutoEnabledTenantPlugins(context.Context) error {
	return nil
}

// Interface guard keeps the fake aligned with the tenant SPI dependency.
var _ tenantspi.ScopeService = sessionTenantScopeService{}
