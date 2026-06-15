## ADDED Requirements

### Requirement: 插件领域能力 Go 方法命名必须避免重复领域名

系统 SHALL 为 `pkg/plugin/capability/*cap` 下的公开 Go 接口使用动作式命名。当 `Service` 所在目录已经表达主资源时，主资源方法 MUST 省略重复领域名，优先使用 `BatchGet`、`Search`、`Current`、`EnsureVisible`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled` 等短名。仅当方法目标是子资源、复合对象或短名会造成歧义时，才允许保留限定词，例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`、`DeleteBySource`。动态 `host service` 的 wire method 名称 MAY 保持显式原样，不受本规则影响。

#### Scenario: 主资源方法使用动作名

- **WHEN** 插件通过 `usercap.Service`、`filecap.Service`、`sessioncap.Service` 或 `plugincap.Service` 调用主资源能力
- **THEN** 方法名 SHOULD 直接表达动作，而不是重复资源名
- **AND** 例如 `Users().Search`、`Files().EnsureVisible`、`Sessions().BatchGet`、`Plugins().BatchGet` 这类命名方式是允许的

#### Scenario: 子资源方法保留限定词

- **WHEN** 方法目标是权限、租户关系、部门树、来源维度或其他复合对象
- **THEN** 方法名 MAY 保留限定词以避免歧义
- **AND** 例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`、`DeleteBySource` 这类命名方式仍然符合规范

#### Scenario: 动态 wire method 名称保持显式

- **WHEN** 动态插件通过 `pluginbridge` 或 WASM host service 调用宿主领域能力
- **THEN** wire method 名称 SHOULD 继续保持显式资源名，例如 `users.batch_get`、`messages.batch_get`
- **AND** 该 wire 命名不要求与 typed Go 方法名完全一致
