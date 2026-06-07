// Package sessioncap adapts host online-session storage and auth revocation to
// plugin-visible session capability contracts.
package sessioncap

import (
	"context"
	"strconv"
	"strings"

	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/hostservices/internal/domaincap"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
)

// AuthSessionRevoker defines the host auth revocation slice required by the adapter.
type AuthSessionRevoker interface {
	// RevokeSession writes a shared revoke marker and removes one online session by token ID.
	RevokeSession(ctx context.Context, tokenID string) error
}

// Service exposes the online-session domain service and management commands.
type Service interface {
	capabilitysessioncap.Service
	capabilitysessioncap.AdminService
}

// adapter bridges host auth/session services into the published session domain
// capability contract.
type adapter struct {
	authSvc      AuthSessionRevoker
	scopeSvc     datascope.Service
	sessionStore session.Store
	tenantSvc    tenantcapsvc.RuntimeService
}

var (
	_ capabilitysessioncap.Service      = (*adapter)(nil)
	_ capabilitysessioncap.AdminService = (*adapter)(nil)
)

// New creates the host-owned online-session capability adapter.
func New(
	authSvc AuthSessionRevoker,
	scopeSvc datascope.Service,
	sessionStore session.Store,
	tenantSvc tenantcapsvc.RuntimeService,
) Service {
	return &adapter{
		authSvc:      authSvc,
		scopeSvc:     scopeSvc,
		sessionStore: sessionStore,
		tenantSvc:    tenantSvc,
	}
}

// SearchSessions returns one bounded visible session page.
func (a *adapter) SearchSessions(ctx context.Context, _ capmodel.CapabilityContext, input capabilitysessioncap.SearchInput) (*capmodel.PageResult[*capabilitysessioncap.Projection], error) {
	if a == nil || a.sessionStore == nil {
		return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: []*capabilitysessioncap.Projection{}, Total: 0}, nil
	}
	pageNum, pageSize := domaincap.NormalizePage(input.Page)
	result, err := a.sessionStore.ListPageScoped(
		ctx,
		toInternalFilter(input),
		pageNum,
		pageSize,
		a.currentScopeSvc(),
		a.currentTenantSvc(),
	)
	if err != nil {
		return nil, err
	}
	return fromInternalListResult(result), nil
}

// BatchGetSessions returns visible sessions and opaque missing IDs.
func (a *adapter) BatchGetSessions(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilitysessioncap.SessionID) (*capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID], error) {
	result := &capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID]{
		Items:      make(map[capabilitysessioncap.SessionID]*capabilitysessioncap.Projection, len(ids)),
		MissingIDs: []capabilitysessioncap.SessionID{},
	}
	if len(ids) == 0 {
		return result, nil
	}
	requested := make(map[string]capabilitysessioncap.SessionID, len(ids))
	for _, id := range ids {
		tokenID := strings.TrimSpace(string(id))
		if tokenID == "" {
			result.MissingIDs = append(result.MissingIDs, id)
			continue
		}
		if _, exists := requested[tokenID]; exists {
			continue
		}
		requested[tokenID] = id
	}
	if a == nil || a.sessionStore == nil {
		for _, id := range ids {
			if _, ok := result.Items[id]; !ok && !domaincap.Contains(result.MissingIDs, id) {
				result.MissingIDs = append(result.MissingIDs, id)
			}
		}
		return result, nil
	}
	items, err := a.batchGetInternalSessions(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		requestID, ok := requested[string(item.ID)]
		if ok {
			result.Items[requestID] = item
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !domaincap.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// RevokeSession invalidates one visible online session.
func (a *adapter) RevokeSession(ctx context.Context, _ capmodel.CapabilityContext, id capabilitysessioncap.SessionID) error {
	if a == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "session"))
	}
	if strings.TrimSpace(string(id)) == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	tokenID := string(id)
	if a.sessionStore != nil {
		sessionItem, err := a.sessionStore.Get(ctx, tokenID)
		if err != nil {
			return err
		}
		if sessionItem != nil {
			if tenantSvc := a.currentTenantSvc(); tenantSvc != nil {
				if err = tenantSvc.EnsureTenantVisible(ctx, tenantcapsvc.TenantID(sessionItem.TenantId)); err != nil {
					return err
				}
			}
			if scopeSvc := a.currentScopeSvc(); scopeSvc != nil {
				if err = scopeSvc.EnsureUsersVisible(ctx, []int{sessionItem.UserId}); err != nil {
					return err
				}
			}
		}
	}
	if a.authSvc == nil {
		return nil
	}
	return a.authSvc.RevokeSession(ctx, tokenID)
}

// batchGetInternalSessions returns requested visible sessions without paging unrelated rows.
func (a *adapter) batchGetInternalSessions(ctx context.Context, ids []capabilitysessioncap.SessionID) ([]*capabilitysessioncap.Projection, error) {
	if a == nil || a.sessionStore == nil {
		return []*capabilitysessioncap.Projection{}, nil
	}
	tokenIDs := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		tokenID := strings.TrimSpace(string(id))
		if tokenID == "" {
			continue
		}
		if _, exists := seen[tokenID]; exists {
			continue
		}
		seen[tokenID] = struct{}{}
		tokenIDs = append(tokenIDs, tokenID)
	}
	if len(tokenIDs) == 0 {
		return []*capabilitysessioncap.Projection{}, nil
	}
	sessions, err := a.sessionStore.BatchGetScoped(ctx, tokenIDs, a.currentScopeSvc(), a.currentTenantSvc())
	if err != nil {
		return nil, err
	}
	items := make([]*capabilitysessioncap.Projection, 0, len(sessions))
	for _, sessionItem := range sessions {
		items = append(items, fromInternalSession(sessionItem))
	}
	return items, nil
}

// currentScopeSvc returns the shared data-scope service for session operations.
func (a *adapter) currentScopeSvc() datascope.Service {
	if a.scopeSvc != nil {
		return a.scopeSvc
	}
	return nil
}

// currentTenantSvc returns the shared tenant capability service for session operations.
func (a *adapter) currentTenantSvc() tenantcapsvc.RuntimeService {
	if a.tenantSvc != nil {
		return a.tenantSvc
	}
	return nil
}

// toInternalFilter converts the session domain filter into the host-internal filter.
func toInternalFilter(input capabilitysessioncap.SearchInput) *session.ListFilter {
	if strings.TrimSpace(input.Username) == "" && strings.TrimSpace(input.IP) == "" {
		return nil
	}
	return &session.ListFilter{Username: input.Username, Ip: input.IP}
}

// fromInternalListResult projects the host-internal paged session result into the session domain contract.
func fromInternalListResult(result *session.ListResult) *capmodel.PageResult[*capabilitysessioncap.Projection] {
	if result == nil {
		return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: []*capabilitysessioncap.Projection{}, Total: 0}
	}
	items := make([]*capabilitysessioncap.Projection, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, fromInternalSession(item))
	}
	return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: items, Total: result.Total}
}

// fromInternalSession copies one host-internal session projection into the plugin-facing DTO.
func fromInternalSession(sessionItem *session.Session) *capabilitysessioncap.Projection {
	if sessionItem == nil {
		return nil
	}
	return &capabilitysessioncap.Projection{
		ID:           capabilitysessioncap.SessionID(sessionItem.TokenId),
		TenantID:     capmodel.DomainID(strconv.Itoa(sessionItem.TenantId)),
		UserID:       strconv.Itoa(sessionItem.UserId),
		Username:     sessionItem.Username,
		ClientType:   sessionItem.ClientType,
		DeptName:     sessionItem.DeptName,
		Ip:           sessionItem.Ip,
		Browser:      sessionItem.Browser,
		Os:           sessionItem.Os,
		LoginAt:      sessionItem.LoginTime,
		LastActiveAt: sessionItem.LastActiveTime,
	}
}
