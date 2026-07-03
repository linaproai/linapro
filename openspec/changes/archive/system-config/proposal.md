## Why

系统配置需要成为宿主运行时行为的受控入口，而不是仅提供可编辑键值记录。JWT 有效期、在线会话超时、上传大小、登录 IP 黑名单、登录页品牌和工作台主题都已经被宿主热路径消费，但这些参数曾分散在数据库种子、配置模板、静态兜底和业务代码中，导致默认值、校验、导入保护和读取成本不一致。

配置治理还牵引了若干邻域问题：公共前端配置必须在未登录阶段安全暴露，认证和在线会话必须在每个受保护请求中消费有效配置，上传链路必须统一大小限制，插件、API 文档、启动编排和首页加载必须避免把配置读取或治理投影变成重复 SQL。`config-management`单元测试覆盖率也需要覆盖快照、修订号、公共前端配置和异常分支，防止配置成为运行期隐患。

## What Changes

- 注册受保护宿主运行时参数和公共前端配置元数据，覆盖默认值、格式说明、校验规则、导入保护、更新保护、删除保护和运行时读取入口。
- 将`sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize`、`sys.login.blackIPList`接入认证、在线会话、上传校验和登录安全路径。
- 将登录页品牌、工作台主题等公共前端设置通过白名单接口暴露，避免匿名读取任意`sys_config`键。
- 将`sys.upload.maxSize`的数据库种子值、配置模板默认值和后端静态兜底值统一为 20 MB，并同步上传校验和友好错误提示。
- 使用本地不可变快照加共享修订号降低受保护配置热路径读取成本，单机模式只本地失效，集群模式通过共享修订号最终收敛。
- 将`sys_config`读取从硬编码 key 白名单升级为数据驱动的有效配置快照读取，源码插件可通过稳定`HostConfig()`读取`sys_config`中的有效 key。
- 调整插件作用域配置读取优先级，优先读取主框架静态配置文件中的`plugin.<plugin-id>`配置段，当该段不存在时继续按现有顺序读取。
- 将`HostConfig.GetRaw(ctx, key)`统一为`sys_config`有效快照、`config.yaml`、系统默认值、`nil`的读取顺序，移除具体配置键判断和`IsManagedSysConfigKey()`分支。
- 优化登录后首页的在线会话校验与插件 release 读取复用，减少同一请求或同一列表投影内的重复 SQL。
- 为`config-management`、运行时配置快照、公共前端配置、会话校验和插件 release 复用补齐自动化测试与覆盖率门禁。

## Capabilities

### New Capabilities

- `login-home-sql-efficiency`

### Modified Capabilities

- `config-management`

## Impact

- 影响宿主配置服务、认证、在线会话、上传链路、登录页和工作台启动配置读取路径。
- 影响配置初始化 SQL、配置模板、上传大小静态兜底值和相关单元测试。
- 交叉影响系统 API 文档、插件生命周期、插件 UI、启动 SQL、定时任务和 OpenSpec 语言治理；这些能力的当前契约由对应 owner 分组或`openspec/specs`承载。
- 不引入新的公开 HTTP API 契约、数据库 schema、插件源码目录、运行时 i18n 资源或生产构建入口。
