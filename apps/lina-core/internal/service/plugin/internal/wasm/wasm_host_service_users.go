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
	"lina-core/pkg/statusflag"
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
	switch method {
	case bridgehostservice.HostServiceMethodUsersCurrent:
		result, err := service.Current(ctx)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersBatchGet:
		request, err := decodeUsersHostServiceRequest[usersBatchGetRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.BatchGet(ctx, userIDsFromStrings(request.UserIDs))
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersBatchResolve:
		request, err := decodeUsersHostServiceRequest[usersBatchResolveRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.BatchResolve(ctx, usercap.BatchResolveInput{
			IDs:       userIDsFromStrings(request.UserIDs),
			Usernames: append([]string(nil), request.Usernames...),
			Contacts:  append([]string(nil), request.Contacts...),
		})
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersList:
		request, err := decodeUsersHostServiceRequest[usersListRequest](payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.List(ctx, usercap.ListInput{
			Keyword:     strings.TrimSpace(request.Keyword),
			Status:      userStatusFlag(request.Status),
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
		if err = service.EnsureVisible(ctx, userIDsFromStrings(request.UserIDs)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodUsersCreate:
		var request usercap.CreateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Create(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodUsersUpdate:
		var request usercap.UpdateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Update(ctx, request)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodUsersDelete:
		var request userIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Delete(ctx, usercap.UserID(strings.TrimSpace(request.UserID)))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodUsersSetStatus:
		var request usersSetStatusRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.SetStatus(ctx, usercap.UserID(strings.TrimSpace(request.UserID)), statusflag.Enabled(request.Status))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodUsersResetPassword:
		var request usersResetPasswordRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.ResetPassword(ctx, usercap.UserID(strings.TrimSpace(request.UserID)), request.Password)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodUsersReplaceRoles:
		var request usersReplaceRolesRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		assignment := service.Assignment()
		if assignment == nil {
			return domainServiceNotScoped("users.assignment")
		}
		err := assignment.ReplaceRoles(ctx, usercap.UserID(strings.TrimSpace(request.UserID)), append([]int(nil), request.RoleIDs...))
		return domainCapabilityResult(true, err)
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

type usersListRequest struct {
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

type usersSetStatusRequest struct {
	UserID string `json:"userId"`
	Status int    `json:"status"`
}

type usersResetPasswordRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
}

type usersReplaceRolesRequest struct {
	UserID  string `json:"userId"`
	RoleIDs []int  `json:"roleIds"`
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

// userStatusFlag converts the wire string status into the shared optional status flag.
func userStatusFlag(status string) *statusflag.Enabled {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		value := statusflag.Enabled(-1)
		return &value
	}
	value := statusflag.Enabled(parsed)
	return &value
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
	if err := service.EnsureVisible(ctx, userIDsFromInts(userIDs)); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusCapabilityDenied, err)
	}
	return nil
}
