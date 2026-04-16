## Why

当前仓库在 GoFrame v2 约定、REST 接口一致性、API 文档标签、模块解耦以及 OpenSpec 主规范结构上已经积累了系统性偏差。该偏差已经开始影响日常开发质量，并直接导致本次 `distributed-locker` 归档过程中出现主规范结构不兼容、归档残留文件和校验阻塞等问题，因此需要单独开启一次规范整改迭代。

## What Changes

- 统一 OpenSpec 主规范到当前 schema 要求，修复归档/校验过程中暴露的主规范结构和残留文件问题
- 建立后端 GoFrame v2 合规基线，整改控制器、服务层、数据库访问和软删除相关的典型违规模式
- 统一后端 REST 路径风格、路径参数绑定方式以及 API DTO 的 `dc`/`eg` 文档标签
- 为业务模块补齐可启用/禁用的解耦约束，确保模块关闭时后端能够平滑降级
- 增加一套可重复执行的规范校验与回归验证流程，降低后续迭代再次偏离规范的概率

## Capabilities

### New Capabilities
- `spec-governance`: 规范 OpenSpec 主规范的文件结构、归档前校验和残留产物治理
- `backend-conformance`: 约束后端控制器、服务层、ORM 使用方式与 GoFrame v2 最佳实践保持一致
- `api-contract-consistency`: 统一 REST 语义、路径参数绑定和 API 文档标签质量
- `module-decoupling`: 定义业务模块启用/禁用时的后端降级和依赖解耦行为

### Modified Capabilities
- 无

## Impact

- `openspec/specs/` 主规范文件结构和部分历史规范内容需要整理
- `openspec/changes/` 的归档流程和变更工件需要按当前 schema 校正
- `apps/lina-core/api/`、`apps/lina-core/internal/controller/`、`apps/lina-core/internal/service/` 会有批量规范整改
- 若接口路径或绑定方式发生调整，`apps/lina-vben/` 与 `hack/tests/` 中相关调用和测试需要同步更新
