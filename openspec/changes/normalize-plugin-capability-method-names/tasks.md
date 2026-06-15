## 1. 文档与规则

- [x] 1.1 更新 OpenSpec 提案、设计和增量规范，固化“主资源方法使用动作名、子资源保留限定词”的命名规则。
- [x] 1.2 更新 `localdocs/plugin-domain-capability-expansion-design.md`、`apps/lina-core/pkg/plugin/README.md`、`apps/lina-core/pkg/plugin/README.zh-CN.md` 和 `.agents/rules/backend-go.md`。

## 2. 宿主能力接口重命名

- [x] 2.1 重命名 `usercap`、`filecap`、`jobcap`、`sessioncap`、`plugincap`、`notifycap` 的公开接口方法和注释。
- [x] 2.2 重命名对应 `internal/service/plugin/internal/capabilityhost` 实现，保持数据权限、缓存和错误语义不变。

## 3. 代理与插件调用点同步

- [x] 3.1 重命名 `pluginbridge/internal/domainhostcall`、`internal/service/plugin/internal/wasm` 和宿主测试替身中的方法名。
- [x] 3.2 重命名 `linapro-content-notice`、`linapro-org-core`、`linapro-tenant-core`、`linapro-monitor-online` 的调用点与测试替身。

## 4. 验证与审查

- [x] 4.1 运行 `gofmt`、`openspec validate normalize-plugin-capability-method-names --strict` 和静态命名扫描。
- [x] 4.2 运行受影响 Go 测试并完成 `lina-review`。
