// This file implements post Excel export helpers.

package post

import (
	"bytes"
	"context"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

// ExportInput defines input for Export function.
type ExportInput struct {
	DeptId *int
	Code   string
	Name   string
	Status *int
}

// Export generates an Excel file with post data based on filters.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysPost.Columns()
	m := dao.SysPost.Ctx(ctx)

	// Apply filters
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			m = m.Where(cols.DeptId, 0)
		} else {
			deptIds, err := s.getDeptAndDescendantIds(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			m = m.WhereIn(cols.DeptId, deptIds)
		}
	}
	if in.Code != "" {
		m = m.WhereLike(cols.Code, "%"+in.Code+"%")
	}
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	var list []*entity.SysPost
	err = m.OrderAsc(cols.Sort).Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"岗位编码", "岗位名称", "排序", "状态", "备注", "创建时间"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, p := range list {
		row := i + 2
		if err = setCellValueByName(f, sheet, cellName(1, row), p.Code); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(2, row), p.Name); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(3, row), p.Sort); err != nil {
			return nil, err
		}
		statusText := "正常"
		if p.Status == 0 {
			statusText = "停用"
		}
		if err = setCellValueByName(f, sheet, cellName(4, row), statusText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(5, row), p.Remark); err != nil {
			return nil, err
		}
		if p.CreatedAt != nil {
			if err = setCellValueByName(f, sheet, cellName(6, row), p.CreatedAt.String()); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
