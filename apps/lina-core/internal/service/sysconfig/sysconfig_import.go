// This file implements Excel import and template generation for system
// configuration records.

package sysconfig

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/configvaluetype"
	"lina-core/pkg/excelutil"
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

// closeExcelFile closes the workbook and folds any close failure into the
// caller-managed error pointer.
func closeExcelFile(ctx context.Context, file *excelize.File, errPtr *error) {
	excelutil.CloseFile(ctx, file, errPtr)
}

// setCellValue writes one value by row and column coordinates.
func setCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	return excelutil.SetCellValue(file, sheet, col, row, value)
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

		var (
			name  = row[0]
			key   = row[1]
			value = row[2]
		)
		if name == "" || key == "" || value == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.localizedConfigImportFailure(ctx, "requiredValues", "Parameter name, key, and value cannot be empty"),
			})
			continue
		}

		valueTypeRaw := ""
		if len(row) > 3 {
			valueTypeRaw = strings.TrimSpace(row[3])
		}
		optionsRaw := ""
		if len(row) > 4 {
			optionsRaw = strings.TrimSpace(row[4])
		}
		remark := ""
		if len(row) > 5 {
			remark = row[5]
		}
		if valueTypeRaw != "" {
			if _, resolveErr := configvaluetype.ResolveCode(valueTypeRaw); resolveErr != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: s.localizedConfigImportError(ctx, bizerr.WrapCode(resolveErr, CodeSysConfigValueTypeInvalid)),
				})
				continue
			}
		}
		if optionsRaw != "" {
			if _, parseErr := configvaluetype.ParseOptions(optionsRaw); parseErr != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: s.localizedConfigImportError(ctx, bizerr.WrapCode(parseErr, CodeSysConfigOptionsInvalid)),
				})
				continue
			}
		}

		var (
			previousValue string
			created       bool
			finalValue    = value
		)
		err = s.withConfigMutation(ctx, func(ctx context.Context) error {
			// Check if key exists (GoFrame auto-adds deleted_at IS NULL)
			var existing *entity.SysConfig
			existingModel := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{
				TenantId: datascope.CurrentTenantID(ctx),
				Key:      key,
			})
			scanErr := existingModel.Scan(&existing)
			if scanErr != nil {
				return bizerr.WrapCode(scanErr, CodeSysConfigImportQueryFailed)
			}

			if existing != nil {
				// Key exists
				if !updateSupport {
					return bizerr.NewCode(CodeSysConfigKeyExists, bizerr.P("key", key))
				}
				previousValue = existing.Value
				finalType := entityValueType(existing.ValueType)
				finalOptions := existing.Options
				data := do.SysConfig{
					Name:   name,
					Remark: remark,
				}
				if isBuiltInConfigRecord(existing) {
					// Built-in type/options stay locked.
					finalValue = normalizePersistedValue(finalType, value)
					if validateErr := validateTypedConfigValue(finalType, finalOptions, finalValue); validateErr != nil {
						return validateErr
					}
				} else {
					resolvedType, resolveErr := configvaluetype.ResolveCode(valueTypeRaw)
					if resolveErr != nil {
						return bizerr.WrapCode(resolveErr, CodeSysConfigValueTypeInvalid)
					}
					finalType = resolvedType
					if optionsRaw != "" || valueTypeRaw != "" {
						parsedOptions, parseErr := configvaluetype.ParseOptions(optionsRaw)
						if parseErr != nil {
							return bizerr.WrapCode(parseErr, CodeSysConfigOptionsInvalid)
						}
						encoded, encodeErr := configvaluetype.EncodeOptions(parsedOptions)
						if encodeErr != nil {
							return bizerr.WrapCode(encodeErr, CodeSysConfigOptionsInvalid)
						}
						finalOptions = encoded
					}
					finalValue = normalizePersistedValue(finalType, value)
					if validateErr := validateTypedConfigValue(finalType, finalOptions, finalValue); validateErr != nil {
						return validateErr
					}
					data.ValueType = finalType.String()
					data.Options = finalOptions
				}
				if validateErr := validateManagedConfigValue(key, finalValue); validateErr != nil {
					return validateErr
				}
				data.Value = finalValue
				if hostconfig.IsManagedSysConfigKey(key) {
					data.IsBuiltin = 1
				}
				_, updateErr := dao.SysConfig.Ctx(ctx).
					Where(do.SysConfig{Id: existing.Id}).
					Data(data).
					Update()
				if updateErr != nil {
					return bizerr.WrapCode(updateErr, CodeSysConfigImportUpdateFailed)
				}
				return nil
			}

			// Create new record (GoFrame auto-fills created_at and updated_at)
			createType, createOptions, inheritErr := resolveCreateTypeMetadata(
				ctx,
				key,
				valueTypeRaw,
				parseEntityOptions(optionsRaw),
			)
			if inheritErr != nil {
				return inheritErr
			}
			finalValue = normalizePersistedValue(createType, value)
			if validateErr := validateTypedConfigValue(createType, createOptions, finalValue); validateErr != nil {
				return validateErr
			}
			if validateErr := validateManagedConfigValue(key, finalValue); validateErr != nil {
				return validateErr
			}

			data := currentTenantConfigDO(ctx)
			data.Name = name
			data.Key = key
			data.Value = finalValue
			data.ValueType = createType.String()
			data.Options = createOptions
			data.IsBuiltin = builtInConfigFlag(key)
			data.Remark = remark
			_, insertErr := dao.SysConfig.Ctx(ctx).Data(data).Insert()
			if insertErr != nil {
				return bizerr.WrapCode(insertErr, CodeSysConfigImportInsertFailed)
			}
			created = true
			return nil
		})
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: s.localizedConfigImportError(ctx, err),
			})
			continue
		}
		if err = s.refreshRuntimeParamSnapshotIfNeeded(ctx, key, previousValue, finalValue, created); err != nil {
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
	if err = setCellValue(f, sheet, 4, 2, configvaluetype.Text.String()); err != nil {
		return nil, err
	}
	if err = setCellValue(f, sheet, 5, 2, ""); err != nil {
		return nil, err
	}
	if err = setCellValue(
		f,
		sheet,
		6,
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
