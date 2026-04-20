# cron-handler-registry Specification

## Purpose
TBD - created by archiving change scheduled-job-management. Update Purpose after archive.
## Requirements
### Requirement: Handler 注册表

系统 SHALL 提供统一的定时任务 Handler 注册表,管理宿主与插件侧的可用 handler,并作为任务调度、UI 下拉与参数校验的唯一真理来源。

#### Scenario: 注册 handler

- **WHEN** 宿主或插件通过注册表 `Register(ref, def)` 注册 handler
- **THEN** 系统 SHALL 将 handler 定义持久化到内存注册表
- **AND** `ref` 形如 `host:<name>` 或 `plugin:<pluginID>/<name>`,且全局唯一
- **AND** `def` 至少包含 `DisplayName / Description / ParamsSchema / Source / Invoke`

#### Scenario: 重复注册冲突

- **WHEN** 尝试注册已存在的 `ref`
- **THEN** 系统 SHALL 返回明确错误
- **AND** 不覆盖已注册 handler

#### Scenario: 注销 handler

- **WHEN** 宿主或插件调用 `Unregister(ref)`
- **THEN** 系统 SHALL 从内存注册表移除该 handler
- **AND** 触发任务状态级联

### Requirement: Handler 参数 JSON Schema

系统 SHALL 要求 handler 声明参数的 `JSON Schema draft-07` 受限标量子集,供参数校验与 UI 动态渲染使用。

#### Scenario: 仅接受受限 Schema 子集

- **WHEN** 宿主或插件注册一个 handler 的 `ParamsSchema`
- **THEN** 根节点 SHALL 为 `type=object`
- **AND** 字段类型仅支持 `string / integer / number / boolean`
- **AND** 关键字仅支持 `properties / required / description / default / enum / format`
- **AND** 系统 SHALL 拒绝 `array`、嵌套 `object`、`$ref`、`allOf`、`anyOf`、`oneOf`、`not`、`patternProperties` 等超出本迭代表单映射能力的结构

#### Scenario: 创建任务时参数校验

- **WHEN** 创建或修改 `task_type=handler` 的任务
- **THEN** 系统 SHALL 按 handler 对应的 JSON Schema 校验 `params`
- **AND** 校验失败时返回具体字段错误信息

#### Scenario: 执行时参数校验

- **WHEN** handler 执行前
- **THEN** 系统 SHALL 再次按最新 Schema 校验 `params_snapshot`
- **AND** 校验失败时记日志 `status=failed` 并终止本次执行

#### Scenario: UI 下拉数据

- **WHEN** 前端调用 `GET /job/handler`
- **THEN** 系统 SHALL 返回所有已注册 handler 列表
- **AND** 列表包含 `ref / displayName / description / source / pluginId`

#### Scenario: UI 详情获取 Schema

- **WHEN** 前端调用 `GET /job/handler/{ref}`
- **THEN** 系统 SHALL 返回该 handler 的完整 `ParamsSchema`
- **AND** 前端可据此动态渲染表单控件

### Requirement: 插件 Handler 生命周期

系统 SHALL 订阅插件安装/启用/禁用/卸载事件,自动同步 handler 注册表与关联任务状态。

#### Scenario: 生命周期回调与响应边界

- **WHEN** 插件安装、启用、禁用或卸载请求成功
- **THEN** 系统 SHALL 在同一请求链路内通过显式生命周期回调完成 handler 注册表与关联任务状态同步
- **AND** 不依赖独立的 best-effort 异步事件总线再补偿

#### Scenario: 插件安装时补齐内置任务投影

- **WHEN** 插件安装成功但尚未启用
- **THEN** 系统 SHALL 先将该插件声明的内置定时任务同步到 `sys_job`
- **AND** 若对应 handler 当前尚未可用,这些任务 SHALL 直接进入 `paused_by_plugin` 状态
- **AND** 后续启用插件时再通过 handler 注册表恢复执行能力

#### Scenario: 插件启用时注册 handler

- **WHEN** 插件启用
- **THEN** 系统 SHALL 读取插件清单声明的 handler 列表
- **AND** 按 `plugin:<pluginID>/<name>` 形式注册
- **AND** 对 `stop_reason=plugin_unavailable` 的任务恢复为 `status=enabled` 并重新注册到调度器

#### Scenario: 插件禁用时注销 handler 并级联任务

- **WHEN** 插件被禁用
- **THEN** 系统 SHALL 注销该插件所有 handler
- **AND** 将 `handler_ref=plugin:<pluginID>/*` 且 `status=enabled` 的任务置为 `paused_by_plugin`
- **AND** 在对应任务上写入 `stop_reason=plugin_unavailable`
- **AND** 从调度器注销这些任务

#### Scenario: 插件卸载时保留任务数据

- **WHEN** 插件被卸载
- **THEN** 系统 SHALL 执行与禁用相同的级联(标记 `paused_by_plugin`)
- **AND** 不删除已有任务与历史日志
- **AND** UI 明显标红任务并提示 handler 不可用

#### Scenario: UI 可见性

- **WHEN** 任务列表返回 `paused_by_plugin` 任务
- **THEN** 前端 SHALL 在状态列显式提示"插件处理器不可用"
- **AND** 禁用"立即触发""启用"按钮

### Requirement: 动态插件定时任务声明契约

系统 SHALL 为动态插件提供独立的定时任务声明契约,并将其与 runtime host service 边界分离。

#### Scenario: 动态插件通过 cron host service 代码注册定时任务

- **WHEN** 动态插件需要提供内置定时任务
- **THEN** 插件 SHALL 通过独立 `cron` host service 在 guest 代码中调用 `pluginbridge.Cron().Register(...)` 提交任务元数据与 guest 处理器绑定信息
- **AND** 宿主 SHALL 通过保留的 guest 注册入口执行一次受控 discovery 来收集这些声明
- **AND** 收集到的声明 SHALL 纳入统一任务投影链路

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

### Requirement: Handler 执行契约

系统 SHALL 在 handler 规范中要求所有 handler 支持 context 取消,并在取消时尽快退出。

#### Scenario: Handler 接受 context

- **WHEN** 系统调用 handler 的 `Invoke(ctx, params)` 方法
- **THEN** handler SHALL 在长阻塞操作中检查 `ctx.Done()` 或把 `ctx` 透传到下游
- **AND** 收到取消信号后应尽快返回 `ctx.Err()`

#### Scenario: Handler 返回结构化结果

- **WHEN** handler 执行成功
- **THEN** handler SHALL 返回可 JSON 序列化的结果
- **AND** 系统 SHALL 将结果写入 `sys_job_log.result_json`

#### Scenario: Handler 抛错

- **WHEN** handler 执行返回非 nil error
- **THEN** 系统 SHALL 记 `status=failed` 且 `err_msg` 为 error.Error() 摘要

