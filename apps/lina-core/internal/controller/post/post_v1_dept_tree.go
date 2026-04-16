package post

import (
	"context"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// DeptTree returns department tree structure (with post count)
func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	nodes, err := c.postSvc.DeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{
		List: convertDeptTreeNodes(nodes),
	}, nil
}

// convertDeptTreeNodes converts service layer DeptTreeNode slice to API layer DeptTreeNode slice
func convertDeptTreeNodes(nodes []*postsvc.DeptTreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &v1.DeptTreeNode{
			Id:        n.Id,
			Label:     n.Label,
			PostCount: n.PostCount,
			Children:  convertDeptTreeNodes(n.Children),
		})
	}
	return result
}
