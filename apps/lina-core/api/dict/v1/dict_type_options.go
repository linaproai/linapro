package v1

import "github.com/gogf/gf/v2/frame/g"

// TypeOptionsReq defines the request for querying dictionary type options.
type TypeOptionsReq struct {
	g.Meta `path:"/dict/type/options" method:"get" tags:"Dictionary Management" summary:"Get all dictionary type options" dc:"Gets a list of all dictionary types in their normal state for use in the Dictionary Maintenance view of the Administration Workbench or other type selector assembly options" permission:"system:dict:query"`
}

// TypeOptionsRes defines the response for querying dictionary type options.
type TypeOptionsRes struct {
	List []*DictTypeOptionItem `json:"list" dc:"Dictionary type options list" eg:"[]"`
}
