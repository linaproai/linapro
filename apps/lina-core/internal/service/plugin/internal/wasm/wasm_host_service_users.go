// This file adapts dynamic-plugin user host-service calls to the ordinary
// usercap.Service contract. The dispatcher exposes only projections, bounded
// search, and visibility checks; user-management commands remain outside the
// dynamic ordinary service surface.

package wasm

import (
	"context"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
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
	case bridgehostservice.HostServiceMethodUsersBatchGet:
		request, err := bridgehostservice.UnmarshalHostServiceUsersBatchGetRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.BatchGetUsers(ctx, capCtx, userIDsFromStrings(request.UserIDs))
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(result)
	case bridgehostservice.HostServiceMethodUsersSearch:
		request, err := bridgehostservice.UnmarshalHostServiceUsersSearchRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.SearchUsers(ctx, capCtx, usercap.SearchInput{
			Keyword: strings.TrimSpace(request.Keyword),
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
		request, err := bridgehostservice.UnmarshalHostServiceUsersEnsureVisibleRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if err = service.EnsureUsersVisible(ctx, capCtx, userIDsFromStrings(request.UserIDs)); err != nil {
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

// usersServiceForHostCall resolves the users service for one host call.
func usersServiceForHostCall(hcc *hostCallContext) usercap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Users()
}

// capabilityContextForHostCall constructs audited domain-call metadata from
// the trusted host-call context.
func capabilityContextForHostCall(hcc *hostCallContext, service string, method string) capmodel.CapabilityContext {
	now := time.Now()
	if hcc == nil {
		return capmodel.CapabilityContext{
			Actor:       capmodel.CapabilityActor{Type: capmodel.ActorTypeSystem, SystemReason: "dynamic plugin host service"},
			Source:      capmodel.CapabilitySourceHost,
			SystemCall:  true,
			Resource:    strings.TrimSpace(service) + "." + strings.TrimSpace(method),
			RequestedAt: now,
		}
	}

	actor := capmodel.CapabilityActor{
		Type:         capmodel.ActorTypeSystem,
		SystemReason: "dynamic plugin host service",
	}
	tenantID := ""
	if hcc.identity != nil {
		tenantID = strconv.Itoa(int(hcc.identity.TenantId))
		if hcc.identity.UserID > 0 {
			actor = capmodel.CapabilityActor{
				Type:   capmodel.ActorTypeUser,
				UserID: int64(hcc.identity.UserID),
				Name:   hcc.identity.Username,
			}
		}
	}
	return capmodel.CapabilityContext{
		PluginID:      strings.TrimSpace(hcc.pluginID),
		Actor:         actor,
		TenantID:      capmodel.DomainID(tenantID),
		Source:        capabilitySourceFromExecution(hcc.executionSource),
		SystemCall:    actor.Type == capmodel.ActorTypeSystem,
		Authorization: capabilityAuthorizationFromHostServices(hcc.hostServices),
		Resource:      strings.TrimSpace(service) + "." + strings.TrimSpace(method),
		TraceID:       strings.TrimSpace(hcc.requestID),
		RequestedAt:   now,
	}
}

func capabilitySourceFromExecution(source bridgecontract.ExecutionSource) capmodel.CapabilitySource {
	switch bridgecontract.NormalizeExecutionSource(source) {
	case bridgecontract.ExecutionSourceRoute:
		return capmodel.CapabilitySourceHTTP
	case bridgecontract.ExecutionSourceHook:
		return capmodel.CapabilitySourceHook
	case bridgecontract.ExecutionSourceJobs:
		return capmodel.CapabilitySourceJobs
	case bridgecontract.ExecutionSourceLifecycle:
		return capmodel.CapabilitySourceLifecycle
	default:
		return capmodel.CapabilitySourceHost
	}
}

func capabilityAuthorizationFromHostServices(specs []*bridgehostservice.HostServiceSpec) capmodel.CapabilityAuthorizationSnapshot {
	authorization := capmodel.CapabilityAuthorizationSnapshot{
		Services:  map[string][]string{},
		Resources: map[string][]string{},
	}
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		service := strings.TrimSpace(spec.Service)
		if service == "" {
			continue
		}
		authorization.Services[service] = append([]string(nil), spec.Methods...)
		if len(spec.Resources) == 0 {
			continue
		}
		for _, resource := range spec.Resources {
			if resource == nil || strings.TrimSpace(resource.Ref) == "" {
				continue
			}
			for _, method := range spec.Methods {
				key := service + "." + method
				authorization.Resources[key] = append(authorization.Resources[key], strings.TrimSpace(resource.Ref))
			}
		}
	}
	return authorization
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
	if err := service.EnsureUsersVisible(ctx, capCtx, userIDsFromInts(userIDs)); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusCapabilityDenied, err)
	}
	return nil
}
