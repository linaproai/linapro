// user_v1_post_options.go implements the host-owned post-options endpoint used
// by user management when the optional organization plugin is available.

package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// PostOptions returns selectable post options for the requested department subtree.
func (c *ControllerV1) PostOptions(ctx context.Context, req *v1.PostOptionsReq) (res *v1.PostOptionsRes, err error) {
	options, err := c.orgCapSvc.ListPostOptions(ctx, req.DeptId)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.UserPostOption, 0, len(options))
	for _, option := range options {
		if option == nil {
			continue
		}
		items = append(items, &v1.UserPostOption{
			PostId:   option.PostID,
			PostName: option.PostName,
		})
	}
	return &v1.PostOptionsRes{List: items}, nil
}
