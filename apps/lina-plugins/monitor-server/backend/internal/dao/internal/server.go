// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ServerDao is the data access object for the table plugin_monitor_server.
type ServerDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ServerColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ServerColumns defines and stores column names for the table plugin_monitor_server.
type ServerColumns struct {
	Id        string // Record ID
	NodeName  string // Node name (hostname)
	NodeIp    string // Node IP address
	Data      string // Monitoring data in structured text format, including CPU, memory, disk, network, Go runtime, and other metrics
	CreatedAt string // Collection time
	UpdatedAt string // Update time
}

// serverColumns holds the columns for the table plugin_monitor_server.
var serverColumns = ServerColumns{
	Id:        "id",
	NodeName:  "node_name",
	NodeIp:    "node_ip",
	Data:      "data",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewServerDao creates and returns a new DAO object for table data access.
func NewServerDao(handlers ...gdb.ModelHandler) *ServerDao {
	return &ServerDao{
		group:    "default",
		table:    "plugin_monitor_server",
		columns:  serverColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ServerDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ServerDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ServerDao) Columns() ServerColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ServerDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ServerDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ServerDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
