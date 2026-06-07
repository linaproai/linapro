// This file maps dictionary data and type list projections into public API DTOs,
// including shared status and tenant-override contract fields.

package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/pkg/apitime"
	"lina-core/pkg/fallbackoverride"
	"lina-core/pkg/statusflag"
)

// DataList returns dictionary data list.
func (c *ControllerV1) DataList(ctx context.Context, req *v1.DataListReq) (res *v1.DataListRes, err error) {
	out, err := c.dictSvc.DataList(ctx, dictsvc.DataListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		DictType: req.DictType,
		Label:    req.Label,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.DictDataItem, 0, len(out.List))
	for _, row := range out.List {
		item := dictDataItem(row)
		list = append(list, &item)
	}
	return &v1.DataListRes{List: list, Total: out.Total}, nil
}

// dictDataItem maps a dictionary data projection to the API-safe response DTO.
func dictDataItem(row *dictsvc.DictDataProjection) v1.DictDataItem {
	if row == nil || row.SysDictData == nil {
		return v1.DictDataItem{}
	}
	return v1.DictDataItem{
		Id:             row.Id,
		DictType:       row.DictType,
		Label:          row.Label,
		Value:          row.Value,
		Sort:           row.Sort,
		TagStyle:       row.TagStyle,
		CssClass:       row.CssClass,
		Status:         statusflag.Enabled(row.Status),
		IsBuiltin:      statusflag.YesNo(row.IsBuiltin),
		Remark:         row.Remark,
		SourceTenantId: row.SourceTenantId,
		IsFallback:     row.IsFallback,
		CanEdit:        row.CanEdit,
		CanOverride:    row.CanOverride,
		OverrideMode:   fallbackoverride.Mode(row.OverrideMode),
		CreatedAt:      apitime.Milli(row.CreatedAt),
		UpdatedAt:      apitime.Milli(row.UpdatedAt),
	}
}

// dictTypeItem maps a dictionary type projection to the API-safe response DTO.
func dictTypeItem(row *dictsvc.DictTypeProjection) v1.DictTypeItem {
	if row == nil || row.SysDictType == nil {
		return v1.DictTypeItem{}
	}
	return v1.DictTypeItem{
		Id:                  row.Id,
		Name:                row.Name,
		Type:                row.Type,
		Status:              statusflag.Enabled(row.Status),
		IsBuiltin:           statusflag.YesNo(row.IsBuiltin),
		AllowTenantOverride: row.AllowTenantOverride,
		Remark:              row.Remark,
		SourceTenantId:      row.SourceTenantId,
		IsFallback:          row.IsFallback,
		CanEdit:             row.CanEdit,
		CanOverride:         row.CanOverride,
		OverrideMode:        fallbackoverride.Mode(row.OverrideMode),
		CreatedAt:           apitime.Milli(row.CreatedAt),
		UpdatedAt:           apitime.Milli(row.UpdatedAt),
	}
}
