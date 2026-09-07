# 深色侧边栏无闪屏实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复亮色主题下开启“深色侧边栏”时菜单区域短暂显示白色背景的问题，并建立可重复执行的端到端回归门禁。

**Architecture:** 所有可见侧边栏菜单根节点负责绘制稳定的主题背景，普通菜单项保持透明，仅在悬停和激活状态绘制局部背景。主题类变化后，端到端测试从菜单进入暗色状态开始逐帧采样实际计算样式，确保稳定背景立即生效且任何菜单项都不会残留不透明浅色层。

**Tech Stack:** Vue 3、SCSS、Tailwind CSS、Playwright、OpenSpec。

**Spec:** `openspec/changes/fix-dark-sidebar-flash/specs/frontend-theme-switching/spec.md`

## 全局约束

- 只修改主题切换、菜单样式、对应`E2E`和现有`OpenSpec`变更，不处理当前工作树中的后端与插件改动。
- 先运行新增`E2E`并确认它因白色背景帧失败，再修改生产样式。
- 保留菜单悬停、激活、折叠和弹出菜单反馈，不增加定时器、设备判断、降级或兜底分支。
- 不修改用户可见文案、`HTTP API`、Go后端、数据库、数据权限、缓存或运行时依赖。
- 实施阶段不执行`git commit`或`git push`；用户明确授权交付后再提交与推送。

## 文件结构

- 修改`hack/tests/pages/MainLayout.ts`：封装深色侧边栏开关定位、状态恢复和逐帧样式探针。
- 修改`hack/tests/e2e/dashboard/TC007-theme-switch-performance.ts`：增加深色侧边栏切换无浅色帧的回归场景与截图。
- 修改`apps/lina-vben/packages/@core/ui-kit/menu-ui/src/components/menu.vue`：由菜单根节点绘制稳定背景，普通菜单项与子菜单使用透明基础背景。
- 新增`openspec/changes/fix-dark-sidebar-flash/design.md`：记录局部侧边栏主题事务的根因与样式职责。
- 新增`openspec/changes/fix-dark-sidebar-flash/specs/frontend-theme-switching/spec.md`：增加深色侧边栏原子换色要求与场景。
- 新增`openspec/changes/fix-dark-sidebar-flash/tasks.md`：记录反馈修复、失败复现和验证结果。

### Task 1：建立深色侧边栏闪白回归用例

**Files:**

- Modify: `hack/tests/pages/MainLayout.ts`
- Test: `hack/tests/e2e/dashboard/TC007-theme-switch-performance.ts`

**Interfaces:**

- Consumes: 现有`MainLayout.openPreferences()`、`MainLayout.prepareThemeModeWithoutAnimation()`和后台分析页认证 fixture。
- Produces: `MainLayout.ensureSemiDarkSidebar(enabled: boolean)`和`MainLayout.captureSemiDarkSidebarThemeFrames()`，返回切换后前`200ms`内菜单根节点及可见菜单项的计算背景色记录。

- [x] **Step 1：编写失败测试**

在`TC007`中新增`TC007c`：先固定亮色主题并关闭深色侧边栏，安装逐帧探针后开启深色侧边栏，断言菜单根节点首帧为暗色，普通菜单项不存在不透明浅色背景，并捕获切换完成截图。

- [x] **Step 2：运行测试并确认正确失败**

Run: `pnpm exec playwright test e2e/dashboard/TC007-theme-switch-performance.ts --grep "TC007c"`

Expected: `FAIL`，失败记录显示`.vben-menu-item`或`.vben-sub-menu-content`在菜单已经进入暗色状态后仍保持不透明浅色背景。

### Task 2：从菜单绘制职责修复白色闪烁

**Files:**

- Modify: `apps/lina-vben/packages/@core/ui-kit/menu-ui/src/components/menu.vue`
- Test: `hack/tests/e2e/dashboard/TC007-theme-switch-performance.ts`

**Interfaces:**

- Consumes: 现有`--menu-background-color`、`--menu-item-hover-background-color`和`--menu-item-active-background-color`变量。
- Produces: 菜单根节点稳定绘制主题背景，普通菜单项和子菜单默认透明，悬停与激活状态继续使用原变量绘制反馈。

- [x] **Step 1：实施最小样式修复**

将菜单根节点背景改为直接消费`--menu-background-color`；将亮色、暗色菜单的普通菜单项与子菜单基础背景设为透明，避免主题变化时大量不透明白色块参与`background`过渡。

- [x] **Step 2：运行新增测试并确认通过**

Run: `pnpm exec playwright test e2e/dashboard/TC007-theme-switch-performance.ts --grep "TC007c"`

Expected: `PASS`，所有暗色状态采样帧均不存在不透明浅色菜单背景。

- [x] **Step 3：运行完整主题切换回归**

Run: `pnpm exec playwright test e2e/dashboard/TC007-theme-switch-performance.ts`

Expected: `PASS`，原有双向主题揭幕、图表实例稳定、立即刷新和新增深色侧边栏场景全部通过。

### Task 3：同步规范并完成验证

**Files:**

- Create: `openspec/changes/fix-dark-sidebar-flash/proposal.md`
- Create: `openspec/changes/fix-dark-sidebar-flash/design.md`
- Create: `openspec/changes/fix-dark-sidebar-flash/specs/frontend-theme-switching/spec.md`
- Create: `openspec/changes/fix-dark-sidebar-flash/tasks.md`

**Interfaces:**

- Consumes: `TC007c`失败与通过证据、最终菜单绘制职责。
- Produces: 可审查的反馈根因、规范场景、任务和验证记录。

- [x] **Step 1：更新现有主题切换规范**

增加“局部深色侧边栏必须原子换色”的要求，明确菜单背景由稳定父层绘制，普通菜单项不得在主题切换时保留不透明浅色过渡层。

- [x] **Step 2：运行前端静态验证**

Run: `pnpm exec vue-tsc --noEmit --skipLibCheck -p packages/@core/ui-kit/menu-ui/tsconfig.json`

Result: `PASS`。另外使用临时限定`tsconfig`对`MainLayout.ts`与`TC007-theme-switch-performance.ts`运行`pnpm exec tsc --noEmit`，并对三个代码文件运行`Prettier --check`，均以退出码`0`通过；临时配置随后删除。

- [x] **Step 3：运行 E2E 治理校验**

Run: `pnpm run test:validate`

Result: Windows 下脚本直接启动`pnpm`时因`stdout`未定义在`.trim()`处崩溃；单独运行前端`i18n:check`通过。使用一次性 Windows 进程适配继续执行原治理脚本后，本次`TC007`、`MainLayout`和菜单组件无治理错误；在最新`origin/main`上，仓库仍因`TC004-page-panel-border-radius.ts`的插件生命周期并行隔离和`TC009-forgot-password-and-register.ts`的系统配置并行隔离两个既有问题退出`1`。

- [x] **Step 4：运行 OpenSpec 严格校验**

Run: `openspec validate fix-dark-sidebar-flash --strict`

Result: `PASS`，输出`Change 'fix-dark-sidebar-flash' is valid`。

- [x] **Step 5：审查截图与工作树**

检查`temp/20260820/`下设置开启截图与悬停后截图，确认侧边栏、菜单文字、图标、选中项和悬停状态可读；使用`git diff --check`、限定路径`git diff`和`git status --short`确认隔离工作树只修改约定的菜单、`E2E`、`OpenSpec`与计划文件，其余文件与最新`origin/main`保持一致。

## 自检结果

- 规范覆盖：原问题复现、根因修复、悬停与激活行为、截图审查和回归验证均有对应步骤。
- 占位符检查：无`TBD`、`TODO`或未定义接口。
- 类型一致性：页面对象方法、测试场景与样式变量名称在各任务中保持一致。
- 影响判断：无`i18n`、接口、缓存、数据权限、数据库、后端、跨平台工具或运行时依赖影响。
