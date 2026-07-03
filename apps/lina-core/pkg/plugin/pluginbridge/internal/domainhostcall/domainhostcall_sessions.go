// This file implements guest-side online-session capability hostcall clients.
// Session search and batch DTOs stay local to the session domain.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// sessionsService adapts online-session reads to host services.
type sessionsService struct{ baseService }

// Sessions creates the online-session domain guest client.
func Sessions(invoker Invoker) sessioncap.Service {
	return sessionsService{baseService: newBaseService(invoker)}
}

// Current returns the visible session projection for the current token.
func (s sessionsService) Current(_ context.Context) (*sessioncap.SessionInfo, error) {
	var out *sessioncap.SessionInfo
	err := s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsCurrent, nil, &out)
	return out, err
}

// Get returns one visible session projection through the registered batch-read method.
func (s sessionsService) Get(ctx context.Context, id sessioncap.SessionID) (*sessioncap.SessionInfo, error) {
	result, err := s.BatchGet(ctx, []sessioncap.SessionID{id})
	if err != nil || result == nil {
		return nil, err
	}
	if item, ok := result.Items[id]; ok {
		return item, nil
	}
	return nil, nil
}

// List returns one bounded visible session page.
func (s sessionsService) List(_ context.Context, input sessioncap.ListInput) (*capmodel.PageResult[*sessioncap.SessionInfo], error) {
	out := &capmodel.PageResult[*sessioncap.SessionInfo]{Items: []*sessioncap.SessionInfo{}}
	err := s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsList, sessionListRequest{
		Username: input.Username,
		IP:       input.IP,
		PageNum:  input.Page.PageNum,
		PageSize: input.Page.PageSize,
	}, out)
	return out, err
}

// BatchGet returns visible sessions and opaque missing IDs.
func (s sessionsService) BatchGet(_ context.Context, ids []sessioncap.SessionID) (*capmodel.BatchResult[*sessioncap.SessionInfo, sessioncap.SessionID], error) {
	out := &capmodel.BatchResult[*sessioncap.SessionInfo, sessioncap.SessionID]{Items: map[sessioncap.SessionID]*sessioncap.SessionInfo{}}
	err := s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsBatchGet, idsRequest{IDs: sessionIDsToStrings(ids)}, out)
	return out, err
}

// BatchGetUserOnlineStatus returns visible users' online status in one bounded call.
func (s sessionsService) BatchGetUserOnlineStatus(_ context.Context, userIDs []string) (*capmodel.BatchResult[*sessioncap.UserOnlineStatus, string], error) {
	out := &capmodel.BatchResult[*sessioncap.UserOnlineStatus, string]{Items: map[string]*sessioncap.UserOnlineStatus{}}
	err := s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsBatchGetUserOnlineStatus, sessionUserOnlineStatusRequest{UserIDs: append([]string(nil), userIDs...)}, out)
	return out, err
}

// EnsureVisible rejects when any requested online session is absent or invisible.
func (s sessionsService) EnsureVisible(_ context.Context, ids []sessioncap.SessionID) error {
	return s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsEnsureVisible, idsRequest{IDs: sessionIDsToStrings(ids)}, nil)
}

// Revoke revokes one visible online session.
func (s sessionsService) Revoke(_ context.Context, id sessioncap.SessionID) error {
	return s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsRevoke, sessionIDRequest{SessionID: string(id)}, nil)
}

// RevokeMany revokes visible online sessions.
func (s sessionsService) RevokeMany(_ context.Context, ids []sessioncap.SessionID) error {
	return s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsRevokeMany, idsRequest{IDs: sessionIDsToStrings(ids)}, nil)
}

// sessionListRequest carries bounded online-session list parameters.
type sessionListRequest struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// sessionUserOnlineStatusRequest carries bounded user online status parameters.
type sessionUserOnlineStatusRequest struct {
	UserIDs []string `json:"userIds"`
}

type sessionIDRequest struct {
	SessionID string `json:"sessionId"`
}

// sessionIDsToStrings converts online-session IDs to transport strings.
func sessionIDsToStrings(ids []sessioncap.SessionID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var _ sessioncap.Service = (*sessionsService)(nil)
