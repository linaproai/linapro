// This file implements user Excel import-template and export helpers.

package user

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

// ExportInput defines input for Export function.
type ExportInput struct {
	Ids []int // User ID list, empty means export all
}

// Export generates an Excel file with user data based on IDs.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	}

	var list []*entity.SysUser
	err = m.FieldsEx(cols.Password).
		OrderAsc(cols.Id).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"用户名", "昵称", "手机号码", "邮箱", "性别", "状态", "备注", "创建时间"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, u := range list {
		row := i + 2
		if err = setCellValueByName(f, sheet, cellName(1, row), u.Username); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(2, row), u.Nickname); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(3, row), u.Phone); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(4, row), u.Email); err != nil {
			return nil, err
		}
		sexText := "未知"
		switch u.Sex {
		case 1:
			sexText = "男"
		case 2:
			sexText = "女"
		}
		if err = setCellValueByName(f, sheet, cellName(5, row), sexText); err != nil {
			return nil, err
		}
		statusText := "正常"
		if u.Status == int(StatusDisabled) {
			statusText = "停用"
		}
		if err = setCellValueByName(f, sheet, cellName(6, row), statusText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(7, row), u.Remark); err != nil {
			return nil, err
		}
		if u.CreatedAt != nil {
			if err = setCellValueByName(f, sheet, cellName(8, row), u.CreatedAt.String()); err != nil {
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

// ImportResult defines the result of import operation.
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

// Import reads an Excel file and creates users from it.
func (s *serviceImpl) Import(ctx context.Context, fileReader io.Reader) (result *ImportResult, err error) {
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
		if len(row) < 2 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "用户名和密码为必填项",
			})
			continue
		}

		username := row[0]
		password := row[1]
		if username == "" || password == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "用户名和密码不能为空",
			})
			continue
		}

		// Check username uniqueness (GoFrame auto-adds deleted_at IS NULL)
		count, err := dao.SysUser.Ctx(ctx).
			Where(do.SysUser{Username: username}).
			Count()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}
		if count > 0 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("用户名 '%s' 已存在", username),
			})
			continue
		}

		hash, err := s.authSvc.HashPassword(password)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "密码加密失败",
			})
			continue
		}

		// Insert user (GoFrame auto-fills created_at and updated_at)
		data := do.SysUser{
			Username: username,
			Password: hash,
			Status:   int(StatusNormal),
		}
		if len(row) > 2 {
			data.Nickname = row[2]
		}
		if len(row) > 3 {
			data.Phone = row[3]
		}
		if len(row) > 4 {
			data.Email = row[4]
		}
		if len(row) > 5 {
			switch row[5] {
			case "男", "1":
				data.Sex = 1
			case "女", "2":
				data.Sex = 2
			default:
				data.Sex = 0
			}
		}
		if len(row) > 6 {
			switch row[6] {
			case "停用", "0":
				data.Status = int(StatusDisabled)
			}
		}
		if len(row) > 7 {
			data.Remark = row[7]
		}

		_, err = dao.SysUser.Ctx(ctx).Data(data).Insert()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("插入失败: %v", err),
			})
			continue
		}

		result.Success++
	}

	return result, nil
}

// GenerateImportTemplate creates an Excel template for user import.
func (s *serviceImpl) GenerateImportTemplate() (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"用户名", "密码", "昵称", "手机号码", "邮箱", "性别", "状态", "备注"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Example row
	if err = setCellValueByName(f, sheet, cellName(1, 2), "zhangsan"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(2, 2), "123456"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(3, 2), "张三"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(4, 2), "13800138000"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(5, 2), "zhangsan@example.com"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(6, 2), "男"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(7, 2), "正常"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, cellName(8, 2), "示例用户"); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
