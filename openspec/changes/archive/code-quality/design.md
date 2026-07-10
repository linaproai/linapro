# Design

## API 合同和响应边界

API 治理从"接口能跑通"收敛为显式外部合同。读取、创建、更新、删除分别使用仓库约定的 REST 方法和资源化路径，路径参数统一使用`g.Meta`的`{param}`和输入 DTO 的`json`标签，避免`p`与`json`混用导致代码生成、文档和前端调用语义漂移。

响应边界禁止直接嵌入数据库实体，控制器必须逐字段映射独立响应 DTO，只暴露当前接口需要且经过权限、数据权限和业务边界评估的字段。用户密码、软删除时间、物理存储路径、哈希、内部租户治理字段、缓存版本等内部信息默认不进入 HTTP JSON 响应。公开瞬时时间字段统一使用 Unix 毫秒时间戳，内部 DAO、Entity、DO 和服务模型仍可使用 Go 时间类型；`birthday`等 date-only 字段保持日历日期字符串并在文档中说明语义。

跨模块稳定值不再各模块重复定义，也不引入大一统`pkg/enums`。排序方向、租户覆盖模式、通用状态标志等按小型公共契约组件维护，已有`menutype`、`pluginbridge`等稳定组件直接复用，领域私有枚举继续留在所属包内。

## GoFrame 后端一致性和源码可读性

后端实现统一遵循 GoFrame v2 分层和 ORM 约定。包含`deleted_at`的表依赖 GoFrame 自动软删除过滤，生产代码不手写`WhereNull(deleted_at)`；写入和更新使用 DO 对象传递数据，不手工维护框架自动处理的时间字段。字典类型和字典数据表补齐软删除字段，避免规范与数据库结构不一致。

事务一致性从"失败写 warning 后继续"调整为"关联清理失败即回滚整笔操作"。用户删除、角色删除、菜单删除和角色授权用户等路径将主记录变更与关联清理放入单事务，拓扑通知只在事务提交后触发。菜单子树判断改为一次加载 parent-child 映射并在内存中遍历，避免按层级循环 SQL。

生产`panic`采用 allowlist，而不是简单全禁。启动不可恢复错误、源码插件静态注册契约失败、`Must*`构造、未知 panic 重抛和 GoFrame 内部行为属于允许边界；普通请求、导入导出、资源关闭、动态插件输入、运行时配置读取和可恢复错误必须显式返回`error`。初始化、注册、路由、Cron、中间件和 host service 配置 API 遇到可预期失败时返回`error`给调用方决策。

源码可读性治理把主文件定义为契约入口。宿主`internal/service`、源码插件`backend/internal/service`和`lina-core/pkg`公共组件的主文件只保留包说明、核心类型、接口、实现结构、编译期断言和构造函数；复杂实现迁移到职责文件。接口方法必须有紧邻注释，说明输入、输出、错误和权限、数据权限、缓存、i18n、事务、幂等、并发等适用约束。接口方法集合不得重复或近义并存；兼容期入口必须说明权威方法、过渡原因和清理条件。

## 显式依赖注入和共享实例

运行期依赖治理选择"类型签名可见"的显式构造方式，不引入通用 DI 容器、全局 service locator 或新的宿主私有组装层。启动期已有编排、路由绑定和插件 registrar 是构造边界；Controller、Middleware、Service、源码插件、插件宿主服务适配器和 WASM host service 通过构造函数逐项接收接口型依赖。

`Dependencies`、`Deps`、`Options`等聚合结构体不得整体承载多个接口型运行期依赖。这个限制让新增、删除或替换依赖可以通过 Go 编译错误暴露所有未同步调用点，避免依赖变化被宽结构体隐藏。

缓存敏感组件必须共享启动期实例或共享后端。认证、session、角色、插件、配置、i18n、cachecoord、kvcache、locker、notify、host service adapter 和插件 runtime cache 只要持有缓存、派生状态、订阅状态、token/session 状态、插件启用快照、运行时配置快照或跨实例协调依赖，就不得在业务构造函数、请求路径、插件回调或 host service 调用中再创建孤立服务图。源码插件通过 registrar 获取宿主发布的服务目录，动态插件 host service handler 只作为 transport 适配层进入同一宿主能力语义。

## 宿主运行能力和数据一致性

宿主运行质量基线包括匿名`GET /api/v1/health`、GoFrame`Server.Run()`优雅停机复用、受保护的上传访问路径和可配置调度默认时区。健康检查公开在未鉴权路由组，通过轻量数据库探针返回`ok`或脱敏不可用原因；HTTP graceful shutdown 交给 GoFrame，返回后按 cron、cluster、数据库连接池顺序清理宿主资源并受`shutdown.timeout`约束。

上传访问从`cmd_http.go`中的本地静态路径处理收口到文件 API 和文件 controller，挂载在受认证和权限中间件保护的路由组下，读取时通过文件服务和存储后端，不拼接本地物理路径。调度默认时区从硬编码`Asia/Shanghai`改为`scheduler.defaultTimezone`，缺省为`UTC`。

SQL 结构按常用访问路径补齐索引：用户状态、手机号、创建时间，用户角色的`role_id`，角色菜单的`menu_id`，在线会话的`last_active_time`和定时任务的`group_id`。`sys_job`去除数据库外键，改由服务层校验分组存在性，保持与仓库其他关联表的应用层一致性策略一致。

## 启动 SQL 效率

启动效率治理把 SQL 明细日志从默认输出改为诊断能力，`database.default.debug`默认关闭。正常启动输出结构化阶段摘要，包含插件扫描数量、同步变更数量、无差异插件数量、startup snapshot 构造数量、内置任务投影数量和各阶段耗时，不输出完整 SQL 文本。

HTTP 启动编排中创建一次共享`StartupContext`，携带 catalog、integration 和 job startup snapshots 以及统计收集器。插件自动启用、HTTP 路由注册、动态前端预热、cron 内置任务同步等阶段复用同一组快照；当注册表、release、菜单、路由权限和资源引用都无差异时，不进入事务、不写数据库，也不为了刷新刚写入的数据反复读取全表。

写入路径同步更新快照。插入使用`InsertAndGetId`构造最新实体并更新快照；更新使用`existing + data`合成最新投影；只有数据库默认值、自动时间或触发器结果无法确定时才执行 post-write read。内置任务执行定义以源码声明为权威，同步后用返回的投影注册 scheduler，持久化任务扫描排除`is_builtin=1`。

## 前端运行性能和 i18n 状态刷新

语言切换只刷新与语言强相关的本地状态，包括公共配置同步和字典缓存重置，不再触发完整权限、菜单和路由重拉。菜单和面包屑通过初次菜单响应中的`meta.i18nKey`或本地`$t(...)`响应式更新，路由生成阶段不得把当前语言文本固化成静态字符串。

服务器监控和用户消息未读轮询使用页面可见性感知：页签隐藏时暂停请求，恢复可见时立即刷新一次并继续周期轮询，组件卸载或退出登录时停止 timer。路由守卫中的`loadedPaths`收敛为有界 LRU，避免长时间会话中无界增长。

## 测试效率和审查治理

Go 单元测试效率治理不以移除`-race`作为默认优化。主单测路径继续覆盖插件运行时、缓存、协调、session、锁、认证、租户和权限等并发敏感代码；耗时优化来自减少重复重型 fixture、真实 dynamic Wasm 样例执行和无测试包完整执行。

普通插件 runtime、catalog、integration 和 lifecycle 测试优先使用 synthetic artifact、fake executor 和轻量 host service 替身；真实 bundled dynamic Wasm 样例收敛为少量 smoke。`linactl test.go`在执行前生成测试计划，区分包含`_test.go`的包、只需编译 smoke 的包和无需执行单测的包，并输出 module 耗时、待测包数量、无测试包数量、race 状态和 smoke 边界。

`lina-review`纳入 API DTO 响应实体暴露、时间字段、后端主文件职责、接口注释、接口方法唯一性、显式依赖注入、缓存敏感实例来源、初始化错误返回、测试效率和归档治理影响判断。历史任务通过 Go 单元测试、E2E、静态扫描、OpenSpec 校验、`make db.init`、`make dao`、`make ctrl`、前端类型和审查闭环验证。

## E2E 质量审查治理

E2E 质量审查采用结果级约束，不绑定具体测试实现方式。审查判断测试是否能够稳定证明本次业务行为、数据状态、权限结果或用户可观察状态变化正确，而不强制要求使用特定定位器类型、封装模式或编码风格。

质量审查按验收风险分级：默认警告；当问题导致本次变更缺少有效验收证据、功能行为类 bugfix 缺少复现验证、或测试不可独立可信运行时，升级为严重问题并阻塞任务完成、反馈完成或归档。审查报告必须说明触发原因、覆盖判断依据、发现问题的严重性和建议补充的业务场景。

触发范围覆盖 E2E 文件变更、用户可观察行为变更、端到端工作流变更和功能行为类反馈修复。断言有效性审查关注断言是否能证明业务结果而非仅验证页面元素存在；稳定性审查关注测试是否依赖其他测试用例的执行结果、残留数据或执行顺序。

## OpenSpec 归档文档治理

归档文档治理将`openspec/specs`确认为当前能力契约唯一事实来源，`openspec/changes/archive`定位为历史摘要、设计演进、反馈闭环和验证证据承载位置。每个能力最多由一个归档分组作为历史 owner，非 owner 分组中的重复能力全文规范迁移为交叉影响摘要，摘要指向当前契约位置和历史 owner。

压缩按样板分组先行验证：先选择`plugin-framework`作为高体量样板，完整处理`proposal.md`、`design.md`、`tasks.md`和`specs/`，验证通过后再批量处理`user-auth`、`distributed-infra`、`devops-tooling`、`code-quality`等分组和低体量分组。文件级压缩边界固定：`proposal.md`只保留背景目标范围影响，`design.md`保留架构决策和交叉影响摘要，`tasks.md`只保留`FB-*`、根因、修复、验证和治理影响最小维护摘要，`specs/`仅保留 owner 能力历史契约。

压缩结果量化：归档体量从约`3.5M`降至约`1.5M`，归档`spec.md`从`277`降至`116`，跨分组重复能力从`61`降至`0`，与主规范完全相同的归档副本从`16`降至`0`。验证以 OpenSpec 严格校验、重复能力扫描、Markdown 格式检查和语义覆盖审查结合。

## Official Plugin Panic And Error Copy Decoupling

Production panic governance keeps precise Path/Function/Count allowlists for `apps/lina-core/**`. Official plugin packages no longer require host tests to enumerate every `apps/lina-plugins/<id>` path: `backend/plugin.go` package `init` fail-fast `panic(err)` patterns are auto-classified as `plugin-registration` when the official workspace is ready. Literal panics, non-init production panics, and non-registration fail-fast patterns still fail.

Capability-constraint error copy uses capability semantics rather than official plugin brand IDs. `PLUGIN_TENANT_PROVISIONING_POLICY_INVALID` keeps a stable error code and message key while English source text and `manifest/i18n/**/error.json` describe multi-tenant / framework tenant governance. Callers must rely on `errorCode` / `messageKey`, not natural-language brand names. Startup provider-consistency rewrite and broader plugin-id fixture fictionalization stay out of scope.

## Cross-Domain Impacts

- `distributed-cache-coordination`为显式依赖注入和缓存敏感实例共享提供一致性背景；当前契约由`openspec/specs/distributed-cache-coordination/spec.md`承载，历史 owner 为`archive/distributed-infra`。
- `plugin-host-service-extension`、`plugin-http-slot-extension`、`plugin-manifest-lifecycle`和`plugin-startup-bootstrap`受显式依赖、共享 host service、差异同步和启动快照复用影响；当前契约由`openspec/specs`对应能力承载，历史 owner 为`archive/plugin-framework`。
- `cron-job-management`只保留调度默认时区、`sys_job`外键移除和应用层一致性的质量影响；当前契约由`openspec/specs/cron-job-management/spec.md`承载，历史 owner 为`archive/scheduled-jobs`。
- `user-management`、`role-management`、`user-role-association`和`menu-management`承载批量删除、事务回滚、关联清理、索引和菜单子树遍历的业务契约；当前契约由`openspec/specs`对应能力承载，历史 owner 为`archive/user-management`。
- `online-user`、`server-monitor`和`user-message`分别受在线会话索引、服务器监控可见性感知轮询和消息未读轮询影响；当前契约由系统治理或通知相关主规范承载，历史 owner 分别为`archive/system-governance`和`archive/notification`。
- `module-decoupling`影响模块禁用时的后端降级边界；当前契约由`openspec/specs/module-decoupling/spec.md`承载，历史 owner 为`archive/host-plugin-boundary`。
- `spec-governance`影响 OpenSpec 主规范结构、归档残留、规则加载矩阵和`i18n`影响记录；当前契约由`openspec/specs/spec-governance/spec.md`承载，历史 owner 为`archive/devops-tooling`。
- `e2e-suite-organization`受 E2E 质量审查结果级要求增强影响；当前契约由`openspec/specs/e2e-suite-organization/spec.md`承载，历史 owner 为`archive/e2e-testing`。

## Governance Notes

本分组历史上触及后端、API、SQL、前端、插件宿主服务、开发工具、测试、OpenSpec 审查规则、归档文档治理和 E2E 质量审查；这些运行时实现已由对应历史任务验证。本次归档压缩只改 OpenSpec 历史文档，不新增运行时行为。当前审查应关注 owner 规范是否保留、非 owner 语义是否进入交叉影响摘要、`i18n`影响是否仅限文档，以及 OpenSpec 校验和 Markdown 空白检查是否通过。
