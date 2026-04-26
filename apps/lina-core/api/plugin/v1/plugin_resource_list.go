package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceListReq is the request for querying plugin-owned backend resources.
type ResourceListReq struct {
	g.Meta   `path:"/plugins/{id}/resources/{resource}" method:"get" tags:"Plugin Management" summary:"Query plugin resource data" dc:"Query the plugin's own backend resource data according to the plugin's general resource contract. The resource interface is verified by the controller based on the plugin permissions declared by the resource or the plugin resource permissions deduced by default, instead of requiring additional query permissions from the plugin management background."`
	Id       string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Resource string `json:"resource" v:"required|length:1,64" dc:"Plugin resource identifier, registered by the plugin itself in the plugin directory backend implementation" eg:"records"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number, starting from 1" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of records per page, maximum 100" eg:"10"`
}

// ResourceListRes is the response for querying plugin resources.
type ResourceListRes struct {
	List  []map[string]interface{} `json:"list" dc:"List of plugin resource records. The specific field structure is determined by the plugin's own resource declaration." eg:"[]"`
	Total int                      `json:"total" dc:"Total number of records" eg:"1"`
}
