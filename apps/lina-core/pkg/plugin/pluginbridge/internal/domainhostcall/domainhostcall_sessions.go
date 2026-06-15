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

// Search returns one bounded visible session page.
func (s sessionsService) Search(_ context.Context, _ capmodel.CapabilityContext, input sessioncap.SearchInput) (*capmodel.PageResult[*sessioncap.Projection], error) {
	out := &capmodel.PageResult[*sessioncap.Projection]{Items: []*sessioncap.Projection{}}
	err := s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsSearch, sessionSearchRequest{
		Username: input.Username,
		IP:       input.IP,
		PageNum:  input.Page.PageNum,
		PageSize: input.Page.PageSize,
	}, out)
	return out, err
}

// BatchGet returns visible sessions and opaque missing IDs.
func (s sessionsService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []sessioncap.SessionID) (*capmodel.BatchResult[*sessioncap.Projection, sessioncap.SessionID], error) {
	out := &capmodel.BatchResult[*sessioncap.Projection, sessioncap.SessionID]{Items: map[sessioncap.SessionID]*sessioncap.Projection{}}
	err := s.callJSONRequest(protocol.HostServiceSessions, protocol.HostServiceMethodSessionsBatchGet, idsRequest{IDs: sessionIDsToStrings(ids)}, out)
	return out, err
}

// sessionSearchRequest carries bounded online-session search parameters.
type sessionSearchRequest struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
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
