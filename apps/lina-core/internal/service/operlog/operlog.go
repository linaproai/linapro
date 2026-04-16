// Package operlog implements operation-log query, cleanup, and export services
// for the Lina core host service.
package operlog

import (
	"bytes"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/pkg/gdbutil"
)

const MaxExportRows = 10000 // Maximum rows for export

// Dict types used in operation log
const (
	DictTypeOperType   = "sys_oper_type"   // Operation type dictionary
	DictTypeOperStatus = "sys_oper_status" // Operation status dictionary
)

// Operation type values (matching sys_oper_type dictionary)
const (
	OperTypeCreate = 1 // 新增
	OperTypeUpdate = 2 // 修改
	OperTypeDelete = 3 // 删除
	OperTypeExport = 4 // 导出
	OperTypeImport = 5 // 导入
	OperTypeOther  = 6 // 其他
)

// OperTag* constants define semantic operLog tag values used in g.Meta and plugin bridge specs.
const (
	OperTagCreate = "create" // 新增
	OperTagUpdate = "update" // 修改
	OperTagDelete = "delete" // 删除
	OperTagExport = "export" // 导出
	OperTagImport = "import" // 导入
	OperTagOther  = "other"  // 其他
)

// operTagToType maps semantic operLog tag to OperType int value.
var operTagToType = map[string]int{
	OperTagCreate: OperTypeCreate,
	OperTagUpdate: OperTypeUpdate,
	OperTagDelete: OperTypeDelete,
	OperTagExport: OperTypeExport,
	OperTagImport: OperTypeImport,
	OperTagOther:  OperTypeOther,
}

// ResolveOperTag converts a semantic operLog tag to the corresponding OperType int value.
// Returns OperTypeOther if the tag is not recognized.
func ResolveOperTag(tag string) (int, bool) {
	v, ok := operTagToType[tag]
	return v, ok
}

// ValidOperTags returns all valid operLog tag values.
func ValidOperTags() []string {
	return []string{OperTagCreate, OperTagUpdate, OperTagDelete, OperTagExport, OperTagImport, OperTagOther}
}

// Operation status values (matching sys_oper_status dictionary)
const (
	OperStatusSuccess = 0 // 成功
	OperStatusFail    = 1 // 失败
)

// Service defines the operlog service contract.
type Service interface {
	// Create inserts a new operation log record.
	Create(ctx context.Context, in CreateInput) error
	// List queries operation log list with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves operation log by ID.
	GetById(ctx context.Context, id int) (*entity.SysOperLog, error)
	// Clean hard-deletes operation logs by time range.
	Clean(ctx context.Context, in CleanInput) (int, error)
	// DeleteByIds hard-deletes operation logs by IDs.
	DeleteByIds(ctx context.Context, ids []int) (int, error)
	// Export generates an Excel file with operation log data (max 10000 rows).
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	dictSvc dictsvc.Service // dictionary service for label lookups
}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		dictSvc: dictsvc.New(),
	}
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Title         string
	OperSummary   string
	OperType      int
	Method        string
	RequestMethod string
	OperName      string
	OperUrl       string
	OperIp        string
	OperParam     string
	JsonResult    string
	Status        int
	ErrorMsg      string
	CostTime      int
}

// Create inserts a new operation log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	_, err := dao.SysOperLog.Ctx(ctx).Data(do.SysOperLog{
		Title:         in.Title,
		OperSummary:   in.OperSummary,
		OperType:      in.OperType,
		Method:        in.Method,
		RequestMethod: in.RequestMethod,
		OperName:      in.OperName,
		OperUrl:       in.OperUrl,
		OperIp:        in.OperIp,
		OperParam:     in.OperParam,
		JsonResult:    in.JsonResult,
		Status:        in.Status,
		ErrorMsg:      in.ErrorMsg,
		CostTime:      in.CostTime,
		OperTime:      gtime.Now(),
	}).Insert()
	return err
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum        int
	PageSize       int
	Title          string
	OperName       string
	OperType       *int
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysOperLog
	Total int
}

// List queries operation log list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	cols := dao.SysOperLog.Columns()
	m := dao.SysOperLog.Ctx(ctx)

	if in.Title != "" {
		m = m.WhereLike(cols.Title, "%"+in.Title+"%")
	}
	if in.OperName != "" {
		m = m.WhereLike(cols.OperName, "%"+in.OperName+"%")
	}
	if in.OperType != nil {
		m = m.Where(cols.OperType, *in.OperType)
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.OperTime, in.BeginTime)
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.OperTime, endTime)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var (
		orderBy           = cols.OperTime
		allowedSortFields = map[string]string{
			"id":       cols.Id,
			"operTime": cols.OperTime,
			"costTime": cols.CostTime,
		}
		direction = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)
	if in.OrderBy != "" {
		if field, ok := allowedSortFields[in.OrderBy]; ok {
			orderBy = field
		}
	}

	var list []*entity.SysOperLog
	err = gdbutil.ApplyModelOrder(
		m.Page(in.PageNum, in.PageSize),
		orderBy,
		direction,
	).Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// GetById retrieves operation log by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysOperLog, error) {
	var record *entity.SysOperLog
	err := dao.SysOperLog.Ctx(ctx).
		Where(do.SysOperLog{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("操作日志不存在")
	}
	return record, nil
}

// CleanInput defines input for Clean function.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// Clean hard-deletes operation logs by time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	cols := dao.SysOperLog.Columns()
	m := dao.SysOperLog.Ctx(ctx)

	hasFilter := false
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.OperTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.OperTime, endTime)
		hasFilter = true
	}
	if !hasFilter {
		m = m.Where(1)
	}

	result, err := m.Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// DeleteByIds hard-deletes operation logs by IDs.
func (s *serviceImpl) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := dao.SysOperLog.Ctx(ctx).WhereIn(dao.SysOperLog.Columns().Id, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Title          string
	OperName       string
	OperType       *int
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
	Ids            []int // Specific IDs to export; if empty, export all matching records
}

// Export generates an Excel file with operation log data (max 10000 rows).
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysOperLog.Columns()
	m := dao.SysOperLog.Ctx(ctx)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.Title != "" {
			m = m.WhereLike(cols.Title, "%"+in.Title+"%")
		}
		if in.OperName != "" {
			m = m.WhereLike(cols.OperName, "%"+in.OperName+"%")
		}
		if in.OperType != nil {
			m = m.Where(cols.OperType, *in.OperType)
		}
		if in.Status != nil {
			m = m.Where(cols.Status, *in.Status)
		}
		if in.BeginTime != "" {
			m = m.WhereGTE(cols.OperTime, in.BeginTime)
		}
		if in.EndTime != "" {
			endTime := in.EndTime
			if len(endTime) == 10 {
				endTime += " 23:59:59"
			}
			m = m.WhereLTE(cols.OperTime, endTime)
		}
	}

	// Limit export to prevent memory issues
	m = m.Limit(MaxExportRows)

	var (
		orderBy   = cols.OperTime
		direction = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)

	var list []*entity.SysOperLog
	err = gdbutil.ApplyModelOrder(m, orderBy, direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"模块名称", "操作名称", "操作类型", "操作人", "请求方式", "请求URL", "操作IP", "请求参数", "响应结果", "状态", "错误信息", "耗时(ms)", "操作时间"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Build label maps from dictionary for batch lookups
	operTypeMap := s.dictSvc.BuildIntLabelMap(ctx, DictTypeOperType)
	statusMap := s.dictSvc.BuildIntLabelMap(ctx, DictTypeOperStatus)

	for i, log := range list {
		row := i + 2
		if err = setCellValueByName(f, sheet, cellName(1, row), log.Title); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(2, row), log.OperSummary); err != nil {
			return nil, err
		}
		// Use dictionary lookup for operation type
		operTypeText, ok := operTypeMap[log.OperType]
		if !ok {
			operTypeText = s.dictSvc.GetLabelByIntValue(ctx, DictTypeOperType, 6) // fallback to "其他"
		}
		if err = setCellValueByName(f, sheet, cellName(3, row), operTypeText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(4, row), log.OperName); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(5, row), log.RequestMethod); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(6, row), log.OperUrl); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(7, row), log.OperIp); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(8, row), log.OperParam); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(9, row), log.JsonResult); err != nil {
			return nil, err
		}
		// Use dictionary lookup for status
		statusText, ok := statusMap[log.Status]
		if !ok {
			statusText = s.dictSvc.GetLabelByIntValue(ctx, DictTypeOperStatus, 0) // fallback to "成功"
		}
		if err = setCellValueByName(f, sheet, cellName(10, row), statusText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(11, row), log.ErrorMsg); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(12, row), log.CostTime); err != nil {
			return nil, err
		}
		if log.OperTime != nil {
			if err = setCellValueByName(f, sheet, cellName(13, row), log.OperTime.String()); err != nil {
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
