# 分布式缓存协调

## Purpose

定义单节点和集群部署的拓扑感知缓存协调，包括持久化共享修订号、范围失效、域策略回退、可观测性和关键缓存路径的故障处理。
## Requirements
### Requirement:宿主必须提供拓扑感知缓存协调

系统 SHALL 为发布显式范围修订号、同步进程内派生缓存以及根据 `cluster.enabled` 区分单节点和集群策略提供统一的缓存协调。

#### Scenario:单节点模式使用本地协调

- **当** `cluster.enabled=false` 且业务写路径发布缓存变更时
- **则** 系统仅更新当前进程中的本地修订号
- **且** 当前进程中对应的缓存域立即失效或刷新
- **且** 系统不得依赖共享修订号表或分布式协调组件

#### Scenario:集群模式使用共享修订号

- **当** `cluster.enabled=true` 且业务写路径发布缓存变更时
- **则** 系统持久化递增对应缓存域和范围的共享修订号
- **且** 所有节点在请求路径或监听路径上观察到新修订号后刷新本地缓存
- **且** 修订号发布必须是幂等的、可重试的和可观测的

### Requirement:共享缓存修订号必须持久化且原子递增

系统 SHALL 将关键缓存域修订号存储在持久化共享存储中，并确保同一缓存域和范围的并发递增不会丢失。

#### Scenario:同一范围的并发修订号发布

- **当** 多个节点并发发布同一缓存域和范围的变更时
- **则** 系统为每次成功发布生成单调递增的修订号
- **且** 最终共享修订号至少增加成功发布的次数
- **且** 任何节点不得通过读-修改-写竞争覆盖其他节点的递增

#### Scenario:数据库重启后修订号仍然可用

- **当** 共享数据库重启并恢复时
- **则** 已提交的缓存修订号仍然存在
- **且** 新启动的节点可使用持久化修订号判断本地缓存是否必须刷新

#### Scenario:有损缓存不得承载关键修订号

- **当** 系统为权限、运行时配置或插件运行时等关键缓存域发布修订号时
- **则** 系统写入持久化修订号表
- **且** 不得将关键修订号存储在 `sys_kv_cache` 或任何其他有损缓存中

### Requirement:缓存域策略配置不得阻塞使用

系统 SHALL 允许调用方直接为任何合法的缓存域字符串发布和读取修订号。不得要求域在参与协调前进行预先注册。关键缓存域 SHALL 在其所属实现代码中声明权威数据源、一致性模型、失效触发点、最大可容忍陈旧时间、跨实例同步机制和故障回退策略。未配置的域 SHALL 使用组件默认策略。

#### Scenario:使用未配置策略的域

- **当** 宿主模块或插件逻辑为新的合法缓存域字符串发布修订号时
- **则** 系统接受该域并使用默认一致性和故障策略
- **且** 调用方无需修改 `cachecoord` 组件源码或交付清单来添加该域

#### Scenario:配置关键缓存域策略

- **当** 宿主关键缓存域需要不同于默认的陈旧窗口或回退行为时
- **则** 该域的实现代码配置权威源和最大可容忍陈旧时间
- **且** 该域的实现代码配置刷新失败回退行为
- **且** 审查可使用该配置判断域是否满足集群一致性要求

#### Scenario:关键缓存超过陈旧窗口

- **当** 集群模式下的节点无法读取共享修订号且本地缓存超过域的最大陈旧窗口时
- **则** 系统按该域的故障策略处理请求
- **且** 权限缓存在超过故障窗口后不得静默允许请求

### Requirement:关键写路径必须可靠发布失效

权限、配置、插件运行时稳定状态和等效域的关键写路径 SHALL 在业务数据变更成功后可靠发布对应的缓存域修订号。如果发布失败，调用方不得收到静默成功。

#### Scenario:在事务内发布缓存修订号

- **当** 业务数据变更和缓存修订号发布可使用同一数据库事务时
- **则** 系统在同一事务中提交业务数据和修订号递增
- **且** 不存在业务数据提交成功但修订号缺失的状态

#### Scenario:发布失败返回错误

- **当** 关键写路径完成业务数据变更但缓存修订号发布失败时
- **则** 系统返回结构化业务错误
- **且** 系统记录可观测日志
- **且** 调用方可重试操作或触发修复流程

### Requirement:缓存协调状态必须可观测

系统 SHALL 暴露缓存协调状态，至少包含缓存域、范围、本地修订号、共享修订号、上次同步时间、最新错误和陈旧秒数。

#### Scenario:查询缓存协调状态

- **当** 运维工具或健康检查查询缓存协调状态时
- **则** 系统返回已配置或已触及缓存域的同步状态
- **且** 集群模式可识别节点是否落后于共享修订号

#### Scenario:缓存同步失败可诊断

- **当** 节点刷新缓存域失败时
- **则** 系统记录最新失败原因和时间
- **且** 后续状态查询可将该域显示为异常或陈旧

### Requirement: 集群模式缓存协调必须使用 Redis revision
系统 SHALL 在 `cluster.enabled=true` 时通过 Redis revision store 协调关键缓存域修订号。系统 MUST 不依赖 `sys_cache_revision` 表完成跨节点一致性。

#### Scenario: 集群模式发布缓存 revision
- **WHEN** 集群模式下业务写路径发布 `runtime-config` 变更
- **THEN** 系统递增 Redis revision key
- **AND** 返回新的 shared revision
- **AND** 不通过 `sys_cache_revision` 行锁递增作为主实现

#### Scenario: 单机模式本地 revision
- **WHEN** `cluster.enabled=false` 且业务写路径发布缓存变更
- **THEN** 系统仅更新进程内 revision
- **AND** 不连接 Redis

### Requirement: 集群模式缓存失效必须发布 Redis event
系统 SHALL 在集群模式下为缓存 revision 变更发布跨节点 event。event MUST 携带 tenant ID、cascade 标记、domain、scope、revision、reason、source node 和创建时间。

#### Scenario: 权限拓扑失效事件
- **WHEN** 管理员修改角色权限
- **THEN** 系统递增 `permission-access` revision
- **AND** 发布 `cache.invalidate` event
- **AND** 其他节点收到事件后清理本地 token access snapshot

#### Scenario: 重复事件幂等
- **WHEN** 节点收到相同 event 两次
- **THEN** 节点最多执行一次有效刷新
- **AND** 最终本地 observed revision 与 shared revision 收敛

### Requirement: revision 读取必须兜底 Pub/Sub 丢失
系统 SHALL 保留请求路径或 watcher 的 revision check。即使 Redis event 没有被某个节点收到，该节点也 MUST 能通过读取 Redis revision 判断本地缓存是否陈旧。

#### Scenario: 节点错过失效事件
- **WHEN** 节点 B 在节点 A 发布失效事件时短暂离线
- **AND** 节点 B 恢复后处理请求
- **THEN** 节点 B 读取 Redis revision
- **AND** 发现本地 observed revision 落后
- **AND** 刷新对应本地缓存后继续处理请求

### Requirement: 租户范围必须在 Redis revision 中显式表达
系统 SHALL 在 Redis revision key 和 event payload 中显式表达 tenant scope。单租户失效、平台默认级联失效和全租户运维失效 MUST 可区分。

#### Scenario: 单租户失效
- **WHEN** 租户 A 修改租户级字典覆盖
- **THEN** Redis revision key 包含租户 A ID
- **AND** event payload `tenantId=A`
- **AND** 其他租户缓存不被失效

#### Scenario: 平台默认级联失效
- **WHEN** 平台管理员修改平台默认配置
- **THEN** event payload 包含 `tenantId=0`
- **AND** event payload 包含 `cascadeToTenants=true`
- **AND** 各节点按平台 fallback 语义清理受影响租户视图

### Requirement: 关键缓存故障必须遵循域策略
系统 SHALL 在 Redis revision 不可读或刷新失败时按缓存域策略处理。权限缓存 MUST fail-closed；插件运行时缓存 MUST conservative-hide；运行时配置 MUST 返回可见错误或等价结构化错误。

#### Scenario: 权限 revision 超过陈旧窗口
- **WHEN** 节点无法读取 `permission-access` Redis revision
- **AND** 本地 observed revision 已超过最大陈旧窗口
- **THEN** 受保护 API 权限校验失败
- **AND** 系统不得静默放行请求

#### Scenario: 插件 runtime revision 不可确认
- **WHEN** 节点无法确认 `plugin-runtime` revision freshness
- **THEN** 动态插件能力按 conservative-hide 处理
- **AND** 不得暴露可能已禁用或已卸载的插件入口

### Requirement: 缓存敏感服务实例必须由拓扑感知构造边界共享
系统 SHALL 确保依赖缓存协调、共享修订号、事件订阅、分布式 KV、分布式锁、会话热状态、token 状态或本地派生缓存的服务实例由拓扑感知构造边界创建并共享。`cluster.enabled=true` 时，相关组件 MUST 使用同一 coordination-backed 后端或同一运行期服务实例，MUST 不创建仅当前节点可见的孤立默认实例。

#### Scenario: 集群模式认证中间件使用共享 token 和 session 状态
- **WHEN** `cluster.enabled=true` 且认证中间件校验请求
- **THEN** 认证中间件使用启动期注入的 auth service 和 session store
- **AND** revoked token、pre-token 和 session hot state 读取使用同一 coordination-backed 事实源
- **AND** 认证中间件不得自行构造仅当前节点可见的 auth/session 服务图

#### Scenario: 插件运行时缓存使用共享插件治理实例
- **WHEN** 插件管理、动态路由、源码插件 route registrar 或插件运行时缓存读取插件启用状态
- **THEN** 这些路径复用同一插件治理实例或同一共享 revision/event 后端
- **AND** 不得因为多个插件服务实例持有不同 enabled snapshot 而暴露已禁用或已卸载插件

#### Scenario: 运行时配置和 i18n 缓存使用共享失效路径
- **WHEN** 运行时配置或 i18n 资源在集群模式下变更
- **THEN** 使用注入的共享 cachecoord/coordination 依赖发布范围化失效
- **AND** 消费方复用同一运行期配置或 i18n 服务实例观察刷新
- **AND** 不得通过局部新实例绕过已配置的失效策略

### Requirement: 缓存一致性审查必须覆盖实例来源
系统 SHALL 在缓存一致性审查中检查缓存敏感服务的实例来源、共享边界和故障策略。审查 MUST 识别新增或修改的 `New()` 调用是否会创建独立缓存状态或独立订阅状态。

#### Scenario: 审查新增缓存服务构造
- **WHEN** 变更新增或修改认证、权限、插件、配置、i18n、session、cachecoord、kvcache、lock 或 notify 相关服务构造
- **THEN** 审查确认该构造来自启动期或 registrar 显式传入的共享依赖
- **AND** 审查标记会导致本地状态分裂的隐式构造

#### Scenario: 审查无缓存影响变更
- **WHEN** 变更确认不涉及缓存、派生状态、失效或跨实例同步
- **THEN** 审查结论明确记录无缓存一致性影响
- **AND** 不得省略该判断

### Requirement:插件管理读模型缓存必须复用 plugin-runtime 协调

系统 SHALL 将插件管理摘要列表读模型和详情读模型视为 `plugin-runtime`派生缓存。缓存 MUST 绑定插件运行时修订号、locale 和运行时翻译包版本，并复用既有单机本地 revision 或集群 Redis revision/event 完成失效；系统 MUST NOT 为插件管理读模型创建仅当前节点可见的独立缓存协调域。

#### Scenario:单节点模式本地失效插件管理读模型
- **WHEN** `cluster.enabled=false` 且插件安装、启用、禁用、卸载、升级、active release 切换或源码插件同步成功
- **THEN** 系统更新本地 `plugin-runtime` revision
- **AND** 当前进程内插件管理摘要列表缓存和对应插件详情缓存失效
- **AND** 下一次插件管理请求基于新的权威状态重建读模型

#### Scenario:集群模式通过 Redis event 失效插件管理读模型
- **WHEN** `cluster.enabled=true` 且某节点发布 `plugin-runtime` Redis revision/event
- **THEN** 其他节点观察到 revision 前进后失效本地插件管理摘要列表缓存和受影响插件详情缓存
- **AND** 后续插件管理请求不得继续返回旧 revision 下的插件安装、启用、版本或授权摘要状态

#### Scenario:语言资源变化区分缓存键
- **WHEN** 当前用户 locale 或 runtime bundle version 与已缓存插件管理读模型不同
- **THEN** 系统使用独立缓存键读取或构建摘要列表和详情读模型
- **AND** 系统不得把旧语言或旧运行时翻译包版本下的插件展示元数据返回给当前请求

#### Scenario:无法确认 freshness 时不返回过期治理状态
- **WHEN** 节点无法确认 `plugin-runtime` revision freshness
- **AND** 本地插件管理读模型超过域策略允许的陈旧窗口
- **THEN** 系统不得返回过期摘要列表或详情
- **AND** 系统按插件运行时故障策略 conservative-hide 或结构化错误处理该请求

### Requirement: 插件缓存敏感依赖不得使用孤立默认服务图

系统 SHALL 确保插件管理、动态插件 runtime、WASM host service、source plugin registrar 和插件能力适配器中的缓存敏感依赖都来自拓扑感知启动构造边界或测试 fixture 显式注入。生产路径 MUST NOT 通过包级变量、构造函数 fallback、setter 默认值或普通业务调用临时创建仅当前节点可见的 cache、config、session、lock、notify、plugin runtime 或 capability service 实例。

#### Scenario: WASM cache host service 不创建本地默认 cache

- **WHEN** 集群模式下动态插件通过 `cache` host service 读写插件缓存
- **THEN** dispatcher 使用启动期注入的共享 `kvcache.Service` 或共享后端
- **AND** 不存在包级默认 `kvcache.New()` 实例参与生产调用

#### Scenario: runtime session store 来自启动期共享实例

- **WHEN** 动态插件路由鉴权需要校验在线 session hot state
- **THEN** runtime 使用启动期注入的 session store 或同一 coordination-backed 事实源
- **AND** 不因 runtime 内部默认 `session.NewDBStore()` fallback 绕过已配置的集群 session 热状态

#### Scenario: 缺失共享依赖时 fail fast

- **WHEN** 插件 runtime 或 WASM host service 缺失缓存敏感共享依赖
- **THEN** 构造、启动配置或首次 host call 返回明确错误
- **AND** 系统不得静默退化为仅当前节点可见的默认实例

### Requirement: 插件复杂度治理审查必须记录实例来源和扫描成本

系统 SHALL 在插件 runtime、host service、pluginbridge 或插件能力适配器变更审查中记录运行期依赖实例来源、共享边界、缓存一致性影响和高频路径扫描成本。涉及插件列表、runtime state、manifest discovery、artifact parsing 或 host service 调用路径时，审查 MUST 说明访问次数如何随插件数、artifact 数、registry 行数或请求数增长。

#### Scenario: 审查 runtime state 列表变更

- **WHEN** 变更修改 runtime state 列表、插件列表、manifest discovery 或 artifact parsing 路径
- **THEN** 审查结论记录扫描次数、artifact parse 次数和批量读取策略
- **AND** 拒绝无固定上限的循环 `ScanManifests` 或逐插件 artifact 重复解析

#### Scenario: 审查 host service 依赖变更

- **WHEN** 变更新增或修改 WASM host service 运行期依赖
- **THEN** 审查结论追溯依赖 owner、创建位置、传递路径和共享实例策略
- **AND** 标记会创建孤立缓存状态、配置状态、session 状态或锁状态的隐式 `New()` 调用

### Requirement: runtime revision controller 必须属于缓存协调边界

系统 SHALL 将运行时缓存 revision controller 作为缓存协调组件能力维护，而不是作为插件领域包的非 internal 子包暴露。插件 runtime、插件管理读模型、runtime reconciler 和 i18n runtime bundle 等消费方 MUST 从缓存协调边界导入 revision controller，并按各自 domain/scope 创建实例。

#### Scenario: plugin 和 i18n 使用 revision controller

- **WHEN** plugin runtime 缓存和 i18n runtime bundle 缓存需要 observed revision
- **THEN** 二者从缓存协调组件导入 revision controller
- **AND** 不导入`internal/service/plugin/runtimecache`
- **AND** 各自实例化独立 domain/scope controller

#### Scenario: revision controller tenant scope

- **WHEN** 调用方需要设置 tenant scope
- **THEN** controller API 使用返回副本的`WithTenantScope`或语义明确的构造期 setter
- **AND** 不得在共享 controller 已被多个调用方使用后通过误导性 fluent API 原地修改作用域

#### Scenario: 审查插件 runtimecache 旧路径

- **WHEN** 生产或测试代码重新导入`internal/service/plugin/runtimecache`
- **THEN** 静态治理测试失败
- **AND** 调用方必须改用缓存协调边界下的 revision controller

### Requirement: 插件生命周期缓存失效必须通过单一变化发布入口

系统 SHALL 为插件同步、动态包上传、安装、卸载、启用、禁用、源码升级、动态升级、租户供应策略更新和启动自动启用等插件治理变化提供单一插件变化发布入口。该入口 MUST 复用`plugin-runtime`revision controller，统一失效 runtime 派生缓存、插件管理读模型、frontend bundle、i18n runtime bundle 和 WASM 相关派生状态，不得创建额外仅当前节点可见的缓存域或分散发布路径。

#### Scenario: 生命周期写入后发布变化

- **WHEN** 插件安装、卸载、启用、禁用或状态变更成功写入治理状态
- **THEN** lifecycle 编排调用统一插件变化发布入口
- **AND** 入口发布`plugin-runtime`revision 并记录 reason
- **AND** 插件管理读模型和 runtime 派生缓存均观察同一 revision 失效

#### Scenario: 租户供应策略变化后发布变化

- **WHEN** 平台管理员更新插件新租户供应策略
- **THEN** 系统通过统一插件变化发布入口失效受影响的插件管理和运行时派生缓存
- **AND** 不绕过`plugin-runtime`revision controller 创建独立本地失效路径

#### Scenario: 审查缓存失效入口

- **WHEN** 变更新增插件治理写路径或迁移生命周期编排
- **THEN** 静态治理或审查确认该路径最终调用统一插件变化发布入口
- **AND** 对无缓存影响路径必须记录无影响判断

### Requirement: kvcache 拓扑选择必须由构造边界显式表达

系统 SHALL 在拓扑感知构造边界显式表达 `kvcache` 后端选择。`cluster.enabled=false` 时构造 SQL table 后端；`cluster.enabled=true` 时构造 coordination KV 后端并复用宿主统一 coordination provider。生产代码不得通过进程级默认 provider 修改或读取来表达当前拓扑。

#### Scenario: 集群模式后端选择可追溯到 coordination provider

- **WHEN** 审查集群模式 HTTP runtime 的 `kvcache` 构造路径
- **THEN** 可以从共享 `coordination.Service` 追溯到 `kvcache` provider 和共享 `kvcache.Service`
- **AND** 该路径不创建独立 Redis client 或仅当前节点可见的本地默认实例

#### Scenario: 单机模式不触碰 coordination 后端

- **WHEN** `cluster.enabled=false` 且宿主创建共享 `kvcache.Service`
- **THEN** 构造路径选择 SQL table provider
- **AND** 不初始化或要求 coordination KV backend

