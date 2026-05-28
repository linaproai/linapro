# Auth Provider Integration Guide

This guide walks you through adding a new authentication provider (Google,
Discord, GitHub, custom OIDC, etc.) as a LinaPro source plugin. It captures
the contract every auth provider plugin must implement, the host services
it can rely on, the OAuth callback handoff, and the SQL conventions the
host enforces.

> 中文镜像见 [auth-provider-integration.zh-CN.md](./auth-provider-integration.zh-CN.md).

## 1. What an Auth Provider Plugin Owns

A source plugin that provides a third-party login does four things:

1. Publishes one **auth provider** to the host registry. The provider
   exposes static metadata (display name, icon, entry URL) plus a few
   runtime fields read live from plugin settings.
2. Exposes one **settings API** under `/x/<plugin-id>/api/v1/plugin/<plugin-id>/settings`
   that admin UIs read/write via the standard plugin route convention.
3. Implements the **OAuth flow** behind two well-known routes:
   - `GET /api/v1/auth/<provider-id>` — initiate authorization.
   - `GET /api/v1/auth/<provider-id>/callback` — handle the provider callback.
4. Delegates final login to the **host auth service** which converts the
   verified external identity into a host JWT (or pre-login handoff token
   when the user belongs to multiple tenants).

The plugin does **not** own:

- A private SQL configuration table. All settings flow through the host's
  `PluginSettingsService` and live in `sys_config` under your plugin's
  namespace.
- The frontend OAuth handoff page. The host workbench owns
  `/oauth-handoff` and consumes whatever query parameters the callback
  produces.
- JWT issuance, session creation, or login lifecycle hook dispatch. The
  host's `AuthService.LoginByExternal` runs the full host login policy.

## 2. Repository Layout

Place every new provider under `apps/lina-plugins/<plugin-id>/` with the
following layout (see `linapro-oidc-google` and `linapro-oidc-discord` for
reference):

```
apps/lina-plugins/linapro-oidc-<provider>/
├── plugin.yaml                       # plugin manifest (id, name, lifecycle hooks)
├── plugin_embed.go                   # static assets (//go:embed)
├── manifest/
│   └── sql/
│       └── uninstall/
│           └── 001-<plugin-id>-schema.sql    # sys_config cleanup (see §6)
├── backend/
│   ├── plugin.go                     # init + registerRoutes
│   ├── api/
│   │   └── settings/v1/settings.go   # admin settings DTO + g.Meta
│   └── internal/
│       ├── controller/
│       │   ├── settings/             # admin settings controller
│       │   └── oauth/                # /auth/<provider> + /auth/<provider>/callback
│       └── service/
│           ├── config/               # typed settings projection on top of PluginSettings
│           ├── oauth/                # OAuth client (state, code exchange, userinfo)
│           └── provider/             # authprovider.Provider implementation
└── frontend/
    ├── constants.ts                  # public route + URL constants used by Vue pages
    └── pages/
        └── <provider>-settings.vue   # admin settings form
```

The host enforces:

- The plugin **must not** declare its own configuration table. All
  configuration flows through `sys_config` via `PluginSettingsService`.
- Auth provider plugins **do not** ship an install SQL file. They only
  ship the uninstall SQL described in §6 so removing the plugin cleans
  up its namespace.

## 3. Plugin Registration

```go
// backend/plugin.go
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/authprovider"
	"lina-core/pkg/pluginhost"
	myplugin "lina-plugin-linapro-oidc-acme"
	oauthctrl "lina-plugin-linapro-oidc-acme/backend/internal/controller/oauth"
	settingsctrl "lina-plugin-linapro-oidc-acme/backend/internal/controller/settings"
	configsvc "lina-plugin-linapro-oidc-acme/backend/internal/service/config"
	provider "lina-plugin-linapro-oidc-acme/backend/internal/service/provider"
)

const pluginID = "linapro-oidc-acme"

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(myplugin.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	hostServices := registrar.HostServices()
	if hostServices == nil {
		return gerror.New("linapro-oidc-acme routes require host services")
	}
	authSvc := hostServices.Auth()
	settingsClient := hostServices.PluginSettings()
	if authSvc == nil || settingsClient == nil {
		return gerror.New("linapro-oidc-acme routes require host auth + plugin settings")
	}
	settingsSvc := configsvc.New(settingsClient)
	authprovider.RegisterProvider(provider.New(settingsSvc))

	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.Bind(
					settingsctrl.NewV1(settingsSvc).GetSettings,
					settingsctrl.NewV1(settingsSvc).SaveSettings,
				)
			})
		})
	})

	routes.Group("/api/v1/auth", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)
		group.GET("/acme", oauthctrl.New(authSvc, settingsSvc).StartLogin)
		group.GET("/acme/callback", oauthctrl.New(authSvc, settingsSvc).HandleCallback)
	})
	_ = ctx
	return nil
}
```

Key conventions:

- `RegisterProvider` runs inside `registerRoutes` (not `init`) because it
  needs the host's `PluginSettings()` adapter.
- Settings routes go under `routes.APIPrefix() + "/api/v1"` and inherit
  the standard auth/tenancy/permission middleware stack.
- OAuth routes go under root `/auth/...`. They must NOT inherit the auth
  middleware because they are reached by anonymous browsers.

## 4. Settings DTO and Validation

`backend/api/settings/v1/settings.go` defines the GET/PUT pair the admin
form talks to.

```go
type SaveSettingsReq struct {
	g.Meta       `path:"/plugin/linapro-oidc-acme/settings" method:"put" tags:"ACME OAuth"
	              summary:"Save ACME OAuth settings"
	              permission:"linapro-oidc-acme:settings:edit"`
	ClientID     string `json:"clientId" v:"required" dc:"OAuth client ID" eg:"acme-123"`
	// ClientSecret is intentionally optional: the GET response masks the
	// stored secret, and the admin form posts an empty value to mean
	// "keep current".
	ClientSecret string `json:"clientSecret"
	                    dc:"OAuth client secret; leave empty to keep stored secret unchanged"
	                    eg:"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`
	RedirectURI  string `json:"redirectUri" v:"required" dc:"Registered OAuth redirect URI" eg:"https://example.com/api/v1/auth/acme/callback"`
	EnableBackendRedirect  bool   `json:"enableBackendRedirect" dc:"Enable state-keyed post-login redirect" eg:"true"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"Fallback redirect URL when state does not match any rule" eg:"/admin"`
	BackendRedirects       string `json:"backendRedirects" dc:"JSON state→URL map" eg:"{\"acme\":\"/admin\"}"`
	Enabled                bool   `json:"enabled" dc:"Whether ACME login is enabled" eg:"true"`
}
```

Always test the "leave secret empty" path with a `gvalid` round-trip
asserting `ClientSecret` is **not** marked `required`.

## 5. Typed Settings Service

`backend/internal/service/config/config.go` projects the namespaced
key-value store into a typed `Settings` struct for the rest of the plugin
to consume.

```go
const PluginID = "linapro-oidc-acme"

type Settings struct {
	ClientID, ClientSecret, RedirectURI string
	EnableBackendRedirect bool
	DefaultBackendRedirect string
	BackendRedirects string
	Enabled bool
}

const (
	keyClientID     = "clientId"
	keyClientSecret = "clientSecret"
	// ...
)

type Service struct {
	settings contract.PluginSettingsService
}

func New(settings contract.PluginSettingsService) *Service { return &Service{settings: settings} }

func (s *Service) Get(ctx context.Context) (*Settings, error) {
	values, err := s.settings.List(ctx, PluginID)
	if err != nil { return nil, err }
	return &Settings{
		ClientID: values[keyClientID],
		// ...
	}, nil
}

func (s *Service) Save(ctx context.Context, in *Settings) error {
	if err := s.settings.SetString(ctx, PluginID, keyClientID, in.ClientID); err != nil { return err }
	// SetSecret skips the row when value is "" so masked round-trips stay safe.
	if err := s.settings.SetSecret(ctx, PluginID, keyClientSecret, in.ClientSecret); err != nil { return err }
	// ...
	return nil
}

func (s *Service) GetMaskedClientSecret(ctx context.Context) (string, error) {
	return s.settings.GetMaskedSecret(ctx, PluginID, keyClientSecret)
}
```

Rules:

- Use `PluginSettingsService.SetSecret` (not `SetString`) for any field
  the admin UI displays masked. Empty values preserve the stored value.
- The admin GET controller must return the **masked** secret via
  `GetMaskedSecret`, never the raw value.
- Never read or write keys outside your `PluginID` namespace. The
  contract enforces this and returns an error on cross-namespace access.

## 6. Install / Uninstall SQL

Auth provider plugins **must not** ship an `manifest/sql/<seq>-*.sql`
install file. Plugin configuration lives entirely in the host's shared
`sys_config` table and is written at runtime via `PluginSettingsService`,
so there is no schema for the install step to create. Adding an install
SQL file solely to "register" the plugin is unnecessary and adds an
upgrade footprint operators have to reason about.

Uninstall SQL is still required so removing the plugin cleans up its
namespace in `sys_config`. Ship exactly one file at
`manifest/sql/uninstall/001-<plugin-id>-schema.sql`:

```sql
DELETE FROM sys_config
WHERE tenant_id = 0
  AND key LIKE 'linapro-oidc-acme.%';
```

The uninstall script must satisfy the SQL idempotency rule: re-running it
on a clean database must not error or change state. The `DELETE` above
already meets that rule because rows that do not exist are simply not
deleted.

## 7. OAuth Flow Implementation

`backend/internal/service/oauth/oauth.go` is provider-specific (Google
uses `accounts.google.com`, Discord uses `discord.com`, etc.). It must
provide three operations:

1. `BuildAuthorizeURL(settings, redirectURI, stateKey)` — return the URL
   the browser should visit. Embed a signed state token that you can
   verify in the callback.
2. `ExchangeCode(ctx, settings, redirectURI, code)` — POST to the
   provider's token endpoint and return the access token.
3. `FetchUserIdentity(ctx, accessToken)` — GET the provider's userinfo
   endpoint and return at least `Subject`, `Email`, `EmailVerified`, and
   `Name`. Email is the authoritative join key against the local user
   table; reject identities without a verified email.

Use HMAC-SHA256 keyed by the client secret to sign the state. Include a
nonce and an expiration timestamp so the state cannot be replayed.

`backend/internal/controller/oauth/oauth.go` wires it all together:

```go
func (c *Controller) StartLogin(req *ghttp.Request) {
	ctx := req.Context()
	settings, err := c.settingsSvc.Get(ctx)
	if err != nil { c.writeError(req, "settings unavailable", err); return }
	if settings == nil || !settings.Enabled {
		c.writeError(req, "provider disabled", gerror.New("disabled"))
		return
	}
	stateKey := strings.TrimSpace(req.Get("state").String())
	redirectURI := c.resolveRedirectURI(req, settings.RedirectURI)
	authorizeURL, _, err := c.oauthSvc.BuildAuthorizeURL(toOAuthSettings(settings), redirectURI, stateKey)
	if err != nil { c.writeError(req, "build authorize url", err); return }
	req.Response.RedirectTo(authorizeURL, 302)
}

func (c *Controller) HandleCallback(req *ghttp.Request) {
	ctx := req.Context()
	code := strings.TrimSpace(req.Get("code").String())
	rawState := strings.TrimSpace(req.Get("state").String())
	settings, _ := c.settingsSvc.Get(ctx)
	payload, err := c.oauthSvc.DecodeState(rawState, settings.ClientSecret)
	if err != nil { c.redirectWithError(req, "/", "invalid_state"); return }
	accessToken, err := c.oauthSvc.ExchangeCode(ctx, toOAuthSettings(settings), c.resolveRedirectURI(req, settings.RedirectURI), code)
	if err != nil { c.redirectWithError(req, "/", "code_exchange_failed"); return }
	identity, err := c.oauthSvc.FetchUserIdentity(ctx, accessToken)
	if err != nil { c.redirectWithError(req, "/", "userinfo_failed"); return }
	if !identity.EmailVerified { c.redirectWithError(req, "/", "email_not_verified"); return }
	out, err := c.authSvc.LoginByExternal(ctx, plugincontract.ExternalLoginInput{
		ProviderID:     "acme",
		PluginID:       configsvc.PluginID,
		ExternalUserID: identity.Subject,
		Email:          identity.Email,
		DisplayName:    identity.Name,
		ClientIP:       req.GetClientIp(),
	})
	if err != nil {
		c.redirectWithError(req, "/", classifyHostLoginError(err))
		return
	}
	target := c.resolveFrontendTarget(req, payload.StateKey, settings.BackendRedirects, settings.DefaultBackendRedirect, settings.EnableBackendRedirect)
	c.redirectToHandoff(req, target, payload.StateKey, out)
}
```

Notes:

- The OAuth callback **must** call `c.authSvc.OAuthHandoffURL(ctx, values)`
  for the final redirect. That helper already accounts for the
  workspace base path and the vue-router history backend (hash vs
  history) so your plugin never hard-codes `/oauth-handoff`.
- When `LoginByExternal` returns a `bizerr.Error`, propagate its
  `RuntimeCode()` to the frontend so the handoff page can show a
  precise error (e.g. `AUTH_EXTERNAL_USER_NOT_PROVISIONED` when the
  email is not linked to a local account).

## 8. Provider Registration

```go
// backend/internal/service/provider/provider.go
type Provider struct { settingsSvc *configsvc.Service }

func New(settingsSvc *configsvc.Service) *Provider {
	return &Provider{settingsSvc: settingsSvc}
}

func (p *Provider) ProviderID() string { return "acme" }
func (p *Provider) PluginID() string { return configsvc.PluginID }
func (p *Provider) Kind() authprovider.Kind { return authprovider.KindOAuth2 }

func (p *Provider) LoginEntry(ctx context.Context) (*authprovider.LoginEntry, error) {
	settings, err := p.settingsSvc.Get(ctx)
	if err != nil { return nil, err }
	return &authprovider.LoginEntry{
		ProviderID:             "acme",
		PluginID:               configsvc.PluginID,
		Kind:                   authprovider.KindOAuth2,
		Name:                   "ACME",
		Description:            "Sign in with an ACME account",
		Icon:                   "logos:acme",
		EntryURL:               "/api/v1/auth/acme",
		BackendRedirectEnabled: settings.EnableBackendRedirect,
		BackendRedirectDefault: settings.DefaultBackendRedirect,
		BackendRedirectRules:   settings.BackendRedirects,
		DisplayOrder:           30,
	}, nil
}
```

## 9. Frontend Settings Page

Frontend pages live under `frontend/pages/`. Use the same conventions as
`google-settings.vue`:

- Use `getPluginIdApiPath(...)` helpers, **not** module-level
  `pluginApiPath(...)`, to avoid ESM TDZ errors.
- On load, the form clears `clientSecret` so the masked value never
  round-trips back through PUT.
- The "Redirect URI" input defaults to `window.location.origin +
  /auth/<provider>/callback`. Operators can override it.
- "Backend redirect rules" is a repeating editor; serialize the rows to
  a JSON object before submitting and parse JSON on load.

## 10. Tests You Must Write

The host's bugfix testing rule applies to every auth provider plugin:

1. **DTO validation** (`backend/api/settings/v1/settings_test.go`):
   - `TestSaveSettingsAllowsEmptyClientSecret`
   - `TestSaveSettingsRequiresClientID`
   - `TestSaveSettingsRequiresRedirectURI`
2. **Typed settings service** (`backend/internal/service/config/config_test.go`)
   using an in-memory `PluginSettingsService` double:
   - "Save with empty ClientSecret preserves stored secret"
   - "Save with new ClientSecret rotates stored value"
   - "Roundtrip Save→Get returns identical fields"
   - "Get returns defaults when no rows exist"
   - "GetMaskedClientSecret never returns the raw secret"
3. **OAuth state signing** (`backend/internal/service/oauth/oauth_test.go`):
   - HMAC roundtrip
   - Tampered MAC rejection
   - Different client secret rejection
   - Expired state rejection
   - Disabled provider rejection
4. **Callback error classification**
   (`backend/internal/controller/oauth/oauth_test.go`):
   - bizerr runtime code extraction
   - Plain error fallback
   - Nil error short-circuit

Run them with `rtk go test ./... -count=1` from your plugin module root.
Also run `cd apps/lina-core && rtk go test ./internal/cmd -count=1` after
adding any `panic` in your plugin's `init()` so the allowlist test
documents the new boundary.

## 11. Frontend Handoff

The host workbench serves `/oauth-handoff` (relative to the workspace
base path) and consumes these query parameters:

| Key | Meaning |
|---|---|
| `provider` | Provider identifier for the UI banner. |
| `state` | The state key the user originated from (used by the redirect rules). |
| `redirect` | The final URL the SPA should navigate to after login. |
| `accessToken` / `refreshToken` | Set when `LoginByExternal` returned a single-tenant token pair. |
| `preToken` / `tenants` | Set when the user has multiple tenants; the SPA opens the tenant picker. |
| `error` | Stable error code (typically the host bizerr `RuntimeCode`). |

Your plugin only assembles the map and calls
`c.authSvc.OAuthHandoffURL(ctx, values)` — the host puts everything in
the right place for hash vs history routing.

## 12. Checklist Before Shipping

- [ ] `plugin.yaml` declares the plugin id, name, and lifecycle hooks.
- [ ] `backend/api/settings/v1/settings.go` defines GET/PUT DTOs with
      proper `dc`, `eg`, and `permission` tags. `ClientSecret` is **not**
      marked `required`.
- [ ] `backend/internal/service/config/config.go` uses
      `PluginSettingsService` (no DAO, no DB driver imports, no
      `gdb.Raw`).
- [ ] OAuth provider supports state signing with HMAC + nonce +
      expiration.
- [ ] Callback handler uses `c.authSvc.LoginByExternal` and
      `c.authSvc.OAuthHandoffURL`.
- [ ] No install SQL file under `manifest/sql/` (fresh installs use
      `sys_config` only; the plugin writes via `PluginSettingsService`).
- [ ] Uninstall SQL deletes every `linapro-oidc-<provider>.*` row in
      `sys_config`.
- [ ] Tests for DTO validation, secret semantics, state signing, and
      error classification all pass.
- [ ] `rtk go test ./... -count=1` passes in both the plugin module and
      `apps/lina-core`.
