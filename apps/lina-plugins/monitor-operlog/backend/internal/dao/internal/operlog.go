// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OperlogDao is the data access object for the table plugin_monitor_operlog.
type OperlogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  OperlogColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// OperlogColumns defines and stores column names for the table plugin_monitor_operlog.
type OperlogColumns struct {
	Id                 string // Log ID
	TenantId           string // Owning tenant ID, 0 means PLATFORM
	ActingUserId       string // Actual acting user ID for platform operations or impersonation
	OnBehalfOfTenantId string // Target tenant ID when a platform administrator acts on behalf of a tenant
	IsImpersonation    string // Whether this log was produced during tenant impersonation
	Title              string // Module title
	OperSummary        string // Operation summary
	RouteOwner         string // Route owner: core or plugin ID
	RouteMethod        string // Route request method
	RoutePath          string // Route path
	RouteDocKey        string // API documentation structured key
	OperType           string // Operation type: create=create, update=update, delete=delete, export=export, import=import, other=other
	Method             string // Method name
	RequestMethod      string // Request method: GET/POST/PUT/DELETE
	OperName           string // Operator
	OperUrl            string // Request URL
	OperIp             string // Operation IP address
	OperParam          string // Request parameters
	JsonResult         string // Response parameters
	Status             string // Operation status: 0=succeeded, 1=failed
	ErrorMsg           string // Error message
	CostTime           string // Duration in milliseconds
	OperTime           string // Operation time
}

// operlogColumns holds the columns for the table plugin_monitor_operlog.
var operlogColumns = OperlogColumns{
	Id:                 "id",
	TenantId:           "tenant_id",
	ActingUserId:       "acting_user_id",
	OnBehalfOfTenantId: "on_behalf_of_tenant_id",
	IsImpersonation:    "is_impersonation",
	Title:              "title",
	OperSummary:        "oper_summary",
	RouteOwner:         "route_owner",
	RouteMethod:        "route_method",
	RoutePath:          "route_path",
	RouteDocKey:        "route_doc_key",
	OperType:           "oper_type",
	Method:             "method",
	RequestMethod:      "request_method",
	OperName:           "oper_name",
	OperUrl:            "oper_url",
	OperIp:             "oper_ip",
	OperParam:          "oper_param",
	JsonResult:         "json_result",
	Status:             "status",
	ErrorMsg:           "error_msg",
	CostTime:           "cost_time",
	OperTime:           "oper_time",
}

// NewOperlogDao creates and returns a new DAO object for table data access.
func NewOperlogDao(handlers ...gdb.ModelHandler) *OperlogDao {
	return &OperlogDao{
		group:    "default",
		table:    "plugin_monitor_operlog",
		columns:  operlogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OperlogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OperlogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OperlogDao) Columns() OperlogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OperlogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OperlogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *OperlogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
