## Why

当前`HostConfig`通用读取链路仍在`GetRaw()`中判断特殊配置键，并在读取`config.yaml`之前对内置受管 key 返回硬编码默认值。这会让运行时配置读取优先级不一致，也让后续新增系统配置默认值必须修改通用读取逻辑。

需要将`HostConfig`读取统一为数据驱动的优先级链路：先读取当前上下文可见的`sys_config`内存快照，再读取`config.yaml`，最后才读取系统已有默认值；任何来源都没有命中时返回`nil`。

## What Changes

- 调整`HostConfig.GetRaw(ctx, key)`的通用读取顺序为`sys_config`有效快照、`config.yaml`、系统默认值、`nil`。
- 移除`GetRaw()`中的特殊配置键分支，禁止通过`IsManagedSysConfigKey()`或具体 key 常量决定读取顺序。
- 将系统已有硬编码默认值收敛为可按 key 查询的通用默认值元数据或等价 resolver，由读取链路统一调用。
- 保持源码插件不需要逐 key 授权读取`HostConfig`，动态插件`hostconfig.get`仍必须先通过 manifest 授权快照校验目标 key。
- 保持现有`sys_config`快照缓存、租户覆盖和 runtime-config revision 机制，不新增并行缓存。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `config-management`：修改宿主配置服务读取`sys_config`、静态配置和系统默认值的优先级与缺失语义。
- `plugin-config-service`：修改`HostServices.HostConfig()`暴露给源码插件和动态插件的宿主配置读取语义，保持动态插件授权边界不变。

## Impact

- 后端：影响`apps/lina-core/internal/service/config`中的`GetRaw()`、默认值元数据、专用配置 getter 与相关单元测试。
- 插件宿主能力：影响`apps/lina-core/internal/service/plugin/internal/hostconfig`适配层及动态 WASM `hostconfig.get`的回归测试语义，但不改变 host service 协议。
- 缓存一致性：继续复用`runtime-config`共享 revision、本地`gcache`快照和租户作用域缓存键；不新增缓存层。
- 数据权限与租户边界：继续使用当前上下文可见的`sys_config`有效快照，租户上下文仍优先读取租户行并回退平台行。
- `i18n`：本变更不新增用户可见文案、菜单、API 文档源文本或翻译资源；实施时需要在任务记录中明确无`i18n`资源影响。
- 数据库与 API：不新增表结构、SQL 迁移、HTTP API、DTO 或动态插件 host service 方法。
