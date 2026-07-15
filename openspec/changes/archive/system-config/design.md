# Design

## Runtime Configuration Contract

宿主配置服务集中注册受保护运行时参数：`sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize`、`sys.login.blackIPList`，以及登录页和工作台启动需要的公共前端设置。注册表同时定义默认值、格式说明、校验规则、运行时覆盖读取方式，以及重命名、删除、导入和更新保护，避免认证、会话、上传和前端启动逻辑各自维护隐式约定。

公共前端配置只通过白名单接口暴露，返回结构化且可审计的键集合。登录页和工作台可以在未登录阶段读取品牌、文案和主题默认值，但匿名调用方不能枚举或读取任意 `sys_config` 记录。

## Upload Size Source Of Truth

`sys.upload.maxSize` 是上传大小限制的唯一业务真源。数据库种子值、`config.template.yaml` 默认值与后端静态兜底值统一为 20 MB；上传请求体保护、文件校验和用户可见错误提示都使用同一有效值。这样可以消除不同启动路径下 10 MB、16 MB、20 MB 并存的默认行为分裂。

## Snapshot And Revision Strategy

宿主热路径不在每次请求中查询 `sys_config`。每个进程维护解析后的本地不可变快照；写入路径立即清理本地快照并递增共享修订号；其他节点通过周期同步感知修订变化后重建快照。单机部署只依赖本地失效和重建，集群部署通过共享修订号保证跨实例最终收敛。

这个模型只服务受保护运行时参数读取，不把普通配置管理列表、搜索、导入或审计路径改造成长期缓存。公共前端设置也必须遵循显式失效边界，避免登录页和工作台启动读取陈旧配置。`value_type`/`options`/`system_manageable` 属于管理面与写路径元数据，不进入 runtime snapshot 热路径。

## sys_config 有效快照泛化

**决策**：将 `config.Service.GetRaw()` 对 `sys_config` 的读取从硬编码 key 白名单升级为数据驱动的有效配置快照读取。快照加载时按当前租户上下文查询 `sys_config` 可见行：平台上下文只加载 `tenant_id=0`，租户上下文加载 `tenant_id IN (0, currentTenantID)`，同一 key 同时存在平台行和租户行时租户行覆盖平台行。

**源码插件与动态插件边界**：源码插件属于受信扩展，通过 `Services.HostConfig()` 读取当前上下文有效 `sys_config` key，不需要逐 key manifest 授权。动态插件仍属于运行时加载产物，必须保持最小授权面，`hostconfig.get` 调度前继续使用 `hostServices.resources.keys` 校验目标 key。

## 插件业务配置优先级

**决策**：将主框架静态配置文件中的 `plugin.<plugin-id>` 作为插件业务配置的最高优先级来源。只要 `plugin.<plugin-id>` 配置段存在，系统就使用该配置段作为该插件的有效配置源；该段内缺失的 key 返回缺失或调用方默认值，不再继续读取 `plugins/<plugin-id>/config.yaml` 补齐单个 key。

**配置工厂共享**：源码插件和动态插件复用同一个 `ConfigServiceFactory`，确保 `Services.Plugins().Config()` 和动态 `plugins.config.get` 语义一致。`plugin.<plugin-id>` 只通过 `Plugins().Config()` 生效，不等同于允许插件通过 `HostConfig()` 读取任意宿主配置。

## HostConfig 读取优先级统一

**决策**：将非 root key 的 `HostConfig.GetRaw(ctx, key)` 统一为固定来源 pipeline：`sys_config` 有效快照 → `config.yaml` → 系统默认值 → `nil`。从 `GetRaw()` 读取流程中移除具体配置键判断和 `IsManagedSysConfigKey()` 分支。

**默认值元数据**：系统已有硬编码默认值收敛到可按 key 查询的默认值元数据或等价 resolver，`GetRaw()` 只调用这个通用入口，不直接引用具体 key 常量。专用 getter（`GetJwtExpire()`、`GetSessionTimeout()` 等）继续负责类型解析、归一化和业务校验，但来源优先级与 `GetRaw()` 一致。

## 编辑详情元数据本地化

列表与按 key 查询已对内置参数 `name`/`remark` 使用 `config.<key>.name` / `config.<key>.remark` 投影。编辑详情若故意回填库内中文 seed，会造成英文环境列表与弹窗不一致；若把投影写回库，又会污染权威存储。

**决策**：

1. 对外 `GetById` 对 `name`/`remark` 做与列表相同的语言投影；**不**对 `value` 做公共前端默认值投影——`value` 始终是库内可编辑真源。
2. 内置（`is_builtin` 或受管系统键）记录的 `Update` **忽略**请求中的 `name`/`remark`，即使请求体携带投影后的文案；非内置自定义参数仍可写 `name`/`remark`。
3. 前端编辑弹窗在内置参数上将 `name`/`remark` 设为只读，主要可编辑字段为参数值（及类型化后的值组件）。
4. mutation 路径使用 **raw** 按 ID 加载（`getByIdRaw`），对外 `GetById` 在 raw 之上再 localize，避免未来详情投影扩展时污染比较与写回。

导出行数据继续使用库内原文策略，不为 name/remark 引入独立多语言持久化列。

## 参数值类型与选项

管理面需要按类型渲染输入组件，并在创建/更新/导入时按类型校验 `value`，但不能改变宿主与插件运行时以字符串 `value` 为权威业务值的契约。

**决策**：

| 主题 | 选择 |
|------|------|
| 类型集合 | 封闭枚举：`text`、`textarea`、`number`、`boolean`、`select`、`radio`、`multi_select`、`richtext`；空/未知归一为 `text` |
| 选项存储 | `sys_config.options` 为 TEXT，权威格式为 JSON 数组 `[{label,value}]`；管理面与 Excel 用户编辑可接受简单行格式（`标签=值` / `标签|值` / 仅值），写入前统一编码为 JSON |
| 业务值 | 所有类型最终序列化为既有字符串 `value`；`multi_select` 用英文分号 `;` 连接，校验时顺序不敏感 |
| 校验分层 | 通用类型校验 + 托管键既有校验（duration、上传上限等）叠加；`boolean` 仅 `true`/`false` |
| 内置锁定 | 内置记录允许改 `value`，**禁止**管理面改 `value_type`/`options`；非内置可完整编辑 |
| 租户覆盖 | 创建 fallback 覆盖时从平台行复制 `value_type`/`options`，仅写租户侧 `value` |
| 前端 | 按 `valueType` 动态组件；枚举类型可编辑 options；`richtext` 使用宿主 Tiptap；弹窗密度按类型分为 compact / spacious |
| SQL | 在 `005-config-management.sql` 建表即含 `value_type`（默认 `text`）与 `options`（默认空串）；内置 seed（含 `011` 定时任务相关键）直接写入匹配类型元数据；不保留独立增量 SQL |

插件 `SetValue` 新建行缺省 `value_type=text`、`options` 空，不强制插件声明类型。不在本能力中引入字典类型外链作为选项来源。

## 管理面 system_manageable 分流

`sys_config` 同时承载宿主内置参数与插件业务参数。管理面曾全量展示，导致系统页与插件设置页双轨维护。

**决策**：

1. 字段 `system_manageable SMALLINT NOT NULL DEFAULT 1`：`1` 允许在系统参数页维护，`0` 不允许。
2. 管理面 List/Export 仅返回 `system_manageable=1` 的可见行；Get/Update/Delete/Import 对 `=0` 拒绝或视为不存在；Create 固定写入 `1`。
3. 能力 API：

```go
type SetSysConfigValueOptions struct {
    SystemManageable *bool // nil insert→0; nil update→keep; non-nil→write
}
type SetSysConfigValueItem struct { Key SysConfigKey; Value string }

SetValue(ctx, key, value, options *SetSysConfigValueOptions) error
BatchSetValue(ctx, items []SetSysConfigValueItem, options *SetSysConfigValueOptions) error
```

- `SetValue` 委托 `BatchSetValue`（单键批）。
- `BatchSetValue`：一事务写全部 items，成功后仅一次 revision；空 items 成功无副作用。
- 多字段插件 settings MUST 用 `BatchSetValue`，不得循环 `SetValue`。
- `options == nil` 等价于未指定 `SystemManageable`；插件入口闭环 MUST 显式传 `false`。
- 宿主 seed 默认 `1`；插件首次插入且未指定时默认 `0`（安全默认）。

不按 `plugin.*` 命名空间硬过滤，也不在本能力中引入参数 category 分组。运行时读路径不受 `system_manageable` 影响。

**演进说明**：曾短暂要求系统参数管理面隐藏/锁定 `plugin.*` 并附加密钥投影掩码与 isBuiltin 强制投影；确认该需求有误后已完整撤销 List/Export 过滤、写拒绝、错误码与相关 i18n。最终边界以 `system_manageable` 分流为准：插件键在 `system_manageable=1` 时仍可出现在系统参数页，在 `=0` 时仅经插件设置页与 `SetValue` 维护。

## 登录 slogan 插画参数

`sys.auth.sloganImage` 为受保护公共前端参数：`value_type=text`，默认 `/slogan.svg`，允许空串；非空最长 500 字符。读取路径必须保留库内空值，不得因空串回退默认。公开投影字段 `auth.sloganImage` 按 URL/路径原样下发，不做文案类 i18n 投影。种子写入 `005-config-management.sql`，静态资源导出为宿主 `public/slogan.svg`。登录页消费语义见登录展示 owner 规范。

## Login Home SQL Efficiency

DB-only 在线会话校验改为读取一条 `sys_online_session` 记录后完成租户、超时和活跃时间判断。会话不存在、租户不匹配或已超时时拒绝请求；过期记录仍被清理；只有超过写入节流窗口才更新 `last_active_time`。这保留强制下线、JWT revoke、租户隔离和会话超时语义，同时减少每个有效鉴权请求的数据库往返。

插件 release 读取只在请求级或列表级作用域复用。同一首页请求、插件管理列表或动态运行态列表中，对同一 release ID 或同一 `plugin_id + release_version` 的等价查询必须命中同一快照；请求结束后释放。跨请求或跨实例缓存不属于本轮设计，若未来引入必须接入明确的修订号或失效机制。

## Test And Delivery Gates

`config-management` 以包级覆盖率不低于 80% 作为交付门槛，重点覆盖插件配置路径、公共前端辅助方法、运行时修订号控制器、快照缓存、默认值、无效值、类型校验、管理面分流和共享 KV 异常分支。登录后首页 SQL 优化必须用单元测试覆盖有效会话、过期会话、租户不匹配、近期活跃不重复写入、release 按 ID 复用、按版本复用和状态刷新后读取更新。

启动和首页 SQL 优化不使用框架元数据探测 SQL 的精确条数作为稳定门禁。门禁只约束项目自身可控行为：默认不输出 SQL 明细、无差异插件同步不写库不启事务、启动快照构造次数在预算内、请求级 release 复用不跨作用域泄漏。

## Cross-Domain Impacts

- `user-auth` 消费 `sys.jwt.expire` 和 `sys.login.blackIPList`，当前契约由 `openspec/specs/user-auth/spec.md` 承载，历史 owner 为 `archive/user-auth`。
- `online-user` 消费 `sys.session.timeout` 并在认证链路清理过期在线会话，当前契约由 `openspec/specs/online-user/spec.md` 承载，历史 owner 为 `archive/system-governance`。
- `startup-sql-efficiency` 承载启动阶段 SQL 明细日志、`StartupContext`、插件 no-op 同步和内置任务投影复用，当前契约由 `openspec/specs/startup-sql-efficiency/spec.md` 承载，历史 owner 为 `archive/code-quality`。
- `plugin-manifest-lifecycle`、`plugin-runtime-loading`、`plugin-startup-bootstrap` 和 `plugin-ui-integration` 承载插件治理、路由元数据、自动启用、授权展示和动态示例分页，当前契约由 `openspec/specs/<capability>/spec.md` 承载，历史 owner 为 `archive/plugin-framework`。
- `cron-job-management` 承载内置任务投影、手动触发、Shell 任务、数据权限、租户分组、集群 watcher 和调度时区，当前契约由 `openspec/specs/cron-job-management/spec.md` 承载，历史 owner 为 `archive/scheduled-jobs`。
- `system-api-docs` 承载宿主管理的 `/api.json`、启用插件路由投影和内部路由排除，当前契约由 `openspec/specs/system-api-docs/spec.md` 承载，历史 owner 为 `archive/system-governance`。
- `spec-governance` 承载 OpenSpec 文档语言和归档语言策略，当前契约由 `openspec/specs/spec-governance/spec.md` 承载，历史 owner 为 `archive/devops-tooling`。

## Risks And Boundaries

- 配置快照和请求级 release 快照都不是长期业务缓存；跨请求缓存必须另行设计失效机制。
- 公共前端配置白名单必须保持最小集合，不能成为匿名读取配置中心的旁路。
- 运行时配置影响认证、会话和上传热路径，错误值必须在写入或导入时被拒绝，而不是延迟到请求处理时暴露。
- 内置参数管理员不能通过管理面改显示名/备注或类型选项；元数据真源为 i18n 与 SQL 种子，避免热修污染。
- 旧插件未传 `SystemManageable` 时首次插入仍为 `0`（安全默认）。
- 本分组不保留插件、定时任务、API 文档、启动 SQL 或认证会话的完整规范全文；这些能力的当前契约以 `openspec/specs` 和对应历史 owner 分组为准。
