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

// importConfigRowFields holds one Excel data row after column extraction.
type importConfigRowFields struct {
	name         string
	key          string
	value        string
	valueTypeRaw string
	optionsRaw   string
	remark       string
}

// appendImportFail records one failed import row on the result aggregate.
func appendImportFail(result *ImportResult, rowNum int, reason string) {
	result.Fail++
	result.FailList = append(result.FailList, ImportFailItem{
		Row:    rowNum,
		Reason: reason,
	})
}

// parseImportConfigRow extracts and validates one Excel data row for import.
// On validation failure it records the fail item and returns ok=false.
func (s *serviceImpl) parseImportConfigRow(
	ctx context.Context,
	row []string,
	rowNum int,
	result *ImportResult,
) (fields importConfigRowFields, ok bool) {
	if len(row) < 3 {
		appendImportFail(result, rowNum, s.localizedConfigImportFailure(
			ctx,
			"requiredColumns",
			"Parameter name, key, and value are required",
		))
		return fields, false
	}

	fields = importConfigRowFields{
		name:  row[0],
		key:   row[1],
		value: row[2],
	}
	if fields.name == "" || fields.key == "" || fields.value == "" {
		appendImportFail(result, rowNum, s.localizedConfigImportFailure(
			ctx,
			"requiredValues",
			"Parameter name, key, and value cannot be empty",
		))
		return fields, false
	}
	if len(row) > 3 {
		fields.valueTypeRaw = strings.TrimSpace(row[3])
	}
	if len(row) > 4 {
		fields.optionsRaw = strings.TrimSpace(row[4])
	}
	if len(row) > 5 {
		fields.remark = row[5]
	}
	if fields.valueTypeRaw != "" {
		if _, resolveErr := configvaluetype.ResolveCode(fields.valueTypeRaw); resolveErr != nil {
			appendImportFail(result, rowNum, s.localizedConfigImportError(
				ctx,
				bizerr.WrapCode(resolveErr, CodeSysConfigValueTypeInvalid),
			))
			return fields, false
		}
	}
	if fields.optionsRaw != "" {
		if _, parseErr := configvaluetype.ParseOptions(fields.optionsRaw); parseErr != nil {
			appendImportFail(result, rowNum, s.localizedConfigImportError(
				ctx,
				bizerr.WrapCode(parseErr, CodeSysConfigOptionsInvalid),
			))
			return fields, false
		}
	}
	return fields, true
}

// applyImportedConfigUpdate updates an existing config row during Excel import.
func applyImportedConfigUpdate(
	ctx context.Context,
	existing *entity.SysConfig,
	fields importConfigRowFields,
	finalValue *string,
) error {
	if !isSystemManageableRecord(existing) {
		return bizerr.NewCode(CodeSysConfigSystemManageDenied)
	}
	var (
		finalType    = entityValueType(existing.ValueType)
		finalOptions = existing.Options
		data         = do.SysConfig{
			Name:   fields.name,
			Remark: fields.remark,
		}
	)
	if isBuiltInConfigRecord(existing) {
		// Built-in type/options stay locked.
		*finalValue = normalizePersistedValue(finalType, fields.value)
		if validateErr := validateTypedConfigValue(finalType, finalOptions, *finalValue); validateErr != nil {
			return validateErr
		}
	} else {
		resolvedType, resolveErr := configvaluetype.ResolveCode(fields.valueTypeRaw)
		if resolveErr != nil {
			return bizerr.WrapCode(resolveErr, CodeSysConfigValueTypeInvalid)
		}
		finalType = resolvedType
		if fields.optionsRaw != "" || fields.valueTypeRaw != "" {
			parsedOptions, parseErr := configvaluetype.ParseOptions(fields.optionsRaw)
			if parseErr != nil {
				return bizerr.WrapCode(parseErr, CodeSysConfigOptionsInvalid)
			}
			encoded, encodeErr := configvaluetype.EncodeOptions(parsedOptions)
			if encodeErr != nil {
				return bizerr.WrapCode(encodeErr, CodeSysConfigOptionsInvalid)
			}
			finalOptions = encoded
		}
		*finalValue = normalizePersistedValue(finalType, fields.value)
		if validateErr := validateTypedConfigValue(finalType, finalOptions, *finalValue); validateErr != nil {
			return validateErr
		}
		data.ValueType = finalType.String()
		data.Options = finalOptions
	}
	if validateErr := validateManagedConfigValue(fields.key, *finalValue); validateErr != nil {
		return validateErr
	}
	data.Value = *finalValue
	if hostconfig.IsManagedSysConfigKey(fields.key) {
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

// applyImportedConfigCreate inserts a new config row during Excel import.
func applyImportedConfigCreate(
	ctx context.Context,
	fields importConfigRowFields,
	finalValue *string,
) error {
	createType, createOptions, inheritErr := resolveCreateTypeMetadata(
		ctx,
		fields.key,
		fields.valueTypeRaw,
		parseEntityOptions(fields.optionsRaw),
	)
	if inheritErr != nil {
		return inheritErr
	}
	*finalValue = normalizePersistedValue(createType, fields.value)
	if validateErr := validateTypedConfigValue(createType, createOptions, *finalValue); validateErr != nil {
		return validateErr
	}
	if validateErr := validateManagedConfigValue(fields.key, *finalValue); validateErr != nil {
		return validateErr
	}

	data := currentTenantConfigDO(ctx)
	data.Name = fields.name
	data.Key = fields.key
	data.Value = *finalValue
	data.ValueType = createType.String()
	data.Options = createOptions
	data.IsBuiltin = builtInConfigFlag(fields.key)
	data.SystemManageable = 1
	data.Remark = fields.remark
	_, insertErr := dao.SysConfig.Ctx(ctx).Data(data).Insert()
	if insertErr != nil {
		return bizerr.WrapCode(insertErr, CodeSysConfigImportInsertFailed)
	}
	return nil
}

// importConfigRow processes one Excel data row and updates the import result.
func (s *serviceImpl) importConfigRow(
	ctx context.Context,
	row []string,
	rowNum int,
	updateSupport bool,
	result *ImportResult,
) {
	fields, ok := s.parseImportConfigRow(ctx, row, rowNum, result)
	if !ok {
		return
	}

	var (
		previousValue string
		created       bool
		finalValue    = fields.value
	)
	err := s.withConfigMutation(ctx, func(ctx context.Context) error {
		// Check if key exists (GoFrame auto-adds deleted_at IS NULL)
		var existing *entity.SysConfig
		existingModel := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{
			TenantId: datascope.CurrentTenantID(ctx),
			Key:      fields.key,
		})
		scanErr := existingModel.Scan(&existing)
		if scanErr != nil {
			return bizerr.WrapCode(scanErr, CodeSysConfigImportQueryFailed)
		}

		if existing != nil {
			if !updateSupport {
				return bizerr.NewCode(CodeSysConfigKeyExists, bizerr.P("key", fields.key))
			}
			previousValue = existing.Value
			return applyImportedConfigUpdate(ctx, existing, fields, &finalValue)
		}

		// Create new record (GoFrame auto-fills created_at and updated_at)
		if createErr := applyImportedConfigCreate(ctx, fields, &finalValue); createErr != nil {
			return createErr
		}
		created = true
		return nil
	})
	if err != nil {
		appendImportFail(result, rowNum, s.localizedConfigImportError(ctx, err))
		return
	}
	if err = s.refreshRuntimeParamSnapshotIfNeeded(ctx, fields.key, previousValue, finalValue, created); err != nil {
		appendImportFail(result, rowNum, s.localizedConfigImportError(ctx, err))
		return
	}
	result.Success++
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
		s.importConfigRow(ctx, row, i+2, updateSupport, result)
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
