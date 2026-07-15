## Why

系统配置需要成为宿主运行时行为的受控入口，而不是仅提供可编辑键值记录。JWT 有效期、在线会话超时、上传大小、登录 IP 黑名单、登录页品牌和工作台主题都已经被宿主热路径消费，但这些参数曾分散在数据库种子、配置模板、静态兜底和业务代码中，导致默认值、校验、导入保护和读取成本不一致。

管理面体验与边界同样需要治理：

- 参数设置页列表已按请求语言本地化内置参数名称与描述，但编辑详情若回填库内中文 seed，英文环境会出现「列表英文、编辑中文」的断裂；若直接把投影写回库，又会污染权威存储与其他语言的 fallback。
- 全部参数值若统一用自由文本编辑，枚举与开关只能依赖备注人工拷贝，易误填；需要稳定的 `value_type` 与 `options` 元数据驱动组件与写路径校验，同时运行时仍以字符串 `value` 为业务真源。
- 插件业务参数经 `HostConfig.SysConfig().SetValue` 落库后若与宿主参数一并出现在系统参数页，会形成系统页与插件设置页双轨维护；持久化可继续统一使用 `sys_config`，但管理入口必须按 `system_manageable` 分流。

配置治理还牵引了邻域问题：公共前端配置必须在未登录阶段安全暴露，认证和在线会话必须消费有效配置，上传链路必须统一大小限制，插件与首页加载必须避免把配置读取变成重复 SQL。`config-management` 单元测试覆盖率也需要覆盖快照、修订号、公共前端配置和异常分支。

## What Changes

- 注册受保护宿主运行时参数和公共前端配置元数据，覆盖默认值、格式说明、校验规则、导入/更新/删除保护与运行时读取入口。
- 将 `sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize`、`sys.login.blackIPList` 接入认证、在线会话、上传校验和登录安全路径。
- 将登录页品牌、工作台主题、登录 slogan 插画（`sys.auth.sloganImage`，默认 `/slogan.svg`，空值隐藏）等公共前端设置通过白名单接口暴露，避免匿名读取任意 `sys_config` 键。
- 将 `sys.upload.maxSize` 的数据库种子值、配置模板默认值和后端静态兜底值统一为 20 MB。
- 使用本地不可变快照加共享修订号降低受保护配置热路径读取成本；单机仅本地失效，集群通过共享修订号最终收敛。
- 将 `sys_config` 读取升级为数据驱动的有效配置快照；`HostConfig.GetRaw` 统一为 `sys_config` 有效快照 → `config.yaml` → 系统默认值 → `nil`。
- 插件作用域配置优先读取主框架 `plugin.<plugin-id>` 静态段；源码插件与动态插件复用同一配置工厂。
- 编辑详情对 `name`/`remark` 按请求语言投影，`value` 始终返回库内原文；内置参数更新忽略 `name`/`remark` 写回；前端内置元数据只读。
- 为 `sys_config` 增加封闭集合 `value_type` 与 JSON `options`；CRUD/导入导出暴露类型元数据；管理面按类型渲染；写路径按类型校验；运行时仍只依赖字符串 `value`。
- 为 `sys_config` 增加 `system_manageable`；系统参数管理面仅面向标记为 1 的行；`SetValue`/`BatchSetValue` 通过 options 控制标记；插件闭环默认 0，宿主 seed 与管理面创建默认 1。不按 `plugin.*` 命名空间硬过滤管理面行集（曾评估并撤销“隐藏/锁定 plugin.*”方案）。
- 优化登录后首页的在线会话校验与插件 release 读取复用，减少重复 SQL。
- 为配置管理、运行时快照、类型校验、管理面分流与会话/release 复用补齐自动化测试与覆盖率门禁。

## Capabilities

### New Capabilities

- `login-home-sql-efficiency`
- `config-param-value-types`

### Modified Capabilities

- `config-management`

## Impact

- 影响宿主配置服务、认证、在线会话、上传链路、登录页和工作台启动配置读取路径。
- 影响 `sys_config` 表结构（`value_type`、`options`、`system_manageable`）、配置初始化 SQL、配置模板、上传大小静态兜底值、参数 CRUD/导入导出契约与管理面表单。
- 影响插件 `HostConfig.SysConfig().SetValue`/`BatchSetValue` 能力契约与插件 settings 写入路径。
- 交叉影响系统 API 文档、插件生命周期、插件 UI、启动 SQL、定时任务和 OpenSpec 语言治理；这些能力的当前契约由对应 owner 分组或 `openspec/specs` 承载。
