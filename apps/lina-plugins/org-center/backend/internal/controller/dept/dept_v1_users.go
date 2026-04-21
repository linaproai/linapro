// dept_v1_users.go implements the controller method that returns department users.

package dept

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/dept/v1"
)

// Users returns the list of users in a department.
func (c *ControllerV1) Users(ctx context.Context, req *v1.UsersReq) (res *v1.UsersRes, err error) {
	users, err := c.deptSvc.Users(ctx, req.Id, req.Keyword, req.Limit)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.DeptUser, 0, len(users))
	for _, item := range users {
		if item == nil {
			continue
		}
		list = append(list, &v1.DeptUser{Id: item.Id, Username: item.Username, Nickname: item.Nickname})
	}
	return &v1.UsersRes{List: list}, nil
}
