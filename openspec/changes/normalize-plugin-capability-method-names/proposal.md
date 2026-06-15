## Why

当前插件领域能力的公开 Go 方法仍有不少“操作 + 领域名”的重复命名，例如 `Users.SearchUsers`、`Files.EnsureFilesVisible`。这些名字在能力目录已经表达主资源的前提下仍然重复领域名，会让插件侧调用面更啰嗦，也会把后续扩展成本反复带回主框架。

## What Changes

- **BREAKING** 重命名 `pkg/plugin/capability/*cap` 中主资源能力方法，改为动作式短名，例如 `BatchGet`、`Search`、`EnsureVisible`、`Current`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled`。
- 保留子资源、目标对象不同或短名会歧义的方法限定词，例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`。
- 保持动态 `host service` 的 wire method 名称不变，例如 `users.batch_get`、`files.visible.ensure`。
- 同步更新宿主适配器、`pluginbridge` 代理、WASM 分发器、插件调用点、测试替身、README、设计文档和项目规则。

## Capabilities

### Modified Capabilities

- `plugin-host-domain-capabilities`: 调整插件可消费领域能力的 Go 方法命名规则，要求主资源能力方法避免重复领域名。

## Impact

- `apps/lina-core/pkg/plugin/capability/*cap`
- `apps/lina-core/internal/service/plugin/internal/capabilityhost`
- `apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall`
- `apps/lina-core/internal/service/plugin/internal/wasm`
- `apps/lina-plugins/linapro-content-notice`
- `apps/lina-plugins/linapro-org-core`
- `apps/lina-plugins/linapro-tenant-core`
- `apps/lina-plugins/linapro-monitor-online`
- `apps/lina-core/pkg/plugin/README.md`
- `apps/lina-core/pkg/plugin/README.zh-CN.md`
- `localdocs/plugin-domain-capability-expansion-design.md`
- `.agents/rules/backend-go.md`
