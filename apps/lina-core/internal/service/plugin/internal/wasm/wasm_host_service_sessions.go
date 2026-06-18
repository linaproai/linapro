// This file adapts online-session host-service calls to the shared session
// capability service.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/sessioncap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchSessionsHostService routes online-session domain host-service calls.
func dispatchSessionsHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := sessionsServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("sessions")
	}
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceSessions, method)
	switch method {
	case bridgehostservice.HostServiceMethodSessionsCurrent:
		result, err := service.Current(ctx, capCtx)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodSessionsSearch:
		var request sessionSearchRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Search(ctx, capCtx, sessioncap.SearchInput{
			Username: request.Username,
			IP:       request.IP,
			Page: capmodel.PageRequest{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodSessionsBatchGet:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGet(ctx, capCtx, sessionIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodSessionsBatchGetUserOnlineStatus:
		var request sessionUserOnlineStatusRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGetUserOnlineStatus(ctx, capCtx, append([]string(nil), request.UserIDs...))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodSessionsEnsureVisible:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.EnsureVisible(ctx, capCtx, sessionIDs(request.IDs))
		return domainCapabilityResult(true, err)
	default:
		return domainMethodNotFound("sessions", method)
	}
}

// sessionsServiceForHostCall resolves the online-session service for one host call.
func sessionsServiceForHostCall(hcc *hostCallContext) sessioncap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Sessions()
}

// sessionSearchRequest carries online-session search filters.
type sessionSearchRequest struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// sessionUserOnlineStatusRequest carries bounded user online status parameters.
type sessionUserOnlineStatusRequest struct {
	UserIDs []string `json:"userIds"`
}

// sessionIDs converts transport string identifiers into typed session IDs.
func sessionIDs(ids []string) []sessioncap.SessionID {
	out := make([]sessioncap.SessionID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, sessioncap.SessionID(value))
		}
	}
	return out
}
