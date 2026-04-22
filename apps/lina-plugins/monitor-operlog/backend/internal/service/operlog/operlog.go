// Package operlog implements operation-log persistence, query, cleanup,
// and export services for the monitor-operlog source plugin. It owns the
// plugin_monitor_operlog table access instead of depending on host-internal operlog
// services.
package operlog

import (
	"bytes"
	"context"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/excelutil"
	"lina-core/pkg/gdbutil"
	"lina-plugin-monitor-operlog/backend/internal/dao"
	"lina-plugin-monitor-operlog/backend/internal/model/do"
	entitymodel "lina-plugin-monitor-operlog/backend/internal/model/entity"
)

// Table, column, and dictionary constants used by the plugin-owned operation-log service.
const (
	colID            = "id"
	colTitle         = "title"
	colOperSummary   = "oper_summary"
	colOperType      = "oper_type"
	colMethod        = "method"
	colRequestMethod = "request_method"
	colOperName      = "oper_name"
	colOperURL       = "oper_url"
	colOperIP        = "oper_ip"
	colOperParam     = "oper_param"
	colJSONResult    = "json_result"
	colStatus        = "status"
	colErrorMsg      = "error_msg"
	colCostTime      = "cost_time"
	colOperTime      = "oper_time"

	colDictType  = "dict_type"
	colDictValue = "value"
	colDictLabel = "label"
	colDictSort  = "sort"
)

// Operation-log export limit and dictionary constants.
const (
	MaxExportRows      = 10000
	DictTypeOperType   = "sys_oper_type"
	DictTypeOperStatus = "sys_oper_status"
)

// Operation type values stored in plugin_monitor_operlog.
const (
	OperTypeCreate = 1
	OperTypeUpdate = 2
	OperTypeDelete = 3
	OperTypeExport = 4
	OperTypeImport = 5
	OperTypeOther  = 6
)

// Operation status values stored in plugin_monitor_operlog.
const (
	OperStatusSuccess = 0
	OperStatusFail    = 1
)

var defaultOperTypeLabels = map[int]string{
	OperTypeCreate: "新增",
	OperTypeUpdate: "修改",
	OperTypeDelete: "删除",
	OperTypeExport: "导出",
	OperTypeImport: "导入",
	OperTypeOther:  "其他",
}

var defaultOperStatusLabels = map[int]string{
	OperStatusSuccess: "成功",
	OperStatusFail:    "失败",
}

// Service defines the monitor-operlog service contract.
type Service interface {
	// Create inserts one operation-log record.
	Create(ctx context.Context, in CreateInput) error
	// List queries the paginated operation-log list.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves one operation-log record by primary key.
	GetById(ctx context.Context, id int) (*OperLogEntity, error)
	// Clean hard-deletes operation logs within one optional time range.
	Clean(ctx context.Context, in CleanInput) (int, error)
	// DeleteByIds hard-deletes operation logs by ID list.
	DeleteByIds(ctx context.Context, ids []int) (int, error)
	// Export generates an Excel workbook for operation logs.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a new monitor-operlog service instance.
func New() Service {
	return &serviceImpl{}
}

// OperLogEntity mirrors the plugin-local generated plugin_monitor_operlog entity.
type OperLogEntity = entitymodel.Operlog

// dictDataRow reuses the plugin-local generated sys_dict_data entity.
type dictDataRow = entitymodel.SysDictData

// CreateInput defines the operation-log create input.
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

// ListInput defines the operation-log list filter input.
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

// ListOutput defines the operation-log list output.
type ListOutput struct {
	List  []*OperLogEntity
	Total int
}

// CleanInput defines the operation-log cleanup input.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// ExportInput defines the operation-log export input.
type ExportInput struct {
	Title          string
	OperName       string
	OperType       *int
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
	Ids            []int
}

// Create inserts one operation-log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	_, err := dao.Operlog.Ctx(ctx).Data(do.Operlog{
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

// List queries the paginated operation-log list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := dao.Operlog.Ctx(ctx)
	model = applyOperLogFilters(model, in.Title, in.OperName, in.OperType, in.Status, in.BeginTime, in.EndTime)

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	allowedSortFields := map[string]string{
		"id":        colID,
		"operTime":  colOperTime,
		"oper_time": colOperTime,
		"costTime":  colCostTime,
		"cost_time": colCostTime,
	}
	orderBy := colOperTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*OperLogEntity, 0)
	err = gdbutil.ApplyModelOrder(
		model.Page(in.PageNum, in.PageSize),
		orderBy,
		direction,
	).Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{List: list, Total: total}, nil
}

// GetById retrieves one operation-log record by primary key.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*OperLogEntity, error) {
	var record *OperLogEntity
	err := dao.Operlog.Ctx(ctx).Where(colID, id).Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("操作日志不存在")
	}
	return record, nil
}

// Clean hard-deletes operation logs within one optional time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	model := dao.Operlog.Ctx(ctx)
	hasFilter := false
	if in.BeginTime != "" {
		model = model.WhereGTE(colOperTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		model = model.WhereLTE(colOperTime, normalizeEndTime(in.EndTime))
		hasFilter = true
	}
	if !hasFilter {
		model = model.Where("1 = 1")
	}

	result, err := model.Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// DeleteByIds hard-deletes operation logs by ID list.
func (s *serviceImpl) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := dao.Operlog.Ctx(ctx).WhereIn(colID, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// Export generates an Excel workbook for operation logs.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	model := dao.Operlog.Ctx(ctx)
	if len(in.Ids) > 0 {
		model = model.WhereIn(colID, in.Ids)
	} else {
		model = applyOperLogFilters(model, in.Title, in.OperName, in.OperType, in.Status, in.BeginTime, in.EndTime)
	}
	model = model.Limit(MaxExportRows)

	allowedSortFields := map[string]string{
		"id":        colID,
		"operTime":  colOperTime,
		"oper_time": colOperTime,
		"costTime":  colCostTime,
		"cost_time": colCostTime,
	}
	orderBy := colOperTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*OperLogEntity, 0)
	err = gdbutil.ApplyModelOrder(model, orderBy, direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer excelutil.CloseFile(file, &err)
	sheet := "Sheet1"
	headers := []string{"模块名称", "操作名称", "操作类型", "操作人", "请求方式", "请求URL", "操作IP", "请求参数", "响应结果", "状态", "错误信息", "耗时(ms)", "操作时间"}
	for index, header := range headers {
		if setErr := excelutil.SetCellValue(file, sheet, index+1, 1, header); setErr != nil {
			return nil, setErr
		}
	}

	operTypeMap := buildIntDictLabelMap(ctx, DictTypeOperType)
	statusMap := buildIntDictLabelMap(ctx, DictTypeOperStatus)
	for index, log := range list {
		row := index + 2
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(1, row), log.Title); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(2, row), log.OperSummary); setErr != nil {
			return nil, setErr
		}
		operTypeText, ok := operTypeMap[log.OperType]
		if !ok {
			operTypeText = defaultOperTypeLabels[log.OperType]
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(3, row), operTypeText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(4, row), log.OperName); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(5, row), log.RequestMethod); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(6, row), log.OperUrl); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(7, row), log.OperIp); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(8, row), log.OperParam); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(9, row), log.JsonResult); setErr != nil {
			return nil, setErr
		}
		statusText, ok := statusMap[log.Status]
		if !ok {
			statusText = defaultOperStatusLabels[log.Status]
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(10, row), statusText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(11, row), log.ErrorMsg); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(12, row), log.CostTime); setErr != nil {
			return nil, setErr
		}
		if log.OperTime != nil {
			if setErr := excelutil.SetCellValueByName(file, sheet, cellName(13, row), log.OperTime.String()); setErr != nil {
				return nil, setErr
			}
		}
	}

	var buffer bytes.Buffer
	if writeErr := file.Write(&buffer); writeErr != nil {
		return nil, writeErr
	}
	return buffer.Bytes(), nil
}

// applyOperLogFilters wires the shared operation-log query filters onto one model.
func applyOperLogFilters(model *gdb.Model, title string, operName string, operType *int, status *int, beginTime string, endTime string) *gdb.Model {
	if title != "" {
		model = model.WhereLike(colTitle, "%"+title+"%")
	}
	if operName != "" {
		model = model.WhereLike(colOperName, "%"+operName+"%")
	}
	if operType != nil {
		model = model.Where(colOperType, *operType)
	}
	if status != nil {
		model = model.Where(colStatus, *status)
	}
	if beginTime != "" {
		model = model.WhereGTE(colOperTime, beginTime)
	}
	if endTime != "" {
		model = model.WhereLTE(colOperTime, normalizeEndTime(endTime))
	}
	return model
}

// buildIntDictLabelMap builds one integer-value dictionary label map.
func buildIntDictLabelMap(ctx context.Context, dictType string) map[int]string {
	rows := make([]*dictDataRow, 0)
	err := dao.SysDictData.Ctx(ctx).
		Fields(colDictValue, colDictLabel).
		Where(colDictType, dictType).
		Where(colStatus, 1).
		OrderAsc(colDictSort).
		Scan(&rows)
	if err != nil || len(rows) == 0 {
		return map[int]string{}
	}

	labels := make(map[int]string, len(rows))
	for _, row := range rows {
		value, convErr := strconv.Atoi(row.Value)
		if convErr != nil {
			continue
		}
		labels[value] = row.Label
	}
	return labels
}

// normalizeEndTime expands date-only end values to the end of day.
func normalizeEndTime(value string) string {
	if len(value) == 10 {
		return value + " 23:59:59"
	}
	return value
}

// cellName converts numeric coordinates into an Excel A1-style cell reference.
func cellName(col int, row int) string {
	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		panic(gerror.Wrap(err, "生成Excel单元格名称失败"))
	}
	return name
}
