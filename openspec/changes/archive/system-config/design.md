# Design

## Runtime Configuration Contract

宿主配置服务集中注册受保护运行时参数：`sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize`、`sys.login.blackIPList`，以及登录页和工作台启动需要的公共前端设置。注册表同时定义默认值、格式说明、校验规则、运行时覆盖读取方式，以及重命名、删除、导入和更新保护，避免认证、会话、上传和前端启动逻辑各自维护隐式约定。

公共前端配置只通过白名单接口暴露，返回结构化且可审计的键集合。登录页和工作台可以在未登录阶段读取品牌、文案和主题默认值，但匿名调用方不能枚举或读取任意`sys_config`记录。

## Upload Size Source Of Truth

`sys.upload.maxSize`是上传大小限制的唯一业务真源。数据库种子值、`config.template.yaml`默认值与后端静态兜底值统一为 20 MB；上传请求体保护、文件校验和用户可见错误提示都使用同一有效值。这样可以消除不同启动路径下 10 MB、16 MB、20 MB 并存的默认行为分裂。

## Snapshot And Revision Strategy

宿主热路径不在每次请求中查询`sys_config`。每个进程维护解析后的本地不可变快照；写入路径立即清理本地快照并递增共享修订号；其他节点通过周期同步感知修订变化后重建快照。单机部署只依赖本地失效和重建，集群部署通过共享修订号保证跨实例最终收敛。

这个模型只服务受保护运行时参数读取，不把普通配置管理列表、搜索、导入或审计路径改造成长期缓存。公共前端设置也必须遵循显式失效边界，避免登录页和工作台启动读取陈旧配置。

## sys_config 有效快照泛化

**决策**：将`config.Service.GetRaw()`对`sys_config`的读取从硬编码 key 白名单升级为数据驱动的有效配置快照读取。快照加载时按当前租户上下文查询`sys_config`可见行：平台上下文只加载`tenant_id=0`，租户上下文加载`tenant_id IN (0, currentTenantID)`，同一 key 同时存在平台行和租户行时租户行覆盖平台行。

**源码插件与动态插件边界**：源码插件属于受信扩展，通过`Services.HostConfig()`读取当前上下文有效`sys_config` key，不需要逐 key manifest 授权。动态插件仍属于运行时加载产物，必须保持最小授权面，`hostconfig.get`调度前继续使用`hostServices.resources.keys`校验目标 key。

## 插件业务配置优先级

**决策**：将主框架静态配置文件中的`plugin.<plugin-id>`作为插件业务配置的最高优先级来源。只要`plugin.<plugin-id>`配置段存在，系统就使用该配置段作为该插件的有效配置源；该段内缺失的 key 返回缺失或调用方默认值，不再继续读取`plugins/<plugin-id>/config.yaml`补齐单个 key。

**配置工厂共享**：源码插件和动态插件复用同一个`ConfigServiceFactory`，确保`Services.Plugins().Config()`和动态`plugins.config.get`语义一致。`plugin.<plugin-id>`只通过`Plugins().Config()`生效，不等同于允许插件通过`HostConfig()`读取任意宿主配置。

## HostConfig 读取优先级统一

**决策**：将非 root key 的`HostConfig.GetRaw(ctx, key)`统一为固定来源 pipeline：`sys_config`有效快照 → `config.yaml` → 系统默认值 → `nil`。从`GetRaw()`读取流程中移除具体配置键判断和`IsManagedSysConfigKey()`分支。

**默认值元数据**：系统已有硬编码默认值收敛到可按 key 查询的默认值元数据或等价 resolver，`GetRaw()`只调用这个通用入口，不直接引用具体 key 常量。专用 getter（`GetJwtExpire()`、`GetSessionTimeout()`等）继续负责类型解析、归一化和业务校验，但来源优先级与`GetRaw()`一致。

## Login Home SQL Efficiency

DB-only 在线会话校验改为读取一条`sys_online_session`记录后完成租户、超时和活跃时间判断。会话不存在、租户不匹配或已超时时拒绝请求；过期记录仍被清理；只有超过写入节流窗口才更新`last_active_time`。这保留强制下线、JWT revoke、租户隔离和会话超时语义，同时减少每个有效鉴权请求的数据库往返。

插件 release 读取只在请求级或列表级作用域复用。同一首页请求、插件管理列表或动态运行态列表中，对同一 release ID 或同一`plugin_id + release_version`的等价查询必须命中同一快照；请求结束后释放。跨请求或跨实例缓存不属于本轮设计，若未来引入必须接入明确的修订号或失效机制。

## Test And Delivery Gates

`config-management`以包级覆盖率不低于 80% 作为交付门槛，重点覆盖插件配置路径、公共前端辅助方法、运行时修订号控制器、快照缓存、默认值、无效值和共享 KV 异常分支。登录后首页 SQL 优化必须用单元测试覆盖有效会话、过期会话、租户不匹配、近期活跃不重复写入、release 按 ID 复用、按版本复用和状态刷新后读取更新。

启动和首页 SQL 优化不使用框架元数据探测 SQL 的精确条数作为稳定门禁。门禁只约束项目自身可控行为：默认不输出 SQL 明细、无差异插件同步不写库不启事务、启动快照构造次数在预算内、请求级 release 复用不跨作用域泄漏。

## Cross-Domain Impacts

- `user-auth`消费`sys.jwt.expire`和`sys.login.blackIPList`，当前契约由`openspec/specs/user-auth/spec.md`承载，历史 owner 为`archive/user-auth`。
- `online-user`消费`sys.session.timeout`并在认证链路清理过期在线会话，当前契约由`openspec/specs/online-user/spec.md`承载，历史 owner 为`archive/system-governance`。
- `startup-sql-efficiency`承载启动阶段 SQL 明细日志、`StartupContext`、插件 no-op 同步和内置任务投影复用，当前契约由`openspec/specs/startup-sql-efficiency/spec.md`承载，历史 owner 为`archive/code-quality`。
- `plugin-manifest-lifecycle`、`plugin-runtime-loading`、`plugin-startup-bootstrap`和`plugin-ui-integration`承载插件治理、路由元数据、自动启用、授权展示和动态示例分页，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/plugin-framework`。
- `cron-job-management`承载内置任务投影、手动触发、Shell 任务、数据权限、租户分组、集群 watcher 和调度时区，当前契约由`openspec/specs/cron-job-management/spec.md`承载，历史 owner 为`archive/scheduled-jobs`。
- `system-api-docs`承载宿主管理的`/api.json`、启用插件路由投影和内部路由排除，当前契约由`openspec/specs/system-api-docs/spec.md`承载，历史 owner 为`archive/system-governance`。
- `spec-governance`承载 OpenSpec 文档语言和归档语言策略，当前契约由`openspec/specs/spec-governance/spec.md`承载，历史 owner 为`archive/devops-tooling`。

## Risks And Boundaries

- 配置快照和请求级 release 快照都不是长期业务缓存；跨请求缓存必须另行设计失效机制。
- 公共前端配置白名单必须保持最小集合，不能成为匿名读取配置中心的旁路。
- 运行时配置影响认证、会话和上传热路径，错误值必须在写入或导入时被拒绝，而不是延迟到请求处理时暴露。
- 本分组不保留插件、定时任务、API 文档、启动 SQL 或认证会话的完整规范全文；这些能力的当前契约以`openspec/specs`和对应历史 owner 分组为准。
