// This file defines public plugin runtime-state API DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DynamicListReq is the request for querying public dynamic-plugin states.
type DynamicListReq struct {
	g.Meta `path:"/plugins/dynamic" method:"get" tags:"Plugin Management" summary:"Query plugin running status" dc:"Returns the minimum set of running states required for the frontend public interface layer to render the plugin Slot, which can be used by the login page and layout interface to determine whether the plugin content should be displayed in the anonymous or logged-in state."`
}

// DynamicListRes is the response for querying public dynamic-plugin states.
type DynamicListRes struct {
	List []*PluginDynamicItem `json:"list" dc:"Plugin running status list" eg:"[]"`
}

// PluginDynamicItem represents public dynamic state of one plugin.
type PluginDynamicItem struct {
	Id           string `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-dynamic"`
	Installed    int    `json:"installed" dc:"Installation status: 1=Installed/Integrated 0=Not installed" eg:"1"`
	Enabled      int    `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"1"`
	Version      string `json:"version" dc:"The current effective version number of the plugin; if it is only uploaded without switching, it will still return to the old version." eg:"v0.1.0"`
	Generation   int64  `json:"generation" dc:"The plugin's current effective generation number; the frontend can use this to determine whether the current plugin page needs to be refreshed." eg:"3"`
	StatusKey    string `json:"statusKey" dc:"The location key name of the plugin status in the system plugin registry" eg:"sys_plugin.status:plugin-demo-dynamic"`
	RuntimeState string `json:"runtimeState" dc:"Plugin runtime upgrade state. Frontend business entries may execute only when this value is normal or empty for old servers." eg:"normal"`
}
