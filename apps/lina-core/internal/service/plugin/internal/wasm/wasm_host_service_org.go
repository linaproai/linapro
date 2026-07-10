// This file adapts dynamic-plugin organization host-service calls to the
// ordinary orgcap.Service consumer contract. The dispatcher intentionally keeps
// host-internal scope, assignment, workspace projection, and database query
// builder seams out of the dynamic-plugin protocol.

package wasm

import (
	"context"

	"lina-core/pkg/plugin/capability/orgcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchOrgHostService routes one organization host-service method to the
// same ordinary orgcap.Service surface exposed to source plugins.
//
//nolint:cyclop // Host-service dispatch is an explicit protocol switch with stable method-level branches.
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
	case bridgehostservice.HostServiceMethodOrgBatchGetUserOrgProfiles:
		var request intUserIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, request.UserIDs); response != nil {
			return response
		}
		result, err := service.Assignment().BatchGetUserProfiles(ctx, request.UserIDs)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgListDeptTree:
		var request orgcap.DeptTreeInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Department().ListTree(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgDepartmentBatchGet:
		var request intDeptIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Department().BatchGet(ctx, request.DeptIDs)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgDepartmentList:
		var request orgcap.DeptListInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Department().List(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgPostBatchGet:
		var request intPostIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Post().BatchGet(ctx, request.PostIDs)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgListPostOptions:
		var request orgcap.PostOptionsInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Post().ListOptions(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgEnsureDepartmentsVisible:
		var request intDeptIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := service.Department().EnsureVisible(ctx, request.DeptIDs)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgEnsurePostsVisible:
		var request intPostIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := service.Post().EnsureVisible(ctx, request.PostIDs)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgDepartmentCreate:
		var request orgcap.DeptCreateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Department().Create(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgDepartmentUpdate:
		var request orgcap.DeptUpdateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Department().Update(ctx, request)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgDepartmentDelete:
		var request orgDeptIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Department().Delete(ctx, request.DeptID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgPostCreate:
		var request orgcap.PostCreateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Post().Create(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodOrgPostUpdate:
		var request orgcap.PostUpdateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Post().Update(ctx, request)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgPostDelete:
		var request orgPostIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Post().Delete(ctx, request.PostID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgAssignmentReplaceByUser:
		var request orgAssignmentReplaceByUserRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, []int{request.UserID}); response != nil {
			return response
		}
		err := service.Assignment().ReplaceByUser(ctx, request.UserID, request.DeptID, append([]int(nil), request.PostIDs...))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodOrgAssignmentCleanupByUser:
		var request intUserIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceOrg, method, []int{request.UserID}); response != nil {
			return response
		}
		err := service.Assignment().CleanupByUser(ctx, request.UserID)
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

type orgDeptIDRequest struct {
	DeptID int `json:"deptId"`
}

type orgPostIDRequest struct {
	PostID int `json:"postId"`
}

type orgAssignmentReplaceByUserRequest struct {
	UserID  int   `json:"userId"`
	DeptID  *int  `json:"deptId,omitempty"`
	PostIDs []int `json:"postIds"`
}
