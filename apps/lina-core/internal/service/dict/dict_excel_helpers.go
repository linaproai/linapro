// This file centralizes Excel generation helpers for dictionary imports and exports.

package dict

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/excelutil"
)

func closeExcelFile(file *excelize.File, errPtr *error) {
	excelutil.CloseFile(file, errPtr)
}

func setCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	return excelutil.SetCellValue(file, sheet, col, row, value)
}

func setCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	return excelutil.SetCellValueByName(file, sheet, cell, value)
}

func setSheetName(file *excelize.File, source string, target string) error {
	return excelutil.SetSheetName(file, source, target)
}

func newSheet(file *excelize.File, name string) error {
	return excelutil.NewSheet(file, name)
}

func cellName(col int, row int) string {
	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		panic(gerror.Wrap(err, "生成Excel单元格名称失败"))
	}
	return name
}
