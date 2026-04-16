package v1

import "github.com/gogf/gf/v2/frame/g"

// SyncReq is the request for synchronizing source plugins.
type SyncReq struct {
	g.Meta `path:"/plugins/sync" method:"post" tags:"插件管理" summary:"同步源码插件" permission:"plugin:install" dc:"扫描apps/lina-plugins目录下的源码插件清单，并将插件元数据同步到系统插件注册表"`
}

// SyncRes is the response for synchronizing source plugins.
type SyncRes struct {
	Total int `json:"total" dc:"同步后的源码插件数量" eg:"1"`
}
