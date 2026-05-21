## 1. Build Tool Consolidation

- [x] 1.1 将 `hack/tools/image-builder` 实现迁移到 `hack/tools/linactl/internal/imagebuilder`，并改造为可由 linactl 直接调用的内部包
- [x] 1.2 更新 `linactl image` 与 `linactl image.build` 命令，移除对子进程 `go run ./hack/tools/image-builder` 的依赖
- [x] 1.3 将 `hack/tools/build-wasm/internal/builder` 实现迁移到 `hack/tools/linactl/internal/wasmbuilder`
- [x] 1.4 更新 `linactl wasm` 命令，直接调用 wasmbuilder 并保留插件工作区准备、dry-run 和输出目录语义
- [x] 1.5 更新 `linactl/go.mod` 与根 `go.work`，移除旧独立工具模块并维护 pluginbridge 相关依赖
- [x] 1.6 将 `hack/tools/runtime-i18n` 治理扫描组件合并到 linactl 内部组件 `runtimei18n`
- [x] 1.7 将 imagebuilder、plugins 和 wasmbuilder 内部组件源码文件统一为组件名前缀命名
- [x] 1.8 更新 CI 夹具、测试辅助、E2E 用例和仓库工具文档中的旧路径引用
- [x] 1.9 删除 `hack/tools/image-builder`、`hack/tools/build-wasm` 与 `hack/tools/runtime-i18n` 旧独立工具目录
- [x] 1.10 迁移或补齐镜像构建、Wasm 打包和 i18n 治理扫描相关单元测试
- [x] 1.11 修复 GitHub Actions `plugin-command-smoke` 隔离夹具缺少 linactl 编译所需的 lina-core 本地替换模块

## 2. Development Environment

- [x] 2.1 新增 `linactl env.check` 命令，检测 Go、Node.js、pnpm、Vite、Playwright、PostgreSQL 并输出带边框的 ASCII 表格
- [x] 2.2 PostgreSQL 检测改为读取 `apps/lina-core/manifest/config/config.yaml` 中的 `database.default.link`，通过 Go `database/sql` 连接执行 `SHOW server_version`
- [x] 2.3 将原 `linactl dev.setup` 实现迁移为 `linactl env.setup`，保持前端依赖与 Playwright Chromium 安装行为一致
- [x] 2.4 从 linactl 注册表和根 Make 目标中移除 `dev.setup`，并新增 `env.check` / `env.setup` Make 目标
- [x] 2.5 更新 Playwright 缺失浏览器提示和相关文档/帮助输出，统一指向 `make env.setup`
- [x] 2.6 新增或更新 linactl 单元测试，覆盖环境命令注册、旧命令移除、环境检查表格和命令文件命名治理
- [x] 2.7 将 linactl 子命令的工具查找依赖收敛为 `app.lookPath`，测试可注入替身

## 3. Release Governance

- [x] 3.1 在 linactl 中新增 `release.tag.check` 命令，读取 `metadata.yaml` 并校验 tag 与 `framework.version`
- [x] 3.2 为 tag 相等、tag 不匹配、版本格式非法、缺少版本字段和环境变量 fallback 增加单元测试
- [x] 3.3 在 `Release Test and Build` workflow 中新增最早执行的 release tag 版本一致性 job
- [x] 3.4 将所有 release 测试和镜像发布 job 依赖版本一致性 job
- [x] 3.5 新增受控 `Create Release Tag` 手动 workflow，通过 GitHub App installation token 创建匹配 framework.version 的 tag
- [x] 3.6 在发布文档中说明 GitHub tag ruleset 配置建议
- [x] 3.7 新增仅支持 `workflow_dispatch` 的手动 nightly 镜像发布 workflow
- [x] 3.8 新增 `hack/deploy/docker-compose.yaml` 内存态 nightly 演示启动入口
- [x] 3.9 新增 `hack/deploy/config.yaml` 独立运行时配置
- [x] 3.10 将演示部署从 SQLite 切换为 Compose 内 PostgreSQL 服务
- [x] 3.11 新增 `hack/deploy/tests/docker-compose.yaml` 部署测试 Compose 开发容器
- [x] 3.12 修复 nightly 插件完整镜像中 Turbo strict env 模式未包含 `LINAPRO_SOURCE_PLUGINS` 环境变量
- [x] 3.13 修复 nightly 镜像中 auto-enabled tenant-scoped 插件缺少租户启用行
- [x] 3.14 修复 Tailwind v4 全局 CSS 构建扫描范围未包含源码插件目录

## 4. Release Test and Build

- [x] 4.1 将 release workflow 调整为 `.github/workflows/release-test-and-build.yml`，复用共享测试验证套件
- [x] 4.2 Release 采用 Main CI 的简要测试范围（不含 E2E），镜像发布等待 tag 校验和共享验证套件成功
- [x] 4.3 在 release workflow 中新增 GitHub Release 创建 job，标题为 `LinaPro Release <tag>`
- [x] 4.4 从 `.github/workflows` 中抽取独立 reusable workflow 和 composite action，减少 nightly/release 编排漂移
- [x] 4.5 补齐 release 的 host-only 与 plugin-full 验证矩阵
- [x] 4.6 为所有 workflow 的关键配置、每个 job 和每个内联 step 补充中英文注释
- [x] 4.7 将 `make image-build` 改为 `make image.build`，同步 linactl 和文档引用
- [x] 4.8 新增可复用 Linux `plugin-command-smoke` 和 `make-command-smoke` workflow
- [x] 4.9 将主 CI 中复杂 job 拆为独立 reusable workflow
- [x] 4.10 修复 backend unit test workflow 中不存在的 `prepare-packed-assets` Make 目标
- [x] 4.11 将资源打包统一为 `make pack.assets` 根目录入口
- [x] 4.12 让 CI 中直接调用 linactl 的步骤改为使用 make 入口
- [x] 4.13 补齐 OpenAPI apidoc 中文翻译资源缺少的共享 DTO 结构化 key
- [x] 4.14 将宿主 `en-US` apidoc 恢复为空对象占位
- [x] 4.15 修复 Make command smoke 的 Linux CI fixture 缺少 linactl 本地替换依赖
- [x] 4.16 修复 E2E workflow 中 npm/pnpm Playwright revision 不一致
- [x] 4.17 将共享验证套件的每个 job 提供独立必填 input 开关
- [x] 4.18 将共享状态密集的 E2E 用例纳入串行隔离清单，parallel-workers 显式使用 1
- [x] 4.19 修复 host-only E2E 中字典类型删除和用户批量编辑的 retry 状态丢失问题
- [x] 4.20 修复 host-only E2E 中多租户能力误判和 loading 遮罩问题
- [x] 4.21 修复 host-only E2E 中角色权限不足导致页面 403 卡住
- [x] 4.22 修复 Go unit tests 中 linactl test.go 使用 `-p=1` 串行化 package 执行
- [x] 4.23 修复 host-only E2E 中租户态 localStorage 不持久的问题
- [x] 4.24 拆分过长 E2E 流程并减少非断言目标的重复 UI 导航/清理
- [x] 4.25 将 Playwright 全局默认 test timeout 从 60 秒调整为 180 秒
- [x] 4.26 修复 Go unit tests 中 runDev 单测缺少 pnpm 时测试失败
- [x] 4.27 在 Main CI 中新增 `pull_request` 触发

## 5. Go Unit Test Runtime Optimization

- [x] 5.1 为 `linactl test.go` 增加 Go package 测试发现能力，区分含 `_test.go` 的包和无测试包
- [x] 5.2 调整 `linactl test.go` 默认执行计划，只对含测试文件的包执行单元测试
- [x] 5.3 为 `linactl test.go` 增加模块耗时统计、race/verbose 状态和无测试包数量摘要
- [x] 5.4 将普通插件 runtime/cron 逻辑测试改为使用 synthetic artifact、fake executor 或轻量 fixture
- [x] 5.5 为真实 bundled dynamic Wasm 样例保留最小 smoke 覆盖，并复用一次性 artifact fixture
- [x] 5.6 修正 GitHub Actions PostgreSQL health check，显式使用 `postgres` 用户和 `linapro` 数据库
- [x] 5.7 评估并启用安全的 Go module/build cache 配置
- [x] 5.8 修复 `linactl test.go` 的 Go workspace/module discovery 改为分别捕获 stdout 与 stderr

## 6. Login Home SQL Optimization

- [x] 6.1 优化 DB-only 在线会话校验，减少有效请求的 `sys_online_session` 查询次数
- [x] 6.2 为会话校验补充单元测试，覆盖有效、过期、租户不匹配和近期活跃场景
- [x] 6.3 优化插件 catalog release 读取复用，减少同一请求或列表投影内的 `sys_plugin_release` 重复查询
- [x] 6.4 为插件 release 读取复用补充单元测试
- [x] 6.5 `/plugins/dynamic` 运行态列表创建请求级 catalog 快照，避免每个插件重复查询 `sys_plugin_release`

## 7. Verification

- [x] 7.1 运行 `cd hack/tools/linactl && go test ./... -count=1`
- [x] 7.2 运行 `cd hack/tools/linactl && go run . test.scripts`
- [x] 7.3 运行 `openspec validate` 对所有相关变更进行严格校验
- [x] 7.4 运行 `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/*.yml`
- [x] 7.5 运行受影响的插件 runtime/integration 包测试，保留 `-race` 覆盖
- [x] 7.6 运行 `cd apps/lina-core && go test ./internal/service/session -count=1`
- [x] 7.7 运行覆盖插件 catalog 变更包的 Go 测试
- [x] 7.8 运行 `cd apps/lina-core && go test ./internal/cmd -count=1` 或等价启动绑定编译烟测
- [x] 7.9 记录 i18n、缓存一致性、数据权限和开发工具跨平台影响评估，并执行 lina-review 审查
