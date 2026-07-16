## ADDED Requirements

### Requirement: storagecap.Provider 必须支持可选的客户端直连访问能力

系统 SHALL 允许 `storagecap.Provider` 实现可选的客户端直连访问能力（能力探测 + 创建直连访问）。未实现或不支持的 provider（含内置 local）MUST 被探测为不支持，由上层降级为服务端 `Put`/`Get` 中转。Provider 在签发直连访问时 MUST 只针对宿主传入的 scoped object key 与操作约束，MUST NOT 解释插件 logical path 或动态 hostServices 授权快照。

#### Scenario: local provider 不支持直连

- **WHEN** 当前 active provider 为内置 local
- **AND** 调用方探测 put/get 直连能力
- **THEN** 探测结果 MUST 为不支持
- **AND** 插件与文件中心 MUST 可继续通过服务端中转完成读写

#### Scenario: 云 provider 为 scoped key 签发 put 访问

- **WHEN** 唯一可服务云 provider 支持直连 put
- **AND** 宿主传入合法 scoped object key 与 size/content-type 约束
- **THEN** provider MUST 返回中立 DirectAccess 描述
- **AND** 该访问 MUST 无法用于修改约束 key 之外的对象（在云侧策略与签名能力范围内）

### Requirement: storagecap.Service 必须提供直连 put/get 与确认语义

系统 SHALL 在 `storagecap.Service` 上提供创建直连 put 访问、确认 put 完成、以及创建直连 get 访问的能力（方法名以实现为准）。创建访问时 MUST 将插件 logical path 映射为含插件 ID 与租户维度的 scoped key。确认 put 时 MUST 校验对象存在后，使后续 `Get`/`Stat` 可见。Service MUST NOT 因直连而向插件返回 provider 私有 key 或永久凭证。

#### Scenario: 插件直连 put 后可 Stat

- **WHEN** 插件对 logical path `reports/a.bin` 创建直连 put 访问
- **AND** 客户端完成上传
- **AND** 插件确认 put 成功
- **THEN** 同一插件在同一租户下对该 path 的 `Stat`/`Get` MUST 可见该对象

#### Scenario: 不同插件 path 隔离在直连下仍生效

- **WHEN** 插件 A 直连写入 logical path `reports/a.bin`
- **AND** 插件 B 尝试对同一 logical path 创建 get 直连或 Get
- **THEN** 插件 B MUST NOT 读取到插件 A 的对象内容

### Requirement: 动态插件直连访问必须保留 path 授权

系统 SHALL 要求动态插件在创建直连 put/get 或确认 put 时，其 logical path 仍通过既有 `hostServices` storage path 授权校验。未授权 path MUST 在进入 provider 签发前被拒绝。

#### Scenario: 未授权 path 拒绝直连 init

- **WHEN** 动态插件仅被授权 `reports/`
- **AND** 插件请求对 `secrets/x` 创建直连 put
- **THEN** 宿主 MUST 拒绝
- **AND** provider MUST NOT 收到签发请求
