package v1

import "github.com/gogf/gf/v2/frame/g"

// SyncReq is the request for synchronizing source plugins.
type SyncReq struct {
	g.Meta `path:"/plugins/sync" method:"post" tags:"Plugin Management" summary:"Sync source plugins" permission:"plugin:install" dc:"Synchronize source plugins discovered by the running host into the system plugin registry"`
}

// SyncRes is the response for synchronizing source plugins.
type SyncRes struct {
	Total int `json:"total" dc:"Number of source plugins after synchronization" eg:"1"`
}
