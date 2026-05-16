## Why

宿主服务、源码插件和 `lina-core/pkg` 公共组件的后端主文件职责尚未完全统一，部分组件把接口契约、类型定义和大量实现逻辑混在同一个主文件中，开发者难以通过主文件快速理解组件边界、能力和使用约束。

本次变更将主文件职责、接口方法注释、文件顶部说明和 `lina-review` 审查要求固化为项目规范，并按业务模块分批整改，提升后端源码的可读性、可维护性和审查可控性。

## What Changes

- 强化后端主文件职责规范：宿主与源码插件 `internal/service/<component>/<component>.go`、`lina-core/pkg/<component>/<component>.go` 主文件只保留组件说明、核心类型、接口契约、实现结构体、构造函数和编译期接口断言，具体实现逻辑迁移到同包其他文件。
- 强化接口方法注释规范：后端接口定义中的每个方法注释必须说明功能作用、关键输入参数、输出结果、错误返回、权限/数据权限、缓存、i18n 或调用注意事项中适用的内容。
- 强化文件顶部注释规范：所有后端源文件顶部必须提供该文件职责、主要实现逻辑和注意事项的说明；主文件说明组件整体边界，非主文件说明当前文件承载的实现切片。
- 将上述要求写入 `AGENTS.md` 和 `lina-review` 审查清单，使后续后端 Go 变更必须接受主文件职责、接口注释完整度和文件顶部说明质量检查。
- 按业务模块拆分整改任务，每个任务只覆盖一组职责明确的模块或公共组件，并在任务完成后运行对应 Go 编译门禁和治理验证。
- 不改变现有 REST API、数据库结构、业务行为、权限语义、缓存语义或 i18n 资源内容；本次主要是源码组织与治理规范变更。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `backend-conformance`: 强化后端主文件职责、接口方法注释、文件顶部说明和 `lina-review` 审查要求，覆盖宿主服务、源码插件后端服务和 `lina-core/pkg` 公共组件。

## Impact

- 影响范围：
  - `AGENTS.md`
  - `.agents/skills/lina-review/SKILL.md`
  - `openspec/specs/backend-conformance/spec.md` 的增量规范
  - `apps/lina-core/internal/service/**`
  - `apps/lina-core/pkg/**`
  - `apps/lina-plugins/*/backend/internal/service/**`
- 不涉及数据库迁移、API 路由变更、前端页面、运行时语言包、manifest i18n 或 apidoc i18n 资源。
- 主要风险是大规模同包迁移带来的审查噪音和潜在漏迁；通过按模块分批、保持同包函数签名不变、逐包 `go test` 和 `lina-review` 分段审查控制风险。
