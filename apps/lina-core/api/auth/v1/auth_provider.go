// This file defines the auth provider list request and response DTOs used by
// the host /auth/providers endpoint. The default workbench login page calls
// this endpoint to render third-party login entries discovered via source
// plugins (Google, Discord, GitHub, custom OIDC, ...). The endpoint is
// publicly accessible because it lists entry metadata only; sensitive provider
// configuration stays on the host.

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
	ProviderID             string `json:"providerId" dc:"Stable provider identifier" eg:"google"`
	PluginID               string `json:"pluginId" dc:"Owning source-plugin identifier" eg:"linapro-oidc-google"`
	Kind                   string `json:"kind" dc:"Provider kind: oauth2, oidc, ldap, saml, cas" eg:"oidc"`
	Name                   string `json:"name" dc:"Display name shown on the login page" eg:"Google"`
	Description            string `json:"description" dc:"Short description of the provider capability" eg:"Sign in with a Google account"`
	Icon                   string `json:"icon" dc:"Icon identifier rendered by the workbench" eg:"logos:google-icon"`
	EntryURL               string `json:"entryUrl" dc:"Provider login entry URL or deep-link route" eg:"/api/v1/auth/google"`
	BackendRedirectEnabled bool   `json:"backendRedirectEnabled" dc:"Whether the provider should append a state token for backend redirect routing" eg:"true"`
	BackendRedirectDefault string `json:"backendRedirectDefault" dc:"Fallback redirect URL used when backend redirect state does not match any rule" eg:"/admin"`
	BackendRedirectRules   string `json:"backendRedirectRules" dc:"JSON object mapping state to redirect URL used by backend redirect routing" eg:"{\"google\":\"/admin\"}"`
	DisplayOrder           int    `json:"displayOrder" dc:"Login entry sort order; smaller values first" eg:"10"`
}

// ListProvidersRes is the /auth/providers response payload.
type ListProvidersRes struct {
	Providers []*ProviderEntity `json:"providers" dc:"Enabled authentication providers visible to the login page."`
}
