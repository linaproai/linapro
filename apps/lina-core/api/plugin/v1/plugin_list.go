package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq is the request for querying plugin list.
type ListReq struct {
	g.Meta    `path:"/plugins" method:"get" tags:"插件管理" summary:"查询插件列表" permission:"plugin:query" dc:"扫描源码插件目录并同步插件基础状态，返回插件清单与启用状态"`
	Id        string `json:"id" dc:"按插件唯一标识筛选，模糊匹配，不传则查询全部" eg:"plugin-demo-source"`
	Name      string `json:"name" dc:"按插件名称筛选，模糊匹配，不传则查询全部" eg:"源码插件示例"`
	Type      string `json:"type" dc:"按插件类型筛选：source=源码插件 dynamic=动态插件，不传则查询全部；当前动态插件实现仅支持WASM" eg:"dynamic"`
	Status    *int   `json:"status" dc:"按启用状态筛选：1=启用 0=禁用，不传则查询全部" eg:"1"`
	Installed *int   `json:"installed" dc:"按安装状态筛选：1=已安装/已集成 0=未安装，不传则查询全部；源码插件默认视为已集成" eg:"1"`
}

// ListRes is the response for querying plugin list.
type ListRes struct {
	List  []*PluginItem `json:"list" dc:"插件列表" eg:"[]"`
	Total int           `json:"total" dc:"插件总数" eg:"1"`
}

// PluginItem represents plugin information.
type PluginItem struct {
	Id                     string                       `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Name                   string                       `json:"name" dc:"插件名称" eg:"源码插件示例"`
	Version                string                       `json:"version" dc:"插件当前清单版本号" eg:"v0.1.0"`
	Type                   string                       `json:"type" dc:"插件一级类型：source=源码插件 dynamic=动态插件" eg:"source"`
	Description            string                       `json:"description" dc:"插件描述" eg:"提供左侧菜单页面与公开/受保护路由示例的源码插件"`
	Installed              int                          `json:"installed" dc:"安装状态：1=已安装/已集成 0=未安装；源码插件默认返回1表示已随宿主集成" eg:"1"`
	InstalledAt            string                       `json:"installedAt" dc:"插件安装或源码接入时间，未安装时返回空字符串" eg:"2026-01-01 12:00:00"`
	Enabled                int                          `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"1"`
	StatusKey              string                       `json:"statusKey" dc:"插件状态在系统插件注册表中的定位键名，前端注册表监听会使用该键判断是否需要刷新插件状态" eg:"sys_plugin.status:plugin-demo-source"`
	UpdatedAt              string                       `json:"updatedAt" dc:"插件注册表最后更新时间" eg:"2026-01-01 12:00:00"`
	AuthorizationRequired  int                          `json:"authorizationRequired" dc:"是否存在需要安装/启用时确认的 hostServices 资源申请：1=是 0=否" eg:"1"`
	AuthorizationStatus    string                       `json:"authorizationStatus" dc:"当前授权状态：not_required=无需确认 pending=待确认 confirmed=已确认" eg:"confirmed"`
	RequestedHostServices  []*HostServicePermissionItem `json:"requestedHostServices,omitempty" dc:"插件当前版本声明的 hostServices 申请清单" eg:"[]"`
	AuthorizedHostServices []*HostServicePermissionItem `json:"authorizedHostServices,omitempty" dc:"宿主对当前 release 最终确认后的 hostServices 授权快照" eg:"[]"`
}
