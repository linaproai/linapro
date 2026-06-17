## Why

插件领域能力的公开 Go 方法仍有不少"操作 + 领域名"的重复命名，例如 `Users.SearchUsers`、`Files.EnsureFilesVisible`。在能力目录已经表达主资源的前提下仍然重复领域名，会让插件调用面更啰嗦，也会把后续扩展成本反复带回主框架。

## What Changes

- 重命名 `pkg/plugin/capability/*cap` 中主资源能力方法为动作式短名，例如 `BatchGet`、`Search`、`EnsureVisible`、`Current`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled`。
- 保留子资源或歧义方法限定词，例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`。
- 保持动态 host service 的 wire method 名称不变，例如 `users.batch_get`、`files.visible.ensure`。
- 同步更新宿主适配器、pluginbridge 代理、WASM 分发器、插件调用点、测试替身和文档。

## Impact

- 影响 `apps/lina-core/pkg/plugin/capability/*cap`、`capabilityhost`、`pluginbridge/internal/domainhostcall`、`wasm`、多个 `apps/lina-plugins` 源码插件和文档。
- 不修改动态 host service wire 字符串、数据权限语义或缓存语义。
- 不为旧方法保留兼容别名。
