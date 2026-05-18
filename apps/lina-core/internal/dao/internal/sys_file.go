// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysFileDao is the data access object for the table sys_file.
type SysFileDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysFileColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysFileColumns defines and stores column names for the table sys_file.
type SysFileColumns struct {
	Id        string // File ID
	TenantId  string // Owning tenant ID, 0 means PLATFORM
	Name      string // Stored file name
	Original  string // Original file name
	Suffix    string // File suffix
	Scene     string // Usage scene: avatar=user avatar, notice_image=notice image, notice_attachment=notice attachment, other=other
	Size      string // File size in bytes
	Hash      string // File SHA-256 hash for deduplication
	Url       string // File access URL
	Path      string // File storage path
	Engine    string // Storage engine: local=local storage
	CreatedBy string // Uploader user ID
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// sysFileColumns holds the columns for the table sys_file.
var sysFileColumns = SysFileColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	Name:      "name",
	Original:  "original",
	Suffix:    "suffix",
	Scene:     "scene",
	Size:      "size",
	Hash:      "hash",
	Url:       "url",
	Path:      "path",
	Engine:    "engine",
	CreatedBy: "created_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysFileDao creates and returns a new DAO object for table data access.
func NewSysFileDao(handlers ...gdb.ModelHandler) *SysFileDao {
	return &SysFileDao{
		group:    "default",
		table:    "sys_file",
		columns:  sysFileColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysFileDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysFileDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysFileDao) Columns() SysFileColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysFileDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysFileDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysFileDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
