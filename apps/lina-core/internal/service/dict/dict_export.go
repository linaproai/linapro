// This file implements shared dictionary export helpers used across dict
// type and data services.

package dict

import (
	"bytes"
	"context"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

// CombinedExportInput defines input for CombinedExport function.
type CombinedExportInput struct {
	Name string // Dictionary name, supports fuzzy search
	Type string // Dictionary type, supports fuzzy search
	Ids  []int  // Specific type IDs to export; if empty, export all matching types
}

// CombinedExport generates an Excel file with both dict types and dict data (max 10000 rows each).
func (s *serviceImpl) CombinedExport(ctx context.Context, in CombinedExportInput) (data []byte, err error) {
	// Query dict types
	typeCols := dao.SysDictType.Columns()
	typeM := dao.SysDictType.Ctx(ctx)

	if len(in.Ids) > 0 {
		typeM = typeM.WhereIn(typeCols.Id, in.Ids)
	} else {
		if in.Name != "" {
			typeM = typeM.WhereLike(typeCols.Name, "%"+in.Name+"%")
		}
		if in.Type != "" {
			typeM = typeM.WhereLike(typeCols.Type, "%"+in.Type+"%")
		}
	}

	typeM = typeM.Limit(10000)

	var typeList []*entity.SysDictType
	err = typeM.OrderAsc(typeCols.Id).Scan(&typeList)
	if err != nil {
		return nil, err
	}

	// Collect dict type strings for querying dict data
	typeStrings := make([]string, 0, len(typeList))
	for _, t := range typeList {
		typeStrings = append(typeStrings, t.Type)
	}

	// Query dict data for the selected types
	var dataList []*entity.SysDictData
	if len(typeStrings) > 0 {
		dataCols := dao.SysDictData.Columns()
		dataM := dao.SysDictData.Ctx(ctx).
			WhereIn(dataCols.DictType, typeStrings).
			Limit(10000)

		err = dataM.OrderAsc(dataCols.Sort).Scan(&dataList)
		if err != nil {
			return nil, err
		}
	}

	// Create Excel file with two sheets
	f := excelize.NewFile()
	defer closeExcelFile(f, &err)

	// Sheet 1: 字典类型
	typeSheet := "字典类型"
	if err = setSheetName(f, "Sheet1", typeSheet); err != nil {
		return nil, err
	}

	typeHeaders := []string{"字典名称", "字典类型", "状态", "备注", "创建时间"}
	for i, h := range typeHeaders {
		if err = setCellValue(f, typeSheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, dt := range typeList {
		row := i + 2
		if err = setCellValueByName(f, typeSheet, cellName(1, row), dt.Name); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, typeSheet, cellName(2, row), dt.Type); err != nil {
			return nil, err
		}
		statusText := "正常"
		if dt.Status == 0 {
			statusText = "停用"
		}
		if err = setCellValueByName(f, typeSheet, cellName(3, row), statusText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, typeSheet, cellName(4, row), dt.Remark); err != nil {
			return nil, err
		}
		if dt.CreatedAt != nil {
			if err = setCellValueByName(f, typeSheet, cellName(5, row), dt.CreatedAt.String()); err != nil {
				return nil, err
			}
		}
	}

	// Sheet 2: 字典数据
	dataSheet := "字典数据"
	if err = newSheet(f, dataSheet); err != nil {
		return nil, err
	}

	dataHeaders := []string{"所属类型", "字典标签", "字典值", "排序", "Tag样式", "CSS类", "状态", "备注", "创建时间"}
	for i, h := range dataHeaders {
		if err = setCellValue(f, dataSheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, dd := range dataList {
		row := i + 2
		if err = setCellValueByName(f, dataSheet, cellName(1, row), dd.DictType); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, dataSheet, cellName(2, row), dd.Label); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, dataSheet, cellName(3, row), dd.Value); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, dataSheet, cellName(4, row), dd.Sort); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, dataSheet, cellName(5, row), dd.TagStyle); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, dataSheet, cellName(6, row), dd.CssClass); err != nil {
			return nil, err
		}
		statusText := "正常"
		if dd.Status == 0 {
			statusText = "停用"
		}
		if err = setCellValueByName(f, dataSheet, cellName(7, row), statusText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, dataSheet, cellName(8, row), dd.Remark); err != nil {
			return nil, err
		}
		if dd.CreatedAt != nil {
			if err = setCellValueByName(f, dataSheet, cellName(9, row), dd.CreatedAt.String()); err != nil {
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
