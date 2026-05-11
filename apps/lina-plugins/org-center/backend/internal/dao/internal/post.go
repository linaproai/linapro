// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PostDao is the data access object for the table plugin_org_center_post.
type PostDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PostColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PostColumns defines and stores column names for the table plugin_org_center_post.
type PostColumns struct {
	Id        string // Post ID
	TenantId  string // Owning tenant ID, 0 means PLATFORM
	DeptId    string // Owning department ID
	Code      string // Post code
	Name      string // Post name
	Sort      string // Display order
	Status    string // Status: 0=disabled, 1=enabled
	Remark    string // Remark
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// postColumns holds the columns for the table plugin_org_center_post.
var postColumns = PostColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	DeptId:    "dept_id",
	Code:      "code",
	Name:      "name",
	Sort:      "sort",
	Status:    "status",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewPostDao creates and returns a new DAO object for table data access.
func NewPostDao(handlers ...gdb.ModelHandler) *PostDao {
	return &PostDao{
		group:    "default",
		table:    "plugin_org_center_post",
		columns:  postColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PostDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PostDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PostDao) Columns() PostColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PostDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PostDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PostDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
