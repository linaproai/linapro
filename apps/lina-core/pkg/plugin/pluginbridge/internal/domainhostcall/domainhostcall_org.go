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
func (s orgAssignmentService) BatchListByUsers(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	result, err := s.BatchGetUserProfiles(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return assignments, nil
	}
	for userID, profile := range result.Items {
		if profile == nil || profile.DeptID <= 0 {
			continue
		}
		assignments[userID] = &orgcap.UserDeptAssignment{DeptID: profile.DeptID, DeptName: profile.DeptName}
	}
	return assignments, nil
}

// BatchGetUserOrgProfiles returns stable organization profiles for provided users.
func (s orgAssignmentService) BatchGetUserProfiles(_ context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	out := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{Items: map[int]*orgcap.UserOrgProfile{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgBatchGetUserOrgProfiles,
		intUserIDsRequest{UserIDs: userIDs},
		out,
	)
	return out, err
}

// ListByUser returns one user's organization profile.
func (s orgAssignmentService) ListByUser(ctx context.Context, userID int) (*orgcap.UserOrgProfile, error) {
	result, err := s.BatchGetUserProfiles(ctx, []int{userID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[userID] == nil {
		return &orgcap.UserOrgProfile{UserID: userID}, nil
	}
	return result.Items[userID], nil
}

// GetUserDeptInfo returns one user's department identifier and name.
func (s orgAssignmentService) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	profile, err := s.ListByUser(ctx, userID)
	if err != nil {
		return 0, "", err
	}
	if profile == nil {
		return 0, "", nil
	}
	return profile.DeptID, profile.DeptName, nil
}

// GetUserDeptIDs returns one user's department identifiers.
func (s orgAssignmentService) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	profile, err := s.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.DeptID <= 0 {
		return []int{}, nil
	}
	return []int{profile.DeptID}, nil
}

// GetUserPostIDs returns one user's post identifiers.
func (s orgAssignmentService) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	profile, err := s.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil || len(profile.PostIDs) == 0 {
		return []int{}, nil
	}
	return append([]int(nil), profile.PostIDs...), nil
}

// ListDeptTree returns a bounded department tree projection.
func (s orgDepartmentService) ListTree(_ context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	out := &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListDeptTree,
		input,
		out,
	)
	return out, err
}

// List returns bounded department candidates.
func (s orgDepartmentService) List(_ context.Context, input orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	out := &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgDepartmentList,
		input,
		out,
	)
	return out, err
}

// ListPostOptionsPage returns bounded post candidates.
func (s orgPostService) ListOptions(_ context.Context, input orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	out := &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListPostOptions,
		input,
		out,
	)
	return out, err
}

// Department returns department resource operations.
func (s orgService) Department() orgcap.DepartmentService {
	return orgDepartmentService{baseService: s.baseService}
}

// Post returns post resource operations.
func (s orgService) Post() orgcap.PostService {
	return orgPostService{baseService: s.baseService}
}

// Assignment returns user organization assignment operations.
func (s orgService) Assignment() orgcap.AssignmentService {
	return orgAssignmentService{baseService: s.baseService}
}

// orgDepartmentService adapts department subresource reads to host services.
type orgDepartmentService struct{ baseService }

// orgPostService adapts post subresource reads to host services.
type orgPostService struct{ baseService }

// orgAssignmentService adapts assignment subresource reads to host services.
type orgAssignmentService struct{ baseService }

// Get returns one visible department projection.
func (s orgDepartmentService) Get(ctx context.Context, deptID int) (*orgcap.DeptInfo, error) {
	result, err := s.BatchGet(ctx, []int{deptID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[deptID] == nil {
		return nil, nil
	}
	return result.Items[deptID], nil
}

// BatchGet returns visible department projections and opaque missing IDs.
func (s orgDepartmentService) BatchGet(_ context.Context, deptIDs []int) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	out := &capmodel.BatchResult[*orgcap.DeptInfo, int]{Items: map[int]*orgcap.DeptInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgDepartmentBatchGet,
		intDeptIDsRequest{DeptIDs: deptIDs},
		out,
	)
	return out, err
}

// ListOptions returns bounded department option projections.
func (s orgDepartmentService) ListOptions(ctx context.Context, input orgcap.DeptOptionsInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return s.List(ctx, orgcap.DeptListInput{Keyword: input.Keyword, Status: input.Status, Page: input.Page})
}

// EnsureVisible verifies all department identifiers are visible.
func (s orgDepartmentService) EnsureVisible(_ context.Context, deptIDs []int) error {
	return s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgEnsureDepartmentsVisible,
		intDeptIDsRequest{DeptIDs: deptIDs},
		nil,
	)
}

// Create creates one department.
func (s orgDepartmentService) Create(context.Context, orgcap.DeptCreateInput) (int, error) {
	return 0, unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// Update updates one department.
func (s orgDepartmentService) Update(context.Context, orgcap.DeptUpdateInput) error {
	return unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// Delete deletes one department.
func (s orgDepartmentService) Delete(context.Context, int) error {
	return unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// Get returns one visible post projection.
func (s orgPostService) Get(ctx context.Context, postID int) (*orgcap.PostInfo, error) {
	result, err := s.BatchGet(ctx, []int{postID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[postID] == nil {
		return nil, nil
	}
	return result.Items[postID], nil
}

// BatchGet returns visible post projections and opaque missing IDs.
func (s orgPostService) BatchGet(_ context.Context, postIDs []int) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	out := &capmodel.BatchResult[*orgcap.PostInfo, int]{Items: map[int]*orgcap.PostInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgPostBatchGet,
		intPostIDsRequest{PostIDs: postIDs},
		out,
	)
	return out, err
}

// List returns bounded post candidates.
func (s orgPostService) List(context.Context, orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{}}, unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// EnsureVisible verifies all post identifiers are visible.
func (s orgPostService) EnsureVisible(_ context.Context, postIDs []int) error {
	return s.callJSONRequest(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgEnsurePostsVisible,
		intPostIDsRequest{PostIDs: postIDs},
		nil,
	)
}

// Create creates one post.
func (s orgPostService) Create(context.Context, orgcap.PostCreateInput) (int, error) {
	return 0, unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// Update updates one post.
func (s orgPostService) Update(context.Context, orgcap.PostUpdateInput) error {
	return unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// Delete deletes one post.
func (s orgPostService) Delete(context.Context, int) error {
	return unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// ReplaceByUser rewrites one user's department and post associations.
func (s orgAssignmentService) ReplaceByUser(context.Context, int, *int, []int) error {
	return unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
}

// CleanupByUser deletes one user's optional organization associations.
func (s orgAssignmentService) CleanupByUser(context.Context, int) error {
	return unsupportedDynamicMethodError(orgcap.CapabilityOrgV1)
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

var _ orgcap.Service = (*orgService)(nil)
