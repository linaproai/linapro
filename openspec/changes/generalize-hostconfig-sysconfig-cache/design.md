## 设计

### 决策 1：将 runtime-param snapshot 泛化为 sys_config 有效快照

现有`runtimeParamSnapshot`已经具备共享 revision、本地`gcache`、单机/集群分支和解析缓存能力。实现应复用这套机制，而不是新增平行缓存。

快照加载时不再仅查询`protectedConfigKeys`，而是按当前租户上下文查询`sys_config`可见行：

- 平台上下文只加载`tenant_id=0`。
- 租户上下文加载`tenant_id IN (0, currentTenantID)`。
- 同一 key 同时存在平台行和租户行时，租户行覆盖平台行。

快照中的通用`values`保存所有有效 key 的字符串值；现有运行时参数和公开前端参数继续在同一个快照内执行强类型预解析和默认值处理。

### 决策 2：源码插件读能力不做 key 白名单，动态插件继续做 key 白名单

源码插件随宿主源码编译交付，属于受信扩展。源码插件应通过`Services.HostConfig()`读取当前上下文有效`sys_config` key，不需要像动态插件一样在 manifest 中逐 key 声明。

动态插件仍属于运行时加载产物，必须保持最小授权面。动态插件`hostconfig.get`调度前继续使用`hostServices.resources.keys`校验目标 key；通过校验后才进入同一个`HostConfig()`读取实现。

### 决策 3：GetRaw 读取顺序

`GetRaw(ctx, key)`应采用以下顺序：

1. 标准化 key。
2. 查询当前上下文的`sys_config`有效快照；命中则返回快照值。
3. 对内置受管 key，若`sys_config`缺失则返回对应默认值，保持现有宿主运行时默认语义。
4. 未命中`sys_config`且不是内置受管 key 时，回退到`g.Cfg().Get(ctx, key)`读取静态配置。

`sys.log.retentionDays`继续保持“缺失 seed 行即错误”的语义，不使用合成默认值掩盖交付数据缺失。

### 决策 4：所有 sys_config 变更都应 bump runtime-config revision

既然快照覆盖所有有效`sys_config` key，创建、更新、导入和删除任何`sys_config`记录都可能影响`HostConfig()`读取结果。`sysconfig`服务写路径应在实际值或可见性变化后 bump runtime-config revision。

删除非内置记录也必须触发 revision bump，避免缓存继续返回已删除 key。

### 决策 5：接口性能和缓存一致性

`HostConfig()`读取是高频能力，不应每次查询数据库。快照加载按租户作用域一次性读取有效行，之后在本地`gcache`按 revision 和 tenant scope 命中。缓存 key 应包含租户 scope，避免平台上下文和租户上下文互相污染。

一致性模型：

- 权威数据源：`sys_config`。
- 本地缓存：进程内`gcache`快照。
- 单机模式：本地 revision bump 后立即清除当前进程快照。
- 集群模式：通过`sys_cache_revision`/`cachecoord`共享 revision，同步周期和请求路径 freshness 复用现有 runtime-config 策略。
- 最大陈旧时间：沿用 runtime-config 的 10 秒 watcher 同步预算；写入节点立即可见。
- 故障策略：revision 或快照加载失败时返回调用方错误，不静默使用可能过期的数据。

### 复杂度判断

该变更不新增抽象层，而是扩大现有配置快照的数据覆盖范围，并保留现有强类型解析逻辑。新增 tenant scope cache key 是为了满足租户隔离，不属于过度设计。

### 验证策略

- 单元测试覆盖源码插件/宿主`GetRaw()`可读取自定义`sys_config` key。
- 单元测试覆盖静态配置 fallback 不被破坏。
- 单元测试覆盖动态插件`hostconfig.get`仍需 manifest key 授权。
- 单元测试覆盖 revision bump 后自定义 key 更新和删除能刷新快照。
- Go 编译门禁覆盖`internal/service/config`、`internal/service/sysconfig`和动态插件 WASM host service 相关包。
- OpenSpec 运行`openspec validate generalize-hostconfig-sysconfig-cache --strict`。

