## ADDED Requirements

### Requirement: Host call 授权快照可以请求内复用但不得改变治理语义

系统 SHALL 允许`WASM`host service handler 在同一次 guest 执行中复用已构建的 host service 授权快照。复用 MUST 仅降低快照装配成本，不得改变当前 active release 授权来源、service/method/resource 校验、数据权限、租户边界、审计字段或错误 envelope。

#### Scenario: 同一次 guest 执行复用授权快照

- **WHEN** 动态插件在一次路由请求中连续调用多个 host service
- **THEN** 宿主可以复用本次`ExecuteBridge`入口构建的授权快照
- **AND** 每次 host call 仍校验 service、method 和资源标识是否已授权

#### Scenario: 授权收缩后新请求使用新快照

- **WHEN** 插件 active release 的 host service 授权被收缩并发布`plugin-runtime`修订号
- **THEN** 后续 guest 执行不得继续使用旧请求中的授权快照
- **AND** 未授权 service、method 或资源调用必须被拒绝

#### Scenario: 系统型调用不伪造用户上下文

- **WHEN** 动态插件在生命周期、hook 或 cron 中调用需要用户上下文的 host service
- **THEN** 即使授权快照命中，handler 也必须按领域契约拒绝或按系统调用边界处理
- **AND** 不得伪造请求型用户身份来绕过数据权限
