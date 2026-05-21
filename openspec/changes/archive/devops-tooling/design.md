## Context

LinaPro 的开发工具链由 `hack/tools/linactl` 统一承载跨平台开发命令。`linactl image` 和 `linactl image.build` 仍通过 `go run ./hack/tools/image-builder` 调用独立工具；`linactl wasm` 仍通过进入 `hack/tools/build-wasm` 目录执行 `go run .` 打包动态插件；`linactl i18n.check` 仍通过进入 `hack/tools/runtime-i18n` 目录分别执行扫描和消息覆盖检查。

开发环境初始化入口 `make dev.setup` 语义与 `make dev` 启动开发服务混在一起，缺少轻量环境检查入口。release 发布链路缺少 tag 与框架版本一致性门禁，tag push 后才运行检测无法阻止错误标签创建。Go 单元测试总体耗时偏高，主要原因包括真实动态 Wasm 执行、完整插件生命周期、共享 PostgreSQL 状态和大量无测试包扫描。登录后首页触发 10 个并发请求共执行 152 条 SQL，其中 `sys_plugin_release` 查询 85 次、`sys_online_session` 查询 18 次。

## Goals / Non-Goals

**Goals:**

- 将镜像构建、Wasm 打包和 i18n 治理扫描实现收敛到 `linactl/internal/` 子组件，保持公开命令入口不变。
- 提供 `env.check` 轻量环境检查和 `env.setup` 环境初始化入口，移除旧 `dev.setup`。
- 在 release 发布链路中统一校验 Git tag 与 `framework.version`，提供受控打标入口。
- 让 release、nightly 和 main CI 复用共享测试验证套件，减少编排漂移。
- 保留 `-race` 并发安全检测，优化测试设计和重复重型 fixture 治理。
- 减少登录后首页运行期重复 SQL，优化会话校验和插件 catalog 读取路径。
- 让 E2E 测试执行更稳定，调整超时和并行策略。

**Non-Goals:**

- 不改变镜像构建参数、Docker 构建语义、动态插件 artifact 格式或插件运行时桥接协议。
- 不新增后端运行时 REST API、数据库迁移或前端页面。
- 不引入新的第三方测试框架或外部服务依赖。
- 不重构登录、菜单、消息、租户或插件管理 API 契约。
- 不通过代码自动修改 GitHub 仓库 settings 或 ruleset。

## Decisions

### 1. 使用 linactl/internal 独立包承载工具实现

镜像构建实现放入 `hack/tools/linactl/internal/imagebuilder`，动态插件打包实现放入 `hack/tools/linactl/internal/wasmbuilder`，运行时 i18n 治理扫描实现放入 `hack/tools/linactl/internal/runtimei18n`。命令文件仍只负责编排参数、插件工作区准备和输出，复杂实现收敛到内部包。

### 2. 允许 linactl 编译期依赖 lina-core/pkg/pluginbridge

`wasmbuilder` 需要使用 `pluginbridge` 中的 artifact section、生命周期、路由和 host service contract 定义。将这部分依赖保留为编译期依赖可以消除子进程工具边界。

### 3. 环境检查只做工具级 smoke 检测

`env.check` 通过 `exec.LookPath` 和版本命令读取当前版本，使用 `text/tabwriter` 输出表格。PostgreSQL 检测通过 Go `database/sql` 连接配置的数据库执行 `SHOW server_version`，不依赖 `psql` 客户端工具。

### 4. Release tag 校验通过跨平台 Go 工具复用

`linactl release.tag.check` 读取 `metadata.yaml` 的 `framework.version`，默认 tag 来源按顺序使用 `tag=<value>` 参数、`GITHUB_REF_NAME` 环境变量。Release tag 格式限定为 Docker tag 兼容的 SemVer 子集：`vMAJOR.MINOR.PATCH` 和 `vMAJOR.MINOR.PATCH-prerelease`。

### 5. 受控打标 workflow 使用 GitHub App installation token

受控打标 workflow 通过 `actions/create-github-app-token@v3` 使用仓库变量 `RELEASE_APP_CLIENT_ID` 和仓库密钥 `RELEASE_APP_PRIVATE_KEY` 生成 GitHub App installation token。ruleset bypass 必须配置到该 GitHub App actor。

### 6. Release 复用共享测试验证套件

Release workflow 通过 `reusable-test-verification-suite.yml` 编排验证阶段，采用与 Main CI 一致的不含 E2E 的简要测试开关。`release-image` 必须声明依赖 release tag 校验和 `verification-suite`。

### 7. 保留 race，优化测试工作量

`-race` 能发现普通断言无法覆盖的并发数据竞争。优化重点放在减少被 race 放大的重复重型工作：将真实 bundled dynamic Wasm 执行收敛为 smoke，插件测试 fixture 复用且按作用域隔离，`linactl test.go` 先发现测试包再执行测试计划。

### 8. 会话校验改为一次读取再判断

DB-only 校验改为读取一条 `sys_online_session` 记录并检查租户和 `last_active_time`，仅在过期时删除，仅在超出写入节流窗口时更新。常规有效请求减少到 1 次 SELECT。

### 9. 插件 release 读取使用请求级快照

插件 release 表是小表，为了避免在集群模式下引入跨节点陈旧状态，本轮不直接增加长 TTL 进程级缓存。优先使用请求级快照或一次性批量读取，在单个首页列表投影内复用 release 行。

### 10. E2E 执行策略调整

将 Playwright 全局默认 `test.timeout` 从 60 秒调整为 180 秒，`expect.timeout` 保持 10 秒。将共享状态密集的 IAM 角色、IAM 用户、文件管理等用例纳入 E2E 串行隔离清单，`parallel-workers` 显式使用 1。

## Risks / Trade-offs

- linactl 依赖变重 → 通过接受项目专用工具定位，并在 `linactl/go.mod` 中显式维护依赖。
- 旧路径引用遗漏 → 通过 `rg` 扫描旧工具路径引用，并更新测试、文档与 CI 夹具。
- 真实 Wasm 用例减少后可能遗漏打包产物与宿主桥接组合问题 → 保留最小 smoke，并在 fixture 测试中覆盖解析、分发、授权、cron 和错误路径。
- 会话校验改为读取记录后再判断，必须确保租户不匹配仍拒绝、过期会话仍删除。
- 请求级 release 快照只能减少单次请求内重复读取，不能消除所有跨请求重复。
- release 不运行完整 E2E，可能晚于 nightly 才发现浏览器回归 → nightly 继续启用完整 E2E。
- 如果未配置 GitHub tag ruleset，人工仍可直接创建错误 tag → 现有 release workflow 前置校验会阻止错误 tag 发布制品。

## Migration Plan

1. 迁移 image-builder、build-wasm、runtime-i18n 源码到 linactl/internal/ 子组件。
2. 新增 env.check 与 env.setup 命令，移除 dev.setup。
3. 新增 release.tag.check 命令和受控打标 workflow。
4. 将 release workflow 调整为复用共享测试验证套件。
5. 优化 Go 单元测试 fixture 和测试发现策略。
6. 优化会话校验和插件 catalog 读取路径。
7. 调整 E2E 超时和并行策略。
8. 运行 linactl 单元测试、命令 smoke、OpenSpec 校验和静态引用扫描。
