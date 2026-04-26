package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UsageScenesReq defines the request for querying file usage scenes.
type UsageScenesReq struct {
	g.Meta `path:"/file/scenes" method:"get" tags:"File Management" summary:"Get a list of file usage scenarios" dc:"Query the list of all used file usage scenario identifiers for reuse by filters, management views or file forms of the management workbench" permission:"system:file:query"`
}

// UsageScenesRes File usage scenes list response
type UsageScenesRes struct {
	List []*UsageSceneItem `json:"list" dc:"Usage scenario list" eg:"[]"`
}

// UsageSceneItem File usage scene item
type UsageSceneItem struct {
	Value string `json:"value" dc:"Usage scene identifier" eg:"avatar"`
	Label string `json:"label" dc:"Usage scene name" eg:"User avatar"`
}
