// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginMigrationDao is the data access object for the table sys_plugin_migration.
type SysPluginMigrationDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  SysPluginMigrationColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// SysPluginMigrationColumns defines and stores column names for the table sys_plugin_migration.
type SysPluginMigrationColumns struct {
	Id             string // Primary key ID
	PluginId       string // Plugin unique identifier (kebab-case)
	ReleaseId      string // Owning plugin release ID
	Phase          string // Migration phase: install/uninstall/upgrade/rollback
	MigrationKey   string // Migration execution key such as install-step-001, without concrete SQL path
	Checksum       string // Migration file checksum
	ExecutionOrder string // Execution order starting from 1
	Status         string // Execution status: pending/succeeded/failed/skipped
	ExecutedAt     string // Execution time
	ErrorMessage   string // Failure reason or additional description
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
}

// sysPluginMigrationColumns holds the columns for the table sys_plugin_migration.
var sysPluginMigrationColumns = SysPluginMigrationColumns{
	Id:             "id",
	PluginId:       "plugin_id",
	ReleaseId:      "release_id",
	Phase:          "phase",
	MigrationKey:   "migration_key",
	Checksum:       "checksum",
	ExecutionOrder: "execution_order",
	Status:         "status",
	ExecutedAt:     "executed_at",
	ErrorMessage:   "error_message",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewSysPluginMigrationDao creates and returns a new DAO object for table data access.
func NewSysPluginMigrationDao(handlers ...gdb.ModelHandler) *SysPluginMigrationDao {
	return &SysPluginMigrationDao{
		group:    "default",
		table:    "sys_plugin_migration",
		columns:  sysPluginMigrationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPluginMigrationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPluginMigrationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPluginMigrationDao) Columns() SysPluginMigrationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPluginMigrationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPluginMigrationDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPluginMigrationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
