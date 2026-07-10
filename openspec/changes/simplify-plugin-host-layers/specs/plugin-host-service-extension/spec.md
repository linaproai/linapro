## ADDED Requirements

### Requirement: Host service 载荷种类必须可治理

系统 SHALL 在 host-service catalog 中为每个方法记录 `PayloadKind`，并区分 JSON envelope 与 dedicated codec。治理测试 MUST 拒绝未授权的 dedicated 方法扩张。

#### Scenario: 校验 catalog payload kind

- **WHEN** 运行 hostservices catalog 治理测试
- **THEN** 每个已发布方法都有非空 `PayloadKind`
- **AND** dedicated 方法必须命中方法级冻结名单
- **AND** 普通 JSON 方法使用 `HostServiceJSONRequest` / `HostServiceJSONResponse` 语义
