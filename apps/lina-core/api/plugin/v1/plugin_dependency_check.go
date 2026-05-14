// This file defines plugin dependency-check API DTOs.
package v1

import "github.com/gogf/gf/v2/frame/g"

// DependencyCheckReq is the request for checking plugin dependencies.
type DependencyCheckReq struct {
	g.Meta `path:"/plugins/{id}/dependencies" method:"get" tags:"Plugin Management" summary:"Check plugin dependencies" permission:"plugin:query" dc:"Return server-side plugin dependency check results, including framework compatibility, automatic install plan, manual requirements, soft dependency notices, blockers, and uninstall reverse dependents."`
	Id     string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
}

// DependencyCheckRes is the response for checking plugin dependencies.
type DependencyCheckRes = PluginDependencyCheckResult
