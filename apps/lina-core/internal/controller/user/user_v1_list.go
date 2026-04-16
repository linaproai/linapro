package user

import (
	"context"

	v1 "lina-core/api/user/v1"
	usersvc "lina-core/internal/service/user"
)

// List queries user list
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.userSvc.List(ctx, usersvc.ListInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		Username:       req.Username,
		Nickname:       req.Nickname,
		Status:         req.Status,
		Phone:          req.Phone,
		Sex:            req.Sex,
		DeptId:         req.DeptId,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
	})
	if err != nil {
		return nil, err
	}
	// Convert to ListItem with dept and role info
	list := make([]*v1.ListItem, 0, len(out.List))
	for _, u := range out.List {
		list = append(list, &v1.ListItem{
			SysUser:   u.SysUser,
			DeptId:    u.DeptId,
			DeptName:  u.DeptName,
			RoleIds:   u.RoleIds,
			RoleNames: u.RoleNames,
		})
	}
	return &v1.ListRes{
		List:  list,
		Total: out.Total,
	}, nil
}