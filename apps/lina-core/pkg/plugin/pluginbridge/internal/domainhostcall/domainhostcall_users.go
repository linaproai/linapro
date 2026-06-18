// This file implements guest-side user capability reads that cross the
// pluginbridge host-service transport. The wrapper keeps dynamic-plugin code on
// the usercap domain contract without exposing host sys_user storage details.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// usersService adapts user projection reads to host services.
type usersService struct{ baseService }

// Users creates the user-domain capability guest client.
func Users(invoker Invoker) usercap.Service {
	return usersService{baseService: newBaseService(invoker)}
}

// Current returns the current actor's visible user projection.
func (s usersService) Current(_ context.Context, _ capmodel.CapabilityContext) (*usercap.UserProjection, error) {
	var result *usercap.UserProjection
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersCurrent,
		nil,
		&result,
	)
	return result, err
}

// BatchGet returns visible user projections and opaque missing IDs.
func (s usersService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []usercap.UserID) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	result := &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{
		Items:      map[usercap.UserID]*usercap.UserProjection{},
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
func (s usersService) BatchResolve(_ context.Context, _ capmodel.CapabilityContext, input usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey]{
		Items:      map[usercap.ResolveKey]*usercap.UserProjection{},
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

// Search searches visible user candidates with bounded paging.
func (s usersService) Search(_ context.Context, _ capmodel.CapabilityContext, input usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	result := &capmodel.PageResult[*usercap.UserProjection]{Items: []*usercap.UserProjection{}}
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersSearch,
		usersSearchRequest{
			Keyword:     input.Keyword,
			Status:      string(input.Status),
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
func (s usersService) EnsureVisible(_ context.Context, _ capmodel.CapabilityContext, ids []usercap.UserID) error {
	return s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersEnsureVisible,
		usersEnsureVisibleRequest{UserIDs: userIDsToStrings(ids)},
		nil,
	)
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

// userIDsToStrings converts user IDs to transport strings.
func userIDsToStrings(ids []usercap.UserID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var _ usercap.Service = (*usersService)(nil)
