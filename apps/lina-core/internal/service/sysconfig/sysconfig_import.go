package sysconfig

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// ImportResult defines the result of config import operation.
type ImportResult struct {
	Success  int              // Number of successful imports
	Fail     int              // Number of failed imports
	FailList []ImportFailItem // Failure list
}

// ImportFailItem defines a single import failure.
type ImportFailItem struct {
	Row    int    // Row number
	Reason string // Failure reason
}

// Import reads an Excel file and creates configs from it.
// If updateSupport is true, existing records (matched by key) will be updated; otherwise, they will be skipped.
func (s *serviceImpl) Import(ctx context.Context, fileReader io.Reader, updateSupport bool) (result *ImportResult, err error) {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return nil, gerror.New("无法解析 Excel 文件")
	}
	defer closeExcelFile(f, &err)

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, gerror.New("无法读取 Sheet1")
	}

	if len(rows) < 2 {
		return &ImportResult{}, nil
	}

	result = &ImportResult{}

	for i, row := range rows[1:] { // Skip header
		rowNum := i + 2
		if len(row) < 3 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "参数名称、参数键名、参数键值为必填项",
			})
			continue
		}

		name := row[0]
		key := row[1]
		value := row[2]
		if name == "" || key == "" || value == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "参数名称、参数键名、参数键值不能为空",
			})
			continue
		}

		// Check if key exists (GoFrame auto-adds deleted_at IS NULL)
		var existing *entity.SysConfig
		err := dao.SysConfig.Ctx(ctx).
			Where(do.SysConfig{Key: key}).
			Scan(&existing)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}

		// Parse remark
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		if existing != nil {
			// Key exists
			if !updateSupport {
				// Ignore mode: skip this record
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("参数键名 '%s' 已存在", key),
				})
				continue
			}
			// Overwrite mode: update existing record (GoFrame auto-fills updated_at)
			_, err = dao.SysConfig.Ctx(ctx).
				Where(do.SysConfig{Id: existing.Id}).
				Data(do.SysConfig{
					Name:   name,
					Value:  value,
					Remark: remark,
				}).
				Update()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("更新失败: %v", err),
				})
				continue
			}
		} else {
			// Create new record (GoFrame auto-fills created_at and updated_at)
			_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
				Name:   name,
				Key:    key,
				Value:  value,
				Remark: remark,
			}).Insert()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("插入失败: %v", err),
				})
				continue
			}
		}

		result.Success++
	}

	return result, nil
}

// GenerateImportTemplate creates an Excel template for config import.
func (s *serviceImpl) GenerateImportTemplate() (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"参数名称", "参数键名", "参数键值", "备注"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Example row
	if err = setCellValueByName(f, sheet, cellName(1, 2), "系统名称"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(2, 2), "sys.app.name"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(3, 2), "Lina"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(4, 2), "系统显示名称"); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
