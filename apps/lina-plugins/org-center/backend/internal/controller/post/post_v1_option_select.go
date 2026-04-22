// post_v1_option_select.go implements the controller method that returns selectable posts.

package post

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/post/v1"
	postsvc "lina-plugin-org-center/backend/internal/service/post"
)

// OptionSelect returns post option list.
func (c *ControllerV1) OptionSelect(ctx context.Context, req *v1.OptionSelectReq) (res *v1.OptionSelectRes, err error) {
	options, err := c.postSvc.OptionSelect(ctx, postsvc.OptionSelectInput{DeptId: req.DeptId})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.PostOption, 0, len(options))
	for _, item := range options {
		list = append(list, &v1.PostOption{PostId: item.PostId, PostName: item.PostName})
	}
	return &v1.OptionSelectRes{List: list}, nil
}
