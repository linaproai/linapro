## ADDED Requirements

### Requirement: 翻译缓存 key 携带租户维度
运行时翻译缓存 key SHALL 形如 `i18n:tenant=<id>:locale=<locale>:sector=<sector>:key=<key>`;不同租户的相同 key 翻译可不同(因为字典/配置覆盖会影响可翻译资源)。

#### Scenario: 租户独立翻译缓存
- **WHEN** 租户 A 与租户 B 同时请求 `dict.business_priority.P0` 翻译
- **AND** 两租户对该字典项有不同覆盖
- **THEN** 缓存中存在两份独立条目
- **AND** 互不影响

### Requirement: 失效作用域的 tenant 维度
翻译缓存失效 SHALL 必须显式指定 `(tenant_id, locale, sector)`;禁止在普通业务路径中无 tenant 维度地清空所有租户翻译。

#### Scenario: 租户覆盖触发本租户失效
- **WHEN** 租户 A 修改某字典覆盖
- **THEN** 失效作用域 `(tenant_id=A, locale=*, sector=dict)`
- **AND** 不影响租户 B 缓存

#### Scenario: 平台默认翻译变更级联失效
- **WHEN** 平台管理员修改 apidoc 翻译
- **THEN** 失效作用域 `(tenant_id=0, locale=*, sector=apidoc, cascade_to_tenants=true)`

### Requirement: 缓存命中率监控按租户维度
翻译缓存监控指标(命中率、miss 计数、失效次数)SHALL 按 `tenant_id` 维度上报,便于按租户级别诊断性能问题。

#### Scenario: 监控按租户聚合
- **WHEN** 运维查看翻译缓存性能 dashboard
- **THEN** 指标可按 `tenant_id` 维度筛选与对比
- **AND** 可定位特定租户的缓存命中异常
