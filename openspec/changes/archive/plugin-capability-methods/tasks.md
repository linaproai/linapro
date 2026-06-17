# Tasks

## Summary

- [x] 重命名 `usercap`、`filecap`、`jobcap`、`sessioncap`、`plugincap`、`notifycap` 的公开接口方法和注释
- [x] 重命名对应 `capabilityhost` 实现，保持数据权限、缓存和错误语义不变
- [x] 重命名 `pluginbridge/internal/domainhostcall`、`wasm` 和宿主测试替身中的方法名
- [x] 重命名 `linapro-content-notice`、`linapro-org-core`、`linapro-tenant-core`、`linapro-monitor-online` 的调用点
- [x] 更新 `localdocs/plugin-domain-capability-expansion-design.md`、`pkg/plugin` README 和 `.agents/rules/backend-go.md`

## Verification

- [x] `openspec validate normalize-plugin-capability-method-names --strict` 通过
- [x] `gofmt` 和静态命名扫描通过
- [x] 受影响 Go 测试通过
- [x] `lina-review` 审查通过

## Governance

- [x] i18n：无运行时用户可见文案变化
- [x] 不修改动态 host service wire 字符串
- [x] 不为旧方法保留兼容别名
