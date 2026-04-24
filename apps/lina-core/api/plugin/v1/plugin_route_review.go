// This file defines dynamic-route review DTOs projected into plugin
// management APIs for install and enable governance dialogs.

package v1

// PluginRouteReviewItem describes one dynamic route exposed by the current
// plugin release during install or enable review.
type PluginRouteReviewItem struct {
	// Method is the normalized HTTP method declared by the dynamic route.
	Method string `json:"method" dc:"动态路由 HTTP 方法" eg:"GET"`
	// PublicPath is the host-visible public URL served for this dynamic route.
	PublicPath string `json:"publicPath" dc:"宿主真实公开路径，固定以 /api/v1/extensions/{pluginId}/ 开头" eg:"/api/v1/extensions/plugin-demo-dynamic/review-summary"`
	// Access identifies whether the route is public or requires login context.
	Access string `json:"access" dc:"访问级别：public=公开访问 login=登录访问" eg:"login"`
	// Permission is the host permission key enforced for authenticated routes.
	Permission string `json:"permission,omitempty" dc:"宿主权限标识；public 路由返回空字符串" eg:"plugin-demo-dynamic:review:query"`
	// Summary is the short review-friendly route summary.
	Summary string `json:"summary,omitempty" dc:"动态路由摘要，来自路由合同中的 summary" eg:"查询插件评审摘要"`
	// Description is the detailed business description declared by the route.
	Description string `json:"description,omitempty" dc:"动态路由说明，来自路由合同中的 description" eg:"返回动态插件当前版本生成的评审摘要信息"`
}
