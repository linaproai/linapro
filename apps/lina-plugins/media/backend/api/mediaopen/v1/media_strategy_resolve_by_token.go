// This file declares the public media strategy resolution DTO authenticated by Tieta token.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResolveStrategyByTokenReq defines the public request for resolving one media strategy through Tieta token authentication.
type ResolveStrategyByTokenReq struct {
	g.Meta   `path:"/media/strategy-authorizations" method:"post" tags:"媒体鉴权" summary:"通过铁塔 token 解析媒体策略" dc:"使用铁塔 token 换取租户身份，并校验租户对设备的访问权限后解析生效媒体策略。"`
	Token    string `json:"token" dc:"铁塔用户Token；为空时读取 Authorization 请求头" eg:"Bearer token-value"`
	TenantId string `json:"tenantId" dc:"可选租户ID；传入时必须与铁塔 token 返回租户一致" eg:"tenant-a"`
	DeviceId string `json:"deviceId" v:"required#设备国标ID不能为空" dc:"设备国标ID" eg:"34020000001320000001"`
}

// ResolveStrategyByTokenRes defines the public response for one Tieta-authenticated strategy resolution.
type ResolveStrategyByTokenRes struct {
	UserId       int64  `json:"userId" dc:"铁塔用户ID" eg:"13"`
	Username     string `json:"username" dc:"铁塔用户名" eg:"wj530"`
	RealName     string `json:"realName" dc:"铁塔用户姓名" eg:"王杰"`
	Mobile       string `json:"mobile" dc:"手机号" eg:"18213268117"`
	TenantId     string `json:"tenantId" dc:"铁塔租户ID" eg:"tenant-a"`
	DeviceId     string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	HasAccess    bool   `json:"hasAccess" dc:"是否通过铁塔租户设备权限校验" eg:"true"`
	Matched      bool   `json:"matched" dc:"是否匹配到策略" eg:"true"`
	Source       string `json:"source" dc:"策略来源：tenantDevice、device、tenant、global或none" eg:"tenantDevice"`
	SourceLabel  string `json:"sourceLabel" dc:"策略来源中文说明" eg:"租户设备策略"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认直播策略"`
	Strategy     string `json:"strategy" dc:"YAML格式策略内容" eg:"record: true"`
}
