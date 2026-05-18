package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	"lina-core/internal/model/entity"
	dictsvc "lina-core/internal/service/dict"
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

// dictDataItem maps a dictionary data entity to the API-safe response DTO.
func dictDataItem(row *entity.SysDictData) v1.DictDataItem {
	if row == nil {
		return v1.DictDataItem{}
	}
	return v1.DictDataItem{
		Id:        row.Id,
		DictType:  row.DictType,
		Label:     row.Label,
		Value:     row.Value,
		Sort:      row.Sort,
		TagStyle:  row.TagStyle,
		CssClass:  row.CssClass,
		Status:    row.Status,
		IsBuiltin: row.IsBuiltin,
		Remark:    row.Remark,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// dictTypeItem maps a dictionary type entity to the API-safe response DTO.
func dictTypeItem(row *entity.SysDictType) v1.DictTypeItem {
	if row == nil {
		return v1.DictTypeItem{}
	}
	return v1.DictTypeItem{
		Id:        row.Id,
		Name:      row.Name,
		Type:      row.Type,
		Status:    row.Status,
		IsBuiltin: row.IsBuiltin,
		Remark:    row.Remark,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
