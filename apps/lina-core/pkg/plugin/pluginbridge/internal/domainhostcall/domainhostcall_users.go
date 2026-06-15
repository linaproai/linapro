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

// Search searches visible user candidates with bounded paging.
func (s usersService) Search(_ context.Context, _ capmodel.CapabilityContext, input usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	result := &capmodel.PageResult[*usercap.UserProjection]{Items: []*usercap.UserProjection{}}
	err := s.callJSONRequest(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersSearch,
		usersSearchRequest{
			Keyword:  input.Keyword,
			PageNum:  input.Page.PageNum,
			PageSize: input.Page.PageSize,
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

type usersSearchRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	PageNum  int    `json:"pageNum,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
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
