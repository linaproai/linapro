## REMOVED Requirements

### Requirement: 安装入口点必须提供一致的跨平台能力
**Reason**: 项目源码获取不再通过仓库内安装脚本封装。该脚本只包装 Git clone/tag checkout 工作流，直接提供 Git 命令更简单、透明且更符合当前交付方式。
**Migration**: 使用 Git 获取源码；需要最新稳定版时使用维护中的稳定分支，需要固定发布版本时使用明确的发布 tag。

### Requirement: 安装流程必须基于源码归档下载而非 Git clone
**Reason**: 项目不再维护独立源码归档安装器。源码获取统一回到 Git 原生命令，不再承诺无 Git 的脚本安装流程。
**Migration**: 使用 `git clone --depth 1 --branch <branch-or-tag> https://github.com/linaproai/linapro.git linapro` 获取指定分支或发布 tag。

### Requirement: 安装脚本必须安全处理当前目录和指定目录模式
**Reason**: 目录覆盖保护属于安装脚本封装行为；脚本移除后不再维护 `LINAPRO_DIR`、`LINAPRO_FORCE` 等安装器参数。
**Migration**: 用户通过 Git 命令显式选择目标目录；若目标目录已存在，由 Git 的原生 clone 行为拒绝覆盖。

### Requirement: 安装后输出必须包含环境健康检查和后续指导
**Reason**: 安装脚本移除后不再提供脚本级安装后输出。环境检查与依赖修复职责归属于 `lina-doctor` 技能和项目常规开发命令。
**Migration**: 获取源码后进入项目目录，按项目 README 和 `lina-doctor` 指引准备环境，再执行 `make init` 与 `make dev`。
