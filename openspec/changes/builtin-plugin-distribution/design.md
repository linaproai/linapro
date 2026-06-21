## Context

当前插件治理有两条生命周期入口：普通插件管理页面，以及宿主配置中的`plugin.autoEnable`。二者都面向“可被管理员治理的插件”。项目内建源码插件的语义不同：它是 LinaPro 项目源码交付的一部分，业务逻辑仍闭环在`apps/lina-plugins/<plugin-id>/`，但安装、启用和安全升级应由启动流程自动收敛，普通插件管理入口不应允许禁用、卸载或手动升级。

现有插件能力已经具备 manifest 扫描、注册表同步、发布 snapshot、启动自动启用、运行时升级、依赖检查、SQL 迁移、菜单权限同步、enabled snapshot 和运行时缓存刷新。本设计不新增第二套生命周期，而是在这些现有入口上增加`distribution`治理维度。

## Goals / Non-Goals

**Goals:**

- 用`distribution: builtin`表达项目内建源码插件，缺省值保持`marketplace`。
- 将`distribution`写入 manifest 投影、注册表、发布 snapshot、列表和详情 API。
- 普通插件管理列表默认隐藏`builtin`插件，写操作在服务端统一拒绝。
- 启动期独立执行`BootstrapBuiltinPlugins(ctx)`，在插件路由、cron、前端包预热前自动安装、启用和安全升级`builtin`源码插件。
- 生命周期变化继续复用现有依赖解析、SQL 迁移、资源同步、缓存失效、enabled snapshot 和集群主节点边界。

**Non-Goals:**

- 不删除或替代`plugin.autoEnable`；它仍用于部署方托管普通插件。
- 不允许动态插件通过 manifest 自声明`builtin`。
- 不新增`managed`、`auto_install`、`auto_enable`、`auto_upgrade`或`visibility`等组合开关。
- 不将具体业务插件逻辑写入`apps/lina-core`。
- 第一版不强制新增独立诊断页面；可通过受控只读查询保留诊断入口扩展空间。

## Decisions

### 决策 1：`distribution`只保留`marketplace`和`builtin`

`plugintypes`新增`PluginDistribution`命名类型和归一化函数。manifest 空值归一化为`marketplace`，非法值在 manifest 校验阶段失败。`catalog.Manifest`和`store.ManifestSnapshot`保存归一化后的字符串。

备选方案是增加`managed`或多开关组合。该方案会与`plugin.autoEnable`语义重叠，并让插件生命周期策略出现组合状态，不利于审查和长期治理，因此不采用。

### 决策 2：`builtin`必须同时满足源码插件和编译期注册

`distribution=builtin`只允许`type=source`，且源码插件注册表中必须存在同 ID 的`SourcePlugin`绑定。只依赖 manifest 声明不足以获得内建治理语义；若动态插件或未注册源码目录声明`builtin`，启动扫描必须失败。

该双因子约束保证`builtin`继续代表“随项目源码编译交付”的业务能力，而不是运行时 artifact 自我提权。

### 决策 3：注册表持久化`distribution`

`sys_plugin`基线表结构新增`distribution varchar(32) not null default 'marketplace'`。管理列表过滤、写操作 guard 和缺少当前 manifest 时的诊断投影都以注册表作为权威状态。发布 snapshot 同步保存目标 release 的`distribution`，便于升级预览和失败诊断。

查询性能判断：普通列表默认过滤`builtin`属于插件管理分页读模型的一部分，应在注册表投影或管理摘要缓存中完成，不在列表请求中重新解析 manifest。

### 决策 4：管理写操作由服务端 guard 统一拒绝

根插件服务在安装、启用、禁用、卸载、手动升级和租户 provisioning 策略更新入口执行`builtin` guard。错误使用稳定`bizerr`码`plugin.builtin.management.action.denied`，前端隐藏操作只是体验优化，不作为安全边界。

数据权限判断：插件管理属于平台治理数据，不接入普通业务数据权限过滤；普通列表默认隐藏`builtin`以避免存在性暴露，显式包含内建插件的诊断查询需要平台治理权限。

### 决策 5：启动收敛独立于`BootstrapAutoEnable`

新增`BootstrapBuiltinPlugins(ctx)`或同等独立入口，启动顺序为：source manifest sync → builtin 收敛 → `plugin.autoEnable` → 租户 provisioning → runtime routes/frontend/cron 接线。这样可以保持`plugin.autoEnable`的部署托管语义不被内建插件逻辑污染。

收敛策略为未安装则安装、待升级则安全升级、未启用则启用。安全升级复用现有源码升级编排、依赖检查、SQL 迁移、治理同步、失败诊断和缓存发布逻辑，不新增第二套升级状态机。

### 决策 6：缓存和集群边界复用现有生命周期副作用

`builtin`安装、升级和启用成功后必须触发现有管理读模型失效、enabled snapshot 刷新、runtime revision 或共享状态发布。`cluster.enabled=true`时仅 primary 执行共享生命周期写入，从节点等待注册表和 runtime 状态收敛并刷新本地投影。

缓存权威来源仍是`sys_plugin`、`sys_plugin_release`、资源引用、菜单权限和已发布 runtime revision。失效必须发生在权威写入成功之后，不在事务回滚前发布不可恢复事件。

## Risks / Trade-offs

- `builtin`自动升级失败会阻塞启动。缓解方式：保留升级失败 ledger 和 release 诊断，错误包含插件 ID、目标版本和失败阶段，修复后重启。
- 列表默认隐藏`builtin`可能让管理员误以为插件不存在。缓解方式：保留受控`includeBuiltin`诊断查询或后续独立诊断页。
- 启动自动升级复用显式升级逻辑时容易重复执行管理 guard。缓解方式：公开管理入口执行`builtin` guard，启动内部入口走窄升级契约并标记启动来源。
- `distribution`字段进入宿主 API 文档和前端类型，涉及 i18n/apidoc 治理。缓解方式：DTO 源文本使用英文并同步宿主非英文 apidoc 翻译。
- 数据库变更需要 DAO 生成和编译验证。缓解方式：将字段合入全新项目的基线建表 SQL，执行生成和变更包测试。

## Migration Plan

1. 新增 manifest 字段、命名类型、校验、注册表基线字段和 release snapshot 序列化。
2. 扩展插件列表和详情投影，默认过滤`builtin`，补齐写操作 guard 和错误翻译。
3. 调整前端插件管理页，基于`distribution`隐藏内建插件和写操作。
4. 新增`BootstrapBuiltinPlugins`并接入启动顺序，复用现有 lifecycle/upgrade/cache/cluster 能力。
5. 补齐单元测试、必要 E2E、`openspec validate builtin-plugin-distribution --strict`、Go 编译门禁、SQL 幂等审查和 i18n 校验。

回滚策略：回滚启动收敛入口和管理 guard 后，`sys_plugin.distribution`字段可作为基线字段保留；所有插件缺省`marketplace`，普通插件治理恢复现有行为。

## Open Questions

- `includeBuiltin=true`是否需要新增`plugin:diagnose`权限，还是第一版复用平台插件查询权限并仅用于超级管理员入口。
- `builtin`的 tenant-scoped 新租户 provisioning 是否完全继承 manifest 安装模式，还是后续需要独立诊断字段展示自动纳入原因。
