// This file adapts dynamic-plugin organization host-service calls to the
// ordinary orgcap.Service consumer contract. The dispatcher intentionally keeps
// host-internal scope, assignment, workspace projection, and database query
// builder seams out of the dynamic-plugin protocol.

package wasm

import (
	"context"
	"encoding/json"

	"lina-core/pkg/plugin/capability/orgcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchOrgHostService routes one organization host-service method to the
// same ordinary orgcap.Service surface exposed to source plugins.
func dispatchOrgHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := orgServiceForHostCall(hcc)
	if service == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"org host service is not scoped",
		)
	}

	switch method {
	case bridgehostservice.HostServiceMethodOrgAvailable:
		return capabilityJSONResponse(service.Available(ctx))
	case bridgehostservice.HostServiceMethodOrgStatus:
		return capabilityJSONResponse(service.Status(ctx))
	case bridgehostservice.HostServiceMethodOrgListUserDeptAssignments:
		var request intUserIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, request.UserIDs); response != nil {
			return response
		}
		assignments, err := service.ListUserDeptAssignments(ctx, request.UserIDs)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(assignments)
	case bridgehostservice.HostServiceMethodOrgBatchGetUserOrgProfiles:
		var request intUserIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, request.UserIDs); response != nil {
			return response
		}
		result, err := service.BatchGetUserOrgProfiles(ctx, request.UserIDs)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgGetUserDeptInfo:
		var request intUserIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, []int{request.UserID}); response != nil {
			return response
		}
		deptID, deptName, err := service.GetUserDeptInfo(ctx, request.UserID)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(orgUserDeptInfoResponse{DeptID: deptID, DeptName: deptName})
	case bridgehostservice.HostServiceMethodOrgGetUserDeptName:
		var request intUserIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, []int{request.UserID}); response != nil {
			return response
		}
		name, err := service.GetUserDeptName(ctx, request.UserID)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(name)
	case bridgehostservice.HostServiceMethodOrgGetUserDeptIDs:
		var request intUserIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, []int{request.UserID}); response != nil {
			return response
		}
		deptIDs, err := service.GetUserDeptIDs(ctx, request.UserID)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(deptIDs)
	case bridgehostservice.HostServiceMethodOrgGetUserPostIDs:
		var request intUserIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, []int{request.UserID}); response != nil {
			return response
		}
		postIDs, err := service.GetUserPostIDs(ctx, request.UserID)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(postIDs)
	case bridgehostservice.HostServiceMethodOrgListDeptTree:
		var request orgcap.DeptTreeInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.ListDeptTree(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgSearchDepartments:
		var request orgcap.DeptSearchInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.SearchDepartments(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgListPostOptions:
		var request orgcap.PostOptionsInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.ListPostOptionsPage(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgEnsureDepartmentsVisible:
		var request intDeptIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := service.EnsureDepartmentsVisible(ctx, request.DeptIDs)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgEnsurePostsVisible:
		var request intPostIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := service.EnsurePostsVisible(ctx, request.PostIDs)
		return domainCapabilityResult(true, err)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"org host service method not implemented: "+method,
		)
	}
}

// orgServiceForHostCall resolves the organization service for one host call.
func orgServiceForHostCall(hcc *hostCallContext) orgcap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Org()
}

// orgUserDeptInfoResponse carries the tuple returned by orgcap.Service.GetUserDeptInfo.
type orgUserDeptInfoResponse struct {
	// DeptID is the department identifier.
	DeptID int `json:"deptId"`
	// DeptName is the department display name.
	DeptName string `json:"deptName"`
}

// intUserIDRequest carries one integer user identifier.
type intUserIDRequest struct {
	// UserID is the user identifier.
	UserID int `json:"userId"`
}

// intUserIDsRequest carries multiple integer user identifiers.
type intUserIDsRequest struct {
	// UserIDs are the user identifiers.
	UserIDs []int `json:"userIds"`
}

// intDeptIDsRequest carries department identifiers.
type intDeptIDsRequest struct {
	// DeptIDs are the department identifiers.
	DeptIDs []int `json:"deptIds"`
}

// intPostIDsRequest carries post identifiers.
type intPostIDsRequest struct {
	// PostIDs are the post identifiers.
	PostIDs []int `json:"postIds"`
}

// capabilityJSONResponse encodes one capability result as a transport-owned
// JSON response without making pluginbridge own capability DTO definitions.
func capabilityJSONResponse(value any) *bridgehostcall.HostCallResponseEnvelope {
	content, err := json.Marshal(value)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	payload := bridgehostservice.MarshalHostServiceCapabilityJSONResponse(
		&bridgehostservice.HostServiceCapabilityJSONResponse{Value: content},
	)
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}
