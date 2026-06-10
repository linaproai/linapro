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
	err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgListUserDeptAssignments,
		protocol.MarshalHostServiceCapabilityUsersRequest(
			&protocol.HostServiceCapabilityUsersRequest{UserIDs: userIDs},
		),
		&assignments,
	)
	return assignments, err
}

// GetUserDeptInfo returns one user's department identifier and name.
func (s orgService) GetUserDeptInfo(_ context.Context, userID int) (int, string, error) {
	var info orgUserDeptInfo
	err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptInfo,
		protocol.MarshalHostServiceCapabilityUserRequest(
			&protocol.HostServiceCapabilityUserRequest{UserID: userID},
		),
		&info,
	)
	return info.DeptID, info.DeptName, err
}

// GetUserDeptName returns one user's department name.
func (s orgService) GetUserDeptName(_ context.Context, userID int) (string, error) {
	var name string
	err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptName,
		protocol.MarshalHostServiceCapabilityUserRequest(
			&protocol.HostServiceCapabilityUserRequest{UserID: userID},
		),
		&name,
	)
	return name, err
}

// GetUserDeptIDs returns one user's department identifiers.
func (s orgService) GetUserDeptIDs(_ context.Context, userID int) ([]int, error) {
	var deptIDs []int
	err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserDeptIDs,
		protocol.MarshalHostServiceCapabilityUserRequest(
			&protocol.HostServiceCapabilityUserRequest{UserID: userID},
		),
		&deptIDs,
	)
	return deptIDs, err
}

// GetUserPostIDs returns one user's post identifiers.
func (s orgService) GetUserPostIDs(_ context.Context, userID int) ([]int, error) {
	var postIDs []int
	err := s.call(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgGetUserPostIDs,
		protocol.MarshalHostServiceCapabilityUserRequest(
			&protocol.HostServiceCapabilityUserRequest{UserID: userID},
		),
		&postIDs,
	)
	return postIDs, err
}

// orgUserDeptInfo carries the tuple returned by orgcap.Service.GetUserDeptInfo.
type orgUserDeptInfo struct {
	// DeptID is the department identifier.
	DeptID int `json:"deptId"`
	// DeptName is the department display name.
	DeptName string `json:"deptName"`
}

var _ orgcap.Service = (*orgService)(nil)
