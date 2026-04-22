// post_v1_dept_tree.go implements the controller method that returns post department trees.

package post

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/post/v1"
	postsvc "lina-plugin-org-center/backend/internal/service/post"
)

// DeptTree returns department tree structure with post counts.
func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	nodes, err := c.postSvc.DeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{List: convertDeptTreeNodes(nodes)}, nil
}

// convertDeptTreeNodes converts service tree nodes to API tree nodes.
func convertDeptTreeNodes(nodes []*postsvc.DeptTreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, item := range nodes {
		if item == nil {
			continue
		}
		result = append(result, &v1.DeptTreeNode{
			Id:        item.Id,
			Label:     item.Label,
			PostCount: item.PostCount,
			Children:  convertDeptTreeNodes(item.Children),
		})
	}
	return result
}
