// This file adapts dynamic-plugin user host-service calls to the ordinary
// usercap.Service contract. The dispatcher exposes only projections, bounded
// search, and visibility checks; user-management commands remain outside the
// dynamic ordinary service surface.

package wasm

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchUsersHostService routes one users-domain host-service method to the same
// ordinary usercap.Service surface exposed to source plugins.
func dispatchUsersHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := usersServiceForHostCall(hcc)
	if service == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"users host service is not scoped",
		)
	}

	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceUsers, method)
	switch method {
	case bridgehostservice.HostServiceMethodUsersCurrent:
		result, err := service.Current(ctx, capCtx)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersBatchGet:
		request, err := decodeUsersHostServiceRequest[usersBatchGetRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.BatchGet(ctx, capCtx, userIDsFromStrings(request.UserIDs))
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersBatchResolve:
		request, err := decodeUsersHostServiceRequest[usersBatchResolveRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.BatchResolve(ctx, capCtx, usercap.BatchResolveInput{
			IDs:       userIDsFromStrings(request.UserIDs),
			Usernames: append([]string(nil), request.Usernames...),
			Contacts:  append([]string(nil), request.Contacts...),
		})
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersSearch:
		request, err := decodeUsersHostServiceRequest[usersSearchRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Search(ctx, capCtx, usercap.SearchInput{
			Keyword:     strings.TrimSpace(request.Keyword),
			Status:      usercap.Status(strings.TrimSpace(request.Status)),
			TenantID:    capmodel.DomainID(strings.TrimSpace(request.TenantID)),
			EnabledOnly: request.EnabledOnly,
			Page: capmodel.PageRequest{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		})
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersEnsureVisible:
		request, err := decodeUsersHostServiceRequest[usersEnsureVisibleRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if err = service.EnsureVisible(ctx, capCtx, userIDsFromStrings(request.UserIDs)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"users host service method not implemented: "+method,
		)
	}
}

type usersBatchGetRequest struct {
	UserIDs []string `json:"userIds"`
}

type usersBatchResolveRequest struct {
	UserIDs   []string `json:"userIds,omitempty"`
	Usernames []string `json:"usernames,omitempty"`
	Contacts  []string `json:"contacts,omitempty"`
}

type usersSearchRequest struct {
	Keyword     string `json:"keyword,omitempty"`
	Status      string `json:"status,omitempty"`
	TenantID    string `json:"tenantId,omitempty"`
	EnabledOnly bool   `json:"enabledOnly,omitempty"`
	PageNum     int    `json:"pageNum,omitempty"`
	PageSize    int    `json:"pageSize,omitempty"`
}

type usersEnsureVisibleRequest struct {
	UserIDs []string `json:"userIds"`
}

func decodeUsersHostServiceRequest[T any](payload []byte) (*T, error) {
	request, err := bridgehostservice.UnmarshalHostServiceJSONRequest(payload)
	if err != nil {
		return nil, err
	}
	out := new(T)
	if len(request.Value) == 0 {
		return out, nil
	}
	if err = json.Unmarshal(request.Value, out); err != nil {
		return nil, err
	}
	return out, nil
}

// usersServiceForHostCall resolves the users service for one host call.
func usersServiceForHostCall(hcc *hostCallContext) usercap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Users()
}

func userIDsFromStrings(ids []string) []usercap.UserID {
	out := make([]usercap.UserID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, usercap.UserID(value))
		}
	}
	return out
}

func userIDsFromInts(ids []int) []usercap.UserID {
	out := make([]usercap.UserID, 0, len(ids))
	for _, id := range ids {
		out = append(out, usercap.UserID(strconv.Itoa(id)))
	}
	return out
}

func ensureHostCallUsersVisible(
	ctx context.Context,
	hcc *hostCallContext,
	serviceName string,
	method string,
	userIDs []int,
) *bridgehostcall.HostCallResponseEnvelope {
	service := usersServiceForHostCall(hcc)
	if service == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"users host service is not scoped",
		)
	}
	capCtx := capabilityContextForHostCall(hcc, serviceName, method)
	if err := service.EnsureVisible(ctx, capCtx, userIDsFromInts(userIDs)); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusCapabilityDenied, err)
	}
	return nil
}
