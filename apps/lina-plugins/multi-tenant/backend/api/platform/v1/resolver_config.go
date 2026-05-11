// This file declares platform tenant resolver policy query DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResolverConfigEntity is the resolver policy API projection.
type ResolverConfigEntity struct {
	Chain              []string `json:"chain" dc:"Resolver chain order" eg:"override,jwt,session,header,subdomain,default"`
	ReservedSubdomains []string `json:"reservedSubdomains" dc:"Reserved subdomain labels" eg:"www,api,admin"`
	RootDomain         string   `json:"rootDomain" dc:"Root domain used by subdomain resolver. The first version keeps it empty and does not support updates." eg:""`
	OnAmbiguous        string   `json:"onAmbiguous" dc:"Built-in ambiguous tenant behavior. The first version always uses prompt." eg:"prompt"`
	Version            int64    `json:"version" dc:"Built-in policy version" eg:"1"`
}

// ResolverConfigGetReq defines the request for retrieving resolver policy.
type ResolverConfigGetReq struct {
	g.Meta `path:"/platform/tenant/resolver-config" method:"get" tags:"Platform Tenants" summary:"Get tenant resolver policy" dc:"Get the built-in tenant resolver chain and ambiguity behavior." permission:"system:tenant:resolver:query"`
}

// ResolverConfigGetRes defines the resolver policy response.
type ResolverConfigGetRes struct {
	*ResolverConfigEntity
}
