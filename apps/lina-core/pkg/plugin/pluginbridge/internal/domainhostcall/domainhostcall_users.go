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

// BatchGetUsers returns visible user projections and opaque missing IDs.
func (s usersService) BatchGetUsers(_ context.Context, _ capmodel.CapabilityContext, ids []usercap.UserID) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	result := &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{
		Items:      map[usercap.UserID]*usercap.UserProjection{},
		MissingIDs: []usercap.UserID{},
	}
	err := s.call(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersBatchGet,
		protocol.MarshalHostServiceUsersBatchGetRequest(
			&protocol.HostServiceUsersBatchGetRequest{UserIDs: userIDsToStrings(ids)},
		),
		result,
	)
	return result, err
}

// SearchUsers searches visible user candidates with bounded paging.
func (s usersService) SearchUsers(_ context.Context, _ capmodel.CapabilityContext, input usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	result := &capmodel.PageResult[*usercap.UserProjection]{Items: []*usercap.UserProjection{}}
	err := s.call(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersSearch,
		protocol.MarshalHostServiceUsersSearchRequest(
			&protocol.HostServiceUsersSearchRequest{
				Keyword:  input.Keyword,
				PageNum:  input.Page.PageNum,
				PageSize: input.Page.PageSize,
			},
		),
		result,
	)
	return result, err
}

// EnsureUsersVisible rejects when any requested user is absent or invisible.
func (s usersService) EnsureUsersVisible(_ context.Context, _ capmodel.CapabilityContext, ids []usercap.UserID) error {
	return s.call(
		protocol.HostServiceUsers,
		protocol.HostServiceMethodUsersEnsureVisible,
		protocol.MarshalHostServiceUsersEnsureVisibleRequest(
			&protocol.HostServiceUsersEnsureVisibleRequest{UserIDs: userIDsToStrings(ids)},
		),
		nil,
	)
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
