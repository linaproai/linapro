// Package monitor implements the online-user governance service for the
// monitor-online source plugin. It consumes the published host session seam
// so the host continues to own authentication and session truth.
package monitor

import (
	"context"

	sessionsvc "lina-core/pkg/pluginservice/contract"
)

// Service defines the monitor-online service contract.
type Service interface {
	// List returns one paginated online-user list.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// ForceLogout invalidates one online-user session by token ID.
	ForceLogout(ctx context.Context, tokenID string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	sessionSvc sessionsvc.SessionService // published host session service
}

// New creates and returns a new monitor-online service instance.
func New(sessionSvc sessionsvc.SessionService) Service {
	return &serviceImpl{sessionSvc: sessionSvc}
}

// ListInput defines the online-user list filter input.
type ListInput struct {
	PageNum  int
	PageSize int
	Username string
	Ip       string
}

// ListOutput defines the online-user list result.
type ListOutput struct {
	Items []*sessionsvc.Session
	Total int
}

// List returns one paginated online-user list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.sessionSvc.ListPage(ctx, &sessionsvc.ListFilter{
		Username: in.Username,
		Ip:       in.Ip,
	}, in.PageNum, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &ListOutput{Items: out.Items, Total: out.Total}, nil
}

// ForceLogout invalidates one online-user session by token ID.
func (s *serviceImpl) ForceLogout(ctx context.Context, tokenID string) error {
	return s.sessionSvc.Revoke(ctx, tokenID)
}
