## ADDED Requirements

### Requirement: 租户参数 fallback 行必须返回来源和动作元数据

租户上下文查询参数设置列表时，系统 SHALL 对平台默认 fallback 行返回来源和动作元数据，使调用方能区分“当前租户覆盖值”和“继承平台默认值”。元数据至少包含 `sourceTenantId`、`isFallback`、`canEdit`、`canOverride` 和 `overrideMode`。租户上下文不得把平台 fallback 行伪装成可直接编辑的本租户记录。

#### Scenario: 租户看到平台 fallback 参数

- **WHEN** 租户 A 未覆盖某个内置参数
- **AND** 参数列表通过平台默认值 fallback 返回该参数
- **THEN** 响应行包含 `sourceTenantId = 0`
- **AND** `isFallback = true`
- **AND** `canEdit = false`
- **AND** 删除和直接编辑操作不得在前端展示为可执行动作

#### Scenario: 租户看到本租户覆盖参数

- **WHEN** 租户 A 已覆盖某个参数
- **THEN** 响应行包含 `sourceTenantId = A`
- **AND** `isFallback = false`
- **AND** `canEdit` 根据当前用户权限和内置保护规则计算

#### Scenario: 平台上下文查看平台默认参数

- **WHEN** 平台管理员在平台上下文查询参数列表
- **THEN** 平台默认参数行不标记为租户 fallback
- **AND** `canEdit` 按平台参数管理权限和内置保护规则计算

### Requirement: 参数 fallback 动作必须避免必失败详情请求

前端 SHALL 使用参数行的动作元数据决定操作按钮。对 `isFallback = true` 且 `canEdit = false` 的行，前端不得显示会调用当前租户详情编辑接口且必然返回 not found 的编辑入口。若 `canOverride = true`，前端 MAY 显示创建租户覆盖入口，但该入口 MUST 调用明确的覆盖创建流程。

#### Scenario: fallback 参数行不显示直接编辑

- **WHEN** 租户用户在参数列表看到平台 fallback 行
- **AND** 该行 `canEdit = false`
- **THEN** 前端不显示直接编辑按钮
- **AND** 用户不会触发返回“参数设置不存在”的详情请求

#### Scenario: fallback 参数行允许创建覆盖

- **WHEN** 租户用户看到 `canOverride = true` 且 `overrideMode = createTenantOverride` 的 fallback 参数行
- **THEN** 前端显示的覆盖入口必须表达为创建租户覆盖
- **AND** 保存后创建或更新当前租户的参数覆盖记录
- **AND** 缓存失效范围限定为当前租户和该参数键
