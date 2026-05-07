// This file verifies plugin-facing online-session operations enforce data scope.

package session

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/datascope"
	internalsession "lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
)

// TestSessionListPageAndRevokeApplyDataScope verifies online-user list and
// force-logout operations are constrained by sys_online_session.user_id.
func TestSessionListPageAndRevokeApplyDataScope(t *testing.T) {
	ctx := context.Background()
	store := &sessionDataScopeStore{
		sessions: []*internalsession.Session{
			{TokenId: "visible-token", UserId: 10, Username: "visible", LoginTime: gtime.Now(), LastActiveTime: gtime.Now()},
			{TokenId: "hidden-token", UserId: 20, Username: "hidden", LoginTime: gtime.Now(), LastActiveTime: gtime.Now()},
		},
	}
	svc := &serviceAdapter{
		authSvc:      nil,
		scopeSvc:     sessionDataScopeService{visibleUserIDs: map[int]bool{10: true}},
		sessionStore: store,
	}

	out, err := svc.ListPage(ctx, nil, 1, 20)
	if err != nil {
		t.Fatalf("list scoped sessions: %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].TokenId != "visible-token" {
		t.Fatalf("expected only visible session, got %#v", out)
	}

	if err = svc.Revoke(ctx, "hidden-token"); err == nil {
		t.Fatal("expected hidden session revoke to be denied")
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected hidden session not to be deleted, got token %q", store.deletedTokenID)
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
func (s *sessionDataScopeStore) DeleteByUserId(context.Context, int) error { return nil }

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
func (s *sessionDataScopeStore) ListPageScoped(ctx context.Context, filter *internalsession.ListFilter, pageNum, pageSize int, scopeSvc datascope.Service) (*internalsession.ListResult, error) {
	items := make([]*internalsession.Session, 0, len(s.sessions))
	for _, sessionItem := range s.sessions {
		if sessionItem == nil {
			continue
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
func (s *sessionDataScopeStore) TouchOrValidate(context.Context, string, time.Duration) (bool, error) {
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
