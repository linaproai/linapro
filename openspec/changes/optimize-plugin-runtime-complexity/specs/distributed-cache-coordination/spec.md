## ADDED Requirements

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
