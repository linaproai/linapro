# Tasks

## Summary

- [x] 交付运行时配置治理：注册并保护 JWT、会话、上传、登录黑名单和公共前端配置键，接入认证、在线会话、上传校验、登录页和工作台启动读取。
- [x] 交付上传默认值统一：`sys.upload.maxSize`在 SQL 种子、配置模板和后端兜底中统一为 20 MB，并同步请求限制、文件校验、错误提示和回归测试。
- [x] 交付配置快照治理：受保护运行时参数使用本地快照加共享修订号；单机仅本地失效，集群通过共享修订号收敛；不新增跨请求业务缓存。
- [x] 交付登录后首页 SQL 优化：在线会话校验单次读取后判定，插件 release 在请求级或列表级复用，状态变化后仍以数据库权威状态为准。
- [x] 交叉影响已迁移：插件、定时任务、API 文档、启动 SQL、认证、在线会话和 OpenSpec 语言治理的完整契约由对应 owner 分组或`openspec/specs`承载，本分组仅保留影响摘要。
- [x] 验证：配置管理覆盖率门禁、会话校验测试、插件 release 复用测试、启动 smoke、OpenSpec 校验和`git diff --check`均作为归档维护证据保留。
- [x] 治理：无新增公开 HTTP API、数据库 schema、运行时代码依赖、插件源码目录或 i18n 资源；缓存一致性、数据权限、测试策略和跨平台影响已按原实施记录审查。
- [x] 泛化 sys_config 有效快照：将`GetRaw()`对`sys_config`的读取从硬编码 key 白名单升级为数据驱动的有效配置快照；源码插件通过稳定`HostConfig()`读取，动态插件仍受 manifest key 授权约束。
- [x] 插件业务配置优先级：将`plugin.<plugin-id>`作为插件作用域配置的最高优先级来源；源码插件和动态插件复用同一个`ConfigServiceFactory`。
- [x] HostConfig 读取优先级统一：将`GetRaw()`统一为`sys_config`快照 → `config.yaml` → 系统默认值 → `nil`的读取顺序；移除`IsManagedSysConfigKey()`分支；默认值通过通用元数据查询。
- [x] FB-1 至 FB-5：自定义 sys_config 改 key 后旧 key 失效、合并运行时参数文件、测试文件命名治理、调度模块内置参数统一为`sys.`前缀、HostConfig.Get 增加默认值参数。
- [x] 验证：`openspec validate`通过；`go test`覆盖 config、sysconfig、plugin hostconfig/wasm、httpstartup 等包；静态检索确认无`IsManagedSysConfigKey`、旧 key 名称残留。
