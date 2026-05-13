// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaStreamAliasDao is the data access object for the table media_stream_alias.
type MediaStreamAliasDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MediaStreamAliasColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MediaStreamAliasColumns defines and stores column names for the table media_stream_alias.
type MediaStreamAliasColumns struct {
	Id         string // ID
	Alias      string // 流别名
	AutoRemove string // 是否自动移除：1是，0否
	StreamPath string // 真实流路径
	CreateTime string // 创建时间
}

// mediaStreamAliasColumns holds the columns for the table media_stream_alias.
var mediaStreamAliasColumns = MediaStreamAliasColumns{
	Id:         "id",
	Alias:      "alias",
	AutoRemove: "auto_remove",
	StreamPath: "stream_path",
	CreateTime: "create_time",
}

// NewMediaStreamAliasDao creates and returns a new DAO object for table data access.
func NewMediaStreamAliasDao(handlers ...gdb.ModelHandler) *MediaStreamAliasDao {
	return &MediaStreamAliasDao{
		group:    "default",
		table:    "media_stream_alias",
		columns:  mediaStreamAliasColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaStreamAliasDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaStreamAliasDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaStreamAliasDao) Columns() MediaStreamAliasColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaStreamAliasDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaStreamAliasDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaStreamAliasDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
