package user

import (
	"context"

	v1 "lina-core/api/user/v1"
	usersvc "lina-core/internal/service/user"
)

// Create creates a user
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	status := usersvc.StatusNormal
	if req.Status != nil {
		status = usersvc.Status(*req.Status)
	}
	sex := 0
	if req.Sex != nil {
		sex = *req.Sex
	}
	id, err := c.userSvc.Create(ctx, usersvc.CreateInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Sex:      sex,
		Status:   status,
		Remark:   req.Remark,
		DeptId:   req.DeptId,
		PostIds:  req.PostIds,
		RoleIds:  req.RoleIds,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}