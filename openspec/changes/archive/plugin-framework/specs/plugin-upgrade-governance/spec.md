# 插件升级治理规范

## Purpose

定义插件文件版本发现、有效版本隔离和运行时升级边界。插件文件覆盖只产生发现版本；数据库有效版本、治理资源和插件数据升级必须通过宿主启动后的运行时升级流程处理。
## Requirements
### Requirement:源码插件必须分离有效版本和发现的源码版本

系统 SHALL 区分当前有效的源码插件版本和在源码树中发现的插件版本。`sys_plugin.version` 和 `release_id` 仅代表有效版本，新发现的版本存储为发布记录或发现快照，在运行时升级完成前不得覆盖有效版本。

#### Scenario: 已安装的源码插件发现更高版本
- **WHEN** 源码插件 `plugin-demo` 有效运行 `v0.1.0` 且其源码中的 `plugin.yaml` 已升级到 `v0.5.0`
- **THEN** `sys_plugin.version` 保持 `v0.1.0`
- **AND** 系统记录 `v0.5.0` 源码插件发布快照
- **AND** 该新发布不被视为当前有效版本，直到运行时升级完成

#### Scenario: 已安装的源码插件发现更低版本
- **WHEN** 源码插件 `plugin-demo` 有效运行 `v0.5.0` 且其源码中的 `plugin.yaml` 是 `v0.1.0`
- **THEN** `sys_plugin.version` 保持 `v0.5.0`
- **AND** 系统将插件标记为异常状态
- **AND** 系统要求管理员人工修复文件或数据库状态后再恢复正常

### Requirement:源码插件升级必须是显式的运行时操作

系统 SHALL 要求源码插件升级通过宿主启动后的运行时管理 API 显式执行，而非通过开发期升级命令执行，也非在宿主启动期间自动修复。开发阶段只能通过文件覆盖让宿主发现新版本。

#### Scenario: 显式升级单个源码插件
- **WHEN** 管理员在插件管理页确认升级 `plugin-demo`
- **AND** 宿主收到 `POST /plugins/{id}/upgrade` 请求
- **THEN** 系统仅为 `plugin-demo` 执行运行时升级流程
- **AND** 不触发其他源码插件或动态插件的升级

#### Scenario: 源码插件文件覆盖后等待运行时升级
- **WHEN** 开发者覆盖 `apps/lina-plugins/plugin-demo` 下的源码插件文件
- **AND** 插件 `plugin.yaml` 版本高于数据库有效版本
- **THEN** 宿主启动后将插件标记为待升级
- **AND** 系统不得要求运行旧的开发期升级命令

### Requirement:宿主启动必须标记源码插件升级状态

宿主 SHALL 在启动期间扫描源码插件，然后比较已安装源码插件的发现版本和有效版本。如果发现版本高于有效版本，宿主必须标记插件为待升级并继续启动；如果发现版本低于有效版本，宿主必须标记插件为异常并继续启动。

#### Scenario: 待处理的源码插件升级不阻塞启动
- **WHEN** 宿主启动并发现 `plugin-demo` 有效运行 `v0.1.0` 而源码发现报告 `v0.5.0`
- **THEN** 启动流程继续
- **AND** 插件运行时状态变为待升级
- **AND** 插件管理页可显示有效版本、发现版本和升级动作

#### Scenario: 源码插件发现版本低于有效版本
- **WHEN** 宿主启动并发现 `plugin-demo` 有效运行 `v0.5.0` 而源码发现报告 `v0.1.0`
- **THEN** 启动流程继续
- **AND** 插件运行时状态变为异常
- **AND** 插件管理页提示管理员人工干预修复

### Requirement:源码插件升级必须记录 `phase=upgrade` 并同步治理资源

源码插件运行时升级 SHALL 执行升级阶段迁移记账并同步菜单、权限、资源引用、i18n、apidoc、路由和 cron 等治理资源。成功运行后，新发布成为有效发布。

#### Scenario: 源码插件升级成功
- **WHEN** 管理员升级已安装的源码插件且所有升级回调、SQL 和治理同步步骤成功
- **THEN** `sys_plugin.version` 和 `release_id` 更新为新发布
- **AND** `sys_plugin_migration` 记录 `phase=upgrade` 条目
- **AND** 新发布成为有效发布
- **AND** 插件运行时状态变为正常

#### Scenario: 源码插件升级失败
- **WHEN** 源码插件升级期间插件回调、升级 SQL 语句或治理同步步骤失败
- **THEN** 运行时升级流程立即停止
- **AND** 保留失败的升级记录和错误信息
- **AND** 插件运行时状态变为升级失败
- **AND** 系统不自动执行回滚

### Requirement:动态插件升级必须进入统一运行时升级模型

系统 SHALL 保持动态插件升级在运行时模型上。动态插件新 artifact 被上传或文件覆盖后，若发现版本高于数据库有效版本，系统必须将插件标记为待升级，并通过同一插件管理页升级流程完成有效 release 切换和治理资源同步。

#### Scenario: 动态插件发现更高版本
- **WHEN** 动态插件 `linapro-demo-dynamic` 有效运行 `v0.1.0`
- **AND** 本地发现或上传的动态插件 artifact 版本为 `v0.2.0`
- **THEN** 系统将插件标记为待升级
- **AND** 有效 release 仍保持 `v0.1.0`
- **AND** 管理员必须通过插件管理页显式确认升级

#### Scenario: 动态插件文件版本低于有效版本
- **WHEN** 动态插件 `linapro-demo-dynamic` 有效运行 `v0.2.0`
- **AND** 本地发现或上传的 artifact 版本为 `v0.1.0`
- **THEN** 系统将插件标记为异常
- **AND** 系统不得自动降级有效 release

### Requirement: 插件升级必须校验新版本依赖约束

源码插件升级命令和动态插件安装/升级路径 SHALL 在切换有效发布前校验新版本 manifest 的依赖约束。新版本的框架版本约束、硬依赖存在性和硬依赖版本范围必须满足，否则升级或发布切换必须失败。

#### Scenario: 源码插件升级前校验依赖
- **WHEN** 开发者升级源码插件 `x` 到新版本
- **AND** 新版本声明硬依赖 `a >=0.2.0`
- **AND** 当前已安装或可用的 `a` 版本不满足
- **THEN** 源码插件升级失败
- **AND** `x` 的有效版本保持升级前版本

#### Scenario: 动态插件同版本刷新前校验依赖
- **WHEN** 动态插件以同版本新产物刷新
- **AND** 新产物 manifest 声明当前环境不满足的框架版本约束
- **THEN** 动态插件刷新失败
- **AND** 当前活跃发布继续指向刷新前产物

### Requirement: 插件升级不得破坏已安装插件的反向依赖

插件升级 SHALL 校验升级后的有效版本不会破坏其他已安装插件对该插件的硬依赖版本范围。如果升级结果使下游插件依赖不满足，系统必须拒绝切换有效发布。

#### Scenario: 目标插件升级后不满足下游依赖
- **WHEN** 已安装插件 `consumer` 硬依赖 `base <0.3.0`
- **AND** 管理员尝试将 `base` 升级到 `v0.3.0`
- **THEN** 升级请求失败
- **AND** 错误包含下游插件 `consumer` 和其依赖版本范围

### Requirement: 插件升级不得自动升级依赖插件

插件升级过程 SHALL 不自动升级依赖插件。若新版本依赖要求高于当前依赖插件版本，系统必须阻断升级并返回需要先升级的依赖列表。

#### Scenario: 新版本要求更高依赖版本
- **WHEN** 插件 `x` 新版本要求 `a >=0.2.0`
- **AND** 当前 `a` 有效版本为 `v0.1.0`
- **THEN** 升级 `x` 失败
- **AND** 错误提示先升级 `a`
- **AND** 系统不得自动升级 `a`

### Requirement: 动态插件升级失败必须保留旧有效发布并记录目标失败诊断

系统 SHALL 在动态插件升级或同版本刷新失败时保留升级前的有效发布和 active release。失败的目标发布、artifact 校验和、生命周期阶段、原始错误和 rollback 错误 MUST 被记录为可诊断状态；系统不得将失败目标发布切换为有效发布。

#### Scenario: 动态插件升级 SQL 失败保留旧发布
- **WHEN** 动态插件 P 从 release A 升级到 release B
- **AND** release B 的升级 SQL 执行失败
- **THEN** P 的有效发布继续指向 release A
- **AND** release B 记录升级失败诊断
- **AND** 系统不得暴露 release B 的动态路由、前端资源或 runtime i18n 作为有效能力

#### Scenario: 同版本刷新 rollback 失败保留旧产物并记录诊断
- **WHEN** 动态插件 P 以同版本新 artifact 刷新
- **AND** 刷新失败后的 rollback 也失败
- **THEN** P 的 active release 继续指向刷新前 artifact
- **AND** 系统记录刷新原始失败和 rollback 失败诊断
- **AND** 后续协调不得把失败 artifact 误判为成功刷新

### Requirement: 动态插件升级失败后的运行时缓存不得指向失败目标

系统 SHALL 在动态插件升级或同版本刷新失败后，确保 runtime revision、enabled snapshot、frontend bundle、runtime i18n 和 Wasm 编译缓存继续以旧有效发布或明确失败状态为准。失败目标发布不得成为派生缓存的权威来源。

#### Scenario: 失败升级不刷新为目标缓存
- **WHEN** 动态插件 P 升级到 release B 失败
- **THEN** 系统不得发布使其他节点加载 release B 为有效发布的 runtime revision
- **AND** 其他节点继续使用 release A 的有效缓存或按失败状态隐藏 P

#### Scenario: rollback 失败时采用保守暴露策略
- **WHEN** 动态插件 P 升级失败且 rollback 恢复失败
- **THEN** 系统不得暴露失败目标发布的能力
- **AND** 系统根据权威 active release 可用性继续使用旧发布或隐藏该插件能力

### Requirement: 源码插件升级治理不得通过公共`pkg/sourceupgrade`暴露

系统 SHALL 将源码插件升级发现、版本对比、升级执行、失败状态和发布切换视为宿主插件运行时内部治理能力。`apps/lina-core/pkg/sourceupgrade`公共入口 MUST 被删除；宿主内部调用方 MUST 通过`internal/service/plugin`服务接口或其内部 sourceupgrade 组件访问该能力。

#### Scenario: 宿主内部查询源码插件升级状态

- **WHEN** 宿主插件管理服务需要查询源码插件有效版本和发现版本差异
- **THEN** 调用方通过`internal/service/plugin`服务接口或其内部 sourceupgrade 组件查询
- **AND** 不得 import `lina-core/pkg/sourceupgrade`

#### Scenario: 源码插件升级执行

- **WHEN** 管理员显式升级一个源码插件
- **THEN** 插件管理 API 委托到宿主插件运行时内部 sourceupgrade 实现
- **AND** 升级流程继续遵守插件升级治理中的依赖检查、生命周期回调、SQL 迁移、治理资源同步和缓存失效要求

#### Scenario: 插件开发者声明升级资源

- **WHEN** 源码插件需要提供升级 SQL、生命周期回调或 manifest 资源
- **THEN** 插件通过`pluginhost`生命周期契约和插件资源目录声明
- **AND** 插件不得依赖公共`pkg/sourceupgrade`SDK

### Requirement: 插件生命周期变化必须刷新 Pluginservice Capability Provider 状态

系统 SHALL 在插件安装、启用、禁用、卸载、升级、同版本刷新和发布切换成功后，重新计算受影响的 pluginservice capability provider 激活状态。若插件提供 provider，则其 provider 激活、撤销或替换 MUST 与插件有效 release、运行时状态和依赖校验结果一致；集群模式下 MUST 通过插件 runtime revision、事件广播、共享缓存或等价机制传播。

#### Scenario: Provider 插件升级成功后切换 Provider

- **WHEN** 提供`framework.org.v1`的插件升级成功并切换到新有效 release
- **THEN** pluginservice capability manager 使用新 release 对应的 provider factory 重新创建或刷新 provider
- **AND** 旧 provider 不再作为 active provider 处理新调用
- **AND** 集群其他节点收到运行时修订后刷新本地 provider 状态

#### Scenario: Provider 插件禁用后能力降级

- **WHEN** 提供`framework.tenant.v1`的插件被禁用
- **THEN** pluginservice capability manager 撤销该 provider 激活状态
- **AND** 消费 service 返回不可用状态、fallback 行为或规范定义的降级结果
- **AND** 通过`dependencies.plugins`硬依赖该 provider 插件的下游插件在后续启用、升级或健康检查中被标记为依赖不满足

### Requirement: 插件升级必须校验下游 Provider 插件依赖

插件升级 SHALL 校验升级后的 provider 插件状态不会破坏其他已安装插件通过既有`dependencies.plugins`声明的硬依赖。如果升级、禁用或发布切换会导致下游插件硬 provider 插件依赖不满足，系统 MUST 拒绝该操作或进入规范明确的阻断状态；pluginservice capability 的可选消费仍通过运行时可用性降级表达，不引入独立 capability 依赖阻断模型。

#### Scenario: Provider 升级后不满足下游插件依赖版本

- **WHEN** 已安装插件`consumer`在`dependencies.plugins`中硬依赖`linapro-org-core`版本范围`>=1.0.0 <2.0.0`
- **AND** 管理员尝试将`linapro-org-core`升级为不满足该范围的新版本
- **THEN** 升级请求失败或要求先处理下游依赖
- **AND** 错误包含下游插件 ID、provider 插件 ID 和所需版本范围

#### Scenario: 禁用唯一 Provider 时保护下游硬依赖

- **WHEN** 插件`consumer`已启用且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求失败
- **AND** 错误包含依赖该 provider 插件的下游插件列表
