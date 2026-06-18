## 1. 方案冻结与任务入口

- [x] 1.1 记录`localdocs/plugin-domain-capability-expansion-design.md`审查结论：无必须用户澄清的问题；当前已完成未归档变更`move-plugin-codegen-config-to-root-hack`不与本变更领域能力协议冲突。
- [x] 1.2 运行`openspec validate expand-plugin-domain-capabilities --strict`，确认阶段 0 和阶段 1 OpenSpec 产物可校验。

## 2. 普通领域契约和宿主适配器

- [x] 2.1 更新`usercap`契约、DTO、上限常量和`capabilityhost`用户 adapter，实现`Users.Current`和`Users.BatchResolve`，并覆盖 actor 缺失、数据权限过滤、批量去重、缺失集合和超限错误。
- [x] 2.2 更新`authcap/authz`契约和`capabilityhost`授权 adapter，实现`BatchHasPermissions`，并证明不循环调用`HasPermission`作为常规装配。
- [x] 2.3 将`dictcap.EnsureValuesVisible`提升到普通`Service`消费面，复用现有字典 owner 校验逻辑，并覆盖不可见或不存在整体拒绝。
- [x] 2.4 更新`sessioncap`契约和`capabilityhost`会话 adapter，实现`Sessions.Current`，复用启动期共享 session store，覆盖缺少 token/session 上下文 fail-closed。

## 3. 动态 Host Service 同步

- [x] 3.1 更新`protocol/hostservices` catalog、protocol 常量和公开 alias，声明`users.current.get`、`users.resolve.batch`、`permissions.batch_has`、`values.visible.ensure`和`sessions.current.get`。
- [x] 3.2 更新`pluginbridge/internal/domainhostcall`动态 guest client，使动态插件消费新增方法时仍通过领域接口语义和 JSON envelope。
- [x] 3.3 更新`internal/service/plugin/internal/wasm`dispatcher，只做授权、解码、委托和错误映射，并覆盖系统上下文、未授权和 service 缺失路径。
- [x] 3.4 更新 catalog、guest client、dispatcher 和 JSON envelope 覆盖测试，确保新增动态方法任一同步点遗漏会失败。

## 4. README 与影响记录

- [x] 4.1 更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，同步普通领域能力和动态 host service 表；确认中英文事实一致。
- [x] 4.2 在任务记录中写明影响分析：无静态 HTTP API、前端 UI、SQL schema、插件清单、语言包、开发工具脚本或 E2E 变化；有后端 Go、动态协议、数据权限、缓存复用和文档影响。
- [x] 4.3 记录 DI 来源检查：新增或修改的 adapter 依赖 owner、创建位置、传递路径、共享实例策略；若无新增运行期依赖则明确无影响。

## 5. 验证与审查

- [x] 5.1 运行`openspec validate expand-plugin-domain-capabilities --strict`。
- [x] 5.2 运行覆盖变更包的 Go 编译门禁：`cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`。
- [x] 5.3 运行宿主启动绑定烟测：`cd apps/lina-core && go test ./internal/cmd -count=1`，若因外部依赖阻断则记录替代验证和剩余风险。
- [x] 5.4 运行文档和静态检查：`git diff --check`，并静态检索确认未新增宿主核心表直连、动态`data`核心表授权、per-domain 专用 codec 或动态`i18n`host service。
- [x] 5.5 执行`lina-review`任务级审查，修复严重问题后再标记实现完成。

### 1.1 实施记录

| 项目 | 记录 |
|------|------|
| `localdocs`审查结论 | `localdocs/plugin-domain-capability-expansion-design.md`已经明确目标、非目标、阶段 0 冻结门禁、阶段 1 范围、数据权限、缓存、测试和影响分析，无必须回问的需求缺口。 |
| 活跃变更冲突判断 | 当前仍有已完成未归档变更`move-plugin-codegen-config-to-root-hack`；该变更聚焦插件本地工具配置、构建聚合和旧 release snapshot 迁移，不新增普通`capability`方法，也不扩展新的动态领域`host service`目录，与本变更的用户、权限、字典和会话能力语义无冲突。 |
| 推进方式 | 已创建正式 OpenSpec 变更`expand-plugin-domain-capabilities`，将`localdocs`方案中的阶段 0 和阶段 1 第一批落为 proposal、design、delta specs 和 tasks。 |

### 1.2 验证记录

| 命令 | 结果 |
|------|------|
| `openspec validate expand-plugin-domain-capabilities --strict` | 通过。 |

### 2.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 2.1 | `usercap.Service`新增`Current`、`BatchResolve`、`BatchResolveInput`、`ResolveKey`和领域上限常量；`capabilityhost_user.go`实现当前 actor 用户投影和按 ID、用户名、邮箱、手机号的一次性集合查询。查询在数据库侧先组合 key 条件，再复用`tenantFilter.Apply`和`datascope.Service.ApplyUserScope`，不可见和不存在统一进入`MissingIDs`。测试覆盖 actor 缺失、数据权限空集合、超限和重复 key 去重。 |
| 2.2 | `authz.Service`新增`BatchHasPermissions`和上限常量；实现从`CapabilityAuthorizationSnapshot.Permissions`构建集合后批量返回布尔结果，不循环调用`HasPermission`作为常规装配。`HasPermission`仅作为单 key 兼容消费面委托到批量方法。测试覆盖快照集合结果和超限错误。 |
| 2.3 | `dictcap.Service`新增`EnsureValuesVisible`和上限常量；`capabilityhost_dict.go`复用`ResolveLabels`的字典类型、租户覆盖和缺失值归一逻辑，任一缺失或不可见整体返回`CAPABILITY_DENIED`。测试覆盖缺失值整体拒绝和超限错误。 |
| 2.4 | `sessioncap.Service`新增`Current`；`capabilityhost_session.go`通过共享`bizctxcap.Service`读取当前`TokenID`，再复用`BatchGet`和共享`session.Store.BatchGetScoped`完成租户和数据权限过滤。缺少 token 上下文返回`CAPABILITY_CONTEXT_REQUIRED`，未找到或不可见返回`CAPABILITY_DENIED`。测试覆盖当前 token 成功路径和缺 token fail-closed。 |

### 3.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 3.1 | `pluginbridge/internal/hostservice`、`protocol_hostservice_contract.go`和`protocol/hostservices/catalog.go`已声明`users.current.get`、`users.resolve.batch`、`permissions.batch_has`、`values.visible.ensure`和`sessions.current.get`，并保持资源授权为`host:users`、`host:authz`、`host:dict`、`host:sessions`。 |
| 3.2 | `pluginbridge/internal/domainhostcall`已为`Users.Current`、`Users.BatchResolve`、`Authz.BatchHasPermissions`、`Dict.EnsureValuesVisible`和`Sessions.Current`补齐动态 guest client，全部继续使用统一`HostServiceJSONRequest`/`HostServiceJSONResponse` envelope。 |
| 3.3 | `internal/service/plugin/internal/wasm`已在 users、authz、dict、sessions dispatcher 和 registry 中注册新增方法；dispatcher 只做授权后解码、构造`CapabilityContext`、委托领域服务和错误映射。`contextWithHostCallBizContext`同步透传动态 identity 的`TokenID`，`CapabilityAuthorizationSnapshot`同步透传动态 identity 的`Permissions`。 |
| 3.4 | 已新增/更新 WASM dispatcher 测试覆盖 user current/resolve、authz batch_has、dict ensure、session current/search/batch，以及 catalog/descriptor 同步和普通 JSON service 无专用 codec 测试；`go test ./pkg/plugin/pluginbridge/... -count=1`和目标范围 Go 测试均已通过。 |

### 4.x 实施记录

| 项目 | 记录 |
|------|------|
| README 同步 | 已更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`中普通领域能力描述与动态`hostServices`表。中英文均同步列出`permissions.batch_has`、`users.current.get`、`users.resolve.batch`、`values.visible.ensure`、`sessions.current.get`。 |
| 影响分析 | 无静态 HTTP API、前端 UI、SQL schema、插件清单、语言包、运行时用户可见文案或 E2E 变化；不新增`.sh`、`.ps1`、`.cmd`、Node 脚本或开发工具入口。本变更有后端 Go、动态协议 catalog/alias/guest/dispatcher、数据权限、缓存复用和 README/OpenSpec 文档影响。`i18n`影响判断：未新增 UI 文案、API 文档源文本、语言包或插件清单文案；README 文档双语同步维护，无运行时`i18n`资源变更。E2E 影响判断：无前端可观察行为或页面工作流变化，使用 Go 单元测试和协议测试覆盖。 |
| 数据权限与性能 | 用户解析和当前用户复用`tenantFilter`与`datascope.Service.ApplyUserScope`数据库侧过滤；会话当前投影复用`session.Store.BatchGetScoped`；字典校验复用`ResolveLabels`的租户可见路径；权限批量判断复用当前动态 route 身份权限快照，集合化 map 判断，无数据库循环和无逐项`HasPermission`常规调用。 |
| 缓存一致性 | 不新增本地缓存或失效机制。权限判断复用请求身份权限快照；字典读取复用现有字典表和既有 refresh revision 机制；会话读取复用启动期共享`session.Store`；用户投影直接读取宿主权威表并经过租户/数据权限过滤。 |
| DI 来源检查 | `Sessions.Current`新增`bizctxcap.Service`依赖。owner 为宿主`bizctx`服务，创建位置为`capabilityhost.New`中的`bizCtxAdapter := newBizCtxAdapter(bizCtxSvc)`；传递路径为`capabilityhost.New(... bizCtxSvc ...) -> newBizCtxAdapter(bizCtxSvc) -> newSessionCapabilityAdapter(authSvc, bizCtxAdapter, scopeSvc, sessionStore, tenantSvc)`；共享实例策略为复用启动期传入的`bizCtxSvc`和已有共享`sessionStore`、`datascope.Service`、`tenantspi.RuntimeService`，未在 adapter 或 dispatcher 内部临时`New()`关键服务。其他新增方法未新增运行期依赖。 |

### 5.x 验证记录

| 任务 | 命令或检查 | 结果 |
|------|------------|------|
| 5.1 | `openspec validate expand-plugin-domain-capabilities --strict` | 通过，输出`Change 'expand-plugin-domain-capabilities' is valid`。 |
| 5.2 | `cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1` | 通过。 |
| 5.3 | `cd apps/lina-core && go test ./internal/cmd -count=1` | 通过。 |
| 5.4 | `git diff --check` | 通过，无输出。 |
| 5.4 | 静态检索新增动态方法、`i18n`host service、per-domain codec、核心表动态`data`授权和宿主核心表直连 | 通过。检索结果仅包含本次冻结的 JSON catalog/alias/README 方法、既有`i18n`host service 拒绝测试、既有`sys_user`拒绝测试和`capabilityhost`用户数据权限过滤；未发现新增动态`i18n`host service、阶段一专用 codec、动态`data`核心表授权或绕过领域 DTO 的核心表暴露。 |

### 5.5 Lina 审查记录

| 项目 | 记录 |
|------|------|
| 审查范围 | 任务级审查。范围来自`git status --short`、`git ls-files --others --exclude-standard`和`expand-plugin-domain-capabilities`上下文；排除当前工作区中与本变更无关的`httpstartup`、`config`、`role`、`manifest/config`和`hack/deploy`既有改动。审查范围共 43 个文件，包含能力契约、`capabilityhost`、`pluginbridge`、`wasm`、README 和本 OpenSpec 变更文件。 |
| 已读取规则 | `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/database.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`、`.agents/rules/testing.md`、`.agents/rules/dev-tooling.md`和`.agents/instructions/markdown-format.instructions.md`。 |
| 发现与修复 | 审查发现`values.visible.ensure`动态 dispatcher 会在进入领域服务前丢弃空白字典值，可能让空白目标绕过字典可见性校验。已修改`wasm_host_service_dict.go`保留 trim 后的请求值并委托领域服务 fail-closed，新增`TestHandleHostServiceInvokeDictEnsurePreservesBlankValues`覆盖该路径。 |
| 规则域结论 | 后端 Go、插件协议、架构边界、数据权限、缓存复用、文档、`i18n`影响判断和测试门禁均通过；未命中前端 UI、静态 HTTP API、SQL schema、插件目录文件、开发工具脚本和 E2E 资产变更。数据库规则无 SQL 变更，仅审查查询路径和核心表暴露边界；开发工具规则无影响，未新增或修改工具脚本。 |
| 验证证据 | 修复后已重新运行`openspec validate expand-plugin-domain-capabilities --strict`、`cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`、`git diff --check`和静态边界检索，均通过。 |
| 审查结论 | 未发现剩余阻塞问题。剩余风险为当前工作区仍存在非本变更的未归档/未提交改动，审查未对这些无关改动给出通过结论。 |
