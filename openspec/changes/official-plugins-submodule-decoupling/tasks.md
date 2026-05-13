## 1. 基线调查与隔离验证

- [ ] 1.1 在隔离工作区或可恢复的临时路径中记录当前 `git status --short`、`go.work`、插件目录状态和官方插件列表
- [ ] 1.2 临时移出 `apps/lina-plugins`，运行宿主后端构建和宿主后端单元测试，记录所有失败点并按 Go workspace、编译期导入、运行时扫描、测试辅助分类
- [ ] 1.3 创建空目录 `apps/lina-plugins`，重复宿主后端构建和宿主后端单元测试，记录与目录缺失状态的差异
- [ ] 1.4 在目录缺失和空目录两种状态下运行宿主前端类型检查或构建，记录 Vite 插件页面扫描、路由注册和访问过滤相关失败
- [ ] 1.5 在目录缺失和空目录两种状态下运行宿主 E2E 发现/校验命令，记录 Playwright 配置、执行治理脚本和插件测试范围解析失败

## 2. Go Workspace 与宿主编译解耦

- [ ] 2.1 调整默认 Go workspace 或生成流程，使 host-only 状态不因 `apps/lina-plugins` 和各 `lina-plugin-*` module 缺失而失败
- [ ] 2.2 移除宿主默认入口对官方插件聚合 module 的无条件编译期依赖，并实现显式启用官方源码插件注册的完整构建路径
- [ ] 2.3 为官方源码插件注册路径增加缺失 submodule 的可操作错误提示，包含 `git submodule update --init --recursive`
- [ ] 2.4 补充或更新 Go 单元测试，覆盖默认宿主构建路径不依赖官方插件 module

## 3. 插件发现与工具入口解耦

- [ ] 3.1 调整源码插件 manifest 扫描，在 `apps/lina-plugins` 不存在或为空时返回空源码插件集合，并保留动态插件发现
- [ ] 3.2 区分普通读取和显式源码插件操作：插件列表/宿主启动可降级为空集合，插件专属同步、wasm 构建和完整插件验证必须快速失败
- [ ] 3.3 更新 `linactl`、`make wasm`、mock SQL 加载和测试辅助中的插件根解析，统一输出插件工作区状态诊断
- [ ] 3.4 补充后端单元测试，覆盖源码插件工作区不存在、为空目录、结构无效和 submodule 正常存在四类状态

## 4. 前端 Host-only 支持

- [ ] 4.1 调整 Vite 源码插件页面扫描，使插件工作区不存在或为空时返回空模块集合
- [ ] 4.2 调整插件页面注册和访问过滤逻辑，使没有源码插件页面时宿主路由、菜单和权限过滤仍正常工作
- [ ] 4.3 运行 `pnpm -F @lina/web-antd typecheck` 和必要的前端构建验证，覆盖 host-only 状态
- [ ] 4.4 若新增或修改用户可见空状态/错误文案，同步维护前端运行时 i18n 和插件 manifest i18n 资源；若无 i18n 影响，在任务记录中明确说明

## 5. E2E 与验证入口

- [ ] 5.1 调整 Playwright 配置和测试治理脚本，使宿主 E2E 范围在 `apps/lina-plugins` 不存在或为空时仍可发现并执行
- [ ] 5.2 调整 `plugins` 与 `plugin:<plugin-id>` 范围，使缺少官方插件工作区时快速失败并提示初始化 submodule
- [ ] 5.3 新增或更新宿主级 E2E 覆盖 host-only 插件管理空状态，按当前最大 TC ID 分配编号
- [ ] 5.4 运行宿主 E2E 校验与目标宿主 E2E，用目录缺失和空目录两种状态验证通过

## 6. Submodule 迁移与完整插件验证

- [ ] 6.1 将官方插件仓库作为单个 submodule 挂载到 `apps/lina-plugins`，补齐 `.gitmodules` 和初始化说明
- [ ] 6.2 submodule 初始化后运行所有官方插件 Go 单元测试，并记录失败修复结果
- [ ] 6.3 submodule 初始化后运行所有官方插件自有 E2E 或插件测试清单中的完整插件范围，并记录失败修复结果
- [ ] 6.4 submodule 初始化后运行动态插件 wasm 构建或等价产物验证，确认完整构建路径仍可用

## 7. 文档、治理与审查

- [ ] 7.1 更新 README/README.zh-CN、CONTRIBUTING 和 AGENTS 中的官方插件工作区、submodule 初始化、host-only 验证和 plugin-full 验证说明
- [ ] 7.2 更新 OpenSpec 任务记录，明确 i18n 影响、缓存一致性影响、数据权限影响和测试覆盖结论
- [ ] 7.3 运行 `openspec validate official-plugins-submodule-decoupling --strict`
- [ ] 7.4 运行 `git diff --check -- openspec/changes/official-plugins-submodule-decoupling .gitmodules go.work Makefile hack apps/lina-core apps/lina-vben README.md README.zh-CN.md CONTRIBUTING.md AGENTS.md`
- [ ] 7.5 调用 `lina-review` 完成代码和规范审查，修正审查发现后再标记本变更完成

## Feedback

- [x] **FB-1**: 多租户插件 E2E 专属场景 helper 和 fixture 仍维护在宿主测试目录，影响官方插件工作区解耦；已迁移到插件 `hack/tests/support/`，本治理迁移不影响 i18n、缓存和数据权限
