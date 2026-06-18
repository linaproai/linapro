// This file implements guest-side organization capability reads that cross the
// pluginbridge host-service transport. Status reads follow source-plugin
// fallback semantics by returning zero values when transport is unavailable.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// orgService adapts organization capability reads to host services.
type orgService struct{ baseService }

// Org creates the organization capability guest client.
func Org(invoker Invoker) orgcap.Service {
	return orgService{baseService: newBaseService(invoker)}
}

// Status returns the current organization capability activation state.
func (s orgService) Status(_ context.Context) capmodel.CapabilityStatus {
	var status capmodel.CapabilityStatus
	if err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgStatus,
		nil,
		&status,
	); err != nil {
		return capmodel.CapabilityStatus{}
	}
	return status
}

// Available reports whether the organization capability has an active provider.
func (s orgService) Available(_ context.Context) bool {
	var available bool
	if err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgAvailable,
		nil,
		&available,
	); err != nil {
		return false
	}
	return available
}

// ListUserDeptAssignments returns user-to-department projections for the provided users.
func (s orgService) ListUserDeptAssignments(_ context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListUserDeptAssignments,
		intUserIDsRequest{UserIDs: userIDs},
		&assignments,
	)
	return assignments, err
}

// BatchGetUserOrgProfiles returns stable organization profiles for provided users.
func (s orgService) BatchGetUserOrgProfiles(_ context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	out := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{Items: map[int]*orgcap.UserOrgProfile{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgBatchGetUserOrgProfiles,
		intUserIDsRequest{UserIDs: userIDs},
		out,
	)
	return out, err
}

// GetUserDeptInfo returns one user's department identifier and name.
func (s orgService) GetUserDeptInfo(_ context.Context, userID int) (int, string, error) {
	var info orgUserDeptInfo
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptInfo,
		intUserIDRequest{UserID: userID},
		&info,
	)
	return info.DeptID, info.DeptName, err
}

// GetUserDeptName returns one user's department name.
func (s orgService) GetUserDeptName(_ context.Context, userID int) (string, error) {
	var name string
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptName,
		intUserIDRequest{UserID: userID},
		&name,
	)
	return name, err
}

// GetUserDeptIDs returns one user's department identifiers.
func (s orgService) GetUserDeptIDs(_ context.Context, userID int) ([]int, error) {
	var deptIDs []int
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptIDs,
		intUserIDRequest{UserID: userID},
		&deptIDs,
	)
	return deptIDs, err
}

// GetUserPostIDs returns one user's post identifiers.
func (s orgService) GetUserPostIDs(_ context.Context, userID int) ([]int, error) {
	var postIDs []int
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserPostIDs,
		intUserIDRequest{UserID: userID},
		&postIDs,
	)
	return postIDs, err
}

// ListDeptTree returns a bounded department tree projection.
func (s orgService) ListDeptTree(_ context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	out := &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListDeptTree,
		input,
		out,
	)
	return out, err
}

// SearchDepartments returns bounded department candidates.
func (s orgService) SearchDepartments(_ context.Context, input orgcap.DeptSearchInput) (*capmodel.PageResult[*orgcap.DeptProjection], error) {
	out := &capmodel.PageResult[*orgcap.DeptProjection]{Items: []*orgcap.DeptProjection{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgSearchDepartments,
		input,
		out,
	)
	return out, err
}

// ListPostOptionsPage returns bounded post candidates.
func (s orgService) ListPostOptionsPage(_ context.Context, input orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	out := &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListPostOptions,
		input,
		out,
	)
	return out, err
}

// EnsureDepartmentsVisible verifies all department identifiers are visible.
func (s orgService) EnsureDepartmentsVisible(_ context.Context, deptIDs []int) error {
	return s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgEnsureDepartmentsVisible,
		intDeptIDsRequest{DeptIDs: deptIDs},
		nil,
	)
}

// EnsurePostsVisible verifies all post identifiers are visible.
func (s orgService) EnsurePostsVisible(_ context.Context, postIDs []int) error {
	return s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgEnsurePostsVisible,
		intPostIDsRequest{PostIDs: postIDs},
		nil,
	)
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

// orgUserDeptInfo carries the tuple returned by orgcap.Service.GetUserDeptInfo.
type orgUserDeptInfo struct {
	// DeptID is the department identifier.
	DeptID int `json:"deptId"`
	// DeptName is the department display name.
	DeptName string `json:"deptName"`
}

var _ orgcap.Service = (*orgService)(nil)
