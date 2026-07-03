// This file implements guest-side user capability reads that cross the
// pluginbridge host-service transport. The wrapper keeps dynamic-plugin code on
// the usercap domain contract without exposing host sys_user storage details.

package domainhostcall

import (
	"context"
	"strconv"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// usersService adapts user projection reads to host services.
type usersService struct{ baseService }

// usersAssignmentService adapts user-role assignment operations to host services.
type usersAssignmentService struct{ baseService }

// Users creates the user-domain capability guest client.
func Users(invoker Invoker) usercap.Service {
	return usersService{baseService: newBaseService(invoker)}
}

// Current returns the current actor's visible user projection.
func (s usersService) Current(_ context.Context) (*usercap.UserInfo, error) {
	var result *usercap.UserInfo
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersCurrent,
		nil,
		&result,
	)
	return result, err
}

// BatchGet returns visible user projections and opaque missing IDs.
func (s usersService) BatchGet(_ context.Context, ids []usercap.UserID) (*capmodel.BatchResult[*usercap.UserInfo, usercap.UserID], error) {
	result := &capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]{
		Items:      map[usercap.UserID]*usercap.UserInfo{},
		MissingIDs: []usercap.UserID{},
	}
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersBatchGet,
		usersBatchGetRequest{UserIDs: userIDsToStrings(ids)},
		result,
	)
	return result, err
}

// BatchResolve resolves visible users by IDs, usernames, email addresses, or phone numbers.
func (s usersService) BatchResolve(_ context.Context, input usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey]{
		Items:      map[usercap.ResolveKey]*usercap.UserInfo{},
		MissingIDs: []usercap.ResolveKey{},
	}
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersBatchResolve,
		usersBatchResolveRequest{
			UserIDs:   userIDsToStrings(input.IDs),
			Usernames: append([]string(nil), input.Usernames...),
			Contacts:  append([]string(nil), input.Contacts...),
		},
		result,
	)
	return result, err
}

// Get returns one visible user projection through the registered batch-read method.
func (s usersService) Get(ctx context.Context, id usercap.UserID) (*usercap.UserInfo, error) {
	result, err := s.BatchGet(ctx, []usercap.UserID{id})
	if err != nil || result == nil {
		return nil, err
	}
	if item, ok := result.Items[id]; ok {
		return item, nil
	}
	return nil, nil
}

// List lists visible user candidates with bounded paging.
func (s usersService) List(_ context.Context, input usercap.ListInput) (*capmodel.PageResult[*usercap.UserInfo], error) {
	result := &capmodel.PageResult[*usercap.UserInfo]{Items: []*usercap.UserInfo{}}
	status := ""
	if input.Status != nil {
		status = strconv.Itoa(int(*input.Status))
	}
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersList,
		usersListRequest{
			Keyword:     input.Keyword,
			Status:      status,
			TenantID:    string(input.TenantID),
			EnabledOnly: input.EnabledOnly,
			PageNum:     input.Page.PageNum,
			PageSize:    input.Page.PageSize,
		},
		result,
	)
	return result, err
}

// EnsureVisible rejects when any requested user is absent or invisible.
func (s usersService) EnsureVisible(_ context.Context, ids []usercap.UserID) error {
	return s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersEnsureVisible,
		usersEnsureVisibleRequest{UserIDs: userIDsToStrings(ids)},
		nil,
	)
}

// Create creates one governed user through the dynamic users host service.
func (s usersService) Create(_ context.Context, input usercap.CreateInput) (usercap.UserID, error) {
	var out usercap.UserID
	err := s.callJSONRequest(protocol.HostServiceUsers, protocol.HostServiceMethodUsersCreate, input, &out)
	return out, err
}

// Update updates one visible user through the dynamic users host service.
func (s usersService) Update(_ context.Context, input usercap.UpdateInput) error {
	return s.callJSONRequest(protocol.HostServiceUsers, protocol.HostServiceMethodUsersUpdate, input, nil)
}

// Delete deletes one visible user through the dynamic users host service.
func (s usersService) Delete(_ context.Context, id usercap.UserID) error {
	return s.callJSONRequest(protocol.HostServiceUsers, protocol.HostServiceMethodUsersDelete, userIDRequest{UserID: string(id)}, nil)
}

// SetStatus changes one visible user's lifecycle status.
func (s usersService) SetStatus(_ context.Context, id usercap.UserID, status statusflag.Enabled) error {
	return s.callJSONRequest(protocol.HostServiceUsers, protocol.HostServiceMethodUsersSetStatus, usersSetStatusRequest{
		UserID: string(id),
		Status: int(status),
	}, nil)
}

// ResetPassword resets one visible user's password.
func (s usersService) ResetPassword(_ context.Context, id usercap.UserID, password string) error {
	return s.callJSONRequest(protocol.HostServiceUsers, protocol.HostServiceMethodUsersResetPassword, usersResetPasswordRequest{
		UserID:   string(id),
		Password: password,
	}, nil)
}

// Assignment returns user role assignment operations.
func (s usersService) Assignment() usercap.AssignmentService {
	return usersAssignmentService{baseService: s.baseService}
}

// ReplaceRoles replaces one visible user's role assignments.
func (s usersAssignmentService) ReplaceRoles(_ context.Context, id usercap.UserID, roleIDs []int) error {
	return s.callJSONRequest(protocol.HostServiceUsers, protocol.HostServiceMethodUsersReplaceRoles, usersReplaceRolesRequest{
		UserID:  string(id),
		RoleIDs: append([]int(nil), roleIDs...),
	}, nil)
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

// userIDsToStrings converts user IDs to transport strings.
func userIDsToStrings(ids []usercap.UserID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var _ usercap.Service = (*usersService)(nil)
var _ usercap.AssignmentService = (*usersAssignmentService)(nil)
