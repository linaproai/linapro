package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// Get returns user details
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	user, err := c.userSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	deptId, deptName, err := c.userSvc.GetUserDeptInfo(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	postIds, err := c.userSvc.GetUserPostIds(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if postIds == nil {
		postIds = []int{}
	}
	roleIds, err := c.userSvc.GetUserRoleIds(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if roleIds == nil {
		roleIds = []int{}
	}
	return &v1.GetRes{
		SysUser:  user,
		DeptId:   deptId,
		DeptName: deptName,
		PostIds:  postIds,
		RoleIds:  roleIds,
	}, nil
}
