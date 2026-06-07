// This file implements menu detail retrieval and projects persisted integer
// flags into typed public API contracts.

package menu

import (
	"context"

	v1 "lina-core/api/menu/v1"
	"lina-core/pkg/menutype"
)

// Get returns the detail of the specified menu.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	// Get menu detail
	m, err := c.menuSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Get parent name
	parentName := c.menuSvc.GetParentName(ctx, m.ParentId)

	return &v1.GetRes{
		MenuItem: &v1.MenuItem{
			Id:         m.Id,
			ParentId:   m.ParentId,
			Name:       m.Name,
			Path:       m.Path,
			Component:  m.Component,
			Perms:      m.Perms,
			Icon:       m.Icon,
			Type:       menutype.Code(m.Type),
			Sort:       m.Sort,
			Visible:    statusflagVisibility(m.Visible),
			Status:     statusflagEnabled(m.Status),
			IsFrame:    statusflagYesNo(m.IsFrame),
			IsCache:    statusflagYesNo(m.IsCache),
			QueryParam: m.QueryParam,
			Remark:     m.Remark,
		},
		ParentName: parentName,
	}, nil
}
