## 1. 宿主 SPI 契约与装配

- [x] 1.1 在 `apps/lina-core/pkg/plugin/capability/authcap/extlogin/extidspi/` 新增 `Provider` 接口与 DTO（`ResolveInput`、`ProvisionInput`、`BindInput`、`UnbindInput`、`BoundIdentity`），契约仅承载解析/开户/绑定，无 token/session/租户内部类型
- [x] 1.2 新增 `BindExternalIdentityProvider` seam + auth `identityProvider` 字段；nil 保持 fail-closed。启动装配注入待组 2 接线到 login 路径。**记述校正（见 design D5b）**：此 seam 保留复用，但"沿用 `BindExternalProvisioner` 直注入范式"仅指 post-construction 绑定动作；注入值必须是 **manager-backed lazy service**（`extidspi.New(...)`，经 `IsProviderEnabled` 惰性解析、disable→fail-closed），**不是**插件 raw provider 直推。原因：`framework-capability-registry` 规范 + D4/D7 要求惰性 + 启用门禁，raw push 违规。插件只在声明期注册 factory，不得从 `backend/plugin.go` 直达 `authSvc`
- [x] 1.3 保留 `ProvisionExternalUser` 于宿主用户域作为最小权限建号原语；契约注释已说明"建号 shape 在宿主、开户策略在插件"的分工

## 1b. 宿主建号契约扩展（D6，Momus #1/#2 阻塞项）

- [x] 1b.1 `ExternalProvisionInput`（`auth.go`）与用户域 `ProvisionExternalInput` 增加可选 `UsernameAnchor` 字段；`ProvisionExternalUser` 放宽 `@` 校验为"邮箱为空且 anchor 提供 → 允许"，anchor 存在时用 `resolveProvisionUsernameFromAnchor` 派生用户名；startup adapter 透传 anchor
- [x] 1b.2 建号幂等以链接表 `(provider,subject)` 唯一索引为权威锚点（`sys_user` 仅 `username` 唯一、`email` 无唯一索引，MUST NOT 假设 email 唯一）；无邮箱 anchor 派生用户名须**确定性可复现**（去掉 `resolveProvisionUsername` 的数字后缀去重对 anchor 路径的干扰，同 anchor 命中已有 username 即复用对应账号），MUST NOT 对同一外部身份重复建立有效链接；唯一索引冲突捕获不冒泡 500。**anchor 须碰撞抗性**：从 `(provider,subject)` 取足够熵派生（username `VARCHAR(64)`、provision 上限 30 字符有余量），或在 username 冲突复用前二次校验 re-resolve 的 `(provider,subject)` 链接一致，避免不同 subject 短哈希碰撞导致账号误合并/接管
  - **落地（插件侧，组 5）**：权威锚点=`plugin_linapro_extid_core_user_external_identity` 的 `(provider,subject)` 部分唯一索引（`WHERE deleted_at IS NULL`，解绑释放键）；anchor 由插件 `deriveUsernameAnchor` 从 `sha256(provider+"\x00"+subject)` 取 16 hex（64 bit 碰撞抗性，`oidc-` 前缀共 21 字符 < 30 上限）确定性派生；插件 `Provision` 幂等快路径 + 插入冲突后 re-resolve 复用胜出链接不冒泡 500。宿主侧确定性 anchor→username 复用（1b.1）已就位。测试：`TestProvisionEmaillessDerivesDeterministicAnchor`、`TestProvisionAbsorbsConcurrentUniqueConflict`、`TestProvisionReusesExistingLinkageIdempotently`
- [x] 1b.3 在 `usercap` 宽接口新增插件可调的最小权限外部建号方法 `ProvisionExternal`（区别于 `usercap.Create` 操作员建号），宿主适配器 `user_capability.go` 委托 `ProvisionExternalUser`。**作用域：源码插件专用**（唯一消费者 `linapro-extid-core` 为 `type: source`）——动态插件侧按 `Auth.ExternalLogin()` 先例提供 **fail-closed guest stub**（`domainhostcall_users.go`），**不**登记 protocol 常量 / descriptor / WASM dispatcher case（设计见 D6a，安全依据见 Non-Goal）。补 `TestUsersProvisionExternalFailsClosed` 断言动态路径 fail-closed 且不触达 transport
  - **DI 来源检查**：无新增运行期依赖。`usercap.Service` 契约新增方法，宿主实现落在既有 `userCapabilityAdapter`（owner=`usersvc.Service`，启动期 `httpstartup_runtime.go` 构造并注入，本任务未改构造函数签名）；guest stub 无依赖。9 处测试 fake（host 5 + 插件 4）新增 no-op 方法以维持接口一致性
  - **影响分析**：i18n=无影响（未新增用户可见文案/错误码，fail-closed stub 用 `gerror` 非 bizerr，不进入本地化响应载荷，与既有 `ExternalLogin` stub 一致）；缓存一致性=无影响（无缓存/派生状态）；数据权限=无影响（建号原语 operator-less，去重锚点由插件侧 `(provider,subject)` 负责，非本任务）；开发工具跨平台=无影响（未改脚本/Makefile）；测试策略=新增 1 个 fail-closed 单测，改动 `usercap`/user service/domainhostcall 3 包 69 测试全通过
  - **验证**：`go build ./...` ✅；`go vet ./...` 仅剩 2 个**预存在** `authcap.New` arity 错误（`wasm_host_service_test.go:664`、`integration_test.go:125`，git stash 对照证明与本任务无关，属组 1 auth SPI 遗留）；`go test ./pkg/plugin/capability/usercap/... ./internal/service/user/... ./pkg/plugin/pluginbridge/internal/domainhostcall/...` 69 passed；两插件 `go vet ./...` clean；`make lint.go` 命中的 4 issue 均在未改动文件（`api/auth/v1/*.go` gofmt、`shellexec_process_windows.go` misspell），本任务改动文件 `gofmt -l` 无输出

## 2. 宿主 LoginByExternalIdentity 重构

- [x] 2.1 重构 `auth_external_identity.go`：`ProvideExternalIdentity` 盖章（不变，位于 capabilityhost adapter）→ 调 `s.identityProvider.Resolve`，未命中按 `AllowAutoProvision` 调 `Provision`；provider 为 nil 或 `ErrProviderUnavailable` 时 fail-closed 返回 not-provisioned（`resolveExternalUserID`）。启动装配注入 manager-backed service（`httpstartup_runtime.go`，见 D5b）
- [x] 2.2 拿到 userID 后保留原有租户解析、pre-token、token 铸造、会话持久化、登录时间更新、登录成功/失败 hook 流程不变（`LoginByExternalIdentity` 自用户加载起逐行未动；`TestLoginByExternalIdentityIssuesTokenPairForLinkedUser`/`TestExternalLoginAutoProvisionIssuesSession` 验证 claims）
- [x] 2.3 移除邮箱冲突判定（`CodeAuthExternalEmailConflict` 常量 + 宿主 i18n `error.auth.external.emailConflict` 键）与 `tryAutoProvision`（策略迁入插件 `identity.Provision`）；`user_provision_external.go` 未动，仅保留建号 shape。**顺带移除死代码**：auth `ExternalProvisioner` seam/`ExternalProvisionInput`/`BindExternalProvisioner`/`provisioner` 字段与 `httpstartup_provision_adapter.go`——建号原语现经 `usercap.ProvisionExternal`（1b.3）由插件反向调用，auth 侧 seam 无消费者

## 3. 宿主移除外部身份表工件

- [x] 3.1 删除 `manifest/sql/013-auth-external-identity.sql`（无 uninstall 对应文件）
- [x] 3.2 删除 `sys_user_external_identity` 的 DAO/DO/Entity（`internal/dao` ×2、`internal/model/do`、`internal/model/entity`）
- [x] 3.3 从 `apps/lina-core/hack/config.yaml` DAO 生成清单移除 `sys_user_external_identity`
- [x] 3.4 宿主单测改造为 provider-mock 驱动：`auth_external_identity_test.go`（`mockIdentityProvider` + 新增无 provider fail-closed 用例）、`auth_external_provision_test.go`（开关关闭 fail-closed、provider 策略错误透传、`ErrProviderUnavailable` 映射、开户后铸 token）。9 测试全过；仅 `sys_user` fixture 触库

## 4. 新建 linapro-extid-core 插件骨架

- [x] 4.1 创建 `apps/lina-plugins/linapro-extid-core/` 目录结构（`plugin.yaml`、`plugin_embed.go`、`go.mod`/`go.sum`、`Makefile`、`hack/config.yaml`、`backend/plugin.go`、`manifest/`），参照 `linapro-demo-source`；无 `frontend/`（capability-only，embed 仅 `plugin.yaml manifest`）
- [x] 4.2 `plugin.yaml` 声明 `type: source`、`distribution: builtin`、`platform_only`/`global`、`i18n` 配置（en-US 默认 + zh-CN）；无独立菜单（仅提供能力 + 自隔离 API）
- [x] 4.3 建表 SQL 与卸载 SQL 落地：`plugin_linapro_extid_core_user_external_identity`（字段按 D2；软删 `deleted_at` + **部分唯一索引** `WHERE deleted_at IS NULL`，解绑释放 `(provider,subject)` 键以支持重绑；`idx_plugin_linapro_extid_core_identity_user` 支撑 List/Unbind；全部 `IF NOT EXISTS` 幂等；无自增 id 显式写入；无 Seed DML）。卸载 SQL 仅 `DROP TABLE`，**不级联删除** `sys_user` 孤儿账号（7.5 语义，注释记录）
- [x] 4.4 插件 `hack/config.yaml` DAO 生成配置 + `make dao dir=apps/lina-plugins/linapro-extid-core/backend` 生成 DAO/DO/Entity（先经临时 SQL 执行器建表，执行后即删）

## 5. 插件实现 provider SPI

- [x] 5.1 `backend/internal/service/identity/` 实现 `extidspi.Provider`：`Resolve` 走 `(provider,subject)` 键查链接表（软删自动过滤），未命中 `found=false` 不报错
- [x] 5.2 `Provision`：幂等快路径（已链接即复用）→ 建号经 1b.3 `usercap.ProvisionExternal` → 写链接；插入唯一索引冲突时 re-resolve 复用胜出链接不冒泡 500（D6 顺序②，孤儿账号由确定性 anchor 复用消解）
  - **反馈修复（linglink 预览实测 Google 登录报 "Not signed in"）**：初版经 `usercap.BatchResolve` 做同邮箱冲突判定，但登录路径**无 actor 上下文**，datascope 层必然拒绝（`USER_NOT_AUTHENTICATED`）。根因=带数据权限的能力不可用于未认证登录路径；未过滤邮箱存在性查询只有宿主内部可做。**修正**：防接管安全不变式下沉到宿主铸号原语 `ProvisionExternalUser`（未过滤 `sys_user` email 计数，命中返回 `USER_PROVISION_EMAIL_CONFLICT`，宿主 i18n 双语词条已补）；capability 适配器把内部码翻译为 `usercap.ErrProvisionEmailConflict` 哨兵（插件模块无法 import internal 错误码）；插件 `errors.Is` 识别哨兵映射为 `PLUGIN_EXTID_CORE_PROVISION_EMAIL_CONFLICT`——冲突**策略语义**（拒绝静默开户、要求先登录再绑定）仍归插件，**查询机制**归宿主。测试同步改造（fake 返回哨兵 + 断言无链接残留），host 9 + 插件 7 测试全过
- [x] 5.3 无邮箱开户：邮箱为空且未传 anchor 时用 `deriveUsernameAnchor`（sha256 16 hex + `oidc-` 前缀）派生确定性 anchor 建号，不触发邮箱冲突判定
- [x] 5.4 `Bind`/`Unbind`/`List`：仅作用于当前会话用户；他人占用 `(provider,subject)` 返回 `PLUGIN_EXTID_CORE_BIND_CONFLICT`，自身重复绑定幂等成功；越权解绑统一 not-found 不泄露存在性；插入竞态经唯一索引吸收
- [x] 5.5 `backend/plugin.go` 经 `plugin.Providers().ProvideExternalIdentityProvider(factory)` 注册引擎工厂（声明期注册、宿主 manager 惰性构造 + 启用门禁，见 D5b）；**不**声明 provider ID 字符串归属（本插件不调 `LoginByVerifiedIdentity`）
- [x] 5.6 `manifest/i18n/{zh-CN,en-US}/error.json` 落地 9 个错误码本地化（键由 `PLUGIN_EXTID_CORE_*` 派生）
- [x] 5.7 Bind/Unbind/List 用户可达交付面：`backend/api/identity/v1/` DTO（GET/POST/DELETE 语义、`dc`/`eg`/`v` 标签齐全）+ `make ctrl` 生成控制器 + 路由注册（Auth+Tenancy+Permission 中间件组）+ `manifest/i18n/zh-CN/apidoc/plugin-api-main.json` 翻译与 `en-US/apidoc` 空占位。**权限标签记录**：端点为当前用户自隔离资源（D3 例外），仅要求登录认证，不声明 `permission` 标签也无菜单权限节点——与 4.2 无菜单声明一致，Permission 中间件对未声明权限的端点放行
- [x] 5.8 **记录：本变更仅交付后端能力**，前端绑定入口（已绑定列表 + 绑定/解绑交互页面）留后续迭代；API 交付面已可支撑后续前端接入
  - **i18n 影响记录**：插件 `i18n.enabled: true`，error.json 双语 + apidoc zh-CN 翻译 + en-US 空占位已交付；无运行时前端页面故无 `$t()` 键；`make i18n.check` 通过（exit 0），无 oidc-core 相关告警

## 6. 既有插件依赖调整

- [x] 6.1 `linapro-oidc-google`/`linapro-oidc-discord` 的 `plugin.yaml` 增加 `dependencies.plugins: [{id: linapro-extid-core, version: ">=0.1.0"}]`。**D5 实施校正**：两插件维持 `distribution: managed`（带占位凭证的 OAuth 参考实现不应被宿主启动强制安装），core 为 `builtin` 保证依赖可满足，已记入 design D5
- [x] 6.2 两插件后端代码零改动，登录路径仍走 `extlogin` seam；core 未启用时 fail-closed 由宿主测试（`TestLoginByExternalIdentityFailsClosedWithoutProvider`）与 manager 测试（`TestManagedProviderFollowsEnablementTransitions`）验证

## 7. 测试与验证

- [x] 7.1 宿主单测（mock provider）：provider 缺失/未启用 fail-closed（`TestLoginByExternalIdentityFailsClosedWithoutProvider` + manager 三测试）、resolve 命中后走 token 铸造并校验 claims（`TestLoginByExternalIdentityIssuesTokenPairForLinkedUser`）。provider ownership 拒绝为既有 capabilityhost 测试覆盖（本次未改盖章路径）。auth 包 9 外部登录测试全过
- [x] 7.2 插件单测（`identity_test.go`，7 测试全过）：resolve 未命中、幂等复用、邮箱冲突、开关关闭、无邮箱确定性 anchor、并发唯一冲突吸收、bind 冲突/幂等/越权解绑/列举自隔离/解绑后重绑/跨 provider 独立
- [x] 7.3 数据权限验证（D3 例外记录）：登录未命中统一 not-provisioned 不泄露邮箱账号存在性（宿主 `TestLoginByExternalIdentityRejectsUnprovisionedIdentity` + 插件邮箱冲突路径无信息泄露断言）；绑定/解绑/列举仅限当前会话用户、越权 unbind 统一 not-found（`TestBindUnbindListSelfIsolation`）；权威边界=当前会话用户 + `(provider,subject)` 唯一键，拒绝策略=越权整体拒绝/重复占用冲突
- [x] 7.4 并发正确性单测（D8）：`TestProvisionAbsorbsConcurrentUniqueConflict`（并发胜出链接复用、唯一索引冲突不冒泡 500）+ `TestProvisionEmaillessDerivesDeterministicAnchor`（确定性 anchor 使并发建号收敛同一账号）+ `TestProvisionReusesExistingLinkageIdempotently`（链接写入失败后幂等复用路径）
- [x] 7.5 生命周期单测：跨 provider 同 subject 不交叉链接（`TestBindUnbindListSelfIsolation` 跨 provider 断言）；禁用 fail-closed、重新启用恢复（`TestManagedProviderFollowsEnablementTransitions`）；禁用保留链接数据=禁用仅门禁 factory 解析、不触数据（manager 机制保证）；卸载删表不级联删除 `sys_user` 孤儿账号=uninstall SQL 仅 `DROP TABLE`（SQL 注释记录）
- [x] 7.6 oidc-core `go build`/`go vet`/`gofmt` 通过（GOWORK=off）；google/discord 本次零 Go 改动；宿主 `go build`/`go vet ./...` 全绿（含顺带修复组 1 遗留的 2 处 `authcap.New` 三参调用）；插件完整聚合（14 manifests 含 oidc-core）`go build ./...` 通过；`make i18n.check` exit 0。**lint 记录**：`make lint.go plugins=1` 的 gofmt 告警为 Windows CRLF checkout 环境噪音（`gofmt -w` 后 `git diff` 为空即证），非代码问题；本任务自有文件 `gofmt -l` 全部干净；顺带修复 1 处预存在 misspell（`shellexec_process_windows.go`）
- [ ] 7.7 端到端探测：安装 core+google，走 One Tap/授权码登录成功；卸载 core 后 external login fail-closed。**[deferred]** 需运行完整宿主栈 + 浏览器流程；fail-closed 转换语义已由 7.1/7.5 单测覆盖，真实登录链路留待下一会话或人工验证后勾选

## 8. 文档同步

- [x] 8.1 更新 `apps/lina-core/pkg/plugin/README.md` 与 `README.zh-CN.md`：Auth 域行改述为 provider SPI 引擎（`ProvideExternalIdentityProvider` 声明、linapro-extid-core 归属、fail-closed 语义），移除 `sys_user_external_identity` 宿主独占表述
- [x] 8.2 更新 `apps/lina-plugins/README.md` 与 `README.zh-CN.md` 插件清单：加入 `linapro-extid-core`，顺带补齐缺失的 `linapro-oidc-google`/`linapro-oidc-discord` 行
- [x] 8.3 编写 `linapro-extid-core` 双语 `README.md`/`README.zh-CN.md`（能力边界、宿主边界、数据权限边界、开户契约、目录结构、审查清单）；同步修正 google/discord 双语 README 的过时表述（"宿主负责开户"→ 委托 core、依赖声明、审查清单）
- [x] 8.4 PR 描述内容（供创建 PR 时使用）：external-identity 存储与开户策略已从 host core 抽离为 `linapro-extid-core` 插件——宿主仅保留薄 `extlogin` 接口 + `extidspi` SPI + manager 惰性注入（启用门禁 fail-closed）；`sys_user_external_identity` 表及 DAO 工件已删除；两个 follow-up（无邮箱确定性 anchor 开户、绑定/解绑/列举自隔离 API）已随插件闭环；google/discord 仅新增对 core 的依赖声明，登录调用路径零改动

## Feedback

- [x] **FB-11**: `linapro-extid-core` 插件私有表名不符合项目规范（应为 `plugin_<plugin-id>_*`）
  - **根因**：实现/设计把宿主 `sys_user_external_identity` 迁出时仅去掉 `sys_`，写成裸表名 `user_external_identity`；与全仓插件私有表约定 `plugin_<plugin-id-with-underscores>_<entity>`（如 `plugin_linapro_org_core_dept`、`plugin_linapro_tenant_core_user_membership`）不一致
  - **修复**：表重命名为 `plugin_linapro_extid_core_user_external_identity`；索引缩短为 `uk_plugin_linapro_extid_core_identity_provider_subject` / `idx_plugin_linapro_extid_core_identity_user`（规避 PG 63 字符标识符上限）；同步安装/卸载 SQL、`hack/config.yaml`（`removePrefix`）、DAO/DO/Entity、identity 单测建表、插件 README、OpenSpec D2/spec/proposal
  - **影响分析**：i18n=无；缓存=无；数据权限=无（仅物理表名，自隔离语义不变）；开发工具=无；测试=identity 单测；DI=无；文档=插件 README 双语文档同步
  - **验证**：`GOWORK=off go test ./backend/internal/service/identity/...` 通过；`openspec validate refactor-external-identity-to-provider-plugin --strict` 通过；静态检索无裸表名 `user_external_identity` 残留（历史宿主 `sys_*` 引用除外）
- [x] **FB-10**: Google/Discord 插件介绍描述去技术化，改为产品化官方插件文案
  - **根因**：`plugin.yaml` / `plugin.json` 仍写 “external-identity seam / 外部身份接缝 / 参考源码插件”，与 `linapro-extid-core` 及内容/监控类官方插件「Official source plugin for … / 提供…的官方源码插件」风格不一致
  - **修复**：重写 en-US 源文本与 zh-CN 翻译；同步插件清单 README 与插件 README 首段；名称保持不变
  - **影响分析**：i18n=有（`plugin.linapro-oidc-*.description` en/zh）；缓存=无；数据权限=无；开发工具=无；测试=JSON 键与 en 对齐 `plugin.yaml` 静态校验；DI=无
  - **验证**：JSON 解析 + en-US 与 `plugin.yaml` description 一致
- [x] **FB-9**: 登录页第三方登录入口对齐 Vben5 样式（分隔线 + 横向图标按钮）
  - **根因**：宿主 `auth.login.after` 使用纵向全宽 Ant Design 按钮；与 Vben5 `ThirdPartyLogin`（「其他登录方式」分隔线 + 横向圆形图标按钮）不一致
  - **修复**：宿主 `login.vue` 增加 Vben 同构分隔线与 `flex-wrap justify-center` 图标行（无插件时 `:has` 隐藏）；Google/Discord 入口改为圆形 `type="text"` 图标按钮 + Tooltip；复用既有 i18n 键作 tooltip
  - **影响分析**：i18n=无新增键（复用 `authentication.thirdPartyLogin` 与插件 `login.button`）；缓存=无；数据权限=无；开发工具=无；测试=宿主 TC006 TC-2a/2e + 插件登录入口 E2E；DI=无
  - **验证**：`pnpm exec playwright test e2e/auth/TC006-login-page-presentation.ts` 5 passed
- [x] **FB-8**: Google/Discord 设置页「工作台落地页」补充简短提示说明
  - **根因**：字段仅有 label/placeholder，用户不易区分其与回调 URL、SSO 投递的差异
  - **修复**：两插件设置页按现有 hint 样式增加 `defaultRedirectHint`；en-US/zh-CN `plugin.json` 同步
  - **影响分析**：i18n=有（`plugin.linapro-oidc-*.settings.defaultRedirectHint`）；缓存=无；数据权限=无；开发工具=无；测试=文案/静态展示，无行为变更；DI=无
  - **验证**：JSON 键存在；页面 `$t` 引用与键一致
- [x] **FB-7**: 管理后台 Google/Discord OIDC 菜单名称去掉「设置」后缀
  - **根因**：zh-CN `menu.json` 将设置页菜单标题写成「Google/Discord OIDC 设置」，与英文侧「Google/Discord OIDC」及产品期望不一致
  - **修复**：zh-CN 设置菜单标题改为「Google OIDC」「Discord OIDC」；同步 embed 单测与 E2E TC001d 断言；页面内 `$t('...settings.title')` 仍可保留「设置」语义，仅侧边栏菜单名去后缀
  - **影响分析**：i18n=有（仅菜单 `menu.plugin:linapro-oidc-*:settings.title` zh-CN）；缓存=无；数据权限=无；开发工具=无；测试=embed 单测 + E2E TC001d；DI=无
  - **验证**：`GOWORK=off go test` 两插件 embed 单测通过
- [x] **FB-6**: 英文环境下 Google/Discord OIDC 设置菜单仍显示中文
  - **根因**：宿主 `localizeMenuEntity` 按 `menu.<menuKey>.title` 投影菜单名；`linapro-extid-core` 的「授权登录」目录已补 `menu.json`，但 `linapro-oidc-google`/`linapro-oidc-discord` 仅有 `plugin.json`/`error.json`，缺少 `manifest/i18n/{en-US,zh-CN}/menu.json`。翻译未命中时回退库内 `sys_menu.name`；历史安装或中文源文案写入后，英文环境仍显示中文
  - **修复**：为两插件补齐 `menu.json`：`settings` + `settings-update` 的 en-US/zh-CN 标题（键分别为 `menu.plugin:linapro-oidc-google:settings.title` 等）
  - **影响分析**：i18n=有（插件运行时菜单键 en-US+zh-CN）；缓存=无（菜单本地化随请求语言投影，非业务缓存）；数据权限=无；开发工具跨平台=无；测试策略=embed 单测 + JSON/`make i18n.check` + 宿主 `TestLocalizeMenuEntity*` 回归 + E2E TC001d（`Accept-Language` en-US/zh-CN 精确断言）；DI=无新增运行期依赖
  - **验证**：JSON 解析通过；`make i18n.check` exit 0；宿主 menu 本地化单测通过；`GOWORK=off go test` google/discord `TestEmbeddedMenuJSONLocalizesSettingsTitle` 通过；E2E TC001d 已写入（生效需重编译宿主以重新 embed 插件 manifest）
- [x] **FB-5**: 第三方授权登录插件设置菜单统一挂到「授权登录」目录（由 extid-core 安装创建，非宿主预留）
  - **根因**：Google/Discord 原挂 `extension`；一度误用宿主种子目录 `auth-login`，违背“领域目录随领域插件安装”的边界
  - **修复**：`linapro-extid-core` 声明 `plugin:linapro-extid-core:auth-login` 目录菜单；Google/Discord `parent_key` 指向该目录；**不新增/不保留宿主 SQL**（已删除误加的 `013-auth-login-catalog.sql`）
  - **影响分析**：i18n=有（extid-core `menu.plugin:linapro-extid-core:auth-login` en/zh）；缓存=无；数据权限=无；开发工具=无；测试=E2E TC-18
  - **验证**：目录仅在 extid-core 安装后出现；OIDC 设置挂其下；E2E 已拆到插件目录（extid-core TC001 / google+discord TC001），宿主 TC018 与 TC006 对 OIDC 的硬耦合已移除
- [x] **FB-4**: 点击 Google/Discord 登录无法进入真实第三方登录页（落地 IdP 错误页）
  - **根因**：默认 ClientID/Secret 为 `REPLACE_ME_*` 占位值且非空，`BuildAuthorizeURL` 视为已配置后 302 到 Google/Discord，IdP 以 invalid_client 拒绝；用户感知为“无法跳转第三方登录页”。另缺 Vite `/portal` 代理、登录回跳路径与 history 模式不一致。
  - **修复**：占位/空凭证 fail-closed；登录入口传 `returnTo` 并写入签名 state；配置错误回登录页并 i18n 提示；Vite 代理 `/portal`；默认回跳 `/admin/auth/login`
  - **影响分析**：i18n=有（`authentication.externalLoginConfigMissing` en/zh）；缓存=无；数据权限=无；开发工具=Vite 代理跨平台无关；测试=单测 + E2E TC-2f
  - **验证**：`go test` google/discord oauth 包通过；curl 未配置凭证 Location 回登录页；`pnpm exec playwright test ... TC-2f` 通过
- [x] **FB-1**: `linapro-extid-core` 插件介绍信息（name/description）缺少 i18n，插件管理页中文环境回退英文
  - **根因**：`plugin.yaml` 已启用 `i18n.enabled: true`，宿主 `localizePluginMetadata` 读取 `plugin.<id>.name` / `plugin.<id>.description`；实现只交付了 `error.json` 与 `apidoc`，遗漏 `manifest/i18n/{en-US,zh-CN}/plugin.json`
  - **修复**：按同类插件约定补齐双语 `plugin.json`（en-US 对齐 `plugin.yaml` 源文本，zh-CN 提供中文名称与介绍）
  - **影响分析**：i18n=有影响（插件介绍键 `plugin.linapro-extid-core.{name,description}`，en-US+zh-CN）；缓存一致性=无影响；数据权限=无影响；开发工具跨平台=无影响；测试策略=JSON 合法性 + `make i18n.check` 治理验证（无运行时逻辑变更，无需 E2E/单测）；DI=无新增运行期依赖
  - **验证**：`python3` JSON 解析通过；键存在性断言通过；`make i18n.check` exit 0
- [x] **FB-2**: `linapro-extid-core` 插件介绍文案过技术化，需更地道、自然、通用的产品描述
  - **根因**：介绍沿用内部实现术语（provider/subject、resolve/provision/bind、external-login seam、插件 ID），与其他官方插件「Official source plugin for … / 提供…的官方源码插件」风格不一致
  - **修复**：按同类插件文案风格重写 `plugin.yaml` 英文源文本与 `manifest/i18n/{en-US,zh-CN}/plugin.json`
  - **影响分析**：i18n=有影响（`plugin.linapro-extid-core.description` en-US/zh-CN + `plugin.yaml` 源文本）；缓存=无；数据权限=无；开发工具=无；测试策略=JSON/文案静态校验（纯文案，无行为变更）
  - **验证**：JSON 解析通过；en-US 对齐 `plugin.yaml`
- [x] **FB-3**: 插件管理详情页 Descriptions 最左标签列会出现换行，影响可读性
  - **根因**：`plugin-detail-modal` 使用两列 bordered Descriptions，标签单元格默认可换行；英文多词标签（如 Authorization Status）会折行
  - **修复**：参照 job 表单 Descriptions 模式，对 `.ant-descriptions-item-label` 设置 `white-space: nowrap`；E2E TC-13b 覆盖
  - **影响分析**：i18n=无新增键；缓存=无；数据权限=无；开发工具=无；测试策略=E2E TC-13b（用户可观察布局）
  - **验证**：`pnpm exec playwright test ...TC013...` TC-13a/13b 均通过
