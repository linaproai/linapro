---
name: openspec-e2e
description: Playwright E2E test case management conventions for the OpenSpec workflow. Defines test file naming (TC{NNNN}), directory organization by module, TC ID allocation, file independence rules, and sub-assertion patterns. Use when creating, planning, or reviewing E2E test cases within an OpenSpec change.
compatibility: Requires Playwright.
---

# OpenSpec E2E Test Case Conventions

Standards for organizing, naming, and writing Playwright E2E test cases in this project.

---

## 1. Directory Structure

```
hack/tests/
├── e2e/
│   ├── auth/                        # Module: authentication
│   │   ├── TC0001-login-verification.ts
│   │   └── TC0007-logout.ts
│   ├── admin/                       # Module: admin features
│   │   ├── TC0002-spec-management.ts
│   │   └── TC0003-user-management.ts
│   ├── notebook/                    # Module: notebook lifecycle
│   │   ├── TC0004-create-notebook.ts
│   │   ├── TC0005-jupyterlab-access.ts
│   │   ├── TC0006-training-execution.ts
│   │   ├── TC0008-multi-image-notebook.ts
│   │   └── TC0009-shared-directory.ts
│   └── {module}/                    # New modules follow the same pattern
│       └── TC{NNNN}-{brief-name}.ts
├── fixtures/
│   ├── auth.ts
│   ├── config.ts
│   └── k8s.ts
├── pages/                           # Page Object Model files
│   ├── LoginPage.ts
│   ├── NotebookPage.ts
│   └── ...
└── playwright.config.ts
```

**Key rules:**
- Directories under `e2e/` are named after **feature modules** (e.g., `auth`, `notebook`, `admin`).
- Each test case file lives under the module it primarily tests.

---

## 2. File Naming Convention

Every test file MUST follow the pattern:

```
TC{NNNN}-{brief-name}.ts
```

| Component      | Format         | Example                         |
|----------------|----------------|---------------------------------|
| Prefix         | `TC`           | `TC`                            |
| ID             | 4-digit, zero-padded | `0001`, `0012`, `0100`   |
| Separator      | `-`            | `-`                             |
| Brief name     | kebab-case     | `login-verification`            |
| Extension      | `.ts`          | `.ts`                           |

**Full examples:**
- `TC0001-login-verification.ts`
- `TC0014-bulk-delete-notebooks.ts`

**Rules:**
- One test case (one `test.describe` block) per file.
- The TC ID is **globally unique** across all modules.
- No `.spec.ts` suffix — use plain `.ts`.

---

## 3. TC ID Allocation

Before adding a new test case:

1. **Scan all existing TC files** across every module directory under `e2e/`:
   ```bash
   find hack/tests/e2e -name 'TC*.ts' | sort
   ```
2. **Identify the highest TC number** currently in use.
3. **Assign the next sequential number** (increment by 1).

**Example:** If the highest existing file is `TC0009-shared-directory.ts`, the next test case gets `TC0010`.

**IMPORTANT:** TC IDs must never be reused, even if a test case is deleted. Always increment from the historical maximum.

---

## 4. Test File Template

Each test file follows this structure:

```typescript
import { test, expect } from '../../fixtures/auth'
import { SomePage } from '../../pages/SomePage'
import { config } from '../../fixtures/config'

test.describe('TC-{N} {Brief Description}', () => {
  // Optional: shared setup
  test.beforeEach(async ({ adminPage }) => {
    // ...
  })

  test('TC-{N}a: {sub-assertion description}', async ({ page }) => {
    // Single focused assertion
  })

  test('TC-{N}b: {sub-assertion description}', async ({ adminPage }) => {
    // Another focused assertion
  })

  test('TC-{N}c: {sub-assertion description}', async ({ adminPage }) => {
    // ...
  })
})
```

**Conventions inside the file:**
- The `test.describe` label uses `TC-{N}` (no zero-padding) followed by a brief description.
- Sub-tests use `TC-{N}{letter}:` as prefix (e.g., `TC-1a:`, `TC-1b:`).
- When multiple sub-tests are combined into one block, use range notation: `TC-{N}a~c:`.
- Each sub-test should be focused on a single assertion or closely related assertions.

---

## 5. Test Independence

Each test file MUST be independently runnable:

- **No cross-file dependencies.** A test file must not rely on state created by another test file.
- **Self-contained setup.** If a test needs preconditions (e.g., a logged-in user, a created resource), it must set them up itself via `beforeEach`, `beforeAll`, fixtures, or inline setup.
- **Clean up after itself.** Tests that create resources should clean them up to avoid polluting other tests.
- **Runnable in isolation:**
  ```bash
  npx playwright test hack/tests/e2e/auth/TC0001-login-verification.ts
  ```

---

## 6. Page Object Model (POM)

All page interactions MUST go through Page Object classes in `pages/`:

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
    // Encapsulate complex interactions
  }
}
```

**Rules:**
- One POM class per page/feature area.
- POM files live in `pages/` (not in `e2e/`).
- Use `data-testid` attributes as the primary locator strategy.
- POM methods should return meaningful values or await expected states.

---

## 7. Fixtures

Shared test setup (authentication, configuration) lives in `fixtures/`:

- `auth.ts` — Extends Playwright `test` with authenticated page fixtures (`adminPage`, etc.)
- `config.ts` — Environment-specific configuration (URLs, credentials, timeouts).
- `k8s.ts` — Kubernetes helpers (pod readiness, exec commands).

Import fixtures instead of `@playwright/test` directly:
```typescript
// Correct
import { test, expect } from '../../fixtures/auth'

// Incorrect
import { test, expect } from '@playwright/test'
```

---

## 8. Mapping TC IDs in OpenSpec Tasks

When writing `tasks.md` in an OpenSpec change, E2E test tasks MUST reference TC IDs:

```markdown
### Task 3: E2E — TC0010 Notebook auto-save

- [ ] Create `hack/tests/e2e/notebook/TC0010-notebook-auto-save.ts`
- [ ] Implement TC-10a: file saved after idle timeout
- [ ] Implement TC-10b: save indicator shown in UI
- [ ] Implement TC-10c: content persists after page reload
```

The TC ID in the task title must match the filename. Sub-assertions (`TC-10a`, `TC-10b`) should be listed as sub-items.

---

## 9. Quick Reference

| Item                  | Convention                                      |
|-----------------------|-------------------------------------------------|
| File name             | `TC{NNNN}-{brief-name}.ts`                     |
| TC ID scope           | Global across all modules                       |
| Directory             | `e2e/{module}/`                                 |
| Describe label        | `TC-{N} {Description}`                          |
| Sub-test label        | `TC-{N}{letter}: {description}`                 |
| Import test/expect    | From `../../fixtures/auth`                      |
| Page interactions     | Via POM classes in `pages/`                     |
| Independence          | Each file runnable in isolation                 |
| ID allocation         | Scan max existing → increment by 1              |
