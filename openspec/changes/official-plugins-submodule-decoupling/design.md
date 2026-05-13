## Context

当前主仓库直接包含 `apps/lina-plugins`，并且多个入口默认该目录一定存在：

- `go.work` 直接 `use ./apps/lina-plugins` 以及每个官方插件模块。
- 宿主后端通过 `_ "lina-plugins"` 在编译期引入官方源码插件聚合模块。
- 插件清单扫描默认从 `repoRoot/apps/lina-plugins` 发现源码插件。
- 前端 Vite 配置与插件页面注册逻辑按 `lina-plugins/<plugin-id>/frontend/pages` 扫描页面。
- Playwright 配置和测试治理脚本把源码插件自有 E2E 固定在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e`。
- `linactl wasm` 和部分测试辅助按固定路径读取动态插件或 mock SQL。

如果直接把 `apps/lina-plugins` 改成 submodule，而不先处理这些耦合，clone 后未初始化 submodule、CI host-only 任务、纯宿主开发和插件仓库版本切换都会出现路径或编译错误。目标不是删除官方插件能力，而是让“宿主独立运行”和“插件完整联动”成为两个显式、可验证的模式。

## Goals / Non-Goals

**Goals:**

- 支持 `apps/lina-plugins` 不存在、为空目录或作为 submodule 存在三种工作区状态。
- 宿主基础命令在 host-only 状态下可运行：后端单元测试、前端类型检查/构建、宿主 E2E、插件管理空列表/空状态。
- 官方插件 submodule 初始化后，完整插件单元测试和插件 E2E 仍按既有插件目录约定运行。
- 将强耦合失败点转化为明确的跳过、空集合、可配置插件根或可操作错误提示。
- 保持 `apps/lina-plugins` 作为默认挂载路径，降低路径迁移风险。

**Non-Goals:**

- 不在本变更中把每个官方插件拆成独立 submodule。
- 不改变源码插件目录内部结构、插件 manifest 契约或插件业务实现。
- 不把官方插件发布方式改造成动态插件 marketplace。
- 不在本变更中引入新的数据库 schema。
- 不修改用户可见业务文案，除非 host-only 插件空状态已有缺失且必须补齐。

## Decisions

### 决策 1：采用“单个官方插件仓库 submodule，挂载路径保持 `apps/lina-plugins`”

理由：现有前后端、测试和文档都以 `apps/lina-plugins` 作为约定路径。保留挂载路径可以把变更重点放在可选化和初始化治理上，避免同时做路径迁移。

备选方案：

- 每个插件一个 submodule：版本粒度更细，但 `go.work`、测试发现、插件聚合模块和 CI 复杂度显著上升。
- 改用任意外部插件根路径：更灵活，但需要一次性重写大量路径约定，风险高于当前需求。

### 决策 2：宿主默认不再编译期强依赖官方插件聚合模块

宿主 `lina-core` 应能在没有 `lina-plugins` Go module 的情况下编译和运行。官方源码插件后端注册应通过仅在完整插件工作区中生效的构建入口、独立聚合包、生成文件或工具链注入方式实现，而不是让宿主主入口无条件 blank import 官方插件。

备选方案：

- 保留无条件 `_ "lina-plugins"` 并要求所有开发者初始化 submodule：无法满足 host-only 验证，也无法证明解耦。
- 使用 Go build tag 切换插件导入：实现成本低，但需要确保默认 tag 是 host-only，完整构建入口显式启用插件。

### 决策 3：源码插件发现统一返回“空集合或明确错误”

运行时扫描插件清单时：

- 插件工作区不存在或为空时，普通列表、启动同步和宿主 E2E 应按空源码插件集合处理，并保留动态插件发现能力。
- 用户显式请求某个插件或执行插件专属构建/测试时，如果插件工作区缺失，应返回包含 submodule 初始化建议的明确错误。
- 插件扫描不得因为目录缺失静默伪装成完整插件环境。

### 决策 4：测试套件分成 host-only 与 plugin-full 两条路径

host-only 路径只运行宿主归属测试和宿主级插件框架测试中不依赖官方源码插件内容的用例。plugin-full 路径要求 submodule 存在，并运行所有官方插件后端单元测试和插件自有 E2E。

测试发现脚本应容忍 `apps/lina-plugins` 缺失；只有用户显式运行 `plugin:<plugin-id>` 或 `plugins` 全量插件范围时，才因为缺少插件工作区而快速失败。

### 决策 5：i18n、缓存和数据权限影响以“无新增运行时业务面”为默认结论

本变更主要是工作区、构建、测试和插件发现治理，不新增业务 REST API、不新增业务数据表、不调整权限模型。若实现中补充用户可见空状态或错误提示，必须同步维护运行时 i18n；若插件发现结果被缓存，必须使用已有显式 scope 失效，不能引入全局无范围清空。

## Risks / Trade-offs

- [风险] 默认 host-only 构建可能让开发者误以为官方插件已启用 → 通过 `make status`/preflight/文档明确当前插件工作区状态。
- [风险] build tag 或聚合导入策略使 CI 命令分叉 → 将命令封装在 `make test`、`make test-host`、`make test-plugins` 或等价入口，避免人工记忆底层参数。
- [风险] 插件目录缺失时过度跳过测试导致覆盖下降 → host-only 与 plugin-full 在 CI 中分别设为显式任务，plugin-full 缺少 submodule 必须失败。
- [风险] 前端动态扫描缺少插件页面后路由数量变化 → 宿主 E2E 只断言宿主能力和插件管理空状态，插件页面断言放入 plugin-full 阶段。
- [风险] go.work 仍引用不存在目录导致最早失败 → 需要改为可生成/可同步 workspace，或改为不把可选插件模块写死在默认 workspace。

## Migration Plan

1. 记录当前引用和失败基线：在隔离工作区中临时移走 `apps/lina-plugins`，运行 host-only 命令，归类失败。
2. 修复宿主强耦合：Go workspace、编译期导入、运行时扫描、前端扫描、测试发现和工具入口逐项收敛。
3. 建立 host-only 验证：确认插件工作区不存在和为空目录两种状态下宿主命令通过。
4. 引入 submodule：把官方插件仓库挂载回 `apps/lina-plugins`，补齐 `.gitmodules`、文档和 CI 初始化步骤。
5. 建立 plugin-full 验证：运行所有官方插件 Go 单测、插件 E2E、动态插件 wasm 构建和完整 `make test`。
6. 归档前审查：确认 i18n、缓存一致性、数据权限和文档镜像要求已记录或无需变更。

Rollback 策略：若 submodule 迁移出现阻塞，可保留已完成的 host-only 解耦改造，暂缓 `.gitmodules` 变更；宿主可选化本身不要求立即删除主仓库插件代码。

## Open Questions

- 官方插件远端仓库名称、URL 和访问权限是否已确定。
- CI 是否需要同时提供“未初始化 submodule”的 host-only job 和“初始化 submodule”的 plugin-full job。
- 官方源码插件聚合导入最终采用 build tag、生成文件还是独立完整构建入口。
