// This file declares platform tenant resolver policy validation DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResolverConfigUpdateReq defines the request for validating resolver policy.
type ResolverConfigUpdateReq struct {
	g.Meta             `path:"/platform/tenant/resolver-config" method:"put" tags:"Platform Tenants" summary:"Validate tenant resolver policy" dc:"The resolver policy is code-owned. This endpoint accepts only the built-in no-op policy and rejects runtime mutations." permission:"system:tenant:resolver:edit"`
	Chain              []string `json:"chain" v:"required#gf.gvalid.rule.required" dc:"Resolver chain order" eg:"override,jwt,session,header,subdomain,default"`
	ReservedSubdomains []string `json:"reservedSubdomains" dc:"Reserved subdomain labels" eg:"www,api,admin"`
	RootDomain         string   `json:"rootDomain" dc:"Reserved root domain field. The first version does not support setting it and accepts only an empty value." eg:""`
	OnAmbiguous        string   `json:"onAmbiguous" v:"required|in:prompt#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Built-in ambiguous tenant behavior" eg:"prompt"`
}

// ResolverConfigUpdateRes defines the resolver policy validation response.
type ResolverConfigUpdateRes struct{}
