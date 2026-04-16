// =================================================================================
// Code generated and maintained by GoFrame CLI tool. Do not modify.
// =================================================================================

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// FileSuffixesReq Returns the list of file suffixes existing in the database
type FileSuffixesReq struct {
	g.Meta `path:"/file/suffixes" method:"get" tags:"文件管理" summary:"获取文件类型列表" dc:"查询数据库中已存在的文件后缀列表，用于前端下拉选择" permission:"system:file:query"`
}

// FileSuffixesRes File suffix list response
type FileSuffixesRes struct {
	List []*FileSuffixItem `json:"list" dc:"文件后缀列表" eg:"[]"`
}

// FileSuffixItem File suffix item
type FileSuffixItem struct {
	Value string `json:"value" dc:"文件后缀值" eg:"txt"`
	Label string `json:"label" dc:"文件后缀显示名" eg:".txt"`
}
