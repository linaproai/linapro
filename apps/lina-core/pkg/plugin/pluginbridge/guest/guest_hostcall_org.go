// This file implements guest-side organization capability reads that cross the
// pluginbridge host-service transport. The wrapper is compiled for ordinary Go
// tests and wasip1 guests; only the lower-level transport selects the real host
// import or the unsupported stub.

package guest

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// OrgService exposes guest-side organization capability reads.
type OrgService interface {
	// Status returns the current organization capability activation state. ctx is
	// accepted for parity with source-plugin capability services; the current
	// guest transport cannot cancel an in-flight WASI host call. The zero value is
	// returned with an error when transport or response decoding fails.
	Status(ctx context.Context) (capmodel.CapabilityStatus, error)
	// Available reports whether the organization capability has an active
	// provider. It returns false with the Status error when the host call cannot
	// complete or returns an invalid response.
	Available(ctx context.Context) (bool, error)
	// ListUserDeptAssignments returns user-to-department projections for the
	// provided users. Callers should prefer this batch method for lists and
	// exports to avoid plugin-side N+1 host calls.
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error)
	// GetUserDeptInfo returns one user's department identifier and name.
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptName returns one user's department name.
	GetUserDeptName(ctx context.Context, userID int) (string, error)
	// GetUserDeptIDs returns one user's department identifiers.
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// GetUserPostIDs returns one user's post identifiers.
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
}

var _ OrgService = (*orgService)(nil)

// orgService implements guest organization capability reads.
type orgService struct{}

// Status returns the current organization capability activation state.
func (orgService) Status(_ context.Context) (capmodel.CapabilityStatus, error) {
	var status capmodel.CapabilityStatus
	err := invokeCapabilityJSON(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgStatus,
		nil,
		&status,
	)
	return status, err
}

// Available reports whether the organization capability has an active provider.
func (orgService) Available(_ context.Context) (bool, error) {
	var available bool
	err := invokeCapabilityJSON(
		protocol.HostServiceOrg,
		protocol.HostServiceMethodOrgAvailable,
		nil,
		&available,
	)
	return available, err
}

// ListUserDeptAssignments returns user-to-department projections for the provided users.
func (orgService) ListUserDeptAssignments(_ context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	err := invokeCapabilityJSON(
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
func (orgService) GetUserDeptInfo(_ context.Context, userID int) (int, string, error) {
	var info orgUserDeptInfo
	err := invokeCapabilityJSON(
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
func (orgService) GetUserDeptName(_ context.Context, userID int) (string, error) {
	var name string
	err := invokeCapabilityJSON(
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
func (orgService) GetUserDeptIDs(_ context.Context, userID int) ([]int, error) {
	var deptIDs []int
	err := invokeCapabilityJSON(
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
func (orgService) GetUserPostIDs(_ context.Context, userID int) ([]int, error) {
	var postIDs []int
	err := invokeCapabilityJSON(
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
