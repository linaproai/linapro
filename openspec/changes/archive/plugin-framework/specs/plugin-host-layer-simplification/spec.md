# 插件宿主分层简化

## Purpose
定义 host service JSON 载荷政策、dedicated codec 冻结、wire 常量单一来源与 lifecycle 升级归属。

## Requirements

### Requirement: 新增 host service 方法必须使用 JSON envelope

系统 SHALL 要求新增 core-owned 动态 host service 方法使用统一 JSON envelope（`HostServiceJSONRequest` / `HostServiceJSONResponse` 或等价空载荷）。系统 MUST NOT 为新增方法引入 dedicated binary payload codec。

#### Scenario: 新增普通领域方法

- **WHEN** 开发者为已有或新建 core-owned host service 增加可发布方法
- **THEN** catalog 将该方法标记为 `PayloadKindJSON` 或 `PayloadKindNone`
- **AND** guest 与 wasm dispatcher 通过 JSON envelope 编解码请求响应
- **AND** 不得新增 `protocol_hostservice_*_codec.go` dedicated 编解码类型给该方法

### Requirement: dedicated codec 方法集合必须冻结

系统 SHALL 维护方法级 dedicated codec 冻结名单。仅名单内存量方法允许 `PayloadKindDedicated`；名单外出现 dedicated 载荷 MUST 使治理测试失败。

#### Scenario: 冻结名单外出现 dedicated 方法

- **WHEN** catalog 中某个方法的 `PayloadKind` 为 `dedicated`
- **AND** 该方法不在冻结名单中
- **THEN** catalog 治理测试失败
- **AND** 失败信息包含 `service.method` 标识

### Requirement: host service wire 常量单一来源

系统 SHALL 在 `protocol/hostservices` 中维护唯一的 host service/method wire Go 常量，catalog 条目 MUST 引用这些常量而不是重复 wire 字符串字面量。治理测试 MUST 校验 catalog 与常量一致；系统 MUST NOT 依赖编译期 `go generate` 生成这些常量。

#### Scenario: catalog 使用 wire 常量

- **WHEN** 运行 hostservices catalog 治理测试
- **THEN** catalog 中每个 service/method wire 值等于对应 `HostService*` / `HostServiceMethod*` 常量值
- **AND** 公开 `protocol` re-export 名称保持 `HostService*` / `HostServiceMethod*` 兼容

### Requirement: lifecycle 门面拥有升级编排

系统 SHALL 由 lifecycle 门面拥有源码升级与动态 runtime 升级的构造与对外编排入口。根 plugin facade MUST NOT 平行持有独立的 upgrade service 字段作为第二生命周期 owner。

#### Scenario: 根门面执行升级

- **WHEN** 管理面调用插件升级预览或执行
- **THEN** 根 plugin facade 委托 lifecycle 拥有的升级能力
- **AND** 不在根 service 结构中保留独立 `upgradeSvc` 装配字段
