package v1

import "github.com/gogf/gf/v2/frame/g"

// DynamicListReq is the request for querying public dynamic-plugin states.
type DynamicListReq struct {
	g.Meta `path:"/plugins/dynamic" method:"get" tags:"插件管理" summary:"查询插件运行状态" dc:"返回前端公共界面层渲染插件 Slot 所需的最小运行状态集合，供登录页和布局界面在匿名或登录态下判断插件内容是否应显示"`
}

// DynamicListRes is the response for querying public dynamic-plugin states.
type DynamicListRes struct {
	List []*PluginDynamicItem `json:"list" dc:"插件运行状态列表" eg:"[]"`
}

// PluginDynamicItem represents public dynamic state of one plugin.
type PluginDynamicItem struct {
	Id         string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-dynamic"`
	Installed  int    `json:"installed" dc:"安装状态：1=已安装/已集成 0=未安装" eg:"1"`
	Enabled    int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"1"`
	Version    string `json:"version" dc:"插件当前生效版本号；若仅上传未切换则仍返回旧版本" eg:"v0.1.0"`
	Generation int64  `json:"generation" dc:"插件当前生效代际号；前端可据此判断当前插件页面是否需要刷新" eg:"3"`
	StatusKey  string `json:"statusKey" dc:"插件状态在系统插件注册表中的定位键名" eg:"sys_plugin.status:plugin-demo-dynamic"`
}
