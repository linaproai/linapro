// This file verifies monitor-online service operations delegate to the host session seam.

package monitor

import (
	"context"
	"testing"

	sessionsvc "lina-core/pkg/pluginservice/contract"
)

// TestListDelegatesToSessionService verifies online-user listing goes through
// the published host session service with the original filter and pagination.
func TestListDelegatesToSessionService(t *testing.T) {
	ctx := context.Background()
	session := &sessionsvc.Session{TokenId: "visible-token", UserId: 10, Username: "visible"}
	sessionSvc := &monitorSessionService{
		listResult: &sessionsvc.ListResult{
			Items: []*sessionsvc.Session{session},
			Total: 1,
		},
	}
	svc := &serviceImpl{sessionSvc: sessionSvc}

	out, err := svc.List(ctx, ListInput{
		PageNum:  2,
		PageSize: 25,
		Username: "visible",
		Ip:       "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("list online sessions: %v", err)
	}
	if sessionSvc.listCalled != 1 {
		t.Fatalf("expected one ListPage call, got %d", sessionSvc.listCalled)
	}
	if sessionSvc.pageNum != 2 || sessionSvc.pageSize != 25 {
		t.Fatalf("expected page 2 size 25, got page %d size %d", sessionSvc.pageNum, sessionSvc.pageSize)
	}
	if sessionSvc.listFilter == nil || sessionSvc.listFilter.Username != "visible" || sessionSvc.listFilter.Ip != "127.0.0.1" {
		t.Fatalf("expected forwarded filter, got %#v", sessionSvc.listFilter)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0] != session {
		t.Fatalf("expected seam list result, got %#v", out)
	}
}

// TestForceLogoutDelegatesToSessionService verifies online-user revocation goes
// through the published host session service.
func TestForceLogoutDelegatesToSessionService(t *testing.T) {
	ctx := context.Background()
	sessionSvc := &monitorSessionService{}
	svc := &serviceImpl{sessionSvc: sessionSvc}

	if err := svc.ForceLogout(ctx, "target-token"); err != nil {
		t.Fatalf("force logout online session: %v", err)
	}
	if sessionSvc.revokeCalled != 1 {
		t.Fatalf("expected one Revoke call, got %d", sessionSvc.revokeCalled)
	}
	if sessionSvc.revokedTokenID != "target-token" {
		t.Fatalf("expected token target-token, got %q", sessionSvc.revokedTokenID)
	}
}

// monitorSessionService records calls to the published host session service.
type monitorSessionService struct {
	listCalled     int
	listFilter     *sessionsvc.ListFilter
	listResult     *sessionsvc.ListResult
	pageNum        int
	pageSize       int
	revokeCalled   int
	revokedTokenID string
}

// ListPage records list arguments and returns the configured list result.
func (s *monitorSessionService) ListPage(_ context.Context, filter *sessionsvc.ListFilter, pageNum, pageSize int) (*sessionsvc.ListResult, error) {
	s.listCalled++
	if filter != nil {
		s.listFilter = &sessionsvc.ListFilter{
			Username: filter.Username,
			Ip:       filter.Ip,
		}
	}
	s.pageNum = pageNum
	s.pageSize = pageSize
	if s.listResult != nil {
		return s.listResult, nil
	}
	return &sessionsvc.ListResult{Items: []*sessionsvc.Session{}, Total: 0}, nil
}

// Revoke records the token ID passed to the published host session service.
func (s *monitorSessionService) Revoke(_ context.Context, tokenID string) error {
	s.revokeCalled++
	s.revokedTokenID = tokenID
	return nil
}
