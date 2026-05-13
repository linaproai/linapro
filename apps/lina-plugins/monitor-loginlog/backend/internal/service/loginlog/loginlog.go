// Package loginlog implements login-log persistence, query, cleanup, and
// export services for the monitor-loginlog source plugin. It owns the
// plugin_monitor_loginlog table access instead of depending on host-internal loginlog
// services.
package loginlog

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/excelutil"
	"lina-core/pkg/gdbutil"
	"lina-core/pkg/pluginhost"
	plugincontract "lina-core/pkg/pluginservice/contract"
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

// Runtime i18n key fragments used by dictionary display projection.
const (
	// dictKeyPrefix is the runtime i18n root for dictionary labels.
	dictKeyPrefix = "dict"
	// labelKeySuffix is the final i18n segment for dictionary display labels.
	labelKeySuffix = "label"
	// loginLogMessagePrefix is the plugin-owned runtime i18n root for auth messages.
	loginLogMessagePrefix = "plugin.monitor-loginlog.logMessage"
)

// Login status values stored in plugin_monitor_loginlog.
const (
	LoginStatusSuccess = 0
	LoginStatusFail    = 1
)

var defaultLoginStatusLabels = map[int]string{
	LoginStatusSuccess: "Success",
	LoginStatusFail:    "Failed",
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
type serviceImpl struct {
	i18nSvc      plugincontract.I18nService         // i18nSvc resolves host runtime translations for plugin data.
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned login-log rows.
}

// New creates and returns a new monitor-loginlog service instance.
func New(i18nSvc plugincontract.I18nService, tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{
		i18nSvc:      i18nSvc,
		tenantFilter: tenantFilter,
	}
}

// LoginLogEntity mirrors the plugin-local generated plugin_monitor_loginlog entity.
type LoginLogEntity = entitymodel.Loginlog

// dictDataRow reuses the plugin-local generated sys_dict_data entity.
type dictDataRow = entitymodel.SysDictData

// CreateInput defines the login-log create input.
type CreateInput struct {
	TenantID           *int
	ActingUserID       *int
	OnBehalfOfTenantID *int
	IsImpersonation    *bool
	UserName           string
	Status             int
	Ip                 string
	Browser            string
	Os                 string
	Msg                string
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

// auditTenantContext stores tenant metadata persisted with one login log.
type auditTenantContext struct {
	TenantID           int  // TenantID owns the log row.
	ActingUserID       int  // ActingUserID is the platform actor during impersonation.
	OnBehalfOfTenantID int  // OnBehalfOfTenantID is the operated tenant.
	IsImpersonation    bool // IsImpersonation marks platform impersonation.
}

// Create inserts one login-log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	auditContext := resolveAuditTenantContext(
		ctx,
		s.tenantFilter,
		in.TenantID,
		in.ActingUserID,
		in.OnBehalfOfTenantID,
		in.IsImpersonation,
	)

	_, err := dao.Loginlog.Ctx(ctx).Data(do.Loginlog{
		TenantId:           auditContext.TenantID,
		ActingUserId:       auditContext.ActingUserID,
		OnBehalfOfTenantId: auditContext.OnBehalfOfTenantID,
		IsImpersonation:    auditContext.IsImpersonation,
		UserName:           in.UserName,
		Status:             in.Status,
		Ip:                 in.Ip,
		Browser:            in.Browser,
		Os:                 in.Os,
		Msg:                in.Msg,
		LoginTime:          gtime.Now(),
	}).Insert()
	return err
}

// resolveAuditTenantContext resolves tenant audit metadata from bizctx and explicit overrides.
func resolveAuditTenantContext(
	ctx context.Context,
	tenantFilter plugincontract.TenantFilterService,
	tenantID *int,
	actingUserID *int,
	onBehalfOfTenantID *int,
	isImpersonation *bool,
) auditTenantContext {
	current := tenantFilter.CurrentContext(ctx)
	result := auditTenantContext{
		TenantID:           current.TenantID,
		ActingUserID:       current.ActingUserID,
		OnBehalfOfTenantID: current.OnBehalfOfTenantID,
		IsImpersonation:    current.IsImpersonation,
	}
	if tenantID != nil {
		result.TenantID = *tenantID
	}
	if actingUserID != nil {
		result.ActingUserID = *actingUserID
	}
	if onBehalfOfTenantID != nil {
		result.OnBehalfOfTenantID = *onBehalfOfTenantID
	}
	if isImpersonation != nil {
		result.IsImpersonation = *isImpersonation
	}
	return result
}

// List queries the paginated login-log list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx))
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
	s.localizeRecords(ctx, list)

	return &ListOutput{List: list, Total: total}, nil
}

// GetById retrieves one login-log record by primary key.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*LoginLogEntity, error) {
	var record *LoginLogEntity
	err := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx)).Where(colID, id).Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeLoginLogNotFound)
	}
	s.localizeRecord(ctx, record)
	return record, nil
}

// Clean hard-deletes login logs within one optional time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	model := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx))
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
	result, err := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx)).WhereIn(colID, ids).Delete()
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
	model := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx))
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
	defer excelutil.CloseFile(ctx, file, &err)
	sheet := "Sheet1"
	headers := s.exportHeaders(ctx)
	for index, header := range headers {
		if setErr := excelutil.SetCellValue(file, sheet, index+1, 1, header); setErr != nil {
			return nil, setErr
		}
	}

	statusMap := buildIntDictLabelMap(ctx, DictTypeLoginStatus)
	for index, log := range list {
		row := index + 2
		if setErr := excelutil.SetCellValue(file, sheet, 1, row, log.UserName); setErr != nil {
			return nil, setErr
		}
		statusText := s.exportStatusText(ctx, log.Status, statusMap)
		if setErr := excelutil.SetCellValue(file, sheet, 2, row, statusText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 3, row, log.Ip); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 4, row, log.Browser); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 5, row, log.Os); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 6, row, s.translateLoginLogMessage(ctx, log.Msg)); setErr != nil {
			return nil, setErr
		}
		if log.LoginTime != nil {
			if setErr := excelutil.SetCellValue(file, sheet, 7, row, log.LoginTime.String()); setErr != nil {
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

// exportHeaders returns localized Excel headers for login-log export.
func (s *serviceImpl) exportHeaders(ctx context.Context) []string {
	return []string{
		s.translate(ctx, "plugin.monitor-loginlog.fields.userName", "User Account"),
		s.translate(ctx, "plugin.monitor-loginlog.fields.status", "Login Status"),
		s.translate(ctx, "plugin.monitor-loginlog.fields.ipAddress", "IP Address"),
		s.translate(ctx, "plugin.monitor-loginlog.fields.browser", "Browser"),
		s.translate(ctx, "plugin.monitor-loginlog.fields.os", "Operating System"),
		s.translate(ctx, "plugin.monitor-loginlog.fields.message", "Message"),
		s.translate(ctx, "plugin.monitor-loginlog.fields.loginTime", "Login Time"),
	}
}

// exportStatusText returns the localized export label for one login status.
func (s *serviceImpl) exportStatusText(ctx context.Context, status int, statusMap map[int]string) string {
	statusText, ok := statusMap[status]
	if !ok {
		statusText = defaultLoginStatusLabels[status]
	}
	return s.translateDictLabel(ctx, DictTypeLoginStatus, strconv.Itoa(status), statusText)
}

// localizeRecords translates backend-owned display fields for login-log rows.
func (s *serviceImpl) localizeRecords(ctx context.Context, records []*LoginLogEntity) {
	for _, record := range records {
		s.localizeRecord(ctx, record)
	}
}

// localizeRecord translates backend-owned display fields for one login-log row.
func (s *serviceImpl) localizeRecord(ctx context.Context, record *LoginLogEntity) {
	if record == nil {
		return
	}
	record.Msg = s.translateLoginLogMessage(ctx, record.Msg)
}

// translateLoginLogMessage resolves stable auth lifecycle reason codes.
func (s *serviceImpl) translateLoginLogMessage(ctx context.Context, message string) string {
	key := loginLogReasonI18nKey(strings.TrimSpace(message))
	if key == "" {
		return message
	}
	return s.translate(ctx, key, message)
}

// loginLogReasonI18nKey maps published auth reason codes to plugin-owned i18n keys.
func loginLogReasonI18nKey(reason string) string {
	switch reason {
	case pluginhost.AuthHookReasonLoginSuccessful:
		return loginLogMessagePrefix + ".loginSuccessful"
	case pluginhost.AuthHookReasonLoginFailed:
		return loginLogMessagePrefix + ".loginFailed"
	case pluginhost.AuthHookReasonLogoutSuccessful:
		return loginLogMessagePrefix + ".logoutSuccessful"
	case pluginhost.AuthHookReasonInvalidCredentials:
		return loginLogMessagePrefix + ".invalidCredentials"
	case pluginhost.AuthHookReasonUserDisabled:
		return loginLogMessagePrefix + ".userDisabled"
	case pluginhost.AuthHookReasonIPBlacklisted:
		return loginLogMessagePrefix + ".ipBlacklisted"
	}
	return ""
}

// translateDictLabel translates one dictionary label through runtime i18n keys.
func (s *serviceImpl) translateDictLabel(ctx context.Context, dictType string, value string, fallback string) string {
	key := strings.Join([]string{dictKeyPrefix, dictType, value, labelKeySuffix}, ".")
	return s.translate(ctx, key, fallback)
}

// translate resolves one runtime i18n key through the host i18n service.
func (s *serviceImpl) translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, key, fallback)
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
