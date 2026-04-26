package v1

import "github.com/gogf/gf/v2/frame/g"

// Online User Force Logout API

// OnlineForceLogoutReq defines the request for forcing an online user offline.
type OnlineForceLogoutReq struct {
	g.Meta  `path:"/monitor/online/{tokenId}" method:"delete" tags:"System Monitoring" summary:"Force logout" dc:"Force a specified online session to log out. Subsequent requests for that session return 401." permission:"monitor:online:forceLogout"`
	TokenId string `json:"tokenId" v:"required#gf.gvalid.rule.required" dc:"Session Token ID to be forced offline" eg:"abc123"`
}

// OnlineForceLogoutRes is the force logout response.
type OnlineForceLogoutRes struct{}
