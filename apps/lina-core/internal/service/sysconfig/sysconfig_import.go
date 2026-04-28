// This file implements Excel import and template generation for system
// configuration records.

package sysconfig

import (
	"bytes"
	"context"
	"io"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/bizerr"
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
		return nil, bizerr.WrapCode(err, CodeSysConfigImportExcelParseFailed)
	}
	defer closeExcelFile(ctx, f, &err)

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodeSysConfigImportSheetReadFailed,
			bizerr.P("sheet", "Sheet1"),
		)
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
				Reason: s.localizedConfigImportFailure(ctx, "requiredColumns", "Parameter name, key, and value are required"),
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
				Reason: s.localizedConfigImportFailure(ctx, "requiredValues", "Parameter name, key, and value cannot be empty"),
			})
			continue
		}
		if validateErr := validateManagedConfigValue(key, value); validateErr != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.localizedConfigImportError(ctx, validateErr),
			})
			continue
		}

		// Parse remark
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		err = s.withConfigMutation(ctx, func(ctx context.Context) error {
			// Check if key exists (GoFrame auto-adds deleted_at IS NULL)
			var existing *entity.SysConfig
			scanErr := dao.SysConfig.Ctx(ctx).
				Where(do.SysConfig{Key: key}).
				Scan(&existing)
			if scanErr != nil {
				return bizerr.WrapCode(scanErr, CodeSysConfigImportQueryFailed)
			}

			if existing != nil {
				// Key exists
				if !updateSupport {
					return bizerr.NewCode(CodeSysConfigKeyExists, bizerr.P("key", key))
				}
				// Overwrite mode: update existing record (GoFrame auto-fills updated_at)
				_, updateErr := dao.SysConfig.Ctx(ctx).
					Where(do.SysConfig{Id: existing.Id}).
					Data(do.SysConfig{
						Name:   name,
						Value:  value,
						Remark: remark,
					}).
					Update()
				if updateErr != nil {
					return bizerr.WrapCode(updateErr, CodeSysConfigImportUpdateFailed)
				}
				return s.refreshRuntimeParamSnapshotIfNeeded(ctx, key, existing.Value, value, false)
			}

			// Create new record (GoFrame auto-fills created_at and updated_at)
			_, insertErr := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
				Name:   name,
				Key:    key,
				Value:  value,
				Remark: remark,
			}).Insert()
			if insertErr != nil {
				return bizerr.WrapCode(insertErr, CodeSysConfigImportInsertFailed)
			}
			return s.refreshRuntimeParamSnapshotIfNeeded(ctx, key, "", value, true)
		})
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.localizedConfigImportError(ctx, err),
			})
			continue
		}

		result.Success++
	}

	return result, nil
}

// GenerateImportTemplate creates an Excel template for config import.
func (s *serviceImpl) GenerateImportTemplate(ctx context.Context) (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)
	sheet := "Sheet1"

	headers := s.buildLocalizedImportTemplateHeaders(ctx)
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Example row
	if err = setCellValue(
		f,
		sheet,
		1,
		2,
		s.localizedConfigName(ctx, hostconfig.RuntimeParamKeyJWTExpire, "Authentication - JWT Expiration"),
	); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 2, 2, hostconfig.RuntimeParamKeyJWTExpire); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 3, 2, "24h"); err != nil {
		return nil, err
	}
	if err = setCellValue(
		f,
		sheet,
		4,
		2,
		s.localizedConfigRemark(
			ctx,
			hostconfig.RuntimeParamKeyJWTExpire,
			"Controls the lifetime of newly issued JWT tokens using Go duration format such as 12h or 24h.",
		),
	); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
