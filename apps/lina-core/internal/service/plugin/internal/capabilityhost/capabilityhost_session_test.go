// This file verifies session capability conversion, visibility filtering, and
// revocation behavior inside the sessioncap component.

package capabilityhost

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	authsvc "lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	internalsession "lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
)

// TestToInternalFilter verifies the published filter contract is converted explicitly.
func TestToInternalFilter(t *testing.T) {
	if result := toInternalFilter(capabilitysessioncap.ListInput{}); result != nil {
		t.Fatalf("expected nil filter, got %#v", result)
	}

	filter := capabilitysessioncap.ListInput{
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
		sessionAuthService{store: store},
		nil,
		sessionUsersService{visibleUserIDs: map[string]bool{"10": true}},
		sessionDataScopeService{visibleUserIDs: map[int]bool{10: true}},
		sessionTenantScopeService{visibleTenantIDs: map[int]bool{22: true}},
	)

	out, err := svc.List(ctx, capabilitysessioncap.ListInput{Page: capmodel.PageRequest{PageNum: 1, PageSize: 20}})
	if err != nil {
		t.Fatalf("list scoped sessions: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].ID != "visible-token" {
		t.Fatalf("expected only visible session, got %#v", out)
	}
	if out.Items[0].TenantID != "22" {
		t.Fatalf("expected visible session tenant projection 22, got %s", out.Items[0].TenantID)
	}

	batch, err := svc.BatchGet(ctx, []capabilitysessioncap.SessionID{"visible-token", "hidden-token", "missing-token"})
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

	if err = svc.Revoke(ctx, "hidden-token"); err == nil {
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
	if err = svc.Revoke(ctx, "hidden-tenant-token"); err == nil {
		t.Fatal("expected hidden tenant session revoke to be denied")
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected hidden tenant session not to be deleted, got token %q", store.deletedTokenID)
	}

	if err = svc.Revoke(ctx, "visible-token"); err != nil {
		t.Fatalf("expected visible non-platform session revoke, got %v", err)
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected adapter without auth service to only authorize visible token, got deleted token %q", store.deletedTokenID)
	}
}

// TestSessionCurrentUsesBizContextToken verifies current session lookup uses
// the request token context and the shared scoped batch path.
func TestSessionCurrentUsesBizContextToken(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := &sessionDataScopeStore{
		sessions: []*internalsession.Session{
			{TokenId: "visible-token", TenantId: 22, UserId: 10, Username: "visible", ClientType: "web", LoginTime: &now, LastActiveTime: &now},
		},
	}
	svc := newSessionCapabilityAdapter(
		sessionAuthService{store: store},
		staticBizCtx{current: bizctxcap.CurrentContext{TokenID: "visible-token"}},
		sessionUsersService{visibleUserIDs: map[string]bool{"10": true}},
		sessionDataScopeService{visibleUserIDs: map[int]bool{10: true}},
		sessionTenantScopeService{visibleTenantIDs: map[int]bool{22: true}},
	)

	current, err := svc.Current(ctx)
	if err != nil {
		t.Fatalf("current session failed: %v", err)
	}
	if current == nil || current.ID != "visible-token" || store.batchRequested == 0 {
		t.Fatalf("expected current session from scoped batch path, current=%#v batchRequested=%d", current, store.batchRequested)
	}
}

// TestSessionCurrentRejectsMissingToken verifies current session lookup fails
// closed when no request token context is available.
func TestSessionCurrentRejectsMissingToken(t *testing.T) {
	svc := newSessionCapabilityAdapter(sessionAuthService{store: &sessionDataScopeStore{}}, nil, nil, nil, nil)
	_, err := svc.Current(context.Background())
	if !bizerr.Is(err, capmodel.CodeCapabilityCurrentUserRequired) {
		t.Fatalf("expected context-required error, got %v", err)
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

type staticBizCtx struct {
	current bizctxcap.CurrentContext
}

func (s staticBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

// sessionAuthService exposes the shared session store required by the session
// capability adapter without exercising auth's full token lifecycle.
type sessionAuthService struct {
	authsvc.Service

	store internalsession.Store
}

// SessionStore returns the test-owned session store.
func (s sessionAuthService) SessionStore() internalsession.Store {
	return s.store
}

// RevokeSession accepts visible revocation requests without mutating the fake store.
func (s sessionAuthService) RevokeSession(context.Context, string) error {
	return nil
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

// BatchGetUserOnlineStatusScoped returns visible session counts for requested users.
func (s *sessionDataScopeStore) BatchGetUserOnlineStatusScoped(
	ctx context.Context,
	userIDs []int,
	scopeSvc datascope.Service,
	tenantSvc tenantspi.ScopeService,
) ([]*internalsession.UserOnlineStatus, error) {
	statusByUserID := make(map[int]int, len(userIDs))
	requested := make(map[int]struct{}, len(userIDs))
	for _, userID := range userIDs {
		requested[userID] = struct{}{}
	}
	for _, sessionItem := range s.sessions {
		if sessionItem == nil {
			continue
		}
		if _, ok := requested[sessionItem.UserId]; !ok {
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
		statusByUserID[sessionItem.UserId]++
	}
	statuses := make([]*internalsession.UserOnlineStatus, 0, len(statusByUserID))
	for userID, count := range statusByUserID {
		statuses = append(statuses, &internalsession.UserOnlineStatus{UserId: userID, SessionCount: count})
	}
	return statuses, nil
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
	tenantspi.Service

	visibleTenantIDs map[int]bool
}

// Available reports an active tenant query-scope provider for visibility tests.
func (s sessionTenantScopeService) Available(context.Context) bool { return true }

// Directory returns tenant directory operations for tests.
func (s sessionTenantScopeService) Directory() tenantcap.DirectoryService {
	return s
}

// Apply is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) Apply(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, error) {
	return model, nil
}

// EnsureTenantVisible verifies the requested tenant is configured as visible.
func (s sessionTenantScopeService) EnsureTenantVisible(_ context.Context, tenantID tenantcapsvc.TenantID) error {
	if s.visibleTenantIDs[int(tenantID)] {
		return nil
	}
	return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
}

// EnsureVisible verifies the requested tenant identifiers are configured as visible.
func (s sessionTenantScopeService) EnsureVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error {
	return s.ensureTenantsVisible(ctx, tenantIDs)
}

// ApplyUserTenantScope is unused by sessioncap data-scope tests.
func (s sessionTenantScopeService) ApplyUserTenantScope(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, bool, error) {
	return model, false, nil
}

// batchGetTenants returns visible tenant projections and opaque missing IDs.
func (s sessionTenantScopeService) batchGetTenants(_ context.Context, tenantIDs []tenantcapsvc.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcapsvc.TenantID], error) {
	result := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcapsvc.TenantID]{
		Items:      map[tenantcapsvc.TenantID]*tenantcap.TenantInfo{},
		MissingIDs: []tenantcapsvc.TenantID{},
	}
	for _, tenantID := range tenantIDs {
		if s.visibleTenantIDs[int(tenantID)] {
			result.Items[tenantID] = &tenantcap.TenantInfo{ID: tenantID, Code: "tenant", Name: "Tenant", Status: "active"}
			continue
		}
		result.MissingIDs = append(result.MissingIDs, tenantID)
	}
	return result, nil
}

// Get returns one visible tenant projection.
func (s sessionTenantScopeService) Get(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error) {
	result, err := s.batchGetTenants(ctx, []tenantcap.TenantID{tenantID})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[tenantID], nil
}

// BatchGet returns visible tenant projections and opaque missing IDs.
func (s sessionTenantScopeService) BatchGet(ctx context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	return s.batchGetTenants(ctx, tenantIDs)
}

// List returns an empty tenant directory page for session capability tests.
func (s sessionTenantScopeService) List(context.Context, tenantcap.ListInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: []*tenantcap.TenantInfo{}}, nil
}

// ensureTenantsVisible validates each tenant against the fake visible set.
func (s sessionTenantScopeService) ensureTenantsVisible(ctx context.Context, tenantIDs []tenantcapsvc.TenantID) error {
	for _, tenantID := range tenantIDs {
		if err := s.EnsureTenantVisible(ctx, tenantID); err != nil {
			return err
		}
	}
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

// sessionUsersService exposes visible users for online-status tests.
type sessionUsersService struct {
	visibleUserIDs map[string]bool
}

// Current is unused by sessioncap tests.
func (s sessionUsersService) Current(context.Context) (*capabilityusercap.UserInfo, error) {
	return nil, nil
}

// BatchGet returns only configured visible users.
func (s sessionUsersService) BatchGet(
	_ context.Context,
	ids []capabilityusercap.UserID,
) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]{
		Items:      map[capabilityusercap.UserID]*capabilityusercap.UserInfo{},
		MissingIDs: []capabilityusercap.UserID{},
	}
	for _, id := range ids {
		if s.visibleUserIDs[string(id)] {
			result.Items[id] = &capabilityusercap.UserInfo{ID: id}
		} else {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// Get returns one visible test user projection.
func (s sessionUsersService) Get(ctx context.Context, id capabilityusercap.UserID) (*capabilityusercap.UserInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityusercap.UserID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// BatchResolve is unused by sessioncap tests.
func (s sessionUsersService) BatchResolve(
	context.Context,
	capabilityusercap.BatchResolveInput,
) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey], error) {
	return &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey]{Items: map[capabilityusercap.ResolveKey]*capabilityusercap.UserInfo{}}, nil
}

// List is unused by sessioncap tests.
func (s sessionUsersService) List(
	context.Context,
	capabilityusercap.ListInput,
) (*capmodel.PageResult[*capabilityusercap.UserInfo], error) {
	return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: []*capabilityusercap.UserInfo{}}, nil
}

// EnsureVisible is unused by sessioncap tests.
func (s sessionUsersService) EnsureVisible(context.Context, []capabilityusercap.UserID) error {
	return nil
}

// Create is unused by sessioncap tests.
func (s sessionUsersService) Create(context.Context, capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	return "", nil
}

// ProvisionExternal is unused by sessioncap tests.
func (s sessionUsersService) ProvisionExternal(context.Context, capabilityusercap.ProvisionExternalInput) (capabilityusercap.UserID, error) {
	return "", nil
}

// Update is unused by sessioncap tests.
func (s sessionUsersService) Update(context.Context, capabilityusercap.UpdateInput) error {
	return nil
}

// Delete is unused by sessioncap tests.
func (s sessionUsersService) Delete(context.Context, capabilityusercap.UserID) error {
	return nil
}

// SetStatus is unused by sessioncap tests.
func (s sessionUsersService) SetStatus(context.Context, capabilityusercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword is unused by sessioncap tests.
func (s sessionUsersService) ResetPassword(context.Context, capabilityusercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations unused by sessioncap tests.
func (s sessionUsersService) Assignment() capabilityusercap.AssignmentService {
	return sessionUserAssignments{}
}

// sessionUserAssignments accepts unused role replacements.
type sessionUserAssignments struct{}

// ReplaceRoles is unused by sessioncap tests.
func (sessionUserAssignments) ReplaceRoles(context.Context, capabilityusercap.UserID, []int) error {
	return nil
}

var _ capabilityusercap.Service = sessionUsersService{}
