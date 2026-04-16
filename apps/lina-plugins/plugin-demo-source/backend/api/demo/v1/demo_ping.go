package v1

import "github.com/gogf/gf/v2/frame/g"

// PingReq is the request for querying plugin-demo-source public ping.
type PingReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/ping" method:"get" tags:"源码插件示例" summary:"查询源码插件示例公开 ping" dc:"返回 plugin-demo-source 的公开 ping 信息，用于验证同一插件可在一个 API 模块内通过分组路由同时注册免鉴权与需鉴权接口"`
}

// PingRes is the response for querying plugin-demo-source public ping.
type PingRes struct {
	Message string `json:"message" dc:"插件公开 ping 返回的固定消息" eg:"pong"`
}
