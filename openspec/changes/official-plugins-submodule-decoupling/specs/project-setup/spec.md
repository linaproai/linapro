## ADDED Requirements

### Requirement: 宿主初始化命令必须支持 Host-only 工作区

系统 SHALL 允许开发者在官方源码插件工作区不存在或为空时执行宿主基础初始化、构建、启动和测试命令。默认宿主命令不得因为可选插件目录缺失而在 Go workspace 加载、前端构建初始化或测试发现阶段失败。

#### Scenario: 插件工作区缺失时后端宿主可构建
- **WHEN** `apps/lina-plugins` 不存在
- **AND** 开发者在 `apps/lina-core` 或仓库根目录执行宿主后端构建命令
- **THEN** 后端宿主构建成功或仅因宿主自身代码错误失败
- **AND** 构建不得因缺少 `apps/lina-plugins`、`lina-plugins` 或 `lina-plugin-*` module 失败

#### Scenario: 插件工作区为空时前端宿主可构建
- **WHEN** `apps/lina-plugins` 是空目录
- **AND** 开发者执行宿主前端类型检查或构建命令
- **THEN** 前端构建成功或仅因宿主前端自身代码错误失败
- **AND** 插件页面扫描返回空集合

#### Scenario: Host-only 开发服务启动
- **WHEN** 开发者在未初始化官方插件 submodule 的工作区运行 `make dev`
- **THEN** 后端和前端开发服务按 host-only 模式启动
- **AND** 源码插件相关能力以空集合或空状态降级
