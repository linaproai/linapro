// Package loginlog implements login-log persistence, query, cleanup, and
// export services for the monitor-loginlog source plugin. It owns the
// plugin_monitor_loginlog table access instead of depending on host-internal loginlog
// services.
package loginlog

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
	"lina-plugin-monitor-loginlog/backend/internal/dao"
	"lina-plugin-monitor-loginlog/backend/internal/model/do"
	entitymodel "lina-plugin-monitor-loginlog/backend/internal/model/entity"
)

// Table, column, and dictionary constants used by the plugin-owned login-log service.
const (
	colID        = "id"
	colUserName  = "user_name"
	colStatus    = "status"
	colIP        = "ip"
	colBrowser   = "browser"
	colOS        = "os"
	colMsg       = "msg"
	colLoginTime = "login_time"

	colDictType  = "dict_type"
	colDictValue = "value"
	colDictLabel = "label"
	colDictSort  = "sort"
)

// Login-log export and dictionary constants.
const (
	MaxExportRows       = 10000
	DictTypeLoginStatus = "sys_login_status"
)

// Login status values stored in plugin_monitor_loginlog.
const (
	LoginStatusSuccess = 0
	LoginStatusFail    = 1
)

var defaultLoginStatusLabels = map[int]string{
	LoginStatusSuccess: "成功",
	LoginStatusFail:    "失败",
}

// Service defines the monitor-loginlog service contract.
type Service interface {
	// Create inserts one login-log record.
	Create(ctx context.Context, in CreateInput) error
	// List queries the paginated login-log list.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves one login-log record by primary key.
	GetById(ctx context.Context, id int) (*LoginLogEntity, error)
	// Clean hard-deletes login logs within one optional time range.
	Clean(ctx context.Context, in CleanInput) (int, error)
	// DeleteByIds hard-deletes login logs by ID list.
	DeleteByIds(ctx context.Context, ids []int) (int, error)
	// Export generates an Excel workbook for login logs.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a new monitor-loginlog service instance.
func New() Service {
	return &serviceImpl{}
}

// LoginLogEntity mirrors the plugin-local generated plugin_monitor_loginlog entity.
type LoginLogEntity = entitymodel.Loginlog

// dictDataRow reuses the plugin-local generated sys_dict_data entity.
type dictDataRow = entitymodel.SysDictData

// CreateInput defines the login-log create input.
type CreateInput struct {
	UserName string
	Status   int
	Ip       string
	Browser  string
	Os       string
	Msg      string
}

// ListInput defines the login-log list filter input.
type ListInput struct {
	PageNum        int
	PageSize       int
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines the login-log list output.
type ListOutput struct {
	List  []*LoginLogEntity
	Total int
}

// CleanInput defines the login-log cleanup input.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// ExportInput defines the login-log export input.
type ExportInput struct {
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
	Ids            []int
}

// Create inserts one login-log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	_, err := dao.Loginlog.Ctx(ctx).Data(do.Loginlog{
		UserName:  in.UserName,
		Status:    in.Status,
		Ip:        in.Ip,
		Browser:   in.Browser,
		Os:        in.Os,
		Msg:       in.Msg,
		LoginTime: gtime.Now(),
	}).Insert()
	return err
}

// List queries the paginated login-log list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := dao.Loginlog.Ctx(ctx)
	model = applyLoginLogFilters(model, in.UserName, in.Ip, in.Status, in.BeginTime, in.EndTime)

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	allowedSortFields := map[string]string{
		"id":         colID,
		"loginTime":  colLoginTime,
		"login_time": colLoginTime,
	}
	orderBy := colLoginTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*LoginLogEntity, 0)
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

// GetById retrieves one login-log record by primary key.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*LoginLogEntity, error) {
	var record *LoginLogEntity
	err := dao.Loginlog.Ctx(ctx).Where(colID, id).Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("登录日志不存在")
	}
	return record, nil
}

// Clean hard-deletes login logs within one optional time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	model := dao.Loginlog.Ctx(ctx)
	hasFilter := false
	if in.BeginTime != "" {
		model = model.WhereGTE(colLoginTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		model = model.WhereLTE(colLoginTime, normalizeEndTime(in.EndTime))
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

// DeleteByIds hard-deletes login logs by ID list.
func (s *serviceImpl) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := dao.Loginlog.Ctx(ctx).WhereIn(colID, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// Export generates an Excel workbook for login logs.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	model := dao.Loginlog.Ctx(ctx)
	if len(in.Ids) > 0 {
		model = model.WhereIn(colID, in.Ids)
	} else {
		model = applyLoginLogFilters(model, in.UserName, in.Ip, in.Status, in.BeginTime, in.EndTime)
	}
	model = model.Limit(MaxExportRows)

	allowedSortFields := map[string]string{
		"id":         colID,
		"loginTime":  colLoginTime,
		"login_time": colLoginTime,
	}
	orderBy := colLoginTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*LoginLogEntity, 0)
	err = gdbutil.ApplyModelOrder(model, orderBy, direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer excelutil.CloseFile(file, &err)
	sheet := "Sheet1"
	headers := []string{"用户名", "状态", "IP地址", "浏览器", "操作系统", "提示消息", "登录时间"}
	for index, header := range headers {
		if setErr := excelutil.SetCellValue(file, sheet, index+1, 1, header); setErr != nil {
			return nil, setErr
		}
	}

	statusMap := buildIntDictLabelMap(ctx, DictTypeLoginStatus)
	for index, log := range list {
		row := index + 2
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(1, row), log.UserName); setErr != nil {
			return nil, setErr
		}
		statusText, ok := statusMap[log.Status]
		if !ok {
			statusText = defaultLoginStatusLabels[log.Status]
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(2, row), statusText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(3, row), log.Ip); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(4, row), log.Browser); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(5, row), log.Os); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValueByName(file, sheet, cellName(6, row), log.Msg); setErr != nil {
			return nil, setErr
		}
		if log.LoginTime != nil {
			if setErr := excelutil.SetCellValueByName(file, sheet, cellName(7, row), log.LoginTime.String()); setErr != nil {
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

// applyLoginLogFilters wires the shared login-log query filters onto one model.
func applyLoginLogFilters(model *gdb.Model, userName string, ip string, status *int, beginTime string, endTime string) *gdb.Model {
	if userName != "" {
		model = model.WhereLike(colUserName, "%"+userName+"%")
	}
	if ip != "" {
		model = model.WhereLike(colIP, "%"+ip+"%")
	}
	if status != nil {
		model = model.Where(colStatus, *status)
	}
	if beginTime != "" {
		model = model.WhereGTE(colLoginTime, beginTime)
	}
	if endTime != "" {
		model = model.WhereLTE(colLoginTime, normalizeEndTime(endTime))
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
