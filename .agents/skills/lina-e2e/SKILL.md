---
name: lina-e2e
description: Playwright E2E test case management standards for the OpenSpec workflow. Defines file naming conventions (TC{NNNN}), module-based directory layout, TC ID assignment, file isolation rules, and sub-assertion patterns. Use when creating, planning, or reviewing E2E test cases within an OpenSpec change.
compatibility: 依赖 Playwright。
---

# Lina E2E 测试用例规范

本项目中 `Playwright E2E` 测试用例的组织、命名和编写标准。

**交互语言**：与用户交互的内容语言以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

---

## 1. 目录结构

```
hack/tests/
├── e2e/
│   ├── auth/                        # 宿主模块：认证
│   │   ├── TC0001-login-verification.ts
│   │   └── TC0007-logout.ts
│   ├── admin/                       # 宿主模块：管理功能
│   │   ├── TC0002-spec-management.ts
│   │   └── TC0003-user-management.ts
│   ├── notebook/                    # 宿主模块：笔记本生命周期
│   │   ├── TC0004-create-notebook.ts
│   │   ├── TC0005-jupyterlab-access.ts
│   │   ├── TC0006-training-execution.ts
│   │   ├── TC0008-multi-image-notebook.ts
│   │   └── TC0009-shared-directory.ts
│   └── {module}/                    # 宿主新模块遵循相同模式
│       └── TC{NNNN}-{brief-name}.ts
├── fixtures/
│   ├── auth.ts
│   ├── config.ts
│   └── k8s.ts
├── pages/                           # 宿主/共享页面对象模型文件
│   ├── LoginPage.ts
│   ├── NotebookPage.ts
│   └── ...
└── playwright.config.ts
```

源码插件的插件专属 E2E 必须闭环在插件自己的目录中：

```
apps/lina-plugins/<plugin-id>/
├── e2e/
│   └── TC{NNNN}-{brief-name}.ts
├── e2e-pages/
│   └── <PluginPageObject>.ts
└── e2e-support/
    └── <plugin-helper>.ts
```

**关键规则：**
- 宿主功能测试放在 `hack/tests/e2e/{module}/`。
- 源码插件专属测试放在 `apps/lina-plugins/<plugin-id>/e2e/`。
- `hack/tests/e2e/extension/plugin/` 只用于宿主插件框架、动态插件运行时、源码插件生命周期这类**宿主级插件能力**测试；禁止把某个源码插件自身功能的 E2E 放到这里。
- 每个测试用例文件放在其主要测试的所有权目录下；谁拥有功能，谁拥有测试。

---

## 2. 文件命名规范

每个测试文件必须遵循以下模式：

```
TC{NNNN}-{brief-name}.ts
```

| 组成部分     | 格式           | 示例                            |
|-------------|----------------|--------------------------------|
| 前缀        | `TC`           | `TC`                           |
| ID          | `4` 位数字，补零  | `0001`、`0012`、`0100`          |
| 分隔符      | `-`            | `-`                            |
| 简短名称    | kebab-case     | `login-verification`           |
| 扩展名      | `.ts`          | `.ts`                          |

**完整示例：**
- `TC0001-login-verification.ts`
- `TC0014-bulk-delete-notebooks.ts`

**规则：**
- 每个文件只包含一个测试用例（一个 `test.describe` 块）。
- `TC ID` 在所有模块中**全局唯一**。
- 不使用 `.spec.ts` 后缀，使用普通的 `.ts`。

---

## 3. TC ID 分配

添加新测试用例前：

1. **扫描所有模块目录下的现有 TC 文件**：
   ```bash
   {
     find hack/tests/e2e -type f -name 'TC*.ts'
     find apps/lina-plugins -type f -path '*/e2e/*' -name 'TC*.ts'
     rg -No 'TC[0-9]{4}' openspec/changes -g 'tasks.md'
   } | rg -No 'TC[0-9]{4}' | sort -u | tail -1
   ```
2. **确定当前已实现和 OpenSpec 任务记录中使用或预留的最大 TC 编号**。
3. **分配下一个顺序编号**（递增 1）。

**示例：** 如果现有最大文件为 `TC0009-shared-directory.ts`，则下一个测试用例为 `TC0010`。

**重要：** TC ID 永不复用，即使测试用例被删除也是如此。始终从历史最大值递增。

---

## 4. 测试文件模板

每个测试文件遵循以下结构：

宿主测试：

```typescript
import { test, expect } from '../../fixtures/auth'
import { SomePage } from '../../pages/SomePage'
import { config } from '../../fixtures/config'

test.describe('TC-{N} {简短描述}', () => {
  // 可选：共享设置
  test.beforeEach(async ({ adminPage }) => {
    // ...
  })

  test('TC-{N}a: {子断言描述}', async ({ page }) => {
    // 单一聚焦断言
  })

  test('TC-{N}b: {子断言描述}', async ({ adminPage }) => {
    // 另一个聚焦断言
  })

  test('TC-{N}c: {子断言描述}', async ({ adminPage }) => {
    // ...
  })
})
```

源码插件测试：

```typescript
import { test, expect } from '../../../../hack/tests/fixtures/auth'
import { SomePluginPage } from '../e2e-pages/SomePluginPage'

test.describe('TC-{N} {插件功能描述}', () => {
  test('TC-{N}a: {子断言描述}', async ({ adminPage }) => {
    // 插件专属流程断言
  })
})
```

**文件内约定：**
- `test.describe` 标签使用 `TC-{N}`（不补零）后跟简短描述。
- 子测试使用 `TC-{N}{字母}:` 作为前缀（如 `TC-1a:`、`TC-1b:`）。
- 当多个子测试合并为一个块时，使用范围表示法：`TC-{N}a~c:`。
- 每个子测试应聚焦于单一断言或紧密相关的断言。

---

## 5. 测试独立性

每个测试文件必须可独立运行：

- **无跨文件依赖。** 测试文件不得依赖其他测试文件创建的状态。
- **自包含设置。** 如果测试需要前置条件（如已登录用户、已创建资源），必须通过 `beforeEach`、`beforeAll`、固件或内联设置自行完成。
- **自行清理。** 创建资源的测试应清理资源以避免污染其他测试。
- **可独立运行：**
  ```bash
  npx playwright test hack/tests/e2e/auth/TC0001-login-verification.ts
  pnpm -C hack/tests test:module -- plugin:<plugin-id>
  ```

---

## 6. 页面对象模型（POM）

所有页面交互必须通过页面对象类进行：

```typescript
import { Page, Locator } from '@playwright/test'

export class SomePage {
  readonly page: Page
  readonly someElement: Locator

  constructor(page: Page) {
    this.page = page
    this.someElement = page.locator('[data-testid="some-element"]')
  }

  async goto() {
    await this.page.goto('/some-path')
    await this.page.waitForLoadState('networkidle')
  }

  async performAction() {
    // 封装复杂交互
  }
}
```

**规则：**
- 每个页面/功能区域一个 POM 类。
- 宿主或跨模块共享 POM 放在 `hack/tests/pages/`。
- 源码插件专属 POM 放在 `apps/lina-plugins/<plugin-id>/e2e-pages/`。
- 源码插件专属定位器禁止加到宿主 `hack/tests/pages/` 中；只有多个宿主测试或多个插件确实复用的通用能力，才提升到宿主共享 POM。
- 优先使用 `data-testid` 属性作为定位策略。
- POM 方法应返回有意义的值或等待预期状态。

---

## 7. 测试固件

共享的测试设置（认证、配置）放在 `fixtures/` 目录中：

- `auth.ts` — 扩展 Playwright `test`，提供已认证的页面固件（`adminPage` 等）
- `config.ts` — 环境相关配置（URL、凭据、超时时间）
- `k8s.ts` — Kubernetes 辅助工具（Pod 就绪检查、执行命令）

使用固件而非直接导入 `@playwright/test`：
```typescript
// 宿主测试
import { test, expect } from '../../fixtures/auth'

// 源码插件测试
import { test, expect } from '../../../../hack/tests/fixtures/auth'

// 错误
import { test, expect } from '@playwright/test'
```

---

## 8. 在 OpenSpec 任务中映射 TC ID

在 OpenSpec 变更中编写 `tasks.md` 时，E2E 测试任务必须引用 TC ID：

```markdown
### 任务 3：E2E — TC0010 笔记本自动保存

- [ ] 创建 `hack/tests/e2e/notebook/TC0010-notebook-auto-save.ts`
- [ ] 实现 TC-10a：空闲超时后文件自动保存
- [ ] 实现 TC-10b：UI 中显示保存指示器
- [ ] 实现 TC-10c：页面重新加载后内容持久化
```

源码插件示例：

```markdown
### 任务 3：E2E — TC0221 插件页面入口

- [ ] 创建 `apps/lina-plugins/example-plugin/e2e/TC0221-example-plugin-entry.ts`
- [ ] 实现 TC-221a：插件公开接口可读取
- [ ] 实现 TC-221b：插件插槽内容可见
- [ ] 实现 TC-221c：插件管理页可访问
```

任务标题中的 TC ID 必须与文件名匹配。子断言（`TC-10a`、`TC-10b`）应列为子项。

---

## 9. 快速参考

| 项目                  | 规范                                                |
|----------------------|----------------------------------------------------|
| 文件名               | `TC{NNNN}-{brief-name}.ts`                         |
| TC ID 范围           | 全局唯一，跨所有模块                                  |
| 宿主测试目录          | `hack/tests/e2e/{module}/`                         |
| 源码插件测试目录       | `apps/lina-plugins/<plugin-id>/e2e/`               |
| Describe 标签        | `TC-{N} {描述}`                                     |
| 子测试标签            | `TC-{N}{字母}: {描述}`                              |
| 宿主导入 test/expect  | 从相对路径 `../../fixtures/auth` 导入                 |
| 插件导入 test/expect  | 从相对路径 `../../../../hack/tests/fixtures/auth` 导入 |
| 页面交互              | 通过宿主 `pages/` 或插件 `e2e-pages/` 中的 POM 类      |
| 独立性               | 每个文件可独立运行                                    |
| ID 分配              | 扫描宿主、插件和 OpenSpec 任务记录已用/预留最大值 → 递增 1 |
