// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package role

import (
	"context"

	"lina-core/api/role/v1"
)

type IRoleV1 interface {
	RoleAssignUsers(ctx context.Context, req *v1.RoleAssignUsersReq) (res *v1.RoleAssignUsersRes, err error)
	RoleCreate(ctx context.Context, req *v1.RoleCreateReq) (res *v1.RoleCreateRes, err error)
	RoleDelete(ctx context.Context, req *v1.RoleDeleteReq) (res *v1.RoleDeleteRes, err error)
	RoleGet(ctx context.Context, req *v1.RoleGetReq) (res *v1.RoleGetRes, err error)
	RoleList(ctx context.Context, req *v1.RoleListReq) (res *v1.RoleListRes, err error)
	RoleOptions(ctx context.Context, req *v1.RoleOptionsReq) (res *v1.RoleOptionsRes, err error)
	RoleStatus(ctx context.Context, req *v1.RoleStatusReq) (res *v1.RoleStatusRes, err error)
	RoleUnassignUser(ctx context.Context, req *v1.RoleUnassignUserReq) (res *v1.RoleUnassignUserRes, err error)
	RoleUnassignUsers(ctx context.Context, req *v1.RoleUnassignUsersReq) (res *v1.RoleUnassignUsersRes, err error)
	RoleUpdate(ctx context.Context, req *v1.RoleUpdateReq) (res *v1.RoleUpdateRes, err error)
	RoleUsers(ctx context.Context, req *v1.RoleUsersReq) (res *v1.RoleUsersRes, err error)
}
