# 认证 Provider 接入指南

本文档描述如何作为 LinaPro 源码插件接入一个新的认证 Provider（Google、
Discord、GitHub、自建 OIDC 等）。内容包含每个 auth provider 插件必须实现
的契约、可依赖的宿主服务、OAuth 回调 handoff 流程以及宿主强制要求的 SQL
规范。

> 英文版本见 [auth-provider-integration.md](./auth-provider-integration.md)。

## 1. 一个 Auth Provider 插件负责什么

一个第三方登录源码插件做四件事：

1. 向宿主注册一个 **auth provider**。Provider 暴露静态元数据（显示名、
   图标、入口 URL）和少量由插件设置实时驱动的运行时字段。
2. 在 `/x/<plugin-id>/api/v1/plugin/<plugin-id>/settings` 下暴露
   **设置 API**，遵循标准插件路由约定，admin UI 读写它。
3. 在两个固定路径下实现 **OAuth 流程**：
   - `GET /api/v1/auth/<provider-id>` —— 发起授权。
   - `GET /api/v1/auth/<provider-id>/callback` —— 处理 provider 回调。
4. 把最终登录委托给 **宿主 auth 服务**，由它把已验证的外部身份转换为
   宿主 JWT（或在用户属于多租户时转换为预登录 handoff token）。

插件**不**负责：

- 私有 SQL 配置表。所有设置走宿主 `PluginSettingsService`，存在
  `sys_config` 表中你插件的命名空间下。
- 前端 OAuth handoff 页面。宿主工作台拥有 `/oauth-handoff`，并消费
  回调产生的查询参数。
- JWT 签发、会话创建或登录生命周期钩子分发。这些都由宿主
  `AuthService.LoginByExternal` 按宿主登录策略统一完成。

## 2. 仓库目录结构

每个新 Provider 放在 `apps/lina-plugins/<plugin-id>/` 下，按下述结构组织
（可参考 `linapro-oidc-google` 与 `linapro-oidc-discord`）：

```
apps/lina-plugins/linapro-oidc-<provider>/
├── plugin.yaml                       # 插件清单（id、名称、生命周期钩子）
├── plugin_embed.go                   # 静态资源（//go:embed）
├── manifest/
│   └── sql/
│       └── uninstall/
│           └── 001-<plugin-id>-schema.sql    # sys_config 清理（见 §6）
├── backend/
│   ├── plugin.go                     # init + registerRoutes
│   ├── api/
│   │   └── settings/v1/settings.go   # 设置 DTO + g.Meta
│   └── internal/
│       ├── controller/
│       │   ├── settings/             # 设置 controller
│       │   └── oauth/                # /auth/<provider> 与 /auth/<provider>/callback
│       └── service/
│           ├── config/               # 基于 PluginSettings 的类型化设置
│           ├── oauth/                # OAuth 客户端（state、code 兑换、userinfo）
│           └── provider/             # authprovider.Provider 实现
└── frontend/
    ├── constants.ts                  # 前端用到的路由与 URL 常量
    └── pages/
        └── <provider>-settings.vue   # 设置表单
```

宿主强制要求：

- 插件**不得**声明自己的配置表。所有配置通过 `sys_config` 经
  `PluginSettingsService` 读写。
- Auth provider 插件**不提供**安装 SQL 文件。只提供 §6 中描述的卸载
  SQL，让删除插件时清掉它在 `sys_config` 中的命名空间。

## 3. 插件注册

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

关键约定：

- `RegisterProvider` 放在 `registerRoutes` 内（而不是 `init`），因为
  它依赖宿主的 `PluginSettings()` 适配器。
- 设置路由放在 `routes.APIPrefix() + "/api/v1"` 下，继承标准的鉴权 /
  租户 / 权限中间件链。
- OAuth 路由放在根路径 `/auth/...` 下，**不能**继承鉴权中间件，因为
  这些路由由匿名浏览器访问。

## 4. 设置 DTO 与校验

`backend/api/settings/v1/settings.go` 定义 admin 表单使用的 GET/PUT：

```go
type SaveSettingsReq struct {
	g.Meta       `path:"/plugin/linapro-oidc-acme/settings" method:"put" tags:"ACME OAuth"
	              summary:"Save ACME OAuth settings"
	              permission:"linapro-oidc-acme:settings:edit"`
	ClientID     string `json:"clientId" v:"required" dc:"OAuth client ID" eg:"acme-123"`
	// ClientSecret 故意不是 required：GET 响应会把已存的 secret 打掩码，
	// admin 表单留空提交表示 "保留原值"。
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

必须用 `gvalid` 编写一个 "secret 留空" 的 round-trip 测试，断言
`ClientSecret` **没有** `required` 标签。

## 5. 类型化设置服务

`backend/internal/service/config/config.go` 把命名空间 key-value 投影成
类型化的 `Settings` 给插件其余部分用：

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
	// SetSecret 在 value 为 "" 时跳过写入，确保掩码值不会污染存储的 secret。
	if err := s.settings.SetSecret(ctx, PluginID, keyClientSecret, in.ClientSecret); err != nil { return err }
	// ...
	return nil
}

func (s *Service) GetMaskedClientSecret(ctx context.Context) (string, error) {
	return s.settings.GetMaskedSecret(ctx, PluginID, keyClientSecret)
}
```

规则：

- admin UI 上以掩码方式显示的字段必须使用
  `PluginSettingsService.SetSecret`（不是 `SetString`），空值表示
  "保留原值"。
- admin GET controller 必须通过 `GetMaskedSecret` 返回**掩码后的** secret，
  绝不能返回原始值。
- 任何时候都不要读写自己 `PluginID` 命名空间之外的 key。契约层会拦截
  跨命名空间访问并返回错误。

## 6. 安装 / 卸载 SQL

Auth provider 插件**不**提供 `manifest/sql/<seq>-*.sql` 安装文件。插件
配置完全存放在宿主共享的 `sys_config` 表中，运行时通过
`PluginSettingsService` 写入，安装阶段没有 schema 需要建表。仅为了
"挂个插件" 而提供一份空 install SQL 既无必要，也会给运维多留一份
要追踪的升级足迹。

卸载 SQL 必须保留，用来在删除插件时清掉它在 `sys_config` 中的
命名空间。固定提供一份位于 `manifest/sql/uninstall/001-<plugin-id>-schema.sql`
的文件：

```sql
DELETE FROM sys_config
WHERE tenant_id = 0
  AND key LIKE 'linapro-oidc-acme.%';
```

卸载脚本必须满足 SQL 幂等规范：在干净库上重复执行不报错、不产生
状态变化。上述 `DELETE` 已经满足该规则——不存在的行自然不会被
"重复删除"。

## 7. OAuth 流程实现

`backend/internal/service/oauth/oauth.go` 是 provider 专属的（Google 用
`accounts.google.com`、Discord 用 `discord.com`）。它必须提供三个操作：

1. `BuildAuthorizeURL(settings, redirectURI, stateKey)` —— 返回浏览器
   应跳转的 URL，并内嵌一个可在回调中验证的签名 state token。
2. `ExchangeCode(ctx, settings, redirectURI, code)` —— 向 provider 的
   token endpoint POST，返回 access token。
3. `FetchUserIdentity(ctx, accessToken)` —— GET provider 的 userinfo
   endpoint，至少返回 `Subject`、`Email`、`EmailVerified`、`Name`。
   Email 是与本地用户表关联的权威键；拒绝没有验证过 email 的身份。

用 HMAC-SHA256（key 用 client secret）对 state 做签名。state 内必须包含
nonce 和过期时间戳，防止重放。

`backend/internal/controller/oauth/oauth.go` 把它们组合起来：

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

注意：

- OAuth 回调**必须**调用 `c.authSvc.OAuthHandoffURL(ctx, values)` 进行
  最终重定向。这个 helper 已经处理了 workspace base path 和 vue-router
  历史模式（hash vs history），插件永远不要硬编码 `/oauth-handoff`。
- 当 `LoginByExternal` 返回 `bizerr.Error` 时，把它的 `RuntimeCode()`
  透传给前端，让 handoff 页可以精确展示原因（例如
  `AUTH_EXTERNAL_USER_NOT_PROVISIONED` 表示该邮箱在本地无账号）。

## 8. Provider 注册

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

## 9. 前端设置页

前端页面放在 `frontend/pages/` 下，参考 `google-settings.vue` 的约定：

- 使用 `getPluginIdApiPath(...)` helper，**不要**在模块顶层调用
  `pluginApiPath(...)`，否则会触发 ESM TDZ 错误。
- 页面加载时清空 `clientSecret`，确保掩码值不会再被 PUT 提交回去。
- "Redirect URI" 输入框默认填充 `window.location.origin +
  /auth/<provider>/callback`，运营可手动覆盖。
- "Backend redirect rules" 是一个可重复编辑器；提交前把每一行组装成
  JSON 对象，加载时再把 JSON 解析回多行。

## 10. 必须编写的测试

宿主的 bugfix 测试规则同样适用于每个 auth provider 插件：

1. **DTO 校验**（`backend/api/settings/v1/settings_test.go`）：
   - `TestSaveSettingsAllowsEmptyClientSecret`
   - `TestSaveSettingsRequiresClientID`
   - `TestSaveSettingsRequiresRedirectURI`
2. **类型化设置服务**（`backend/internal/service/config/config_test.go`），
   用内存 `PluginSettingsService` 替身：
   - "空 ClientSecret 保留已存 secret"
   - "新 ClientSecret 覆盖已存 secret"
   - "Save→Get 字段完全 round-trip"
   - "Get 在没有行时返回默认值"
   - "GetMaskedClientSecret 不返回原始值"
3. **OAuth state 签名**（`backend/internal/service/oauth/oauth_test.go`）：
   - HMAC round-trip
   - 篡改 MAC 拒绝
   - 不同 client secret 拒绝
   - 过期 state 拒绝
   - Provider disabled 拒绝
4. **回调错误分类**
   （`backend/internal/controller/oauth/oauth_test.go`）：
   - 提取 bizerr runtime code
   - 普通 error 兜底
   - nil error 短路

在插件 module 根目录运行 `rtk go test ./... -count=1`。如果你在插件
`init()` 里新增了 `panic`，还要跑
`cd apps/lina-core && rtk go test ./internal/cmd -count=1` 让 allowlist
测试记录新增边界。

## 11. 前端 handoff

宿主工作台暴露 `/oauth-handoff`（相对于 workspace base path），消费下列
查询参数：

| Key | 含义 |
|---|---|
| `provider` | provider 标识，UI 横幅用 |
| `state` | 用户来源的 state key（驱动后台跳转规则） |
| `redirect` | SPA 登录完成后应跳转的最终 URL |
| `accessToken` / `refreshToken` | 单租户用户的 token 对 |
| `preToken` / `tenants` | 多租户用户的 pre-login 信息，SPA 打开租户选择器 |
| `error` | 稳定的错误码（通常是宿主 bizerr `RuntimeCode`） |

插件只负责组装 map 并调用
`c.authSvc.OAuthHandoffURL(ctx, values)` —— 宿主负责把它放在正确的
hash/history 路由结构里。

## 12. 上线前检查清单

- [ ] `plugin.yaml` 声明插件 id、名称和生命周期钩子。
- [ ] `backend/api/settings/v1/settings.go` 定义 GET/PUT DTO，
      `dc`、`eg`、`permission` 标签齐全。`ClientSecret` **未**被标记
      `required`。
- [ ] `backend/internal/service/config/config.go` 使用
      `PluginSettingsService`（不导入 DAO、不直接连数据库驱动、不写
      `gdb.Raw`）。
- [ ] OAuth provider 支持 HMAC + nonce + 过期的 state 签名。
- [ ] 回调 handler 使用 `c.authSvc.LoginByExternal` 与
      `c.authSvc.OAuthHandoffURL`。
- [ ] `manifest/sql/` 下没有安装 SQL 文件（新部署只走 `sys_config`，
      由 `PluginSettingsService` 在运行时写入）。
- [ ] 卸载 SQL 删除 `sys_config` 里所有 `linapro-oidc-<provider>.*`
      命名空间的行。
- [ ] DTO 校验、secret 语义、state 签名、错误分类的测试全部通过。
- [ ] `rtk go test ./... -count=1` 在插件 module 和 `apps/lina-core`
      两侧都通过。
