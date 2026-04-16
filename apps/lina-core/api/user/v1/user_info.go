package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetInfoReq defines the request for querying current frontend user info.
type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"用户管理" summary:"获取宿主用户上下文" dc:"获取当前登录用户的基础身份、角色、菜单树与权限集合，供默认管理工作台完成启动装配、导航渲染与权限控制"`
}

// GetInfoRes defines the response for querying current frontend user info.
type GetInfoRes struct {
	UserId      int         `json:"userId" dc:"用户ID" eg:"1"`
	Username    string      `json:"username" dc:"用户名" eg:"admin"`
	RealName    string      `json:"realName" dc:"真实姓名（昵称）" eg:"管理员"`
	Email       string      `json:"email" dc:"邮箱地址" eg:"admin@example.com"`
	Avatar      string      `json:"avatar" dc:"头像地址" eg:"/upload/avatar/default.png"`
	Roles       []string    `json:"roles" dc:"用户角色标识列表" eg:"['admin','user']"`
	HomePath    string      `json:"homePath" dc:"首页路径" eg:"/dashboard"`
	Menus       []*MenuTree `json:"menus" dc:"宿主菜单树，供默认管理工作台装配导航、路由与工作区入口" eg:"[]"`
	Permissions []string    `json:"permissions" dc:"用户有效权限标识列表，包含菜单权限与按钮权限，用于接口声明权限校验和按钮级权限控制" eg:"['system:user:list','system:user:add','system:user:edit']"`
}

// MenuTree represents a menu node in the user menu tree.
type MenuTree struct {
	Id        int         `json:"id" dc:"菜单ID" eg:"1"`
	ParentId  int         `json:"parentId" dc:"父菜单ID" eg:"0"`
	Name      string      `json:"name" dc:"菜单名称" eg:"系统管理"`
	Path      string      `json:"path" dc:"路由路径" eg:"/system"`
	Component string      `json:"component" dc:"组件路径" eg:"LAYOUT"`
	Perms     string      `json:"perms" dc:"权限标识" eg:"system:user:list"`
	Icon      string      `json:"icon" dc:"菜单图标" eg:"ant-design:setting-outlined"`
	Type      string      `json:"type" dc:"菜单类型：D=目录 M=菜单 B=按钮" eg:"D"`
	Sort      int         `json:"sort" dc:"排序" eg:"1"`
	Visible   int         `json:"visible" dc:"是否显示：1=显示 0=隐藏" eg:"1"`
	Status    int         `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	IsFrame   int         `json:"isFrame" dc:"是否外链：1=是 0=否" eg:"0"`
	IsCache   int         `json:"isCache" dc:"是否缓存：1=是 0=否" eg:"0"`
	Children  []*MenuTree `json:"children" dc:"子菜单" eg:"[]"`
}
