# 需求文档

## 引言

本需求文档基于对 linapro 项目（包含后端 `apps/lina-core`、前端 `apps/lina-vben`、插件 `apps/lina-plugins`）的代码评审结果，梳理出 12 项可改进点，覆盖**后端架构**、**前端工程化**、**插件系统**三大领域。

改进目标：
- **降低耦合**：通过依赖注入、构造函数注入等方式替代当前的全局单例和 Setter 注入
- **统一规范**：统一错误处理、API 响应结构、状态管理方式
- **提升体验**：启用 Token 静默续期，补充请求取消/重试/去重等高级 HTTP 能力
- **增强扩展性**：插件依赖声明、槽位动态注册等

需求按优先级分为三档：
- **P0（高）**：影响内存占用、错误响应规范、用户体验，必须优先解决
- **P1（中）**：影响安全性与可扩展性，应在 P0 之后跟进
- **P2（低）**：架构优化与开发体验，可作为长期演进项

本次规划仅产出需求与任务清单文档，不在本工作流内进行编码实现。

## 分支拆分约定（全局约束）

后续所有修复任务**必须**遵守以下分支策略，以保证改动可独立评审、可选择性合入、可单独回滚：

1. **一需求 ↔ 一主分支**：每一条需求（需求 1‒12）都 SHALL 对应一个独立的 fix 分支，命名为 `fix/<short-slug>`（这里列出推荐名称，仅供对齐）
2. **同一需求拆多阶段**：若一条需求体量较大需拆分，可采用后缀序号，如 `fix/dao-injection-part1`、`fix/dao-injection-part2`，但 SHALL 在 PR 描述中明确关联同一需求 ID
3. **不允许跨需求提交**：一个分支 / 一个 PR SHALL 仅解决一条需求的验收标准，不得混入其它需求的改动（以免造成评审难度、回滚负担）
4. **已存在的补齐**：对于未覆盖完全的需求（需求 1、2、3），剩余补齐 SHALL 在原分支上继续提交（如原分支已合入主干，则另开同名后缀 `-followup` 分支，如 `fix/auth-error-handling-followup`）
5. **推荐分支名称对照表**（可调整）：
   - 需求 1 → `fix/service-instance-management`（已存在，补齐在同分支继续）
   - 需求 2 → `fix/auth-error-handling`（已存在，补齐在同分支继续）
   - 需求 3 → `fix/enable-token-refresh`（已存在，补齐在同分支继续）
   - 需求 4 → `fix/api-dto-isolation`
   - 需求 5 → `fix/plugin-service-constructor-injection`
   - 需求 6 → `fix/plugin-dependency-declaration`
   - 需求 7 → `fix/dao-interface-injection`
   - 需求 8 → `fix/frontend-adapter-simplify`
   - 需求 9 → `fix/plugin-state-pinia`
   - 需求 10 → `fix/http-client-advanced`
   - 需求 11 → `fix/plugin-slot-dynamic-register`
   - 需求 12 → `fix/plugin-state-event-bus`
6. **提交信息格式**：提交 SHOULD 遵循 Conventional Commits，使用 `fix:`、`feat:`、`refactor:` 等前缀，并在描述中引用需求编号（如 `Refs: 需求 1`）

## 已处理进度（基于本地 fix 分支）

截至当前，作者已基于本需求清单创建并提交了以下分支，**对 P0 三项需求做了初步处理但尚未全部覆盖**：

| 分支 | 提交信息 | 对应需求 | 覆盖范围 |
|---|---|---|---|
| `fix/service-instance-management` | `fix: use pluginsvc singleton Instance() across all service and controller constructors` | 需求 1 | 🟡 部分 — 仅对 `pluginsvc` 引入 `Instance()`，其它 Service（auth/role/tenantcap/session/orgcap/config 等）仍是 `New()` 多实例 |
| `fix/auth-error-handling` | `fix: replace gerror.Wrap with bizerr in auth service HashPassword` | 需求 2 | 🟡 部分 — 仅替换了 `HashPassword` 的 `gerror.Wrap`，其它仍存在的 `gerror.New/Wrap`（如登录、token 解析包装错误）尚未全面梳理 |
| `fix/enable-token-refresh` | `feat: add token refresh endpoint and enable frontend token refresh` | 需求 3 | 🟢 主体已完成 — 后端新增 refresh endpoint，前端启用刷新；仍需补齐并发去重、Refresh Token 失效处理等子项验收 |

> 这三个分支应**与本需求清单关联评审**：合入主干前需对照需求中的全部验收标准逐条确认未覆盖项，未覆盖项应在任务清单中作为补齐项继续推进。

---

## 需求

### 需求 1（P0）：Service 实例统一管理

> **当前进度：🟡 部分完成** — `fix/service-instance-management` 已为 `pluginsvc` 引入 `Instance()` 单例并替换其在 service/controller 构造器中的引用；剩余 `auth`、`role`、`tenantcap`、`session`、`orgcap`、`config` 等 Service 仍以 `service.New()` 多实例方式创建。

**用户故事：** 作为后端开发者，我希望所有 `Service` 实例由一个统一的注册中心或单例机制创建和管理，以便消除重复实例化、保证缓存状态一致、让依赖关系显式可见。

#### 验收标准

1. WHEN 系统启动 THEN 后端 SHALL 在启动阶段一次性构建所有 `Service` 实例并注册到统一的容器（`Service Registry`）或通过 `xxx.Instance()` 暴露单例
2. WHEN `Controller` 或 `Middleware` 需要使用某个 `Service` THEN 它们 SHALL 从统一容器中获取实例，而不再调用 `service.New()` 创建新实例
3. WHEN 同一个 `Service` 被多处使用 THEN 系统 SHALL 保证所有调用方持有的是同一个实例（内存唯一）
4. IF 某个 `Service` 依赖其他 `Service` THEN 这种依赖关系 SHALL 在容器构建阶段或单例初始化函数中显式声明
5. WHEN 项目重构完成 THEN 所有现有功能（认证、角色、插件等） SHALL 保持原有行为不变，单元测试与 E2E 测试全部通过
6. WHEN 本次需求最终完成 THEN 至少 `auth`、`role`、`tenantcap`、`session`、`orgcap`、`config` 这些核心 Service SHALL 与 `pluginsvc` 一样提供 `Instance()` 单例入口，并替换其在 controller / middleware / 其它 service 中的 `New()` 调用点

---

### 需求 2（P0）：认证错误统一走 bizerr

> **当前进度：🟡 部分完成** — `fix/auth-error-handling` 已替换 `auth.HashPassword` 中的 `gerror.Wrap`；当前 `auth.go` 中登录主流程的核心业务错误（如 IP 黑名单、凭据无效、用户禁用、租户不可用、token 无效等）已使用 `bizerr.NewCode()`，但仍需对整个 auth 包及 controller（`auth_v1_refresh_token.go` 等）做一次全量梳理，确认无残留的 `gerror.New/Wrap` 用于业务错误。

**用户故事：** 作为前端开发者，我希望后端认证模块返回的错误带有标准的 `errorCode` 和 `i18n` 元数据，以便前端可以统一处理错误并展示国际化消息。

#### 验收标准

1. WHEN 认证模块（`apps/lina-core/internal/service/auth`）抛出业务错误 THEN 它 SHALL 使用 `bizerr.NewCode()` 而非 `gerror.New()`
2. WHEN 认证错误被序列化到 HTTP 响应 THEN 响应体 SHALL 包含 `errorCode` 字段以及对应的 i18n key
3. WHEN 用户语言偏好为中文/英文 THEN 错误消息 SHALL 根据语言环境正确翻译
4. IF 某条错误已有现成的 `bizerr` 错误码 THEN 代码 SHALL 复用该错误码而非新建重复码
5. WHEN 改造完成 THEN 认证模块（含 `auth.go`、`auth_v1_refresh_token.go` 等） SHALL 不再出现 `gerror.New()` 用于业务错误的写法
6. WHEN 全量梳理完成 THEN 团队 SHALL 输出一份遗留 `gerror` 调用清单（`grep -R "gerror\.\(New\|Wrap\)" apps/lina-core/internal/service/auth apps/lina-core/internal/controller/auth`），并对每一条做"改 bizerr"或"保留为内部错误"的明确决策

---

### 需求 3（P0）：启用前端 Token 刷新机制

> **当前进度：🟢 主体已完成** — `fix/enable-token-refresh` 已新增后端 refresh endpoint（`auth_v1_refresh_token.go`）并在前端启用了 Token 刷新；剩余需补齐**并发去重**与 **Refresh Token 失效兜底**等子项验收，以及合入主干前的端到端测试。

**用户故事：** 作为系统使用者，我希望在 Access Token 过期时前端能够自动静默续期，以便我不会被频繁踢回登录页，提升使用流畅度。

#### 验收标准

1. WHEN 前端发起 API 请求且 Access Token 已过期 THEN 请求客户端 SHALL 自动调用刷新接口 `/api/v1/auth/refresh` 续期
2. WHEN Token 刷新成功 THEN 客户端 SHALL 用新的 Access Token 重放原始失败请求，对业务调用方透明
3. WHEN 同时多个请求并发遇到 401 THEN 系统 SHALL 只触发一次刷新请求，其余请求 SHALL 等待该刷新结果后统一重放
4. IF Refresh Token 也已失效 THEN 系统 SHALL 清空登录态并跳转至登录页
5. WHEN 前端配置项 `enableRefreshToken` 设为 `true` 且 `refreshTokenApi` 已正确配置 THEN 上述行为 SHALL 生效，无需额外修改业务代码
6. WHEN 合入主干前 THEN E2E 测试 SHALL 覆盖以下场景：(a) 单请求 401 自动刷新成功并重放；(b) 多请求并发 401 仅触发一次刷新；(c) Refresh Token 失效时清空登录态并跳转登录页

---

### 需求 4（P1）：API 响应使用独立 DTO

**用户故事：** 作为安全负责人，我希望 API 响应不会直接嵌入 `entity` 结构，避免泄漏 `password`、内部时间戳等敏感字段。

#### 验收标准

1. WHEN 任何 `Controller` 返回数据 THEN 响应结构 SHALL 使用独立定义的 `DTO`（位于 `api` 目录），而非内嵌 `*entity.XXX`
2. WHEN `DTO` 定义被生成或编写 THEN 它 SHALL 仅包含外部可见字段，明确排除 `password`、`secretKey` 等敏感信息
3. WHEN `entity` 转换为 `DTO` THEN 转换函数 SHALL 集中放置（例如 `internal/logic` 或 `internal/service` 转换层）以便维护
4. IF 某个字段在 `entity` 中存在但不应暴露 THEN 该字段 SHALL 不出现在对应的 `DTO` 中
5. WHEN 改造完成 THEN 接口契约（字段名、类型） SHALL 与现有前端调用保持兼容，必要时同步更新前端类型定义

---

### 需求 5（P1）：Plugin Service 改用构造函数注入

**用户故事：** 作为后端开发者，我希望 `Plugin Service` 的依赖在构造时一次性传入，以便在编译期就能发现缺失依赖，避免顺序错误的运行时 bug。

#### 验收标准

1. WHEN `Plugin Service` 被实例化 THEN 它 SHALL 通过构造函数（或建造者模式 `Builder.Build()`）一次性接收所有必需依赖（如 `BackendLoader`、`ArtifactParser`、`DynamicManifestLoader` 等）
2. WHEN 缺少任一必需依赖 THEN 编译期或构造调用 SHALL 立即报错/返回错误，而不是运行时空指针
3. WHEN 现有 `SetXxx()` 方法存在 THEN 它们 SHALL 被构造函数注入替代或仅保留为可选项的语义化方法
4. IF 依赖关系发生变化 THEN 修改点 SHALL 集中在构造函数签名，不需要在散落各处修改 `SetXxx` 调用顺序
5. WHEN 改造完成 THEN 现有插件加载、热更新、缓存等行为 SHALL 保持不变

---

### 需求 6（P1）：插件依赖声明与校验

**用户故事：** 作为插件开发者，我希望能在 `plugin.yaml` 中声明对其他插件的依赖（含版本约束），以便在安装/启用时由系统自动校验和提示。

#### 验收标准

1. WHEN `plugin.yaml` 中存在 `dependencies` 字段 THEN 系统 SHALL 解析其中的 `id` 与 `version`（支持 `>=`、`^`、`~` 等语义化版本约束）
2. WHEN 安装/启用某个插件且其依赖未满足 THEN 系统 SHALL 阻止启用并返回结构化错误（含缺失依赖列表）
3. WHEN 多个插件存在依赖关系 THEN 加载顺序 SHALL 按依赖图拓扑排序
4. IF 检测到循环依赖 THEN 系统 SHALL 报错并明确指出环路
5. WHEN 依赖检查通过 THEN 插件 SHALL 正常进入后续生命周期（注册、启动、可见）

---

### 需求 7（P2）：DAO 接口化与可注入

**用户故事：** 作为后端开发者，我希望 `DAO` 通过接口定义并由 `Service` 构造时注入，以便单元测试能用 Mock 替换真实数据库。

#### 验收标准

1. WHEN 每个 `DAO` 模块定义 THEN 它 SHALL 同时暴露接口（如 `SysUserDAO`）和默认实现
2. WHEN `Service` 需要使用 `DAO` THEN 它 SHALL 通过构造函数接收 `DAO` 接口而非直接使用 `dao.SysUser` 全局变量
3. WHEN 单元测试需要隔离数据库 THEN 测试代码 SHALL 能够注入 Mock DAO 实现
4. IF 项目使用 GoFrame 的 `gf gen dao` 工具 THEN 应通过扩展自定义模板（位于 `hack/tools/` 之类位置）生成接口，不修改框架源码
5. WHEN 改造完成 THEN 现有 `dao/internal/` 自动生成产物 SHALL 仍可由命令行工具重新生成

---

### 需求 8（P2）：前端 Monorepo 适配层简化

**用户故事：** 作为前端维护者，由于项目只保留 `web-antd` 一个应用，我希望简化掉冗余的多 UI 库适配抽象层，以降低代码理解成本。

#### 验收标准

1. WHEN 评审 `apps/lina-vben` 下的 `adapter` 相关目录 THEN 团队 SHALL 输出一份"删除清单"列出可移除的多 UI 库适配代码
2. WHEN 简化执行 THEN 仅保留 Ant Design Vue 直接使用的代码路径，删除其他 UI 库相关的抽象封装
3. IF 存在 `View` 组件直接 `import 'ant-design-vue'` THEN 该写法 SHALL 在简化后被允许并形成统一规范，不再要求绕道适配器
4. WHEN 简化完成 THEN 前端构建产物体积 SHALL 不增加，运行行为不变
5. WHEN 简化完成 THEN 项目 README 或 CONTRIBUTING SHALL 同步更新组件引用规范

---

### 需求 9（P2）：插件全局状态使用 Pinia 管理

**用户故事：** 作为前端开发者，我希望插件注册表不再挂在 `globalThis` 上，而是通过类型安全的 Pinia Store 管理，以减少全局污染并获得更好的 IDE 智能提示。

#### 验收标准

1. WHEN 插件系统初始化 THEN 注册表 `__linaPluginRegistry` 与 `__linaPluginSlotRegistry` SHALL 由 Pinia Store（或 `provide/inject`）替代
2. WHEN 任意组件需要访问插件注册信息 THEN 它 SHALL 通过 Store 的 getter/action 访问，且具有完整 TypeScript 类型
3. WHEN 多个插件并发注册 THEN Store 的状态变更 SHALL 是响应式且可追踪的（Vue DevTools 可见）
4. IF 当前代码存在 `globalThis.__linaPlugin*` 直接读写 THEN 改造后 SHALL 全部替换
5. WHEN 改造完成 THEN 插件加载、卸载、热更新等行为 SHALL 与之前一致

---

### 需求 10（P2）：HTTP 客户端补充高级能力

**用户故事：** 作为前端开发者，我希望 `RequestClient` 内置请求取消、重试、去重能力，以便在业务侧避免重复造轮子。

#### 验收标准

1. WHEN 业务调用 `requestClient.get(url, { signal })` THEN 客户端 SHALL 支持通过 `AbortController` 取消进行中的请求
2. WHEN 请求遇到网络错误或 5xx THEN 客户端 SHALL 按配置的 `retry: { count, delay }` 进行指数/固定延迟重试
3. WHEN 同一 `(method, url, params)` 组合的请求并发触发 THEN 客户端 SHALL 仅发起一次实际请求，其余调用复用同一 Promise 结果（请求去重）
4. IF 业务侧未传入 `signal` 或未配置 `retry` THEN 客户端 SHALL 保持原行为，无破坏性变更
5. WHEN 改造完成 THEN 既有调用点 SHALL 无需修改即可继续工作，新能力按需启用

---

### 需求 11（P2）：插件 UI 槽位动态注册

**用户故事：** 作为插件开发者，我希望能由插件自行声明并动态注册槽位位置，而不需要修改宿主代码，以提升扩展性。

#### 验收标准

1. WHEN 插件提供 UI 扩展 THEN 它 SHALL 通过统一 API（如 `registerSlot(name, component)`）注册槽位
2. WHEN 宿主渲染界面 THEN 它 SHALL 根据已注册槽位动态渲染，无需为新槽位修改宿主代码
3. WHEN 多个插件向同一槽位注册 THEN 系统 SHALL 按声明顺序或权重渲染所有内容
4. IF 槽位被卸载 THEN 宿主 SHALL 立即移除对应渲染
5. WHEN 改造完成 THEN 现有内置槽位 SHALL 通过同样的注册机制工作（自一致性）

---

### 需求 12（P2）：插件状态变更响应式机制统一

**用户故事：** 作为前端开发者，我希望插件启用/禁用/卸载等状态变更触发的页面刷新逻辑只走一条统一通道，以降低维护复杂度。

#### 验收标准

1. WHEN 插件状态发生变更（启用、禁用、版本切换、卸载） THEN 前端 SHALL 通过单一事件总线或 Store 变更触发响应
2. WHEN UI 侧依赖插件状态 THEN 它 SHALL 通过订阅该统一通道接收变更通知
3. WHEN 多个模块需要刷新 THEN 触发顺序 SHALL 是确定且可追踪的
4. IF 当前存在多条独立的刷新路径 THEN 改造后 SHALL 收敛为一条
5. WHEN 改造完成 THEN 插件页面在状态变更后的刷新表现 SHALL 与改造前一致或更好（无闪烁/无遗漏）

---

## 优先级与里程碑建议

| 阶段 | 包含需求 | 目标 | 当前状态 |
|------|---------|------|----------|
| 里程碑 1（P0） | 需求 1、2、3 | 解决最影响用户体验与规范一致性的核心问题 | 🟡 进行中（3 个 fix 分支已提交，需补齐剩余覆盖范围并合入主干） |
| 里程碑 2（P1） | 需求 4、5、6 | 收敛安全风险，提升插件系统可扩展性 | ⚪ 未开始 |
| 里程碑 3（P2） | 需求 7-12 | 长期架构演进与开发体验优化 | ⚪ 未开始 |

## 范围之外（Out of Scope）

- 数据库方言切换、多租户模型变更等非本评审涉及的事项
- 框架（GoFrame、Vben）本身的版本升级，仅在必要时同步进行
- 业务功能层面的新增需求（如新插件、新菜单等）

