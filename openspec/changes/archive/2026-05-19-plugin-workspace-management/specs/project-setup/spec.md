## MODIFIED Requirements

### Requirement: 开发环境一键启动

系统 SHALL 提供 Makefile 命令，支持一键启动前后端开发环境。仓库级开发工具配置 SHALL 统一放在 `hack/config.yaml`，其中插件来源配置 SHALL 使用 `plugins.sources` 声明，并由跨平台 `linactl` 命令读取执行；默认开发入口不得依赖 Bash、PowerShell 或平台专属脚本实现插件工作区管理。

#### Scenario: 启动开发环境
- **WHEN** 在项目根目录执行 `make dev`
- **THEN** 前端和后端服务同时启动

#### Scenario: 停止开发环境
- **WHEN** 在项目根目录执行 `make stop`
- **THEN** 前端和后端服务同时停止

#### Scenario: 通过配置管理插件来源
- **WHEN** `hack/config.yaml` 配置 `plugins.sources`
- **AND** 用户运行 `make plugins.install`、`make plugins.update` 或 `make plugins.status`
- **THEN** 命令通过 `linactl` 读取该配置
- **AND** 命令按 `apps/lina-plugins` 固定目录执行插件工作区管理
- **AND** 命令不要求用户维护额外的插件路径配置
