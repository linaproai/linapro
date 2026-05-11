// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// LoginlogDao is the data access object for the table plugin_monitor_loginlog.
type LoginlogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  LoginlogColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// LoginlogColumns defines and stores column names for the table plugin_monitor_loginlog.
type LoginlogColumns struct {
	Id                 string // Log ID
	TenantId           string // Owning tenant ID, 0 means PLATFORM
	ActingUserId       string // Actual acting user ID for platform operations or impersonation
	OnBehalfOfTenantId string // Target tenant ID when a platform administrator acts on behalf of a tenant
	IsImpersonation    string // Whether this log was produced during tenant impersonation
	UserName           string // Login account
	Status             string // Login status: 0=succeeded, 1=failed
	Ip                 string // Login IP address
	Browser            string // Browser type
	Os                 string // Operating system
	Msg                string // Prompt message
	LoginTime          string // Login time
}

// loginlogColumns holds the columns for the table plugin_monitor_loginlog.
var loginlogColumns = LoginlogColumns{
	Id:                 "id",
	TenantId:           "tenant_id",
	ActingUserId:       "acting_user_id",
	OnBehalfOfTenantId: "on_behalf_of_tenant_id",
	IsImpersonation:    "is_impersonation",
	UserName:           "user_name",
	Status:             "status",
	Ip:                 "ip",
	Browser:            "browser",
	Os:                 "os",
	Msg:                "msg",
	LoginTime:          "login_time",
}

// NewLoginlogDao creates and returns a new DAO object for table data access.
func NewLoginlogDao(handlers ...gdb.ModelHandler) *LoginlogDao {
	return &LoginlogDao{
		group:    "default",
		table:    "plugin_monitor_loginlog",
		columns:  loginlogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *LoginlogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *LoginlogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *LoginlogDao) Columns() LoginlogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *LoginlogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *LoginlogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *LoginlogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
