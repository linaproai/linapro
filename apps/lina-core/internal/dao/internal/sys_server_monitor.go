// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysServerMonitorDao is the data access object for the table sys_server_monitor.
type SysServerMonitorDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SysServerMonitorColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SysServerMonitorColumns defines and stores column names for the table sys_server_monitor.
type SysServerMonitorColumns struct {
	Id        string // 记录ID
	NodeName  string // 节点名称（hostname）
	NodeIp    string // 节点IP地址
	Data      string // 监控数据（JSON格式，包含CPU、内存、磁盘、网络、Go运行时等指标）
	CreatedAt string // 采集时间
	UpdatedAt string // 更新时间
}

// sysServerMonitorColumns holds the columns for the table sys_server_monitor.
var sysServerMonitorColumns = SysServerMonitorColumns{
	Id:        "id",
	NodeName:  "node_name",
	NodeIp:    "node_ip",
	Data:      "data",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewSysServerMonitorDao creates and returns a new DAO object for table data access.
func NewSysServerMonitorDao(handlers ...gdb.ModelHandler) *SysServerMonitorDao {
	return &SysServerMonitorDao{
		group:    "default",
		table:    "sys_server_monitor",
		columns:  sysServerMonitorColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysServerMonitorDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysServerMonitorDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysServerMonitorDao) Columns() SysServerMonitorColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysServerMonitorDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysServerMonitorDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysServerMonitorDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
