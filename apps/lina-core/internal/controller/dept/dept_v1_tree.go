package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
	deptsvc "lina-core/internal/service/dept"
)

// Tree returns department tree structure.
func (c *ControllerV1) Tree(ctx context.Context, req *v1.TreeReq) (res *v1.TreeRes, err error) {
	nodes, err := c.deptSvc.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TreeRes{
		List: convertTreeNodes(nodes),
	}, nil
}

// convertTreeNodes converts service layer TreeNode slice to API layer TreeNode slice.
func convertTreeNodes(nodes []*deptsvc.TreeNode) []*v1.TreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.TreeNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &v1.TreeNode{
			Id:       n.Id,
			Label:    n.Label,
			Children: convertTreeNodes(n.Children),
		})
	}
	return result
}
