package v1

import "github.com/gogf/gf/v2/frame/g"

// SyncReq is the request for synchronizing source plugins.
type SyncReq struct {
	g.Meta `path:"/plugins/sync" method:"post" tags:"Plugin Management" summary:"Sync source plugin" permission:"plugin:install" dc:"Scan the source plugin list in the apps/lina-plugins directory and synchronize plugin metadata to the system plugin registry"`
}

// SyncRes is the response for synchronizing source plugins.
type SyncRes struct {
	Total int `json:"total" dc:"Number of source plugins after synchronization" eg:"1"`
}
