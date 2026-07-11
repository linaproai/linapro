## 1. 插件骨架与清单

- [x] 1.1 创建 `apps/lina-plugins/linapro-oidc-generic/` 骨架：`plugin.yaml`（managed、platform_only、依赖 `linapro-extlogin-core >=0.1.0`）、`go.mod`、`Makefile`、`plugin_embed.go`、目录结构对齐 google 协议插件
- [x] 1.2 声明菜单 `settings` + `settings-update`，`parent_key: plugin:linapro-extlogin-core:auth-login`；权限码 `linapro-oidc-generic:settings:view|update`
- [x] 1.3 双语 i18n：`plugin.json` / `menu.json` / `error.json`（及 apidoc 占位）；embed 菜单本地化单测（对齐 google/discord）
- [x] 1.4 将插件纳入源码插件聚合/构建路径（与现有 oidc 插件同一登记方式）；`GOWORK=off go build` 插件包通过

## 2. Settings 与配置

- [x] 2.1 实现 settings 服务：sys_config 键（issuer、client_id、client_secret、redirect_url、scopes、display_name、allow_auto_provision 默认关、default_backend_redirect、connection_key 固定 `default`）
- [x] 2.2 API：`GET/PUT .../settings`；secret 脱敏投影；空/掩码 secret 保留原值；请求时 `Load` 无需重启
- [x] 2.3 前端 `frontend/pages/settings.vue`：表单、复制回调 URL 提示、自动开户默认关文案
- [x] 2.4 如 google 使用 SQL 种子 sys_config：补充 `manifest/sql` 与 uninstall；否则文档说明首次保存创建（与既有惯例一致）

## 3. OIDC 协议服务

- [x] 3.1 引入 OIDC/JWT 校验依赖（优先 go-oidc 或等价）；实现 Discovery 解析与短缓存
- [x] 3.2 实现 authorize：PKCE S256、state HMAC、nonce、scopes（强制含 openid）；凭证未配置 fail-closed
- [x] 3.3 实现 callback：校验 state/nonce；code 换 token；JWKS 验 id_token（iss/aud/exp/sub）；可选 userinfo 补 displayName
- [x] 3.4 `provider=oidc:default` 调用 `LoginByVerifiedIdentity`；`AllowAutoProvision` 读设置默认 false；`CreateLoginHandoffFromHost` 后仅 handoff 回跳 SPA
- [x] 3.5 错误回跳使用错误码/i18n，禁止内部 err 原文；无 sub 拒绝

## 4. 路由、ownership 与 catalog

- [x] 4.1 `plugin.go`：`ProvideExternalIdentity("oidc:default")`；注册 portal login/callback 与受保护 settings 路由
- [x] 4.2 `extidcap.RegisterProviderDescriptor`（ID `oidc:default`、协议 oidc、Login/Bind 能力、LoginEntryPath）
- [x] 4.3 `frontend/slots/auth.login.after` 登录入口：显示 display_name 或默认 i18n；未启用/未配置时不展示有效跳转

## 5. 测试与验证

- [x] 5.1 单元测试：authorize/PKCE/state、未配置 fail-closed、id_token 校验 fixture、无 sub 失败、auto-provision 默认 false
- [x] 5.2 `go test` / `go vet` / `gofmt` 插件包通过；聚合构建含新插件
- [x] 5.3 E2E（插件 `hack/tests/e2e`，TC 编号遵循 lina-e2e）：设置菜单可见（依赖 core）、未配置不跳假 IdP；可选 mock IdP 成功路径
- [x] 5.4 `make i18n.check` 无本插件相关失败

## 6. 文档与影响分析

- [x] 6.1 双语 `README.md` / `README.zh-CN.md`：能力边界、依赖、provider 编码、安装顺序、安全清单、审查清单
- [x] 6.2 更新 `apps/lina-plugins/README.md` 与 `README.zh-CN.md` 插件清单行
- [x] 6.3 同步 `linapro-extlogin-core` README 中「协议插件」列举加入 generic（一句）
- [x] 6.4 影响分析记录：i18n（有）、缓存（Discovery/JWKS 进程内短缓存说明）、数据权限（登录无 actor；settings 管理权限）、开发工具（无跨平台脚本变更则记无）、测试策略、DI（无新增宿主运行期依赖则记无；插件内 OIDC 客户端依赖说明）
- [x] 6.5 `openspec validate add-generic-oidc-plugin --strict` 通过；任务完成后执行 `lina-review`


## 实施记录

### 影响分析（6.4）

| 域 | 判断 |
| --- | --- |
| i18n | **有**：插件 plugin/menu/error 双语 + apidoc 占位；登录/设置文案 |
| 缓存 | **有（插件内）**：OIDC Discovery 进程内 TTL 15m；JWKS 按 URL 缓存 1h；非集群共享；失败不永久缓存 |
| 数据权限 | **无新增列表数据域**：登录路径无 actor；settings 走管理权限码；链接解析仍在 extlogin-core |
| 开发工具 | **无**：未改 Makefile/CI 跨平台脚本 |
| 测试 | 单元：authorize fail-closed、PKCE RFC 例、scopes、auto-provision 默认 false、sanitizeReturnTo；embed menu 单测；E2E TC001 |
| DI | **无新增宿主运行期依赖**；插件依赖 host `Auth.ExternalLogin` + `HostConfig.SysConfig` 构造期注入；OIDC 出站为标准库 `net/http` + `golang-jwt` |

### 验证

- `GOWORK=off go test ./...`（插件包）通过
- `openspec validate add-generic-oidc-plugin --strict`

## Feedback

- [x] **FB-1**：将企业`OIDC`登录入口迁移到工作台统一`Vben`按钮样式，并补齐全宽纵向布局与响应式`E2E`验证

### FB-1 影响分析

| 域 | 判断 |
| --- | --- |
| 架构与插件边界 | 插件内前端展示适配；复用宿主公开的`@vben/common-ui`与主题`token`，不新增宿主契约、共享抽象或跨模块调用 |
| `i18n` | 有运行时展示影响但无资源变更；复用现有`en-US`/`zh-CN`登录键，`make i18n.check`无本插件新增告警 |
| 缓存一致性 | 无`Discovery`/`JWKS`缓存或失效策略变化 |
| 数据权限 | 无数据查询、写入、存在性暴露或权限边界变化 |
| `API`与性能 | 无接口契约、请求次数或后端装配路径变化；保留原`login-start`与`returnTo`行为 |
| 开发工具与跨平台 | 无持久脚本、配置或开发入口变化 |
| 测试 | 更新插件`POM`与`E2E TC001`，覆盖统一全宽按钮、移动视口、实际中文文案和未配置`fail-closed`；回归宿主认证`TC006` |
| `DI` | 无新增运行期依赖或服务构造变化 |

### FB-1 验证

- `pnpm --filter @lina/web-antd typecheck`通过
- `pnpm --filter @lina/web-antd build`通过
- `make i18n.check`通过
- 插件`E2E TC001`在执行后已移除的临时`ESM`边界下通过（4/4）；宿主认证回归`TC006`通过（5/5）
- 桌面/移动、亮色/暗色截图审查通过，证据位于`temp/20260711/`
- `openspec validate add-generic-oidc-plugin --strict`通过

当前插件测试运行器缺少持久`ESM`包边界，直接加载插件测试会在宿主`fixture`执行前失败；本次测试已真实执行通过，但该既有运行器风险不属于本反馈的持久变更范围。

- [x] **FB-2**：授权登录目录下通用 OIDC 菜单改名为「OIDC 设置」，设置页布局与表单边距对齐 Google/Discord OIDC（`p-4` + Card + 统一 Ant/VBen 表单样式）
  - **根因**：侧栏菜单 i18n 仍为「通用 OIDC」/「Generic OIDC」；设置页未包 `p-4` 外层，表单贴边，样式细节与 Google/Discord 参考页不一致
  - **修复**：`menu.json` zh-CN=`OIDC 设置`、en-US=`OIDC Settings`；`settings.vue` 对齐 `div.p-4` + Card + Form model/name、描述与提示 class、Switch 行与保存按钮样式；同步 embed 单测、E2E TC001b 精确文案断言、plugin.yaml 菜单 fallback 名
  - **验证**：`GOWORK=off go test`；插件 E2E TC001 4/4；`make i18n.check` 无本插件新增失败；`openspec validate --strict` 通过

### FB-2 影响分析

| 域 | 判断 |
| --- | --- |
| 架构与插件边界 | 仅插件内菜单文案与设置页布局；不改宿主契约 |
| `i18n` | 有：`manifest/i18n/{en-US,zh-CN}/menu.json` 菜单标题；页面文案键未改 |
| 缓存一致性 | 无；菜单翻译依赖运行时 i18n，重建宿主后生效 |
| 数据权限 | 无 |
| API 与性能 | 无接口契约变更 |
| 开发工具与跨平台 | 无脚本变更 |
| 测试 | embed 单测 + E2E TC001a/b；登录入口回归 TC001c/d |
| `DI` | 无 |

