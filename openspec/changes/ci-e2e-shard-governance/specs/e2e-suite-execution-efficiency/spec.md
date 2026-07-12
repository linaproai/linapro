## MODIFIED Requirements

### Requirement: Plugin-full E2E 必须支持通用插件入口分片执行
E2E CI workflow SHALL allow plugin-full browser regression to execute plugin management, plugin-owned tests, and plugin host seam tests as independent shards while preserving the same plugin-full startup semantics. Plugin-owned shards SHALL be planned from discovered source-plugin test ownership with load-balanced packing, not only equal Playwright file-count sharding of the whole `plugins` tree.

#### Scenario: Plugin-full 分片选择模块范围
- **当** workflow 启动 plugin-full E2E 分片
- **则** 每个分片必须使用 plugin-full 服务启动命令
- **且** 源码插件自有测试分片必须通过 CI 分片规划选择一个或多个 `plugins/<plugin-id>` 通用入口集合
- **且** 根目录分片只能选择宿主插件框架通用测试范围，不得选择依赖具体官方源码插件业务实现的根测试别名集合
- **且** 分片日志必须显示选择的 scope/entries、并行文件数、串行文件数和串行隔离类别

#### Scenario: Plugin-full 不维护官方插件业务别名 scope
- **当** 开发者需要运行源码插件自有 E2E
- **则** runner 必须支持 `plugins` 运行全部源码插件自有用例
- **且** runner 必须支持 `plugin:<plugin-id>` 运行单个源码插件自有用例
- **且** E2E manifest 不应为官方插件业务模块维护长期别名 scope

#### Scenario: Plugin-full 分片失败阻止下游发布
- **当** 任一 plugin-full E2E 分片失败
- **则** 完整验证套件必须失败
- **且** 依赖完整验证成功的镜像发布或后续 job 不得执行

#### Scenario: Plugin-full 分片上传独立诊断证据
- **当** plugin-full E2E 分片完成或失败
- **则** workflow 必须上传该分片的 Playwright report、test-results、后端日志和前端日志
- **且** artifact 名称必须包含调用方前缀和分片标识，避免覆盖其他分片证据

#### Scenario: Plugin 分片按插件负载装箱
- **当** 官方插件工作区就绪且 CI 规划 plugin-full 源码插件分片
- **则** 规划器必须按每个源码插件的 TC 文件数（或显式 weightOverrides）装箱到目标分片数
- **且** 所有源码插件 TC 必须恰好属于一个 plugin 分片
- **且** 不得要求根目录硬编码官方插件业务模块别名

## ADDED Requirements

### Requirement: Host-only E2E 必须按能力边界 CI 分片执行
启用 host-only 浏览器 E2E 的完整验证 workflow SHALL 将宿主用例按 `execution-manifest` 中声明的 host CI 分片并行执行，而不是在单个 job 中串行跑完整 `pnpm test:host`。

#### Scenario: Host-only 分片覆盖全部宿主 TC
- **当** 验证器检查 host CI 分片声明
- **则** 所有 host-only 可发现的宿主 TC 必须恰好属于一个 host 分片
- **且** 分片之间不得重复包含同一 TC 文件

#### Scenario: Host-only 分片独立启动服务
- **当** workflow 运行某个 host CI 分片
- **则** 该 job 必须使用 host-only 启动命令
- **且** 必须通过 `ci-shard host <name>` 或等价入口仅执行该分片 entries
- **且** 必须上传带分片名的独立 artifact

#### Scenario: Host-only 分片保留 serial/parallel 治理
- **当** 某个 host 分片启动
- **则** runner 必须对该分片解析出的文件应用与全量相同的 serial/parallel 拆分
- **且** 并行池使用配置的 parallel workers，串行池保持单 worker

### Requirement: E2E CI 分片规划必须可本地复现并进入验证门禁
E2E 套件 SHALL 提供可在本地执行的 CI 分片规划与发射工具，并将分片完备性检查纳入 `pnpm test:validate`（或等价治理入口）。

#### Scenario: 本地发射 host/plugin matrix
- **当** 开发者运行 emit-ci-shards 工具
- **则** 输出必须包含每个分片的稳定 `name` 与可直接用于 CI 的 `command`
- **且** host 与 plugin-full-extra 分片不依赖未声明的临时 YAML 列表作为唯一来源

#### Scenario: 分片配置错误被验证器拒绝
- **当** host 分片漏覆盖、重复覆盖、引用空 entries 或非法 name
- **则** 验证必须失败并指出具体分片与文件
- **当** 插件工作区就绪且 plugin 装箱结果未覆盖全部源码插件 TC
- **则** 验证必须失败

### Requirement: 完整 E2E 默认并行池 worker 与安装缓存
启用完整浏览器 E2E 的 Nightly（及同类完整验证调用方）SHALL 使用不低于 2 的 E2E 并行池 worker 默认值，并在 E2E job 中缓存 Go modules 与 Playwright 浏览器安装产物。

#### Scenario: Nightly 使用提升后的 parallel workers
- **当** Nightly verification suite 运行 E2E
- **则** 传递给 E2E runner 的 parallel workers 默认值必须 ≥ 2
- **且** serial 池仍以单 worker 执行

#### Scenario: E2E job 复用工具链缓存
- **当** E2E job 安装 Go 依赖与 Playwright Chromium
- **则** workflow 必须启用 Go module cache
- **且** 必须缓存 Playwright 浏览器目录以降低重复安装成本
