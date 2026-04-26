// =================================================================================
// Code generated and maintained by GoFrame CLI tool. Do not modify.
// =================================================================================

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// FileSuffixesReq Returns the list of file suffixes existing in the database
type FileSuffixesReq struct {
	g.Meta `path:"/file/suffixes" method:"get" tags:"File Management" summary:"Get a list of file types" dc:"Query the file suffix list that already exists in the database for frontend dropdown selection" permission:"system:file:query"`
}

// FileSuffixesRes File suffix list response
type FileSuffixesRes struct {
	List []*FileSuffixItem `json:"list" dc:"File suffix list" eg:"[]"`
}

// FileSuffixItem File suffix item
type FileSuffixItem struct {
	Value string `json:"value" dc:"File suffix value" eg:"txt"`
	Label string `json:"label" dc:"file suffix display name" eg:".txt"`
}
