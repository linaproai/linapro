package user

import (
	"context"

	v1 "lina-core/api/user/v1"
	"lina-core/internal/service/orgcap"
)

// DeptTree returns user department tree structure
func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	nodes, err := c.orgCapSvc.UserDeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{List: convertDeptTreeNodes(nodes)}, nil
}

// convertDeptTreeNodes converts the host-owned orgcap tree projection into the API DTO layer.
func convertDeptTreeNodes(nodes []*orgcap.DeptTreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		result = append(result, &v1.DeptTreeNode{
			Id:        n.Id,
			Label:     n.Label,
			UserCount: n.UserCount,
			Children:  convertDeptTreeNodes(n.Children),
		})
	}
	return result
}
