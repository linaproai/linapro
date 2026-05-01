## MODIFIED Requirements

### Requirement: 插件内置任务 Handler 生命周期

系统 SHALL 订阅插件安装/启用/禁用/卸载事件,自动同步 handler 注册表、插件内置任务调度 entry 与关联任务投影状态。插件内置任务的执行来源 SHALL 是插件声明和 handler 注册表；`sys_job.is_builtin=1` 投影记录 SHALL 用于展示、日志关联和状态治理，不得作为启动期持久化扫描的注册来源。

#### Scenario: 生命周期回调与响应边界

- **WHEN** 插件安装、启用、禁用或卸载请求成功
- **THEN** 系统 SHALL 在同一请求链路内通过显式生命周期回调完成 handler 注册表、内置任务投影和调度 entry 同步
- **AND** 不依赖独立的 best-effort 异步事件总线再补偿

#### Scenario: 插件安装时补齐内置任务投影

- **WHEN** 插件安装成功但尚未启用
- **THEN** 系统 SHALL 先将该插件声明的内置定时任务同步到 `sys_job`
- **AND** 若对应 handler 当前尚未可用,这些任务 SHALL 直接进入 `paused_by_plugin` 状态
- **AND** 后续启用插件时再通过 handler 注册表恢复执行能力
- **AND** 安装阶段不得仅凭 `sys_job` 投影注册可执行调度 entry

#### Scenario: 插件启用时恢复内置任务 handler

- **WHEN** 插件启用
- **THEN** 系统 SHALL 读取该插件声明的内置定时任务
- **AND** 仅按 `plugin:<pluginID>/cron:<name>` 形式注册对应 handler
- **AND** 对 `stop_reason=plugin_unavailable` 的任务恢复为 `status=enabled`
- **AND** 通过插件声明和当前 `sys_job.id` 投影注册调度 entry

#### Scenario: 插件禁用时注销 handler 并级联任务

- **WHEN** 插件被禁用
- **THEN** 系统 SHALL 注销该插件所有 handler
- **AND** 注销该插件所有内置任务调度 entry
- **AND** 将 `handler_ref=plugin:<pluginID>/cron:*` 且 `status=enabled` 的任务置为 `paused_by_plugin`
- **AND** 在对应任务上写入 `stop_reason=plugin_unavailable`

#### Scenario: 插件卸载时保留任务数据

- **WHEN** 插件被卸载
- **THEN** 系统 SHALL 执行与禁用相同的级联(标记 `paused_by_plugin` 并注销调度 entry)
- **AND** 不删除已有任务与历史日志
- **AND** UI 明显标红任务并提示 handler 不可用

#### Scenario: UI 可见性

- **WHEN** 任务列表返回 `paused_by_plugin` 任务
- **THEN** 前端 SHALL 在状态列显式提示"插件处理器不可用"
- **AND** 禁用"立即触发""启用"按钮

### Requirement: 动态插件定时任务声明契约

系统 SHALL 为动态插件提供独立的定时任务声明契约,并将其与 runtime host service 边界分离。动态插件声明的定时任务 SHALL 作为插件内置任务的执行来源，宿主 SHALL 将其同步为 `sys_job.is_builtin=1` 投影并按插件生命周期注册或注销调度 entry。

#### Scenario: 动态插件通过 cron host service 代码注册定时任务

- **WHEN** 动态插件需要提供内置定时任务
- **THEN** 插件 SHALL 通过独立 `cron` host service 在 guest 代码中调用 `pluginbridge.Cron().Register(...)` 提交任务元数据与 guest 处理器绑定信息
- **AND** 宿主 SHALL 通过保留的 guest 注册入口执行一次受控 discovery 来收集这些声明
- **AND** 收集到的声明 SHALL 纳入统一任务投影链路
- **AND** 收集到的声明 SHALL 作为插件内置任务的运行期注册来源

#### Scenario: runtime host service 保持聚焦

- **WHEN** 宿主暴露 guest-side `runtime` host service SDK
- **THEN** 其职责 SHALL 继续限定为运行时日志、状态与轻量信息查询
- **AND** 不直接承担动态插件定时任务的注册治理入口
- **AND** 若未来需要开放插件侧任务治理能力,应以独立 `cron`/`job` host service 形式扩展

#### Scenario: 动态插件授权页展示 cron 注册明细

- **WHEN** 动态插件在安装或启用前打开宿主服务授权审查页，且当前 release 已声明 `cron` host service 并成功发现定时任务合同
- **THEN** 前端 SHALL 将该服务名称展示为“任务服务”
- **AND** 在该卡片下展示已注册定时任务的名称、表达式、调度范围与并发策略等摘要信息
- **AND** 数据服务、存储服务、网络服务等治理目标摘要标签 SHALL 使用“数据表名”“存储路径”“访问地址”等不带“申请/授权”前缀的等长中文文案
- **AND** `runtime` 服务卡片 SHALL 排在授权页宿主服务列表的最底部，位于“任务服务”之后
- **AND** “任务服务”卡片 SHALL 直接展示任务面板，不额外渲染“定时任务”摘要标签行
- **AND** 每个任务面板中的属性标题（如表达式、调度范围、并发策略）SHALL 使用粗体强调显示
- **AND** 授权页中“申请清单”与“数据表名/存储路径/访问地址”等资源类型标签的背景色 SHALL 与详情页对应标签保持一致
- **AND** 插件详情页中的“当前生效范围”文案 SHALL 统一改为“生效范围”
