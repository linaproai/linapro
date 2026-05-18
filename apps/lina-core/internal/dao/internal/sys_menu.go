// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysMenuDao is the data access object for the table sys_menu.
type SysMenuDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysMenuColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysMenuColumns defines and stores column names for the table sys_menu.
type SysMenuColumns struct {
	Id         string // Menu ID
	ParentId   string // Parent menu ID, 0 means root menu
	MenuKey    string // Stable menu business key
	Name       string // Menu name with i18n support
	Path       string // Route path
	Component  string // Component path
	Perms      string // Permission identifier
	Icon       string // Menu icon
	Type       string // Menu type: D=directory, M=menu, B=button
	Sort       string // Display order
	Visible    string // Visibility: 1=visible, 0=hidden
	Status     string // Status: 0=disabled, 1=enabled
	IsFrame    string // External link flag: 1=yes, 0=no
	IsCache    string // Cache flag: 1=yes, 0=no
	QueryParam string // Route parameters in JSON format
	Remark     string // Remark
	CreatedAt  string // Creation time
	UpdatedAt  string // Update time
	DeletedAt  string // Deletion time
}

// sysMenuColumns holds the columns for the table sys_menu.
var sysMenuColumns = SysMenuColumns{
	Id:         "id",
	ParentId:   "parent_id",
	MenuKey:    "menu_key",
	Name:       "name",
	Path:       "path",
	Component:  "component",
	Perms:      "perms",
	Icon:       "icon",
	Type:       "type",
	Sort:       "sort",
	Visible:    "visible",
	Status:     "status",
	IsFrame:    "is_frame",
	IsCache:    "is_cache",
	QueryParam: "query_param",
	Remark:     "remark",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewSysMenuDao creates and returns a new DAO object for table data access.
func NewSysMenuDao(handlers ...gdb.ModelHandler) *SysMenuDao {
	return &SysMenuDao{
		group:    "default",
		table:    "sys_menu",
		columns:  sysMenuColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysMenuDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysMenuDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysMenuDao) Columns() SysMenuColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysMenuDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysMenuDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysMenuDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
