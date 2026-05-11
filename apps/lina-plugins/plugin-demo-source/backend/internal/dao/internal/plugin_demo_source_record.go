// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginDemoSourceRecordDao is the data access object for the table plugin_demo_source_record.
type PluginDemoSourceRecordDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  PluginDemoSourceRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// PluginDemoSourceRecordColumns defines and stores column names for the table plugin_demo_source_record.
type PluginDemoSourceRecordColumns struct {
	Id             string // Primary key ID
	TenantId       string // Owning tenant ID, 0 means PLATFORM
	Title          string // Record title
	Content        string // Record content
	AttachmentName string // Original attachment file name
	AttachmentPath string // Relative attachment storage path
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
}

// pluginDemoSourceRecordColumns holds the columns for the table plugin_demo_source_record.
var pluginDemoSourceRecordColumns = PluginDemoSourceRecordColumns{
	Id:             "id",
	TenantId:       "tenant_id",
	Title:          "title",
	Content:        "content",
	AttachmentName: "attachment_name",
	AttachmentPath: "attachment_path",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewPluginDemoSourceRecordDao creates and returns a new DAO object for table data access.
func NewPluginDemoSourceRecordDao(handlers ...gdb.ModelHandler) *PluginDemoSourceRecordDao {
	return &PluginDemoSourceRecordDao{
		group:    "default",
		table:    "plugin_demo_source_record",
		columns:  pluginDemoSourceRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginDemoSourceRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginDemoSourceRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginDemoSourceRecordDao) Columns() PluginDemoSourceRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginDemoSourceRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginDemoSourceRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginDemoSourceRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
