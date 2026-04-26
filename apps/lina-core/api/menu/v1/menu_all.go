package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetAllReq defines the request for querying all user menu routes.
type GetAllReq struct {
	g.Meta `path:"/menus/all" method:"get" tags:"Menu Management" summary:"Get host menu route" dc:"Get the host menu routing list accessible to the currently logged in user and return the routing projection of the default management workbench for dynamic routing assembly."`
}

// MenuRouteItem represents a menu route item for Vben frontend
type MenuRouteItem struct {
	Id        int              `json:"id" dc:"Menu ID" eg:"1"`
	ParentId  int              `json:"parentId" dc:"Parent menu ID" eg:"0"`
	Name      string           `json:"name" dc:"Route name (unique)" eg:"System"`
	Path      string           `json:"path" dc:"routing path" eg:"/system"`
	Component string           `json:"component" dc:"component path" eg:"#/views/system/user/index.vue"`
	Redirect  string           `json:"redirect,omitempty" dc:"redirect path" eg:"/system/user"`
	Meta      *MenuRouteMeta   `json:"meta" dc:"Routing meta information" eg:""`
	Children  []*MenuRouteItem `json:"children,omitempty" dc:"Sub-route list" eg:"[]"`
}

// MenuRouteMeta represents route metadata for Vben
type MenuRouteMeta struct {
	Title            string            `json:"title" dc:"menu title" eg:"System management"`
	Icon             string            `json:"icon,omitempty" dc:"menu icon" eg:"ant-design:setting-outlined"`
	ActiveIcon       string            `json:"activeIcon,omitempty" dc:"Activation icon" eg:"ant-design:setting-filled"`
	HideInMenu       bool              `json:"hideInMenu,omitempty" dc:"Whether to hide in the menu" eg:"false"`
	HideInBreadcrumb bool              `json:"hideInBreadcrumb,omitempty" dc:"Whether to hide in breadcrumbs" eg:"false"`
	HideInTab        bool              `json:"hideInTab,omitempty" dc:"Whether to hide in tabs" eg:"false"`
	KeepAlive        bool              `json:"keepAlive,omitempty" dc:"Whether to cache the page" eg:"true"`
	IframeSrc        string            `json:"iframeSrc,omitempty" dc:"The target address in iframe mode, usually used to host inline hosted pages" eg:"/plugin-assets/plugin-runtime-demo/v0.1.0/index.html"`
	Link             string            `json:"link,omitempty" dc:"The target address in new tab mode, which will be opened directly by the host when the menu is clicked." eg:"/plugin-assets/plugin-runtime-demo/v0.1.0/index.html"`
	OpenInNewWindow  bool              `json:"openInNewWindow,omitempty" dc:"Whether to open the current menu in a new window or tab" eg:"true"`
	Query            map[string]string `json:"query,omitempty" dc:"Query parameters attached when the menu is opened, for host inline mounting or normal page recovery context" eg:"{\"pluginAccessMode\":\"embedded-mount\"}"`
	Order            int               `json:"order" dc:"Sorting number, 0 represents the highest priority, the smaller the value, the higher it is." eg:"1"`
	Authority        string            `json:"authority,omitempty" dc:"Permission ID" eg:"system:user:list"`
	IgnoreAccess     bool              `json:"ignoreAccess,omitempty" dc:"Whether to ignore permissions" eg:"false"`
}

// GetAllRes defines the wrapped response for user menu routes.
type GetAllRes struct {
	List []*MenuRouteItem `json:"list" dc:"User menu routing list" eg:"[]"`
}
