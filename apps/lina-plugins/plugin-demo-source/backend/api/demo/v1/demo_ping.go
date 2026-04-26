package v1

import "github.com/gogf/gf/v2/frame/g"

// PingReq is the request for querying plugin-demo-source public ping.
type PingReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/ping" method:"get" tags:"Source Plugin Demo" summary:"Query source plugin example public ping" dc:"Return public ping information for plugin-demo-source, verifying that one plugin can register both public and authenticated routes within the same API module."`
}

// PingRes is the response for querying plugin-demo-source public ping.
type PingRes struct {
	Message string `json:"message" dc:"Plugin exposes fixed messages returned by ping" eg:"pong"`
}
