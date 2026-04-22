// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OperlogDao is the data access object for the table plugin_monitor_operlog.
type OperlogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  OperlogColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// OperlogColumns defines and stores column names for the table plugin_monitor_operlog.
type OperlogColumns struct {
	Id            string // 日志ID
	Title         string // 模块标题
	OperSummary   string // 操作摘要
	OperType      string // 操作类型（create新增 update修改 delete删除 export导出 import导入 other其他）
	Method        string // 方法名称
	RequestMethod string // 请求方式（GET/POST/PUT/DELETE）
	OperName      string // 操作人员
	OperUrl       string // 请求URL
	OperIp        string // 操作IP地址
	OperParam     string // 请求参数
	JsonResult    string // 返回参数
	Status        string // 操作状态（0成功 1失败）
	ErrorMsg      string // 错误消息
	CostTime      string // 耗时（毫秒）
	OperTime      string // 操作时间
}

// operlogColumns holds the columns for the table plugin_monitor_operlog.
var operlogColumns = OperlogColumns{
	Id:            "id",
	Title:         "title",
	OperSummary:   "oper_summary",
	OperType:      "oper_type",
	Method:        "method",
	RequestMethod: "request_method",
	OperName:      "oper_name",
	OperUrl:       "oper_url",
	OperIp:        "oper_ip",
	OperParam:     "oper_param",
	JsonResult:    "json_result",
	Status:        "status",
	ErrorMsg:      "error_msg",
	CostTime:      "cost_time",
	OperTime:      "oper_time",
}

// NewOperlogDao creates and returns a new DAO object for table data access.
func NewOperlogDao(handlers ...gdb.ModelHandler) *OperlogDao {
	return &OperlogDao{
		group:    "default",
		table:    "plugin_monitor_operlog",
		columns:  operlogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OperlogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OperlogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OperlogDao) Columns() OperlogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OperlogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OperlogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *OperlogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
