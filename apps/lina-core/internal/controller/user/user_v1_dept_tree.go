package user

import (
	"context"
	"fmt"

	v1 "lina-core/api/user/v1"
	"lina-core/internal/service/orgcap"
)

// DeptTree returns user department tree structure
func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	nodes, err := c.orgCapSvc.UserDeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{List: c.convertDeptTreeNodes(ctx, nodes)}, nil
}

// convertDeptTreeNodes converts the host-owned orgcap tree projection into the API DTO layer.
func (c *ControllerV1) convertDeptTreeNodes(ctx context.Context, nodes []*orgcap.DeptTreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		label := n.Label
		if n.LabelKey != "" && c.i18nSvc != nil {
			label = fmt.Sprintf("%s (%d)", c.i18nSvc.Translate(ctx, n.LabelKey, n.Label), n.UserCount)
		}
		result = append(result, &v1.DeptTreeNode{
			Id:        n.Id,
			Label:     label,
			UserCount: n.UserCount,
			Children:  c.convertDeptTreeNodes(ctx, n.Children),
		})
	}
	return result
}
