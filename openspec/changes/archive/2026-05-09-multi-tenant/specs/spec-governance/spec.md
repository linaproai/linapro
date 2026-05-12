## ADDED Requirements

### Requirement: 多租户敏感能力 spec 必须显式声明四类行为
凡涉及租户敏感业务数据、缓存、审计的能力 spec(无论新建或修改),其 Requirements 中 SHALL 显式声明对 `tenant_id` 的处理方式;凡涉及平台控制面或全局配置表,也 SHALL 显式声明不使用 `tenant_id` 的理由与替代隔离边界。相关说明覆盖以下四类行为:
1. **读路径**:列表/详情/聚合查询是否注入 `tenant_id` 过滤,是否走 platform fallback。
2. **写路径**:创建/更新是否自动写入当前 `tenant_id`,是否需平台权限才能跨租户写。
3. **缓存**:cache key 是否携带 `tenant_id`,失效作用域如何确定。
4. **审计**:操作日志是否记录 `tenant_id` / `acting_user_id` / `on_behalf_of_tenant_id`。

#### Scenario: spec 四类行为完整
- **WHEN** 评审某 tenancy-aware 能力 spec
- **THEN** 至少包含上述四类中相关项的 Requirement
- **AND** 缺失项需在 design.md 显式说明"该能力不涉及该行为"

### Requirement: lina-review 审查清单纳入 tenancy 检查
`/lina-review` 在审查 spec 与代码时 SHALL 检查:
- DAO 调用是否经 `tenantcap.Apply`。
- 带 `tenant_id` 表的 DAO 调用是否经 `tenantcap.Apply`,平台控制面表是否通过治理 service 访问。
- 租户敏感缓存 key 是否携带 tenant 维度,全局交付资源缓存是否保持明确 scope。
- 带 `tenant_id` 业务表写入是否填 tenant_id。
- 审计日志是否记录 tenant 字段。

任何违反项视为审查不通过。

#### Scenario: 审查阻断
- **WHEN** 提交的 PR 中存在 `dao.SysUser.Ctx(ctx).Where(...).Scan` 未经 `tenantcap.Apply`
- **THEN** lina-review 标记为不通过,要求修复
