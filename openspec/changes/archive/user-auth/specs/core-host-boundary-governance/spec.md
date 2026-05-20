## ADDED Requirements

### Requirement: 租户能力作为宿主稳定接缝
`pkg/tenantcap` SHALL 与 `pkg/orgcap` 并列,被列为宿主"稳定能力接缝(Stable Capability Seam)"之一;接缝必须满足:
1. 接口契约稳定,版本化变更,不随插件细节调整。
2. 默认 no-op 实现,允许在无插件状态下系统正常运行。
3. 通过 Provider + Register 模式让插件接入。
4. 宿主侧不持有具体业务实现,仅透传与默认值。

#### Scenario: 接缝完整性
- **WHEN** 评审 `pkg/tenantcap` 设计
- **THEN** Provider 接口、no-op default、RegisterProvider、Service wrapper 四要素齐备
- **AND** 与 orgcap 形态对齐

### Requirement: 工作台展示变化禁止下沉到核心
凡仅来源于工作台展示需求的变化(列、过滤项、菜单项布局等)SHALL 不允许通过修改 `tenantcap.Provider` 接口或核心 service 解决;应通过工作台适配层处理。

#### Scenario: 拒绝核心污染
- **WHEN** 某需求只是"工作台租户切换器多展示一列"
- **THEN** 修改限定在前端 / 适配层
- **AND** 核心契约不变

### Requirement: 接缝评审纳入 lina-review
`/lina-review` SHALL 在 spec 与代码审查中检查:任何对 `pkg/tenantcap` / `internal/service/tenantcap` 的修改是否符合"稳定接缝"四要素。

#### Scenario: 不符合接缝四要素被阻断
- **WHEN** 提交的 PR 在 `pkg/tenantcap` 中加入了具体业务实现而非纯接口
- **THEN** lina-review 标记不通过
- **AND** 要求把业务实现下沉到 `multi-tenant` 插件内部
