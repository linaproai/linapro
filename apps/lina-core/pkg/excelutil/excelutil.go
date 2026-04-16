// Package excelutil provides shared helpers for generating Excel files with
// explicit error handling.
package excelutil

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"
)

// CloseFile closes one Excel file and folds any close failure into errPtr.
func CloseFile(file *excelize.File, errPtr *error) {
	if file == nil {
		return
	}
	closeErr := file.Close()
	if closeErr == nil {
		return
	}
	wrapped := gerror.Wrap(closeErr, "关闭Excel文件失败")
	if errPtr == nil {
		// Export and import flows must treat close failures as fatal when the
		// caller cannot merge the error into its return path.
		panic(wrapped)
	}
	if *errPtr == nil {
		// Keep any earlier business error intact and only report the close error
		// when it is the first failure in the call chain.
		*errPtr = wrapped
	}
}

// SetSheetName renames one Excel worksheet and returns a wrapped error when it fails.
func SetSheetName(file *excelize.File, source string, target string) error {
	if err := file.SetSheetName(source, target); err != nil {
		return gerror.Wrapf(err, "重命名Excel工作表失败 source=%s target=%s", source, target)
	}
	return nil
}

// NewSheet creates one Excel worksheet and returns a wrapped error when it fails.
func NewSheet(file *excelize.File, name string) error {
	if _, err := file.NewSheet(name); err != nil {
		return gerror.Wrapf(err, "创建Excel工作表失败 sheet=%s", name)
	}
	return nil
}

// SetCellValue writes one Excel cell addressed by coordinates.
func SetCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return gerror.Wrap(err, "生成Excel单元格名称失败")
	}
	return SetCellValueByName(file, sheet, cell, value)
}

// SetCellValueByName writes one Excel cell addressed by A1 notation.
func SetCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	if err := file.SetCellValue(sheet, cell, value); err != nil {
		return gerror.Wrapf(err, "写入Excel单元格失败 sheet=%s cell=%s", sheet, cell)
	}
	return nil
}
