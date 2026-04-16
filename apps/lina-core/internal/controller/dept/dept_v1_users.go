package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
)

// Users returns the list of users in a department.
func (c *ControllerV1) Users(ctx context.Context, req *v1.UsersReq) (res *v1.UsersRes, err error) {
	users, err := c.deptSvc.Users(ctx, req.Id, req.Keyword, req.Limit)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.DeptUser, 0, len(users))
	for _, u := range users {
		list = append(list, &v1.DeptUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}
	return &v1.UsersRes{
		List: list,
	}, nil
}
