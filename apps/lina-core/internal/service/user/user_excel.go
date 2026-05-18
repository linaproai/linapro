// This file implements user Excel import-template and export helpers.

package user

import (
	"bytes"
	"context"
	"io"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
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
		if err = s.ensureUsersVisible(ctx, in.Ids); err != nil {
			return nil, err
		}
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		var scopeEmpty bool
		m, scopeEmpty, err = s.applyUserDataScope(ctx, m)
		if err != nil {
			return nil, err
		}
		if scopeEmpty {
			m = nil
		}
	}

	var list []*entity.SysUser
	if m != nil {
		err = m.FieldsEx(cols.Password).
			OrderAsc(cols.Id).
			Scan(&list)
		if err != nil {
			return nil, err
		}
	}

	// Create Excel file
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)
	sheet := "Sheet1"

	headers := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.user.header.username", Fallback: "Username"},
		{Key: "artifact.user.header.nickname", Fallback: "Nickname"},
		{Key: "artifact.user.header.phone", Fallback: "Mobile Number"},
		{Key: "artifact.user.header.email", Fallback: "Email"},
		{Key: "artifact.user.header.sex", Fallback: "Gender"},
		{Key: "artifact.user.header.status", Fallback: "Status"},
		{Key: "artifact.user.header.remark", Fallback: "Remark"},
		{Key: "artifact.user.header.createdAt", Fallback: "Created At"},
	})
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, u := range list {
		row := i + 2
		if err = setCellValue(f, sheet, 1, row, u.Username); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 2, row, u.Nickname); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 3, row, u.Phone); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 4, row, u.Email); err != nil {
			return nil, err
		}
		sexText := s.userSexText(ctx, u.Sex)
		if err = setCellValue(f, sheet, 5, row, sexText); err != nil {
			return nil, err
		}
		statusText := s.userStatusText(ctx, Status(u.Status))
		if err = setCellValue(f, sheet, 6, row, statusText); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 7, row, u.Remark); err != nil {
			return nil, err
		}
		if u.CreatedAt != nil {
			if err = setCellValue(f, sheet, 8, row, u.CreatedAt.String()); err != nil {
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
		return nil, bizerr.WrapCode(err, CodeUserImportExcelParseFailed)
	}
	defer closeExcelFile(ctx, f, &err)

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodeUserImportSheetReadFailed,
			bizerr.P("sheet", "Sheet1"),
		)
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
				Reason: s.runtimeText(ctx, "artifact.user.import.failure.usernamePasswordRequired", "Username and password are required"),
			})
			continue
		}

		username := row[0]
		password := row[1]
		if username == "" || password == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.runtimeText(ctx, "artifact.user.import.failure.usernamePasswordNotEmpty", "Username and password cannot be empty"),
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
				Reason: s.runtimeText(ctx, "artifact.user.import.failure.queryFailed", "Database query failed: {error}", bizerr.P("error", err)),
			})
			continue
		}
		if count > 0 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.runtimeText(ctx, "artifact.user.import.failure.usernameExists", "Username {username} already exists", bizerr.P("username", username)),
			})
			continue
		}

		hash, err := s.authSvc.HashPassword(password)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.runtimeText(ctx, "artifact.user.import.failure.passwordHashFailed", "Failed to hash password"),
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
			case "1":
				data.Sex = 1
			case "2":
				data.Sex = 2
			default:
				if s.isUserSexInput(ctx, row[5], 1) {
					data.Sex = 1
				} else if s.isUserSexInput(ctx, row[5], 2) {
					data.Sex = 2
				} else {
					data.Sex = 0
				}
			}
		}
		if len(row) > 6 {
			if s.isUserDisabledStatusInput(ctx, row[6]) {
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
				Reason: s.runtimeText(ctx, "artifact.user.import.failure.insertFailed", "Insert failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		result.Success++
	}

	return result, nil
}

// GenerateImportTemplate creates an Excel template for user import.
func (s *serviceImpl) GenerateImportTemplate(ctx context.Context) (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)
	sheet := "Sheet1"

	headers := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.user.header.username", Fallback: "Username"},
		{Key: "artifact.user.header.password", Fallback: "Password"},
		{Key: "artifact.user.header.nickname", Fallback: "Nickname"},
		{Key: "artifact.user.header.phone", Fallback: "Mobile Number"},
		{Key: "artifact.user.header.email", Fallback: "Email"},
		{Key: "artifact.user.header.sex", Fallback: "Gender"},
		{Key: "artifact.user.header.status", Fallback: "Status"},
		{Key: "artifact.user.header.remark", Fallback: "Remark"},
	})
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Example row
	if err = setCellValue(f, sheet, 1, 2, "zhangsan"); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 2, 2, "123456"); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 3, 2, s.runtimeText(ctx, "artifact.user.importTemplate.example.nickname", "Zhang San")); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 4, 2, "13800138000"); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 5, 2, "zhangsan@example.com"); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 6, 2, s.userSexText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 7, 2, s.userStatusText(ctx, StatusNormal)); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 8, 2, s.runtimeText(ctx, "artifact.user.importTemplate.example.remark", "Sample user")); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
