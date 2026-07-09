# Internal Utility Packages

`internal/utility` groups `Lina Core` internal utility packages that are not business modules, runtime services, public `pkg` contracts, generated models, or packed assets.

## Directory Rules

| Rule | Requirement |
|------|-------------|
| Root package | Do not place `Go` files directly under `internal/utility`. |
| Subpackage naming | Use a narrow responsibility name such as `testsupport`, `repoenv`, or `sourceasset`. |
| Generic names | Do not create second-level packages named `util`, `utils`, `common`, `helper`, or `misc`. |
| Reuse threshold | Add a package here only when it is reused by multiple `Lina Core` packages or clearly belongs to cross-package governance. |
| Dependency direction | Utility packages must not depend on `internal/service`, `internal/controller`, or `internal/cmd` business implementations. |
| Public boundary | Utility packages are `Lina Core` internal helpers, not stable `pkg` contracts for plugins or external callers. |

## Placement Guide

Use `internal/utility` for internal helpers that support tests, repository environment detection, static governance checks, or local source asset discovery.

Do not use `internal/utility` for business services, `API` contracts, plugin protocols, cache-bearing runtime components, permission boundaries, generated code, or one-off helpers used by a single package. Keep single-package helpers inside the package that owns them.

## Current Packages

| Package | Responsibility |
|---------|----------------|
| `testsupport` | Shared test helpers for detecting repository and official plugin workspace state. |
