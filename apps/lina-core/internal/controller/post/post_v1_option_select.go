package post

import (
	"context"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// OptionSelect returns post option list
func (c *ControllerV1) OptionSelect(ctx context.Context, req *v1.OptionSelectReq) (res *v1.OptionSelectRes, err error) {
	options, err := c.postSvc.OptionSelect(ctx, postsvc.OptionSelectInput{
		DeptId: req.DeptId,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.PostOption, 0, len(options))
	for _, o := range options {
		list = append(list, &v1.PostOption{
			PostId:   o.PostId,
			PostName: o.PostName,
		})
	}
	return &v1.OptionSelectRes{
		List: list,
	}, nil
}
