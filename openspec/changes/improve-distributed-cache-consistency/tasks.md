## 1. 数据模型与协调基础

- [x] 1.1 调整宿主交付 SQL，新增持久化 `sys_cache_revision` 表，并保留 `sys_kv_cache` 的 `MEMORY` 可丢失缓存语义
- [x] 1.2 为 `sys_cache_revision` 补齐幂等索引、唯一键和并发递增所需约束，复核 `sys_kv_cache` 现有 TTL 查询索引与唯一键
- [x] 1.3 执行 `make init` 和 `make dao`，仅通过代码生成更新 DAO/DO/Entity 工件
- [x] 1.4 复核 `cluster.enabled` 与 `cluster.Service` 当前拓扑抽象，确定 `cachecoord` 的单节点和集群模式分支接入点

## 2. 统一缓存协调组件

- [x] 2.1 新增 `internal/service/cachecoord` 组件，定义缓存域、作用域、权威数据源、一致性模型、陈旧窗口和故障策略
- [x] 2.2 实现 `ConfigureDomain`、`MarkChanged`、`EnsureFresh` 和 `Snapshot` 等核心接口
- [x] 2.3 在 `cluster.enabled=false` 时实现进程内 revision、本地失效和同步刷新，且不依赖共享协调表
- [x] 2.4 在 `cluster.enabled=true` 时实现共享修订号原子递增、幂等发布、请求路径 freshness check 和后台 watcher 同步
- [x] 2.5 为缓存协调失败定义 `bizerr` 错误码，并使用项目 `logger` 组件记录带 `ctx` 的可观测日志

## 3. 关键缓存域接入

- [x] 3.1 将受保护运行时参数缓存接入 `cachecoord`，在参数写路径事务后可靠发布 `runtime-config` 修订号
- [x] 3.2 在运行时参数读取路径执行 freshness check，超过故障窗口时按配置缓存域策略返回可见错误
- [x] 3.3 将角色、菜单、用户角色和插件权限拓扑写路径接入 `permission-access` 修订号发布
- [x] 3.4 在受保护 API 权限校验前执行权限快照 freshness check，无法确认且超过故障窗口时失败关闭
- [x] 3.5 将插件安装、启用、禁用、卸载、升级和同版本刷新接入 `plugin-runtime` 修订号发布

## 4. 插件运行时派生缓存

- [x] 4.1 重构 `pluginruntimecache` 的共享 revision 逻辑，复用 `cachecoord` 统一刷新插件启用快照
- [x] 4.2 将插件前端包缓存、运行时 i18n 包缓存和动态路由派生缓存纳入 `plugin-runtime` 作用域失效
- [x] 4.3 将 Wasm 编译缓存 key 绑定 active release 的 checksum 或 generation，避免同版本刷新后继续命中旧缓存
- [x] 4.4 调整动态插件 artifact 归档或缓存键策略，保证 active release 能被其他节点用 checksum 或 generation 校验
- [x] 4.5 增加旧 artifact 与旧 Wasm 编译缓存清理逻辑，确保不删除当前 active release 引用的内容

## 5. 插件 Host-Cache 可靠性

- [x] 5.1 从宿主关键修订号链路中移除 `sys_kv_cache`，确保它只承载可丢失插件/模块 KV 缓存
- [x] 5.2 使用单 SQL 原子更新或等价机制实现 `incr`，确保共享数据库存活期间多节点并发递增不丢失增量
- [x] 5.3 对非整数递增、超长命名空间、超长键和值写入返回结构化错误，禁止截断或部分写入
- [x] 5.4 将过期清理改为单 key 懒清理和后台批量清理，避免普通读写路径全表扫描
- [x] 5.5 在集群模式下通过主节点协调或幂等批处理限制过期批量清理的重复压力，并将数据库重启后的缓存丢失按缓存未命中处理

## 6. 观测、测试与验收

- [x] 6.1 暴露缓存协调状态快照，至少包含缓存域、作用域、本地修订号、共享修订号、最后同步时间、最近错误和陈旧秒数
- [x] 6.2 若新增 HTTP 诊断接口或 API 文档字段，同步维护 apidoc i18n 资源；若只接入健康检查或日志，记录本次无运行时 UI i18n 影响
- [x] 6.3 增加 `sys_cache_revision` 并发发布测试，验证修订号持久、原子递增且数据库重启后不丢失
- [x] 6.4 增加双实例服务级测试，覆盖运行时参数、权限拓扑和插件运行时缓存的跨节点失效与有界陈旧
- [x] 6.5 增加插件 host-cache 并发 `incr`、TTL、数据库重启后缓存未命中、超长输入和非整数递增测试
- [x] 6.6 增加动态插件同版本刷新测试，验证 checksum 或 generation 变化后旧 Wasm、前端包和 i18n 派生缓存失效
- [x] 6.7 执行后端单元测试、必要的服务级测试和 `openspec status --change improve-distributed-cache-consistency`，确认变更进入可应用状态

主任务阶段 i18n 影响判断：新增 `/system/info` 响应中的缓存协调诊断字段，并已同步维护 `zh-CN` 与 `zh-TW` 的 apidoc i18n JSON；未新增前端页面、按钮、菜单或运行时 UI 文案。反馈阶段的运行时错误文案影响单独记录在 FB-2。

## Feedback

- [x] **FB-1**: 清理关键缓存修订号路径中遗留的 `kvcache` fallback，确保 `role`、`config` 与 `pluginruntimecache` 只通过 `cachecoord` 协调分布式修订号
- [x] **FB-2**: 移除运行时参数快照读取中的 `recover` 兜底，确保 runtime-config freshness 与加载失败通过显式错误传播
- [x] **FB-3**: 收敛 `cachecoord` 多实例创建路径，确保进程内关键缓存协调状态由统一协调器管理
- [x] **FB-4**: 去除 `cachecoord` 对内置 domain 清单和事前注册的依赖，允许宿主模块与插件逻辑直接使用新的合法缓存域

FB-2 i18n 影响判断：本次不新增或调整 API 文档、前端 UI 文案、菜单、按钮或插件清单资源；由于新增 caller-visible `bizerr` 运行时错误码，已同步维护 `en-US`、`zh-CN`、`zh-TW` 的 `error.json` 与 packed manifest 资源。

FB-3 i18n 影响判断：本次只调整后端缓存协调对象的创建方式与进程内拓扑复用策略，不新增或修改 API 文档、前端 UI 文案、菜单、按钮、插件清单或运行时翻译资源。

FB-3 缓存一致性判断：真实缓存数据仍由各业务域的进程级缓存保存，权威数据源不变；`cachecoord.Default` 仅统一关键缓存域的修订号观察状态、拓扑视图和诊断状态。单节点模式继续使用本地 revision 和本地失效；集群模式继续通过 `sys_cache_revision` 共享修订号、请求路径 freshness check 与 watcher 同步实现跨实例失效。最大陈旧窗口和故障降级策略沿用已配置的 `runtime-config`、`permission-access`、`plugin-runtime` 域策略；新增默认协调器避免同一进程内不同服务实例重复维护 freshness 状态，同时保留 `cachecoord.New` 供测试或显式隔离场景使用。

FB-4 i18n 影响判断：本次只调整后端缓存协调组件的 domain 准入和策略配置方式，不新增或修改 API 文档、前端 UI 文案、菜单、按钮、插件清单、运行时错误码或翻译资源；无需维护运行时语言包、manifest i18n 或 apidoc i18n。

FB-4 缓存一致性判断：`cachecoord` 不再把内置 domain 清单或事前注册作为使用准入，任意合法 domain/scope 仍在单节点模式下使用进程内 revision、本地失效和同步刷新，在集群模式下继续使用持久 `sys_cache_revision` 共享修订号、请求路径 freshness check 与 watcher 同步实现跨实例失效。未显式配置策略的 domain 使用默认权威说明、`shared-revision` 一致性模型、5 秒最大陈旧窗口和可见错误降级；`runtime-config`、`permission-access`、`plugin-runtime` 的权威数据源、最大陈旧窗口和故障降级策略改由各自业务实现代码配置，避免插件或后续模块新增 domain 时修改 `cachecoord` 或 manifest，同时保留关键域的可观测和可审查策略。
