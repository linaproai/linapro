package user

import (
	"context"

	v1 "lina-core/api/user/v1"
	"lina-core/internal/model/entity"
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
		TenantId:       req.TenantId,
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
			UserItem:    userItem(u.SysUser),
			DeptId:      u.DeptId,
			DeptName:    u.DeptName,
			RoleIds:     u.RoleIds,
			RoleNames:   u.RoleNames,
			TenantIds:   u.TenantIds,
			TenantNames: u.TenantNames,
		})
	}
	return &v1.ListRes{
		List:  list,
		Total: out.Total,
	}, nil
}

// userItem maps a user entity to the API-safe response DTO.
func userItem(user *entity.SysUser) v1.UserItem {
	if user == nil {
		return v1.UserItem{}
	}
	return v1.UserItem{
		Id:        user.Id,
		TenantId:  user.TenantId,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
		Sex:       user.Sex,
		Avatar:    user.Avatar,
		Status:    user.Status,
		Remark:    user.Remark,
		LoginDate: user.LoginDate,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
