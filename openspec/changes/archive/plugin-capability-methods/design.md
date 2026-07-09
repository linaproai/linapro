## 命名规则

主资源能力方法使用动作名即可，例如 `BatchGet`、`Search`、`EnsureVisible`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled`。子资源或歧义方法保留限定词，例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`、`DeleteBySource`。

## 不变契约

- 动态 host service wire method 名称不变（`users.batch_get`、`files.visible.ensure` 等）
- 数据权限语义不变
- 缓存语义不变
- 不保留旧方法兼容别名
