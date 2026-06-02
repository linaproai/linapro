// This file defines the auth provider list request and response DTOs used by
// the host /auth/providers endpoint. The default workbench login page calls
// this endpoint to render third-party login entries discovered via source
// plugins (Google, Discord, GitHub, custom OIDC, ...). The endpoint is
// publicly accessible and therefore returns button metadata only; sensitive
// redirect rules and token-delivery settings stay behind authenticated plugin
// settings APIs and provider callback logic.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListProvidersReq defines the request for listing enabled authentication
// providers. It carries no query parameters because the list is filtered by
// host plugin enablement state alone.
type ListProvidersReq struct {
	g.Meta `path:"/auth/providers" method:"get" tags:"Authentication" summary:"List enabled auth providers" dc:"Lists enabled third-party authentication providers exposed by source plugins so the default workbench login page can render their entries."`
}

// ProviderEntity is one auth provider entry returned by /auth/providers.
type ProviderEntity struct {
	ProviderID   string `json:"providerId" dc:"Stable provider identifier" eg:"google"`
	PluginID     string `json:"pluginId" dc:"Owning source-plugin identifier" eg:"linapro-oidc-google"`
	Kind         string `json:"kind" dc:"Provider kind: oauth2, oidc, ldap, saml, cas" eg:"oidc"`
	Name         string `json:"name" dc:"Display name shown on the login page" eg:"Google"`
	Description  string `json:"description" dc:"Short description of the provider capability" eg:"Sign in with a Google account"`
	Icon         string `json:"icon" dc:"Icon identifier rendered by the workbench" eg:"logos:google-icon"`
	EntryURL     string `json:"entryUrl" dc:"Provider login entry URL or deep-link route" eg:"/api/v1/auth/google"`
	DisplayOrder int    `json:"displayOrder" dc:"Login entry sort order; smaller values first" eg:"10"`
}

// ListProvidersRes is the /auth/providers response payload.
type ListProvidersRes struct {
	Providers []*ProviderEntity `json:"providers" dc:"Enabled authentication providers visible to the login page."`
}
