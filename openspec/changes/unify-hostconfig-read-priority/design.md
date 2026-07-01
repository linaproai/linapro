## Context

`apps/lina-core/internal/service/config`已经通过`runtime-config`共享 revision、本地`gcache`和租户作用域缓存键，为`sys_config`提供数据驱动的有效快照。`HostConfig.GetRaw()`当前仍包含两类特殊读取逻辑：对`sys.log.retentionDays`直接走必需 seed 行校验，对`IsManagedSysConfigKey()`命中的 key 在读取`config.yaml`之前返回内置默认值。

该行为让`HostConfig`的通用读取顺序和配置中心语义不一致：`sys_config`确实优先，但系统默认值会抢在静态配置之前生效，并且新增默认值可能继续把具体 key 写进通用读取路径。本变更属于`lina-core`宿主通用能力和插件 HostConfig 能力，不属于工作台展示适配，也不修改业务插件目录。

## Goals / Non-Goals

**Goals:**

- 将非 root key 的`HostConfig.GetRaw(ctx, key)`统一为`sys_config`有效快照、`config.yaml`、系统默认值、`nil`的读取顺序。
- 从`GetRaw()`读取流程中移除具体配置键判断和`IsManagedSysConfigKey()`分支。
- 将系统已有硬编码默认值收敛到可按 key 查询的默认值元数据或等价 resolver，读取流程只依赖通用查询。
- 保持`sys_config`快照的租户覆盖、共享 revision、本地缓存和故障可见错误策略。
- 保持源码插件 HostConfig 无白名单、动态插件 HostConfig 有 manifest 授权快照的既有安全边界。

**Non-Goals:**

- 不新增配置写入、保存、热重载或运行时修改`config.yaml`的能力。
- 不把插件自身业务配置纳入`HostConfig`读取链路，插件自身配置仍通过`Services.Plugins().Config()`和`plugins.config.get`读取。
- 不新增 HTTP API、DTO、SQL 迁移、数据表字段或动态 host service 方法。
- 不改变空 key 和`.`的 root snapshot 语义；它们不是普通配置键，继续按宿主配置组件语义返回完整静态配置快照。

## Decisions

### 决策 1：将`GetRaw()`实现为固定来源 pipeline

非 root key 的读取流程固定为：

1. 标准化 key。
2. 查询当前上下文可见的`sys_config`有效快照；命中即返回该原始值。
3. 查询 GoFrame 当前静态配置源中的同名 key；命中即返回静态配置值。
4. 查询宿主系统默认值元数据；命中即返回默认值。
5. 以上均未命中时返回`nil`。

备选方案是在现有`IsManagedSysConfigKey()`分支中调整顺序。该方案仍把“哪些 key 有默认值”的判断留在读取路径里，无法满足通用性要求，因此不采用。

### 决策 2：默认值通过通用元数据查询，不进入读取分支

实现应提供配置包内部的默认值查询入口，例如`lookupHostConfigDefaultValue(key)`或等价 resolver。该入口可以聚合已有默认值来源，包括运行时参数默认值、公开前端设置默认值，以及已经存在于静态配置 getter 中的宿主默认值。`GetRaw()`只调用这个通用入口，不直接引用具体 key 常量。

不采用运行时反射扫描结构体默认值。当前配置默认值包含路径归一化、时长解析、字段校验和专用读取语义，反射会隐藏 key 与默认值的 owner，增加调试成本。显式默认值元数据更直接，也便于测试覆盖所有公开默认 key。

### 决策 3：专用 getter 复用相同来源顺序但保留类型校验

`GetJwtExpire()`、`GetSessionTimeout()`、`GetUploadMaxSize()`、`GetCronLogRetention()`等专用 getter 应继续负责类型解析、归一化和业务校验，但它们的来源优先级需要和`GetRaw()`保持一致：`sys_config`覆盖静态配置，静态配置覆盖系统默认值。

`sys.log.retentionDays`不应在`GetRaw()`中保留缺失即错误的特殊分支。若缺少`sys_config`和`config.yaml`值，通用读取返回系统默认值；专用 getter 在取得值后继续执行正整数校验。这样可以保留运行时安全校验，同时避免通用 HostConfig 读取路径硬编码该 key。

### 决策 4：缓存与租户边界沿用 runtime-config 快照

`sys_config`仍是运行时配置的权威数据源。读取`sys_config`阶段继续复用当前 runtime-config 共享 revision、本地`gcache`快照、租户作用域 cache key 和租户行覆盖平台行的规则。

静态配置和系统默认值不引入新的运行时缓存失效机制。静态配置仍以进程当前 GoFrame 配置源为准，系统默认值随代码部署生效。`sys_config`快照 freshness 无法确认时继续向调用方返回可见错误，不静默回退到静态配置或默认值掩盖一致性故障。

### 决策 5：插件授权边界不随优先级变化

源码插件属于受信扩展，继续通过`Services.HostConfig()`读取宿主配置，不需要逐 key manifest 授权。动态插件仍必须在`hostconfig.get`进入读取 pipeline 前完成`hostServices.resources.keys`授权校验。授权通过后，动态插件读取到的值和源码插件使用同一优先级链路。

## Risks / Trade-offs

- 默认值元数据遗漏某个既有硬编码默认值 -> 通过单元测试覆盖关键默认 key，并用静态检索核对`default*`常量和`DefaultValue`规格是否进入默认值查询入口。
- GoFrame 静态配置缺失判断不准确 -> 单元测试覆盖静态配置存在、静态配置缺失、静态配置空值和默认值回退；空字符串按“已命中值”处理，typed helper 自行决定空白默认值语义。
- `sys_config` freshness 异常被 fallback 掩盖 -> 读取 snapshot 返回错误时立即向上返回错误，不进入静态配置或默认值 fallback。
- 动态插件误以为默认值绕过授权 -> WASM host service 测试继续覆盖未授权 key 被拒绝，授权 key 才能进入统一读取链路。
- 多个活跃 OpenSpec 变更同时涉及`HostConfig` -> 本变更需要在实现前确认`generalize-hostconfig-sysconfig-cache`和`prioritize-host-plugin-config`的未归档内容仍然作为上下文考虑，避免回退既有能力。

## Migration Plan

1. 调整配置服务内部读取 pipeline 和默认值元数据，不改外部接口签名。
2. 更新`HostConfig`相关单元测试，覆盖`sys_config`、静态配置、系统默认值和`nil`四级优先级。
3. 更新专用 getter 相关测试，确认静态配置优先于系统默认值，`sys_config`仍优先于静态配置。
4. 更新动态插件 HostConfig 回归测试，确认授权边界不变。
5. 运行变更范围 Go 测试、静态检索和`openspec validate unify-hostconfig-read-priority --strict`。

回滚方式：恢复`GetRaw()`原读取顺序和默认值分支；不涉及数据迁移或外部 API 兼容处理。

## Open Questions

无。
