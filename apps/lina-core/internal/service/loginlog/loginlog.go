// Package loginlog implements login-log query, cleanup, and export services
// for the Lina core host service.
package loginlog

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

// Dict type used in login log
const DictTypeLoginStatus = "sys_login_status" // Login status dictionary

// Login status values (matching sys_login_status dictionary)
const (
	LoginStatusSuccess = 0 // 成功
	LoginStatusFail    = 1 // 失败
)

// Service defines the loginlog service contract.
type Service interface {
	// Create inserts a new login log record.
	Create(ctx context.Context, in CreateInput) error
	// List queries login log list with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves login log by ID.
	GetById(ctx context.Context, id int) (*entity.SysLoginLog, error)
	// Clean hard-deletes login logs by time range.
	Clean(ctx context.Context, in CleanInput) (int, error)
	// DeleteByIds hard-deletes login logs by IDs.
	DeleteByIds(ctx context.Context, ids []int) (int, error)
	// Export generates an Excel file with login log data (max 10000 rows).
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
	UserName string
	Status   int
	Ip       string
	Browser  string
	Os       string
	Msg      string
}

// Create inserts a new login log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	_, err := dao.SysLoginLog.Ctx(ctx).Data(do.SysLoginLog{
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

// ListInput defines input for List function.
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

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysLoginLog
	Total int
}

// List queries login log list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	cols := dao.SysLoginLog.Columns()
	m := dao.SysLoginLog.Ctx(ctx)

	if in.UserName != "" {
		m = m.WhereLike(cols.UserName, "%"+in.UserName+"%")
	}
	if in.Ip != "" {
		m = m.WhereLike(cols.Ip, "%"+in.Ip+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.LoginTime, in.BeginTime)
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.LoginTime, endTime)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var (
		orderBy           = cols.LoginTime
		allowedSortFields = map[string]string{
			"id":        cols.Id,
			"loginTime": cols.LoginTime,
		}
		direction = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)
	if in.OrderBy != "" {
		if field, ok := allowedSortFields[in.OrderBy]; ok {
			orderBy = field
		}
	}

	var list []*entity.SysLoginLog
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

// GetById retrieves login log by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysLoginLog, error) {
	var record *entity.SysLoginLog
	err := dao.SysLoginLog.Ctx(ctx).
		Where(do.SysLoginLog{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("登录日志不存在")
	}
	return record, nil
}

// CleanInput defines input for Clean function.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// Clean hard-deletes login logs by time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	cols := dao.SysLoginLog.Columns()
	m := dao.SysLoginLog.Ctx(ctx)

	hasFilter := false
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.LoginTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.LoginTime, endTime)
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

// DeleteByIds hard-deletes login logs by IDs.
func (s *serviceImpl) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := dao.SysLoginLog.Ctx(ctx).WhereIn(dao.SysLoginLog.Columns().Id, ids).Delete()
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
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
	Ids            []int // Specific IDs to export; if empty, export all matching records
}

// Export generates an Excel file with login log data (max 10000 rows).
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysLoginLog.Columns()
	m := dao.SysLoginLog.Ctx(ctx)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.UserName != "" {
			m = m.WhereLike(cols.UserName, "%"+in.UserName+"%")
		}
		if in.Ip != "" {
			m = m.WhereLike(cols.Ip, "%"+in.Ip+"%")
		}
		if in.Status != nil {
			m = m.Where(cols.Status, *in.Status)
		}
		if in.BeginTime != "" {
			m = m.WhereGTE(cols.LoginTime, in.BeginTime)
		}
		if in.EndTime != "" {
			endTime := in.EndTime
			if len(endTime) == 10 {
				endTime += " 23:59:59"
			}
			m = m.WhereLTE(cols.LoginTime, endTime)
		}
	}

	// Limit export to prevent memory issues
	m = m.Limit(MaxExportRows)

	var (
		orderBy   = cols.LoginTime
		direction = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)

	var list []*entity.SysLoginLog
	err = gdbutil.ApplyModelOrder(m, orderBy, direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"用户名", "状态", "IP地址", "浏览器", "操作系统", "提示消息", "登录时间"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	// Build label map from dictionary for batch lookups
	statusMap := s.dictSvc.BuildIntLabelMap(ctx, DictTypeLoginStatus)

	for i, log := range list {
		row := i + 2
		if err = setCellValueByName(f, sheet, cellName(1, row), log.UserName); err != nil {
			return nil, err
		}
		// Use dictionary lookup for status
		statusText, ok := statusMap[log.Status]
		if !ok {
			statusText = s.dictSvc.GetLabelByIntValue(ctx, DictTypeLoginStatus, 0) // fallback to "成功"
		}
		if err = setCellValueByName(f, sheet, cellName(2, row), statusText); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(3, row), log.Ip); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(4, row), log.Browser); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(5, row), log.Os); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(6, row), log.Msg); err != nil {
			return nil, err
		}
		if log.LoginTime != nil {
			if err = setCellValueByName(f, sheet, cellName(7, row), log.LoginTime.String()); err != nil {
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
