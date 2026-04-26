package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetInfoReq defines the request for querying current frontend user info.
type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"User Management" summary:"Get host user context" dc:"Obtain the basic identity, role, menu tree and permission set of the currently logged in user for the default management workbench to complete startup assembly, navigation rendering and permission control."`
}

// GetInfoRes defines the response for querying current frontend user info.
type GetInfoRes struct {
	UserId      int         `json:"userId" dc:"User ID" eg:"1"`
	Username    string      `json:"username" dc:"Username" eg:"admin"`
	RealName    string      `json:"realName" dc:"Real name (nickname)" eg:"Administrator"`
	Email       string      `json:"email" dc:"Email address" eg:"admin@example.com"`
	Avatar      string      `json:"avatar" dc:"Avatar address" eg:"/upload/avatar/default.png"`
	Roles       []string    `json:"roles" dc:"List of user role identifiers" eg:"['admin','user']"`
	HomePath    string      `json:"homePath" dc:"Home path" eg:"/dashboard"`
	Menus       []*MenuTree `json:"menus" dc:"Host menu tree for default management workbench assembly navigation, routing and workspace entry" eg:"[]"`
	Permissions []string    `json:"permissions" dc:"List of user effective permission identifiers, including menu permissions and button permissions, used for interface declaration permission verification and button-level permission control" eg:"['system:user:list','system:user:add','system:user:edit']"`
}

// MenuTree represents a menu node in the user menu tree.
type MenuTree struct {
	Id        int         `json:"id" dc:"Menu ID" eg:"1"`
	ParentId  int         `json:"parentId" dc:"Parent menu ID" eg:"0"`
	Name      string      `json:"name" dc:"Menu name" eg:"System management"`
	Path      string      `json:"path" dc:"routing path" eg:"/system"`
	Component string      `json:"component" dc:"component path" eg:"LAYOUT"`
	Perms     string      `json:"perms" dc:"Permission ID" eg:"system:user:list"`
	Icon      string      `json:"icon" dc:"menu icon" eg:"ant-design:setting-outlined"`
	Type      string      `json:"type" dc:"Menu type: D=Directory M=Menu B=Button" eg:"D"`
	Sort      int         `json:"sort" dc:"sort" eg:"1"`
	Visible   int         `json:"visible" dc:"Whether to display: 1=show 0=hide" eg:"1"`
	Status    int         `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	IsFrame   int         `json:"isFrame" dc:"Whether to external link: 1=yes 0=no" eg:"0"`
	IsCache   int         `json:"isCache" dc:"Whether to cache: 1=yes 0=no" eg:"0"`
	Children  []*MenuTree `json:"children" dc:"submenu" eg:"[]"`
}
