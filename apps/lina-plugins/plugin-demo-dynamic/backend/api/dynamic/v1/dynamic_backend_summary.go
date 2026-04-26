// This file defines the backend-summary route DTOs for the dynamic plugin
// sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// BackendSummaryReq is the request for querying the dynamic plugin backend execution summary.
type BackendSummaryReq struct {
	gmeta.Meta `path:"/backend-summary" method:"get" tags:"Dynamic Plugin Demo" summary:"Query the dynamic plugin backend execution summary" dc:"Return the current bridge execution summary for plugin-demo-dynamic when dispatched through the host prefix /api/v1/extensions/{pluginId}/..., including plugin ID, route information, and current user context." access:"login" permission:"plugin-demo-dynamic:backend:view" operLog:"other"`
}

// BackendSummaryRes is the response for querying the dynamic plugin backend execution summary.
type BackendSummaryRes struct {
	Message       string  `json:"message" dc:"Dynamic plugin backend execution instructions, describing the path and method of processing the current request through the Wasm bridge runtime" eg:"This backend example is executed through the plugin-demo-dynamic Wasm bridge runtime."`
	PluginID      string  `json:"pluginId" dc:"The unique identifier of the dynamic plugin currently executing the request" eg:"plugin-demo-dynamic"`
	PublicPath    string  `json:"publicPath" dc:"The currently hit host's public routing path" eg:"/api/v1/extensions/plugin-demo-dynamic/backend-summary"`
	Access        string  `json:"access" dc:"The current access level of dynamic routing: login=requires login public=anonymous accessible" eg:"login"`
	Permission    string  `json:"permission" dc:"The permission identifier of the current dynamic route; an empty string for anonymous routing" eg:"plugin-demo-dynamic:backend:view"`
	Authenticated bool    `json:"authenticated" dc:"Whether the current request has host authentication identity: true=authenticated false=anonymous" eg:"true"`
	Username      *string `json:"username,omitempty" dc:"Current login username; empty when requesting anonymously" eg:"admin"`
	IsSuperAdmin  *bool   `json:"isSuperAdmin,omitempty" dc:"Whether the current identity is a super administrator; empty when requesting anonymously" eg:"true"`
}
