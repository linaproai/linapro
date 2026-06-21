## 背景

当前`Services.HostConfig()`通过`config.Service.GetRaw()`读取宿主配置。`GetRaw()`只对`IsManagedSysConfigKey()`硬编码白名单命中的`sys_config`记录读取运行时值，其他 key 仍回退到静态`config.yaml`。这会导致后续通过`sys_config`新增的运行时配置无法被源码插件通过稳定宿主能力读取，除非同步修改 Go 常量和白名单。

该设计不符合配置中心应由数据驱动扩展的方向，也让缓存一致性边界被硬编码 key 列表限制。需要将`sys_config`读取升级为基于共享 revision 的有效配置快照：源码插件可通过稳定`HostConfig()`读取`sys_config`中的有效 key；动态插件仍必须通过`hostconfig.get`的`hostServices.resources.keys`白名单授权读取。

## 目标

- 将`config.Service.GetRaw()`对`sys_config`的读取从硬编码 key 白名单升级为数据驱动的有效配置快照读取。
- 复用现有 runtime-config 共享 revision 与本地`gcache`缓存机制，保证`sys_config`变更后缓存可失效、可刷新。
- 保持源码插件与宿主同信任边界：源码插件通过稳定`Services.HostConfig()`可读取当前上下文可见的`sys_config`有效 key，不需要逐 key manifest 授权。
- 保持动态插件安全边界：动态插件仍只能读取`plugin.yaml`/manifest 中声明并经宿主确认的`hostconfig.get` key。
- 保留宿主内置运行时参数的强类型解析、默认值、校验和错误语义。
- 保持租户上下文语义：租户上下文优先读取租户覆盖值，否则回退平台值；平台上下文只读取平台值。

## 非目标

- 不将动态插件扩展为可读任意`sys_config` key。
- 不允许插件直接访问`dao.SysConfig`、`entity.SysConfig`或`internal/service/sysconfig`内部实现。
- 不新增前端页面、HTTP API 或工作台交互。
- 不改变插件业务配置`Services.Plugins().Config()`的来源优先级。
- 不引入新的外部分布式缓存后端。

## 影响分析

- 架构：属于宿主通用配置能力和插件 host capability 边界调整，不是工作台适配。
- 插件：源码插件与动态插件读取边界需要在规范和测试中区分。
- 缓存一致性：命中。权威数据源为`sys_config`，一致性依赖 runtime-config 共享 revision、本地`gcache`快照和写入后的 revision bump。
- 数据权限：命中。读取必须遵守当前租户上下文和平台 fallback 语义，动态插件还必须遵守 manifest key 授权。
- 数据库：不计划新增表字段或 SQL；复用现有`sys_config`表和`sys_cache_revision`机制。
- i18n：无新增用户可见文案、API 文档源文本或语言包资源。
- 开发工具跨平台：无脚本、Makefile、CI 或工具入口影响。

