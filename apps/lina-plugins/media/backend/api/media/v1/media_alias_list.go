// This file declares media stream alias list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListAliasesReq defines the request for querying stream aliases.
type ListAliasesReq struct {
	g.Meta   `path:"/media/stream-aliases" method:"get" tags:"媒体管理" summary:"查询流别名列表" dc:"分页查询流别名，支持按别名或真实流路径模糊筛选。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按别名或真实流路径模糊筛选" eg:"live"`
}

// ListAliasesRes defines the response for querying stream aliases.
type ListAliasesRes struct {
	List  []*AliasListItem `json:"list" dc:"流别名列表" eg:"[]"`
	Total int              `json:"total" dc:"匹配总数" eg:"1"`
}

// AliasListItem defines one stream alias row.
type AliasListItem struct {
	Id         int64  `json:"id" dc:"ID" eg:"1"`
	Alias      string `json:"alias" dc:"流别名" eg:"camera-01"`
	AutoRemove int    `json:"autoRemove" dc:"是否自动移除：1是，0否" eg:"0"`
	StreamPath string `json:"streamPath" dc:"真实流路径" eg:"live/camera-01"`
	CreateTime  string `json:"createTime" dc:"创建时间" eg:"2026-05-13 10:00:00"`
}
