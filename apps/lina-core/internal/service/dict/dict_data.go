// This file implements dictionary-data query, option, import, and export
// helpers.

package dict

import (
	"bytes"
	"context"
	"sort"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// DataListInput defines input for DataList function.
type DataListInput struct {
	PageNum  int
	PageSize int
	DictType string
	Label    string
}

// DataListOutput defines output for DataList function.
type DataListOutput struct {
	List  []*entity.SysDictData
	Total int
}

// DataList queries dict data list with pagination and filters.
func (s *serviceImpl) DataList(ctx context.Context, in DataListInput) (*DataListOutput, error) {
	var (
		cols = dao.SysDictData.Columns()
		m    = dao.SysDictData.Ctx(ctx)
	)
	m = applyDictFallbackScope(ctx, m)

	// Apply filters
	if in.DictType != "" {
		m = m.Where(do.SysDictData{DictType: in.DictType})
	}
	if in.Label != "" {
		m = m.WhereLike(cols.Label, "%"+in.Label+"%")
	}

	var rows []*entity.SysDictData
	err := m.OrderAsc(cols.Sort).Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := visibleDictData(ctx, rows)
	sort.SliceStable(list, func(i int, j int) bool {
		if list[i].Sort == list[j].Sort {
			return list[i].Id < list[j].Id
		}
		return list[i].Sort < list[j].Sort
	})
	total := len(list)
	list = paginateDictData(list, in.PageNum, in.PageSize)
	s.localizeDictDataEntities(ctx, list)

	return &DataListOutput{
		List:  list,
		Total: total,
	}, nil
}

// DataCreateInput defines input for DataCreate function.
type DataCreateInput struct {
	DictType string
	Label    string
	Value    string
	Sort     int
	TagStyle string
	CssClass string
	Status   int
	Remark   string
}

// DataCreate creates a new dict data entry.
func (s *serviceImpl) DataCreate(ctx context.Context, in DataCreateInput) (int, error) {
	if err := assertDictTenantOverrideAllowed(ctx, in.DictType); err != nil {
		return 0, err
	}

	data := currentTenantDictDataDO(ctx)
	data.DictType = in.DictType
	data.Label = in.Label
	data.Value = in.Value
	data.Sort = in.Sort
	data.TagStyle = in.TagStyle
	data.CssClass = in.CssClass
	data.Status = in.Status
	data.Remark = in.Remark

	id, err := dao.SysDictData.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// DataGetById retrieves dict data by ID.
func (s *serviceImpl) DataGetById(ctx context.Context, id int) (*entity.SysDictData, error) {
	var dictData *entity.SysDictData
	model := dao.SysDictData.Ctx(ctx).Where(do.SysDictData{Id: id})
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&dictData)
	if err != nil {
		return nil, err
	}
	if dictData == nil {
		return nil, bizerr.NewCode(CodeDictDataNotFound)
	}
	return dictData, nil
}

// DataUpdateInput defines input for DataUpdate function.
type DataUpdateInput struct {
	Id       int
	DictType *string
	Label    *string
	Value    *string
	Sort     *int
	TagStyle *string
	CssClass *string
	Status   *int
	Remark   *string
}

// DataUpdate updates dict data information.
func (s *serviceImpl) DataUpdate(ctx context.Context, in DataUpdateInput) error {
	// Check dict data exists
	existing, err := s.DataGetById(ctx, in.Id)
	if err != nil {
		return err
	}

	finalType := existing.DictType
	if in.DictType != nil {
		finalType = *in.DictType
	}
	if err := assertDictTenantOverrideAllowed(ctx, finalType); err != nil {
		return err
	}

	data := do.SysDictData{}
	if in.DictType != nil {
		data.DictType = *in.DictType
	}
	if in.Label != nil {
		data.Label = *in.Label
	}
	if in.Value != nil {
		data.Value = *in.Value
	}
	if in.Sort != nil {
		data.Sort = *in.Sort
	}
	if in.TagStyle != nil {
		data.TagStyle = *in.TagStyle
	}
	if in.CssClass != nil {
		data.CssClass = *in.CssClass
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err = dao.SysDictData.Ctx(ctx).Where(do.SysDictData{Id: in.Id}).Data(data).Update()
	return err
}

// DataDelete hard-deletes a dict data entry.
func (s *serviceImpl) DataDelete(ctx context.Context, id int) error {
	// Check dict data exists
	dictData, err := s.DataGetById(ctx, id)
	if err != nil {
		return err
	}
	if dictData.IsBuiltin == 1 {
		return bizerr.NewCode(CodeDictDataBuiltinDeleteDenied)
	}

	// Hard delete
	_, err = dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{Id: id}).
		Delete()
	return err
}

// DataExportInput defines input for DataExport function.
type DataExportInput struct {
	DictType string
	Label    string
	Ids      []int // Specific IDs to export; if empty, export all matching records
}

// DataExport generates an Excel file with dict data (max 10000 rows).
func (s *serviceImpl) DataExport(ctx context.Context, in DataExportInput) (data []byte, err error) {
	cols := dao.SysDictData.Columns()
	m := dao.SysDictData.Ctx(ctx)
	m = applyDictFallbackScope(ctx, m)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.DictType != "" {
			m = m.Where(cols.DictType, in.DictType)
		}
		if in.Label != "" {
			m = m.WhereLike(cols.Label, "%"+in.Label+"%")
		}
	}

	// Limit export to prevent memory issues
	m = m.Limit(10000)

	var rows []*entity.SysDictData
	err = m.OrderAsc(cols.Sort).Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := visibleDictData(ctx, rows)
	s.localizeDictDataEntities(ctx, list)

	// Create Excel file
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)
	sheet := "Sheet1"

	headers := s.runtimeTexts(ctx, []runtimeTextItem{
		{Key: "artifact.dict.data.header.label", Fallback: "Dictionary Label"},
		{Key: "artifact.dict.data.header.value", Fallback: "Dictionary Value"},
		{Key: "artifact.dict.data.header.sort", Fallback: "Sort"},
		{Key: "artifact.dict.data.header.tagStyle", Fallback: "Tag Style"},
		{Key: "artifact.dict.data.header.cssClass", Fallback: "CSS Class"},
		{Key: "artifact.dict.data.header.status", Fallback: "Status"},
		{Key: "artifact.dict.data.header.remark", Fallback: "Remark"},
		{Key: "artifact.dict.data.header.createdAt", Fallback: "Created At"},
	})
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, dd := range list {
		row := i + 2
		if err = setCellValue(f, sheet, 1, row, dd.Label); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 2, row, dd.Value); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 3, row, dd.Sort); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 4, row, dd.TagStyle); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 5, row, dd.CssClass); err != nil {
			return nil, err
		}
		statusText := s.dictStatusText(ctx, dd.Status)
		if err = setCellValue(f, sheet, 6, row, statusText); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 7, row, dd.Remark); err != nil {
			return nil, err
		}
		if dd.CreatedAt != nil {
			if err = setCellValue(f, sheet, 8, row, dd.CreatedAt.String()); err != nil {
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

// DataByType returns all non-deleted dict data for a given dict type with status=1, ordered by sort ASC.
func (s *serviceImpl) DataByType(ctx context.Context, dictType string) ([]*entity.SysDictData, error) {
	cols := dao.SysDictData.Columns()
	var list []*entity.SysDictData
	model := dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{DictType: dictType, Status: 1}).
		OrderAsc(cols.Sort)
	model = applyDictFallbackScope(ctx, model)
	err := model.Scan(&list)
	if err != nil {
		return nil, err
	}
	list = visibleDictData(ctx, list)
	s.localizeDictDataEntities(ctx, list)
	return list, nil
}

// paginateDictData returns one page from an already materialized effective
// dictionary-data view.
func paginateDictData(rows []*entity.SysDictData, pageNum int, pageSize int) []*entity.SysDictData {
	if pageNum <= 0 || pageSize <= 0 {
		return rows
	}
	start := (pageNum - 1) * pageSize
	if start >= len(rows) {
		return []*entity.SysDictData{}
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}
