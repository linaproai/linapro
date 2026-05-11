// This file verifies plugin-facing online-session operations enforce data scope.

package session

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/datascope"
	internalsession "lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// TestSessionListPageAndRevokeApplyDataScope verifies online-user list and
// force-logout operations are constrained by sys_online_session.user_id.
func TestSessionListPageAndRevokeApplyDataScope(t *testing.T) {
	ctx := context.Background()
	store := &sessionDataScopeStore{
		sessions: []*internalsession.Session{
			{TokenId: "visible-token", TenantId: 22, UserId: 10, Username: "visible", LoginTime: gtime.Now(), LastActiveTime: gtime.Now()},
			{TokenId: "hidden-token", TenantId: 33, UserId: 20, Username: "hidden", LoginTime: gtime.Now(), LastActiveTime: gtime.Now()},
		},
	}
	svc := &serviceAdapter{
		authSvc:      nil,
		scopeSvc:     sessionDataScopeService{visibleUserIDs: map[int]bool{10: true}},
		sessionStore: store,
		tenantSvc:    sessionTenantScopeService{visibleTenantIDs: map[int]bool{22: true}},
	}

	out, err := svc.ListPage(ctx, nil, 1, 20)
	if err != nil {
		t.Fatalf("list scoped sessions: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].TokenId != "visible-token" {
		t.Fatalf("expected only visible session, got %#v", out)
	}
	if out.Items[0].TenantId != 22 {
		t.Fatalf("expected visible session tenant projection 22, got %d", out.Items[0].TenantId)
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
		LoginTime:      gtime.Now(),
		LastActiveTime: gtime.Now(),
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

// sessionDataScopeStore is an in-memory session store for pluginservice tests.
type sessionDataScopeStore struct {
	sessions       []*internalsession.Session
	deletedTokenID string
}

// Set persists one session in memory.
func (s *sessionDataScopeStore) Set(_ context.Context, session *internalsession.Session) error {
	s.sessions = append(s.sessions, session)
	return nil
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

// Delete records the deleted token ID.
func (s *sessionDataScopeStore) Delete(_ context.Context, tokenID string) error {
	s.deletedTokenID = tokenID
	return nil
}

// DeleteByUserId is unused by pluginservice data-scope tests.
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

// ListPageScoped returns only sessions whose users are visible to the supplied scope service.
func (s *sessionDataScopeStore) ListPageScoped(
	ctx context.Context,
	filter *internalsession.ListFilter,
	pageNum, pageSize int,
	scopeSvc datascope.Service,
	tenantSvc tenantcapsvc.Service,
) (*internalsession.ListResult, error) {
	items := make([]*internalsession.Session, 0, len(s.sessions))
	for _, sessionItem := range s.sessions {
		if sessionItem == nil {
			continue
		}
		if tenantSvc != nil {
			if err := tenantSvc.EnsureTenantVisible(ctx, tenantcapsvc.TenantID(sessionItem.TenantId)); err != nil {
				if bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
					continue
				}
				return nil, err
			}
		}
		if scopeSvc != nil {
			if err := scopeSvc.EnsureUsersVisible(ctx, []int{sessionItem.UserId}); err != nil {
				if bizerr.Is(err, datascope.CodeDataScopeDenied) {
					continue
				}
				return nil, err
			}
		}
		items = append(items, sessionItem)
	}
	return &internalsession.ListResult{Items: items, Total: len(items)}, nil
}

// Count returns the number of configured sessions.
func (s *sessionDataScopeStore) Count(context.Context) (int, error) { return len(s.sessions), nil }

// TouchOrValidate is unused by pluginservice data-scope tests.
func (s *sessionDataScopeStore) TouchOrValidate(context.Context, int, string, time.Duration) (bool, error) {
	return true, nil
}

// CleanupInactive is unused by pluginservice data-scope tests.
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
func (s sessionDataScopeService) ApplyUserScopeWithBypass(context.Context, *gdb.Model, string, string, any) (*gdb.Model, bool, error) {
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

// Enabled reports multi-tenancy as enabled for tenant visibility tests.
func (s sessionTenantScopeService) Enabled(context.Context) bool { return true }

// Current returns the first configured tenant ID.
func (s sessionTenantScopeService) Current(context.Context) tenantcapsvc.TenantID {
	for tenantID := range s.visibleTenantIDs {
		return tenantcapsvc.TenantID(tenantID)
	}
	return pkgtenantcap.PLATFORM
}

// Apply is unused by pluginservice data-scope tests.
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
	return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
}

// ResolveTenant is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ResolveTenant(ctx context.Context, _ *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: s.Current(ctx), Matched: true}, nil
}

// ReadWithPlatformFallback is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ReadWithPlatformFallback(context.Context, tenantcapsvc.FallbackScanner[any]) ([]any, error) {
	return nil, nil
}

// ApplyUserTenantScope is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ApplyUserTenantScope(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenants is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return []pkgtenantcap.TenantInfo{}, nil
}

// ApplyUserTenantFilter is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ tenantcapsvc.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return map[int]*pkgtenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ResolveUserTenantAssignment(
	context.Context,
	[]tenantcapsvc.TenantID,
	pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) EnsureUsersInTenant(context.Context, []int, tenantcapsvc.TenantID) error {
	return nil
}

// ValidateUserMembershipStartupConsistency is unused by pluginservice data-scope tests.
func (s sessionTenantScopeService) ValidateUserMembershipStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// Interface guard keeps the fake aligned with the tenantcap dependency.
var _ tenantcapsvc.Service = sessionTenantScopeService{}
