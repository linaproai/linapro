## ADDED Requirements

### Requirement: 租户字典 fallback 行必须返回来源和动作元数据

租户上下文查询字典类型和字典数据时，系统 SHALL 对平台默认 fallback 行返回来源和动作元数据。元数据至少包含 `sourceTenantId`、`isFallback`、`canEdit`、`canOverride` 和 `overrideMode`。租户上下文不得把平台 fallback 字典类型或字典数据伪装成可直接编辑的本租户记录。

#### Scenario: 租户看到平台 fallback 字典类型

- **WHEN** 租户 A 未覆盖某个字典类型
- **AND** 字典类型列表通过平台默认值 fallback 返回该类型
- **THEN** 响应行包含 `sourceTenantId = 0`
- **AND** `isFallback = true`
- **AND** `canEdit = false`
- **AND** 前端不得显示直接编辑或删除该平台行的可执行入口

#### Scenario: 租户看到本租户字典类型

- **WHEN** 租户 A 已覆盖某个字典类型
- **THEN** 响应行包含 `sourceTenantId = A`
- **AND** `isFallback = false`
- **AND** `canEdit` 根据当前用户权限和内置保护规则计算

#### Scenario: 租户看到平台 fallback 字典数据

- **WHEN** 租户 A 查询某个 fallback 字典类型下的字典数据
- **AND** 字典数据来自平台默认行
- **THEN** 每条 fallback 数据行均返回来源和动作元数据
- **AND** 租户上下文不得直接更新或删除这些平台默认数据行

### Requirement: 字典 fallback 动作必须避免必失败详情请求

前端 SHALL 使用字典行的动作元数据决定字典类型和字典数据的操作按钮。对 `isFallback = true` 且 `canEdit = false` 的行，前端不得显示会调用当前租户详情编辑接口且必然返回 not found 的编辑入口。若 `canOverride = true`，前端 MAY 显示创建租户覆盖入口，但该入口 MUST 调用明确的覆盖创建流程。

#### Scenario: fallback 字典类型不显示直接编辑

- **WHEN** 租户用户在字典类型列表看到平台 fallback 行
- **AND** 该行 `canEdit = false`
- **THEN** 前端不显示直接编辑按钮
- **AND** 用户不会触发返回“字典类型不存在”的详情请求

#### Scenario: fallback 字典数据不显示直接编辑

- **WHEN** 租户用户在字典数据列表看到平台 fallback 行
- **AND** 该行 `canEdit = false`
- **THEN** 前端不显示直接编辑按钮
- **AND** 用户不会触发返回“字典数据不存在”的详情请求

#### Scenario: fallback 字典行允许创建覆盖

- **WHEN** 租户用户看到 `canOverride = true` 且 `overrideMode = createTenantOverride` 的 fallback 字典行
- **THEN** 前端显示的覆盖入口必须表达为创建租户覆盖
- **AND** 保存后创建或更新当前租户的字典覆盖记录
- **AND** 字典缓存失效范围限定为当前租户和对应字典键
