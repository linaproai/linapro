package dict

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"strconv"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
)

// Valid type format: lowercase letters, numbers, underscores, starting with letter
var dictTypeRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// isValidDictType checks if the dict type string is valid.
func isValidDictType(typeStr string) bool {
	return dictTypeRegex.MatchString(typeStr)
}

// isValidDictValue checks if the dict value is valid (non-empty, no leading/trailing spaces).
func isValidDictValue(value string) bool {
	return len(value) > 0 && value == regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(value, "")
}

// CombinedImportResult represents the result of combined import.
type CombinedImportResult struct {
	TypeSuccess int
	TypeFail    int
	DataSuccess int
	DataFail    int
	FailList    []ImportFailItem
}

// ImportFailItem represents a failed import record.
type ImportFailItem struct {
	Sheet  string
	Row    int
	Reason string
}

// CombinedImport imports dictionary types and data from an Excel file.
// If updateSupport is true, existing records will be updated; otherwise, they will be skipped.
func (s *serviceImpl) CombinedImport(ctx context.Context, fileData []byte, updateSupport bool) (result *CombinedImportResult, err error) {
	result = &CombinedImportResult{
		FailList: make([]ImportFailItem, 0),
	}

	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDictImportExcelParseFailed)
	}
	defer closeExcelFile(ctx, f, &err)

	// Get existing dict types for validation (GoFrame auto-adds deleted_at IS NULL)
	typeCols := dao.SysDictType.Columns()
	existingTypes := make(map[string]bool)
	var existingTypeList []*struct {
		Type string
	}
	err = dao.SysDictType.Ctx(ctx).
		Fields(typeCols.Type).
		Scan(&existingTypeList)
	if err != nil {
		return nil, err
	}
	for _, t := range existingTypeList {
		existingTypes[t.Type] = true
	}

	// Import sheet 1 for dictionary types.
	typeSheet := s.dictTypeSheetName(ctx)
	typeRows, err := f.GetRows(typeSheet)
	if err != nil {
		// Sheet might not exist, skip
		typeRows = nil
	}

	// Track imported types for data import
	importedTypes := make(map[string]bool)

	for i, row := range typeRows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 3 {
			result.TypeFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  typeSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.incompleteData", "Incomplete data"),
			})
			continue
		}

		name := row[0]
		typeStr := row[1]

		// Validate name is not empty
		if name == "" {
			result.TypeFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  typeSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeNameRequired", "Dictionary name cannot be empty"),
			})
			continue
		}

		// Validate type format
		if !isValidDictType(typeStr) {
			result.TypeFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  typeSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeFormatInvalid", "Dictionary type must start with a lowercase letter and contain only lowercase letters, numbers, and underscores"),
			})
			continue
		}

		status := 1
		if len(row) > 2 && s.isDictDisabledStatusInput(ctx, row[2]) {
			status = 0
		}
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		// Check if type already exists
		if existingTypes[typeStr] {
			if updateSupport {
				// Update existing record (GoFrame auto-fills updated_at)
				_, err := dao.SysDictType.Ctx(ctx).
					Where(do.SysDictType{Type: typeStr}).
					Data(do.SysDictType{
						Name:   name,
						Status: status,
						Remark: remark,
					}).Update()
				if err != nil {
					result.TypeFail++
					result.FailList = append(result.FailList, ImportFailItem{
						Sheet:  typeSheet,
						Row:    i + 1,
						Reason: s.runtimeText(ctx, "artifact.dict.import.failure.updateFailed", "Update failed: {error}", bizerr.P("error", err)),
					})
					continue
				}
				importedTypes[typeStr] = true
				result.TypeSuccess++
			} else {
				result.TypeFail++
				result.FailList = append(result.FailList, ImportFailItem{
					Sheet:  typeSheet,
					Row:    i + 1,
					Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeExists", "Dictionary type already exists"),
				})
			}
			continue
		}

		// Insert dict type (GoFrame auto-fills created_at and updated_at)
		_, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
			Name:   name,
			Type:   typeStr,
			Status: status,
			Remark: remark,
		}).InsertAndGetId()
		if err != nil {
			result.TypeFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  typeSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.insertFailed", "Insert failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		existingTypes[typeStr] = true
		importedTypes[typeStr] = true
		result.TypeSuccess++
	}

	// Import sheet 2 for dictionary data.
	dataSheet := s.dictDataSheetName(ctx)
	dataRows, err := f.GetRows(dataSheet)
	if err != nil {
		// Sheet might not exist, skip
		dataRows = nil
	}

	for i, row := range dataRows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 4 {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.incompleteData", "Incomplete data"),
			})
			continue
		}

		dictType := row[0]
		label := row[1]
		value := row[2]

		// Validate label is not empty
		if label == "" {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.dataLabelRequired", "Dictionary label cannot be empty"),
			})
			continue
		}

		// Validate value is not empty and has no leading/trailing spaces
		if value == "" {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.dataValueRequired", "Dictionary value cannot be empty"),
			})
			continue
		}

		sort := 0
		if len(row) > 3 && row[3] != "" {
			// Parse sort using strconv for better validation
			var parseErr error
			sort, parseErr = strconv.Atoi(row[3])
			if parseErr != nil {
				result.DataFail++
				result.FailList = append(result.FailList, ImportFailItem{
					Sheet:  dataSheet,
					Row:    i + 1,
					Reason: s.runtimeText(ctx, "artifact.dict.import.failure.sortInvalid", "Sort value must be a valid integer"),
				})
				continue
			}
		}
		tagStyle := ""
		if len(row) > 4 {
			tagStyle = row[4]
		}
		cssClass := ""
		if len(row) > 5 {
			cssClass = row[5]
		}
		status := 1
		if len(row) > 6 && s.isDictDisabledStatusInput(ctx, row[6]) {
			status = 0
		}
		remark := ""
		if len(row) > 7 {
			remark = row[7]
		}

		// Check if dict_type exists
		if !existingTypes[dictType] {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeNotFound", "Dictionary type does not exist"),
			})
			continue
		}

		// Check if dict_data already exists (dict_type + value unique)
		var existingData *entity.SysDictData
		err = dao.SysDictData.Ctx(ctx).
			Where(do.SysDictData{DictType: dictType, Value: value}).
			Scan(&existingData)
		if err != nil {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.queryFailed", "Query failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		if existingData != nil {
			if updateSupport {
				// Update existing record (GoFrame auto-fills updated_at)
				_, err := dao.SysDictData.Ctx(ctx).
					Where(do.SysDictData{Id: existingData.Id}).
					Data(do.SysDictData{
						Label:    label,
						Sort:     sort,
						TagStyle: tagStyle,
						CssClass: cssClass,
						Status:   status,
						Remark:   remark,
					}).Update()
				if err != nil {
					result.DataFail++
					result.FailList = append(result.FailList, ImportFailItem{
						Sheet:  dataSheet,
						Row:    i + 1,
						Reason: s.runtimeText(ctx, "artifact.dict.import.failure.updateFailed", "Update failed: {error}", bizerr.P("error", err)),
					})
					continue
				}
				result.DataSuccess++
			} else {
				result.DataFail++
				result.FailList = append(result.FailList, ImportFailItem{
					Sheet:  dataSheet,
					Row:    i + 1,
					Reason: s.runtimeText(ctx, "artifact.dict.import.failure.dataValueExists", "Dictionary value already exists"),
				})
			}
			continue
		}

		// Insert dict data (GoFrame auto-fills created_at and updated_at)
		_, err = dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
			DictType: dictType,
			Label:    label,
			Value:    value,
			Sort:     sort,
			TagStyle: tagStyle,
			CssClass: cssClass,
			Status:   status,
			Remark:   remark,
		}).InsertAndGetId()
		if err != nil {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.insertFailed", "Insert failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		result.DataSuccess++
	}

	return result, nil
}

// CombinedImportTemplate generates an Excel template for dictionary import.
func (s *serviceImpl) CombinedImportTemplate(ctx context.Context) (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)

	// Sheet 1 stores dictionary type metadata.
	typeSheet := s.dictTypeSheetName(ctx)
	if err = setSheetName(f, "Sheet1", typeSheet); err != nil {
		return nil, err
	}

	typeHeaders := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.dict.type.header.name", Fallback: "Dictionary Name"},
		{Key: "artifact.dict.type.header.type", Fallback: "Dictionary Type"},
		{Key: "artifact.dict.type.header.status", Fallback: "Status"},
		{Key: "artifact.dict.type.header.remark", Fallback: "Remark"},
	})
	for i, h := range typeHeaders {
		if err = setCellValue(f, typeSheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Add example row
	if err = setCellValueByName(f, typeSheet, "A2", s.runtimeText(ctx, "dict.sys_user_sex.name", "User Gender")); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, typeSheet, "B2", "sys_user_sex"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, typeSheet, "C2", s.dictStatusText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, typeSheet, "D2", s.runtimeText(ctx, "dict.sys_user_sex.remark", "User gender options")); err != nil {
		return nil, err
	}

	// Sheet 2 stores dictionary data entries.
	dataSheet := s.dictDataSheetName(ctx)
	if err = newSheet(f, dataSheet); err != nil {
		return nil, err
	}

	dataHeaders := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.dict.data.header.dictType", Fallback: "Dictionary Type"},
		{Key: "artifact.dict.data.header.label", Fallback: "Dictionary Label"},
		{Key: "artifact.dict.data.header.value", Fallback: "Dictionary Value"},
		{Key: "artifact.dict.data.header.sort", Fallback: "Sort"},
		{Key: "artifact.dict.data.header.tagStyle", Fallback: "Tag Style"},
		{Key: "artifact.dict.data.header.cssClass", Fallback: "CSS Class"},
		{Key: "artifact.dict.data.header.status", Fallback: "Status"},
		{Key: "artifact.dict.data.header.remark", Fallback: "Remark"},
	})
	for i, h := range dataHeaders {
		if err = setCellValue(f, dataSheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Add example rows
	if err = setCellValueByName(f, dataSheet, "A2", "sys_user_sex"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "B2", s.runtimeText(ctx, "dict.sys_user_sex.1.label", "Male")); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "C2", "1"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "D2", "1"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "E2", "primary"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "F2", ""); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "G2", s.dictStatusText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "H2", s.runtimeText(ctx, "artifact.dict.importTemplate.example.maleRemark", "Male")); err != nil {
		return nil, err
	}

	if err = setCellValueByName(f, dataSheet, "A3", "sys_user_sex"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "B3", s.runtimeText(ctx, "dict.sys_user_sex.2.label", "Female")); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "C3", "2"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "D3", "2"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "E3", "danger"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "F3", ""); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "G3", s.dictStatusText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, dataSheet, "H3", s.runtimeText(ctx, "artifact.dict.importTemplate.example.femaleRemark", "Female")); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}

// ImportResult represents the result of import operation.
type ImportResult struct {
	Success  int
	Fail     int
	FailList []ImportFailItemRecord
}

// ImportFailItemRecord represents a failed import record.
type ImportFailItemRecord struct {
	Row    int
	Reason string
}

// TypeImport imports dictionary types from an Excel file.
func (s *serviceImpl) TypeImport(ctx context.Context, file io.Reader, updateSupport bool) (result *ImportResult, err error) {
	result = &ImportResult{
		FailList: make([]ImportFailItemRecord, 0),
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDictImportExcelParseFailed)
	}
	defer closeExcelFile(ctx, f, &err)

	// Get existing dict types (dict types use hard delete, no deleted_at filter needed)
	existingTypes := make(map[string]bool)
	var existingTypeList []*entity.SysDictType
	err = dao.SysDictType.Ctx(ctx).
		Scan(&existingTypeList)
	if err != nil {
		return nil, err
	}
	for _, t := range existingTypeList {
		existingTypes[t.Type] = true
	}

	// Read Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDictImportExcelReadFailed)
	}

	for i, row := range rows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 2 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.incompleteData", "Incomplete data"),
			})
			continue
		}

		name := row[0]
		typeStr := row[1]

		// Validate name is not empty
		if name == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeNameRequired", "Dictionary name cannot be empty"),
			})
			continue
		}

		// Validate type format
		if !isValidDictType(typeStr) {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeFormatInvalid", "Dictionary type must start with a lowercase letter and contain only lowercase letters, numbers, and underscores"),
			})
			continue
		}

		status := 1
		if len(row) > 2 && s.isDictDisabledStatusInput(ctx, row[2]) {
			status = 0
		}
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		// Check if type already exists
		if existingTypes[typeStr] {
			if updateSupport {
				// Update existing record (GoFrame auto-fills updated_at)
				_, err := dao.SysDictType.Ctx(ctx).
					Where(do.SysDictType{Type: typeStr}).
					Data(do.SysDictType{
						Name:   name,
						Status: status,
						Remark: remark,
					}).Update()
				if err != nil {
					result.Fail++
					result.FailList = append(result.FailList, ImportFailItemRecord{
						Row:    i + 1,
						Reason: s.runtimeText(ctx, "artifact.dict.import.failure.updateFailed", "Update failed: {error}", bizerr.P("error", err)),
					})
					continue
				}
				result.Success++
			} else {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItemRecord{
					Row:    i + 1,
					Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeExists", "Dictionary type already exists"),
				})
			}
			continue
		}

		// Insert new record (GoFrame auto-fills created_at and updated_at)
		_, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
			Name:   name,
			Type:   typeStr,
			Status: status,
			Remark: remark,
		}).InsertAndGetId()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.insertFailed", "Insert failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		existingTypes[typeStr] = true
		result.Success++
	}

	return result, nil
}

// DataImport imports dictionary data from an Excel file.
func (s *serviceImpl) DataImport(ctx context.Context, file io.Reader, updateSupport bool) (result *ImportResult, err error) {
	result = &ImportResult{
		FailList: make([]ImportFailItemRecord, 0),
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDictImportExcelParseFailed)
	}
	defer closeExcelFile(ctx, f, &err)

	// Get existing dict types (dict types use hard delete, no deleted_at filter needed)
	existingTypes := make(map[string]bool)
	var existingTypeList []*entity.SysDictType
	err = dao.SysDictType.Ctx(ctx).
		Scan(&existingTypeList)
	if err != nil {
		return nil, err
	}
	for _, t := range existingTypeList {
		existingTypes[t.Type] = true
	}

	// Read Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDictImportExcelReadFailed)
	}

	for i, row := range rows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 4 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.incompleteData", "Incomplete data"),
			})
			continue
		}

		dictType := row[0]
		label := row[1]
		value := row[2]

		// Validate label is not empty
		if label == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.dataLabelRequired", "Dictionary label cannot be empty"),
			})
			continue
		}

		// Validate value is not empty
		if value == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.dataValueRequired", "Dictionary value cannot be empty"),
			})
			continue
		}

		sort := 0
		if len(row) > 3 && row[3] != "" {
			var parseErr error
			sort, parseErr = strconv.Atoi(row[3])
			if parseErr != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItemRecord{
					Row:    i + 1,
					Reason: s.runtimeText(ctx, "artifact.dict.import.failure.sortInvalid", "Sort value must be a valid integer"),
				})
				continue
			}
		}
		tagStyle := ""
		if len(row) > 4 {
			tagStyle = row[4]
		}
		cssClass := ""
		if len(row) > 5 {
			cssClass = row[5]
		}
		status := 1
		if len(row) > 6 && s.isDictDisabledStatusInput(ctx, row[6]) {
			status = 0
		}
		remark := ""
		if len(row) > 7 {
			remark = row[7]
		}

		// Check if dict_type exists
		if !existingTypes[dictType] {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.typeNotFound", "Dictionary type does not exist"),
			})
			continue
		}

		// Check if dict_data already exists
		var existingData *entity.SysDictData
		err = dao.SysDictData.Ctx(ctx).
			Where(do.SysDictData{DictType: dictType, Value: value}).
			Scan(&existingData)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.queryFailed", "Query failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		if existingData != nil {
			if updateSupport {
				// Update existing record (GoFrame auto-fills updated_at)
				_, err := dao.SysDictData.Ctx(ctx).
					Where(do.SysDictData{Id: existingData.Id}).
					Data(do.SysDictData{
						Label:    label,
						Sort:     sort,
						TagStyle: tagStyle,
						CssClass: cssClass,
						Status:   status,
						Remark:   remark,
					}).Update()
				if err != nil {
					result.Fail++
					result.FailList = append(result.FailList, ImportFailItemRecord{
						Row:    i + 1,
						Reason: s.runtimeText(ctx, "artifact.dict.import.failure.updateFailed", "Update failed: {error}", bizerr.P("error", err)),
					})
					continue
				}
				result.Success++
			} else {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItemRecord{
					Row:    i + 1,
					Reason: s.runtimeText(ctx, "artifact.dict.import.failure.dataValueExists", "Dictionary value already exists"),
				})
			}
			continue
		}

		// Insert new record (GoFrame auto-fills created_at and updated_at)
		_, err = dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
			DictType: dictType,
			Label:    label,
			Value:    value,
			Sort:     sort,
			TagStyle: tagStyle,
			CssClass: cssClass,
			Status:   status,
			Remark:   remark,
		}).InsertAndGetId()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: s.runtimeText(ctx, "artifact.dict.import.failure.insertFailed", "Insert failed: {error}", bizerr.P("error", err)),
			})
			continue
		}

		result.Success++
	}

	return result, nil
}

// GenerateTypeImportTemplate generates an Excel template for dictionary type import.
func (s *serviceImpl) GenerateTypeImportTemplate(ctx context.Context) (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)

	sheet := "Sheet1"
	headers := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.dict.type.header.name", Fallback: "Dictionary Name"},
		{Key: "artifact.dict.type.header.type", Fallback: "Dictionary Type"},
		{Key: "artifact.dict.type.header.status", Fallback: "Status"},
		{Key: "artifact.dict.type.header.remark", Fallback: "Remark"},
	})
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Add example row
	if err = setCellValueByName(f, sheet, "A2", s.runtimeText(ctx, "dict.sys_user_sex.name", "User Gender")); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "B2", "sys_user_sex"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "C2", s.dictStatusText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "D2", s.runtimeText(ctx, "dict.sys_user_sex.remark", "User gender options")); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}

// GenerateDataImportTemplate generates an Excel template for dictionary data import.
func (s *serviceImpl) GenerateDataImportTemplate(ctx context.Context) (data []byte, err error) {
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)

	sheet := "Sheet1"
	headers := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.dict.data.header.dictType", Fallback: "Dictionary Type"},
		{Key: "artifact.dict.data.header.label", Fallback: "Dictionary Label"},
		{Key: "artifact.dict.data.header.value", Fallback: "Dictionary Value"},
		{Key: "artifact.dict.data.header.sort", Fallback: "Sort"},
		{Key: "artifact.dict.data.header.tagStyle", Fallback: "Tag Style"},
		{Key: "artifact.dict.data.header.cssClass", Fallback: "CSS Class"},
		{Key: "artifact.dict.data.header.status", Fallback: "Status"},
		{Key: "artifact.dict.data.header.remark", Fallback: "Remark"},
	})
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Add example rows
	if err = setCellValueByName(f, sheet, "A2", "sys_user_sex"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "B2", s.runtimeText(ctx, "dict.sys_user_sex.1.label", "Male")); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "C2", "1"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "D2", "1"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "E2", "primary"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "F2", ""); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "G2", s.dictStatusText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "H2", s.runtimeText(ctx, "artifact.dict.importTemplate.example.maleRemark", "Male")); err != nil {
		return nil, err
	}

	if err = setCellValueByName(f, sheet, "A3", "sys_user_sex"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "B3", s.runtimeText(ctx, "dict.sys_user_sex.2.label", "Female")); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "C3", "2"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "D3", "2"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "E3", "danger"); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "F3", ""); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "G3", s.dictStatusText(ctx, 1)); err != nil {
		return nil, err
	}
	if err = setCellValueByName(f, sheet, "H3", s.runtimeText(ctx, "artifact.dict.importTemplate.example.femaleRemark", "Female")); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
