// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserExternalIdentityDao is the data access object for the table sys_user_external_identity.
type SysUserExternalIdentityDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  SysUserExternalIdentityColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// SysUserExternalIdentityColumns defines and stores column names for the table sys_user_external_identity.
type SysUserExternalIdentityColumns struct {
	Id            string // External identity linkage ID
	UserId        string // Linked local sys_user ID
	Provider      string // Stable external provider ID owned by the declaring plugin, e.g. google, discord
	Subject       string // Immutable provider-issued subject identifier, e.g. OIDC sub
	PluginId      string // Source-plugin ID that owns the provider and created the linkage
	EmailSnapshot string // Email captured at link time for audit only, never used as a resolution key
	CreatedAt     string // Creation time
	UpdatedAt     string // Update time
}

// sysUserExternalIdentityColumns holds the columns for the table sys_user_external_identity.
var sysUserExternalIdentityColumns = SysUserExternalIdentityColumns{
	Id:            "id",
	UserId:        "user_id",
	Provider:      "provider",
	Subject:       "subject",
	PluginId:      "plugin_id",
	EmailSnapshot: "email_snapshot",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewSysUserExternalIdentityDao creates and returns a new DAO object for table data access.
func NewSysUserExternalIdentityDao(handlers ...gdb.ModelHandler) *SysUserExternalIdentityDao {
	return &SysUserExternalIdentityDao{
		group:    "default",
		table:    "sys_user_external_identity",
		columns:  sysUserExternalIdentityColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysUserExternalIdentityDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysUserExternalIdentityDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysUserExternalIdentityDao) Columns() SysUserExternalIdentityColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysUserExternalIdentityDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysUserExternalIdentityDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysUserExternalIdentityDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
