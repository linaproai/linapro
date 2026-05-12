// Package session exposes a narrowed host authentication-session governance
// contract to source plugins so online-user management can query session
// projections and revoke sessions without depending on host-internal
// packages.
package session

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	internalauth "lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/orgcap"
	"lina-core/internal/service/role"
	internalsession "lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
)

// Session is the stable online-session projection published to source plugins.
type Session struct {
	// TokenId is the unique token identifier.
	TokenId string
	// TenantId is the owning tenant identifier, where 0 means platform.
	TenantId int
	// UserId is the authenticated user identifier.
	UserId int
	// Username is the authenticated username.
	Username string
	// DeptName is the projected department display name.
	DeptName string
	// Ip is the login IP address.
	Ip string
	// Browser is the login browser fingerprint.
	Browser string
	// Os is the login operating system fingerprint.
	Os string
	// LoginTime is the first login time of this session.
	LoginTime *gtime.Time
	// LastActiveTime is the most recent activity time tracked by the host.
	LastActiveTime *gtime.Time
}

// ListFilter is the stable session-list filter contract published to plugins.
type ListFilter struct {
	// Username filters sessions by username using fuzzy matching.
	Username string
	// Ip filters sessions by login IP using fuzzy matching.
	Ip string
}

// ListResult is the stable paged session-list result published to plugins.
type ListResult struct {
	// Items is the current page of online sessions.
	Items []*Session
	// Total is the total number of matching sessions.
	Total int
}

// Service defines the online-session operations published to source plugins.
type Service interface {
	// ListPage returns one paginated online-session list for the optional filter.
	ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error)
	// Revoke invalidates one online session by token ID.
	Revoke(ctx context.Context, tokenID string) error
}

// serviceAdapter bridges host auth/session services into the published plugin contract.
type serviceAdapter struct {
	authSvc      internalauth.Service
	scopeSvc     datascope.Service
	sessionStore internalsession.Store
	tenantSvc    tenantcapsvc.Service
}

// New creates and returns the published session service adapter.
func New() Service {
	authSvc := internalauth.Instance()
	return &serviceAdapter{
		authSvc:      authSvc,
		scopeSvc:     datascope.New(datascope.Dependencies{RoleSvc: role.Instance(), OrgCapSvc: orgcap.Instance()}),
		sessionStore: authSvc.SessionStore(),
		tenantSvc:    tenantcapsvc.Instance(),
	}
}

// ListPage returns one paginated online-session list for the optional filter.
func (s *serviceAdapter) ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error) {
	if s == nil || s.sessionStore == nil {
		return &ListResult{Items: []*Session{}, Total: 0}, nil
	}
	result, err := s.sessionStore.ListPageScoped(
		ctx,
		toInternalFilter(filter),
		pageNum,
		pageSize,
		s.currentScopeSvc(),
		s.currentTenantSvc(),
	)
	if err != nil {
		return nil, err
	}
	return fromInternalListResult(result), nil
}

// Revoke invalidates one online session by token ID.
func (s *serviceAdapter) Revoke(ctx context.Context, tokenID string) error {
	if s == nil {
		return nil
	}
	if s.sessionStore != nil {
		sessionItem, err := s.sessionStore.Get(ctx, tokenID)
		if err != nil {
			return err
		}
		if sessionItem != nil {
			if err = s.currentTenantSvc().EnsureTenantVisible(ctx, tenantcapsvc.TenantID(sessionItem.TenantId)); err != nil {
				return err
			}
			if err = s.currentScopeSvc().EnsureUsersVisible(ctx, []int{sessionItem.UserId}); err != nil {
				return err
			}
		}
	}
	if s.authSvc == nil {
		return nil
	}
	return s.authSvc.RevokeSession(ctx, tokenID)
}

// currentScopeSvc returns the shared data-scope service for plugin-facing session operations.
func (s *serviceAdapter) currentScopeSvc() datascope.Service {
	if s.scopeSvc != nil {
		return s.scopeSvc
	}
	return datascope.New(datascope.Dependencies{RoleSvc: role.Instance()})
}

// currentTenantSvc returns the shared tenant capability service for plugin-facing session operations.
func (s *serviceAdapter) currentTenantSvc() tenantcapsvc.Service {
	if s.tenantSvc != nil {
		return s.tenantSvc
	}
	return tenantcapsvc.Instance()
}

// toInternalFilter converts the published filter contract into the host-internal
// session filter without exposing internal types to plugins.
func toInternalFilter(filter *ListFilter) *internalsession.ListFilter {
	if filter == nil {
		return nil
	}
	return &internalsession.ListFilter{
		Username: filter.Username,
		Ip:       filter.Ip,
	}
}

// fromInternalListResult projects the host-internal paged session result into
// the published plugin contract.
func fromInternalListResult(result *internalsession.ListResult) *ListResult {
	if result == nil {
		return &ListResult{Items: []*Session{}, Total: 0}
	}
	items := make([]*Session, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, fromInternalSession(item))
	}
	return &ListResult{Items: items, Total: result.Total}
}

// fromInternalSession copies one host-internal session projection into the
// published plugin-facing session DTO.
func fromInternalSession(session *internalsession.Session) *Session {
	if session == nil {
		return nil
	}
	return &Session{
		TokenId:        session.TokenId,
		TenantId:       session.TenantId,
		UserId:         session.UserId,
		Username:       session.Username,
		DeptName:       session.DeptName,
		Ip:             session.Ip,
		Browser:        session.Browser,
		Os:             session.Os,
		LoginTime:      session.LoginTime,
		LastActiveTime: session.LastActiveTime,
	}
}
