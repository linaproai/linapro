// dept_v1_tree.go implements the controller method that returns department trees.

package dept

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/dept/v1"
	deptsvc "lina-plugin-org-center/backend/internal/service/dept"
)

// Tree returns department tree structure.
func (c *ControllerV1) Tree(ctx context.Context, req *v1.TreeReq) (res *v1.TreeRes, err error) {
	nodes, err := c.deptSvc.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TreeRes{List: convertTreeNodes(nodes)}, nil
}

// convertTreeNodes converts service layer tree nodes to API tree nodes.
func convertTreeNodes(nodes []*deptsvc.TreeNode) []*v1.TreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.TreeNode, 0, len(nodes))
	for _, item := range nodes {
		if item == nil {
			continue
		}
		result = append(result, &v1.TreeNode{Id: item.Id, Label: item.Label, Children: convertTreeNodes(item.Children)})
	}
	return result
}
