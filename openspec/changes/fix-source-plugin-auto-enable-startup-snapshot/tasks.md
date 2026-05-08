## 1. 规范与定位

- [x] 1.1 新建 OpenSpec 变更并记录源码插件自动启用启动快照一致性要求
- [x] 1.2 复查源码插件自动启用链路，确认启动快照陈旧是 `Plugin is not installed` 的根因

## 2. 后端修复

- [x] 2.1 修复源码插件稳定状态写入后的启动快照同步
- [x] 2.2 增加携带启动快照的 `plugin.autoEnable` 源码插件自动安装并启用回归测试

## 3. 验证与审查

- [x] 3.1 运行相关后端单元测试，覆盖 `apps/lina-core/internal/service/plugin`
- [x] 3.2 明确记录 i18n 影响判断：本变更不涉及运行时语言包、插件 manifest i18n 或 apidoc i18n
- [x] 3.3 明确记录缓存一致性判断：本变更只同步单次启动上下文快照，不新增业务缓存
- [x] 3.4 调用 `lina-review` 完成代码与规范审查

## Feedback

- [x] **FB-1**: 插件管理列表中 `plugin.autoEnable` 标识不应因名称列换行而被单元格裁剪

## 实施记录

- 根因：HTTP 启动上下文已携带 catalog 启动快照，源码插件自动安装后 `applySourcePluginStableState` 只更新数据库，未刷新当前快照；随后的启用检查从快照读取到 `installed=0`，触发 `Plugin is not installed`。
- 修复：源码插件稳定状态写入成功后调用 `RefreshStartupRegistry`，让同一启动上下文后续启用、路由接线和预热阶段读取最新 registry 投影。
- 测试：新增 `TestBootstrapAutoEnableSourcePluginUpdatesStartupSnapshot`，先构造 `WithStartupDataSnapshot` 再执行 `BootstrapAutoEnable`，覆盖真实 HTTP 启动共享快照路径。
- 已运行：`go test ./internal/service/plugin -run 'TestBootstrapAutoEnableSourcePluginUpdatesStartupSnapshot|TestBootstrapAutoEnableInstallsAndEnablesSourcePlugin|TestBootstrapAutoEnableHonorsPerEntryMockDataOptIn' -count=1`。
- 已运行：`go test ./internal/service/plugin -count=1`。
- 已运行：`go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/lifecycle -count=1`。
- 已运行：`go test ./internal/service/plugin/internal/datahost ./internal/service/plugin/internal/frontend ./internal/service/plugin/internal/openapi ./internal/service/plugin/internal/sourceupgrade ./internal/service/plugin/internal/wasm -count=1`。
- 已尝试：`go test ./internal/service/plugin/... -count=1`。该命令失败在既有跨包测试夹具污染：`internal/catalog` 的重复插件 ID 测试临时创建 `apps/lina-plugins/plugin-duplicate-id/plugin.yaml`，根包自动启用测试并发扫描真实插件目录时与 `plugin-demo-source` 冲突；分包单独运行均通过，本次修复路径测试通过。
- i18n 影响判断：本变更不新增、修改或删除用户可见前端文案、菜单、按钮、表单、表格、API DTO 文档源文本、插件 manifest i18n 或 apidoc i18n 资源；无需同步运行时语言包。
- 缓存一致性判断：本变更只维护单次 HTTP 启动编排 context 内的短生命周期启动快照，权威数据源仍为数据库；不新增跨请求、跨进程或分布式业务缓存。集群模式仍按现有拓扑由主节点执行生命周期写入，从节点轮询共享数据库状态收敛。
- `lina-review` 审查结论：本变更只修改源码插件生命周期状态写后快照同步和对应单元测试；未修改 API DTO、SQL、前端 UI、运行时 i18n 或 apidoc i18n 资源；未修改生成代码；未新增数据操作接口，不涉及角色数据权限接入变化。新增启动快照刷新仅服务单次启动 context，不改变跨实例一致性模型。

## Feedback 实施记录

- FB-1 根因：插件管理列表名称列使用可换行的 `Space wrap` 同时承载插件名称和 `plugin.autoEnable` 标识；当名称和标识总宽超过名称列时，标识会换到第二行，而 VXE 单元格高度固定且 `overflow: hidden`，导致标识底部存在被裁剪风险。
- FB-1 修复：名称列改为不换行的 `inline-flex` 布局，插件名称与自动启用标识均固定为 `shrink-0`，并将名称列最小宽度提升到 280px，确保标识不参与换行布局。
- FB-1 测试：更新 `TC0104-plugin-auto-enable-management-hints.ts`，使用较长插件名称模拟风险场景，断言自动启用标识与名称同一行、未越出 VXE 单元格上下边界。
- FB-1 已运行：`pnpm exec playwright test e2e/extension/plugin/TC0104-plugin-auto-enable-management-hints.ts --workers=1`，结果 2 passed。
- FB-1 已使用 Playwright 手动验证：在插件管理页拦截插件列表响应并模拟长名称，测得 `white-space=nowrap`、`clipped=false`，局部截图显示名称与“自动启用”标识同一行完整显示。
- FB-1 i18n 影响判断：本次只调整现有布局和测试，复用已有 `pages.system.plugin.autoEnableBadge` 文案，不新增、修改或删除运行时语言包、插件 manifest i18n 或 apidoc i18n 资源。
- FB-1 缓存一致性判断：本次仅修改前端展示布局和 E2E 测试，不涉及后端数据、缓存权威源、失效策略或集群一致性模型。
