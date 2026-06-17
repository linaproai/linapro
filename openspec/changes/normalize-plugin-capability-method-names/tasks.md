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

## Feedback

- [x] **FB-1**: `TC005` 仍按数字断言 `notifycap.MessageID`，与插件能力字符串 ID 契约不一致。
- [x] **FB-2**: 动态插件安装入口在已缓存 manifest 后未重新读取被替换的 staging artifact，低版本回退文件可能被错误放行。

执行记录：

- 根因：`notifycap.MessageID` 是插件能力层的字符串 ID，而 `TC005` 仍按 JSON number 断言；动态插件公开安装入口先读取 desired manifest 再调用低层安装逻辑，低层逻辑只在未传入 manifest 时失效 catalog 缓存，导致 staging artifact 被直接替换为低版本时仍可能使用旧高版本 manifest 完成降级校验。
- 修复：`TC005` 改为验证 `messageId` 可解析为正整数；`lifecycle.Install` 在读取 desired manifest 前按插件 ID 失效 manifest 缓存，使人工安装动作重新读取当前 staging artifact；单元测试改为覆盖公开 `Lifecycle.Install` 入口，并在替换 artifact 前显式建立 desired manifest 缓存。
- 影响分析：`i18n` 无影响，未新增或修改运行时文案、错误码、API 文档源文本、插件清单或语言包资源；缓存一致性有影响但范围限定为插件 manifest 读模型缓存，权威源为 staging artifact 文件，失效点为人工安装入口，单机和集群下均在本次安装请求内重新读取，不改变 runtime cache、权限快照或 datahost 缓存策略；数据权限无影响，未改变数据读写接口、租户边界、授权关系或 host service/data service 数据访问；开发工具跨平台无影响，未修改 Makefile、脚本、CI 或 `linactl`；测试策略为 Go 单元回归、TypeScript 静态检查、Playwright 用例枚举和 OpenSpec 严格校验。
- 验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/lifecycle -count=1` 通过；`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/lifecycle ./internal/cmd/internal/httpstartup -count=1 -p=1` 通过；`cd apps/lina-core && go test ./internal/cmd -count=1` 通过；`pnpm -C hack/tests exec tsc --noEmit --pretty false` 通过；`pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC005-runtime-wasm-host-services-low-priority.ts e2e/extension/plugin/TC008-runtime-wasm-lifecycle-boundaries.ts --list` 枚举到 5 个相关测试；`openspec validate normalize-plugin-capability-method-names --strict` 通过；`git diff --check` 通过。
- E2E 运行说明：本地完整执行目标 E2E 被默认端口占用阻断，`9120/5666` 当前由另一个工作区 `/Users/john/Workspace/github/gqcn/agentbox` 的 `lina` 和 `node` 进程监听；未停止外部工作区进程，未修改本仓库开发端口配置。
