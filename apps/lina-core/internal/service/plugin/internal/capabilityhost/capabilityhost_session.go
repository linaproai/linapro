// This file adapts host online-session storage and auth revocation to
// plugin-visible session capability contracts.
package capabilityhost

import (
	"context"
	"strconv"
	"strings"

	"lina-core/internal/service/datascope"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	tenantcap "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

// AuthSessionRevoker defines the host auth revocation slice required by the adapter.
type sessionAuthRevoker interface {
	// RevokeSession writes a shared revoke marker and removes one online session by token ID.
	RevokeSession(ctx context.Context, tokenID string) error
}

// Service exposes the online-session domain service and management commands.
type sessionCapabilityService interface {
	capabilitysessioncap.Service
	capabilitysessioncap.AdminService
}

// adapter bridges host auth/session services into the published session domain
// capability contract.
type sessionCapabilityAdapter struct {
	authSvc      sessionAuthRevoker
	bizCtx       bizctxcap.Service
	users        capabilityusercap.Service
	scopeSvc     datascope.Service
	sessionStore session.Store
	tenantSvc    tenantspi.RuntimeService
}

var (
	_ capabilitysessioncap.Service      = (*sessionCapabilityAdapter)(nil)
	_ capabilitysessioncap.AdminService = (*sessionCapabilityAdapter)(nil)
)

// New creates the host-owned online-session capability adapter.
func newSessionCapabilityAdapter(
	authSvc AuthSessionRevoker,
	bizCtx bizctxcap.Service,
	users capabilityusercap.Service,
	scopeSvc datascope.Service,
	sessionStore session.Store,
	tenantSvc tenantspi.RuntimeService,
) sessionCapabilityService {
	return &sessionCapabilityAdapter{
		authSvc:      authSvc,
		bizCtx:       bizCtx,
		users:        users,
		scopeSvc:     scopeSvc,
		sessionStore: sessionStore,
		tenantSvc:    tenantSvc,
	}
}

// Current returns the visible session projection for the current token.
func (a *sessionCapabilityAdapter) Current(ctx context.Context, capCtx capmodel.CapabilityContext) (*capabilitysessioncap.Projection, error) {
	if a == nil || a.sessionStore == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "session"))
	}
	tokenID := ""
	if a.bizCtx != nil {
		tokenID = strings.TrimSpace(a.bizCtx.Current(ctx).TokenID)
	}
	if tokenID == "" {
		tokenID = strings.TrimSpace(bizctxcap.CurrentFromContext(ctx).TokenID)
	}
	if tokenID == "" {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityContextRequired)
	}
	result, err := a.BatchGet(ctx, capCtx, []capabilitysessioncap.SessionID{capabilitysessioncap.SessionID(tokenID)})
	if err != nil {
		return nil, err
	}
	sessionItem := result.Items[capabilitysessioncap.SessionID(tokenID)]
	if sessionItem == nil || len(result.MissingIDs) > 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return sessionItem, nil
}

// Search returns one bounded visible session page.
func (a *sessionCapabilityAdapter) Search(ctx context.Context, _ capmodel.CapabilityContext, input capabilitysessioncap.SearchInput) (*capmodel.PageResult[*capabilitysessioncap.Projection], error) {
	if a == nil || a.sessionStore == nil {
		return &capmodel.PageResult[*capabilitysessioncap.Projection]{Items: []*capabilitysessioncap.Projection{}, Total: 0}, nil
	}
	pageNum, pageSize := NormalizePage(input.Page)
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

// BatchGet returns visible sessions and opaque missing IDs.
func (a *sessionCapabilityAdapter) BatchGet(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilitysessioncap.SessionID) (*capmodel.BatchResult[*capabilitysessioncap.Projection, capabilitysessioncap.SessionID], error) {
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
			if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
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
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// BatchGetUserOnlineStatus returns visible users' online status in one bounded call.
func (a *sessionCapabilityAdapter) BatchGetUserOnlineStatus(
	ctx context.Context,
	capCtx capmodel.CapabilityContext,
	userIDs []string,
) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string], error) {
	result := &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatusProjection, string]{
		Items:      make(map[string]*capabilitysessioncap.UserOnlineStatusProjection, len(userIDs)),
		MissingIDs: []string{},
	}
	if len(userIDs) > capabilitysessioncap.MaxBatchGetUserOnlineStatus {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitysessioncap.MaxBatchGetUserOnlineStatus))
	}
	if len(userIDs) == 0 {
		return result, nil
	}
	requested := make(map[int]string, len(userIDs))
	parsedIDs := make([]int, 0, len(userIDs))
	for _, id := range userIDs {
		normalizedID := strings.TrimSpace(id)
		parsedID, err := strconv.Atoi(normalizedID)
		if err != nil || parsedID <= 0 {
			if !Contains(result.MissingIDs, id) {
				result.MissingIDs = append(result.MissingIDs, id)
			}
			continue
		}
		if _, exists := requested[parsedID]; exists {
			continue
		}
		requested[parsedID] = id
		parsedIDs = append(parsedIDs, parsedID)
	}
	if len(parsedIDs) == 0 {
		return result, nil
	}
	if a == nil || a.users == nil {
		for _, id := range userIDs {
			if !Contains(result.MissingIDs, id) {
				result.MissingIDs = append(result.MissingIDs, id)
			}
		}
		return result, nil
	}
	visibleUsers, err := a.users.BatchGet(ctx, capCtx, sessionUserIDs(parsedIDs))
	if err != nil {
		return nil, err
	}
	visibleParsedIDs := make([]int, 0, len(visibleUsers.Items))
	for _, parsedID := range parsedIDs {
		requestID := requested[parsedID]
		if _, ok := visibleUsers.Items[capabilityusercap.UserID(strconv.Itoa(parsedID))]; !ok {
			if !Contains(result.MissingIDs, requestID) {
				result.MissingIDs = append(result.MissingIDs, requestID)
			}
			continue
		}
		visibleParsedIDs = append(visibleParsedIDs, parsedID)
	}
	if len(visibleParsedIDs) == 0 {
		return result, nil
	}
	if a == nil || a.sessionStore == nil {
		for _, id := range userIDs {
			if !Contains(result.MissingIDs, id) {
				result.MissingIDs = append(result.MissingIDs, id)
			}
		}
		return result, nil
	}
	statuses, err := a.sessionStore.BatchGetUserOnlineStatusScoped(
		ctx,
		visibleParsedIDs,
		a.currentScopeSvc(),
		a.currentTenantSvc(),
	)
	if err != nil {
		return nil, err
	}
	for _, status := range statuses {
		if status == nil {
			continue
		}
		requestID, ok := requested[status.UserId]
		if !ok {
			continue
		}
		result.Items[requestID] = &capabilitysessioncap.UserOnlineStatusProjection{
			UserID:       requestID,
			Online:       status.SessionCount > 0,
			SessionCount: status.SessionCount,
		}
	}
	for _, id := range userIDs {
		normalizedID := strings.TrimSpace(id)
		parsedID, err := strconv.Atoi(normalizedID)
		if err != nil || parsedID <= 0 {
			continue
		}
		if _, ok := result.Items[id]; ok {
			continue
		}
		if _, requestedVisible := requested[parsedID]; !requestedVisible {
			if !Contains(result.MissingIDs, id) {
				result.MissingIDs = append(result.MissingIDs, id)
			}
			continue
		}
		result.Items[id] = &capabilitysessioncap.UserOnlineStatusProjection{
			UserID:       id,
			Online:       false,
			SessionCount: 0,
		}
	}
	return result, nil
}

// sessionUserIDs converts parsed user IDs to user capability IDs.
func sessionUserIDs(ids []int) []capabilityusercap.UserID {
	out := make([]capabilityusercap.UserID, 0, len(ids))
	for _, id := range ids {
		out = append(out, capabilityusercap.UserID(strconv.Itoa(id)))
	}
	return out
}

// EnsureVisible rejects when any requested online session is absent or invisible.
func (a *sessionCapabilityAdapter) EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilitysessioncap.SessionID) error {
	if len(ids) > capabilitysessioncap.MaxEnsureVisible {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitysessioncap.MaxEnsureVisible))
	}
	result, err := a.BatchGet(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Revoke invalidates one visible online session.
func (a *sessionCapabilityAdapter) Revoke(ctx context.Context, _ capmodel.CapabilityContext, id capabilitysessioncap.SessionID) error {
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
				if err = tenantSvc.EnsureTenantVisible(ctx, tenantcap.TenantID(sessionItem.TenantId)); err != nil {
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
func (a *sessionCapabilityAdapter) batchGetInternalSessions(ctx context.Context, ids []capabilitysessioncap.SessionID) ([]*capabilitysessioncap.Projection, error) {
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
func (a *sessionCapabilityAdapter) currentScopeSvc() datascope.Service {
	if a.scopeSvc != nil {
		return a.scopeSvc
	}
	return nil
}

// currentTenantSvc returns the shared tenant capability service for session operations.
func (a *sessionCapabilityAdapter) currentTenantSvc() tenantspi.RuntimeService {
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
