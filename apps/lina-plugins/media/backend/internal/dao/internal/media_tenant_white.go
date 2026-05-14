// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaTenantWhiteDao is the data access object for the table media_tenant_white.
type MediaTenantWhiteDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MediaTenantWhiteColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MediaTenantWhiteColumns defines and stores column names for the table media_tenant_white.
type MediaTenantWhiteColumns struct {
	TenantId    string // 租户ID
	Ip          string // 白名单地址
	Description string // 白名单描述
	Enable      string // 1开启，0关闭
	CreatorId   string // 创建人ID
	CreateTime  string // 创建时间
	UpdaterId   string // 修改人ID
	UpdateTime  string // 修改时间
}

// mediaTenantWhiteColumns holds the columns for the table media_tenant_white.
var mediaTenantWhiteColumns = MediaTenantWhiteColumns{
	TenantId:    "tenant_id",
	Ip:          "ip",
	Description: "description",
	Enable:      "enable",
	CreatorId:   "creator_id",
	CreateTime:  "create_time",
	UpdaterId:   "updater_id",
	UpdateTime:  "update_time",
}

// NewMediaTenantWhiteDao creates and returns a new DAO object for table data access.
func NewMediaTenantWhiteDao(handlers ...gdb.ModelHandler) *MediaTenantWhiteDao {
	return &MediaTenantWhiteDao{
		group:    "default",
		table:    "media_tenant_white",
		columns:  mediaTenantWhiteColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaTenantWhiteDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaTenantWhiteDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaTenantWhiteDao) Columns() MediaTenantWhiteColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaTenantWhiteDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaTenantWhiteDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaTenantWhiteDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
